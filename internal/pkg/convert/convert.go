package convert

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/lemonnekogh/guolai"
)

// https://www.wolai.com/wolai/o2v1vrLkP2qUuZTH6iDZY9
var frontColorHex = map[string]string{
	"gray":      "#8C8C8C",
	"dark_gray": "#5C5C5C",
	"brown":     "#A3431F",
	"orange":    "#F06B05",
	"yellow":    "#DFAB01",
	"green":     "#038766",
	"blue":      "#0575C5",
	"indigo":    "#4A52C7",
	"purple":    "#8831CC",
	"pink":      "#C815B6",
	"red":       "#E91E2C",
	"default":   "#000000",
}

// https://www.wolai.com/wolai/fNb4SHWY1bV2s8Xg5JYUE4
var backColorHex = map[string]string{
	"cultured_background":            "#F3F3F3",
	"light_gray_background":          "#E3E3E3",
	"apricot_background":             "#EFDFDB",
	"vivid_tangerine_background":     "#FCE5D7",
	"blond_background":               "#FCF5D6",
	"aero_blue_background":           "#D7EAE5",
	"uranian_blue_background":        "#D7E7F4",
	"lavender_blue_background":       "#E0E2F5",
	"pale_purple_background":         "#EADDF6",
	"pink_lavender_background":       "#F5D9F2",
	"light_pink_background":          "#FBDADC",
	"fluorescent_yellow_background":  "#FFF784",
	"fluorescent_green_background":   "#CDF7AD",
	"fluorescent_green2_background":  "#A6F9CB",
	"fluorescent_blue_background":    "#A8FFFF",
	"fluorescent_purple_background":  "#FDB7FF",
	"fluorescent_purple2_background": "#CCC4FF",
	"default":                        "#FFFFFF",
}

const (
	ExitCodeParamError = iota + 1
	ExitCodeTokenError
	ExitCodePermissionError
	ExitCodeOutputError
	ExitCodeUnknownError
)

type BlockContent struct {
	Content  []string
	Children []BlockContent
}

type PageToMarkdownContext struct {
	FootNotes  []string
	Result     []BlockContent // if use map, we can fetch child blocks outside the convert package, but map is unsorted
	ChildPages map[string]string
	Images     map[string]string
}

// FIXME: dirty code
func ProcessWolaiError(err guolai.WolaiError, blockId string) {
	if err.Code == 17003 {
		fmt.Println("token is invalid")
		os.Exit(ExitCodeTokenError)
	}

	if err.Code == 17011 {
		fmt.Println("failed to get content of block " + blockId + ": permission denied")
		os.Exit(ExitCodePermissionError)
	}

	if err.Code == 17007 {
		fmt.Println("failed to fetch block: " + blockId + ", API rate limit exceeded, waiting for 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func richTextStyleToMarkdown(text guolai.RichText) string {
	ret := strings.TrimSpace(text.Title)
	if ret == "" {
		return ""
	}
	if text.Bold {
		ret = fmt.Sprintf("**%s**", ret)
	}
	if text.Italic {
		ret = fmt.Sprintf("*%s*", ret)
	}
	if text.Underline {
		ret = fmt.Sprintf(`<span style="text-decoration:underline;">%s</span>`, ret)
	}
	if text.StrikeThrough {
		ret = fmt.Sprintf(`~~%s~~`, ret)
	}
	if text.InlineCode {
		ret = fmt.Sprintf("`%s`", ret)
	}
	frontColor := frontColorHex[string(text.FrontColor)]
	if text.FrontColor != "" {
		ret = fmt.Sprintf(`<span style="color:%s;">%s</span>`, frontColor, ret)
	}
	backColor := backColorHex[string(text.BackColor)]
	if text.BackColor != "" {
		ret = fmt.Sprintf(`<span style="background-color:%s;">%s</span>`, backColor, ret)
	}
	// line break in wolai would be '\n', not '<br>' nor '  '
	if strings.Contains(ret, "\n") {
		ret = strings.ReplaceAll(ret, "\n", "<br>")
	}

	return ret
}

// richTextToMarkdown
//
// These types are not support:
// - mention_member
func (ctx *PageToMarkdownContext) richTextToMarkdown(text []guolai.RichText) string {
	ret := ""
	for _, t := range text {
		switch t.Type {
		case "text":
			txt := richTextStyleToMarkdown(t)
			// link type will be 'text'
			if t.Link != nil {
				txt = fmt.Sprintf("[%s](<%s>)", txt, *t.Link)
			}
			ret += txt
		case "equation":
			ret += fmt.Sprintf("$%s$", t.Title)
		case "footnote":
			ret += fmt.Sprintf(`[^%d]`, len(ctx.FootNotes)+1)
			ctx.FootNotes = append(ctx.FootNotes, ctx.richTextToMarkdown(t.Content))
		case "bi_link":
			ret += fmt.Sprintf("<a href=\"#%s\" style=\"color:inherit;text-decoration:underline dashed;\">%s</a>", *t.BlockId, t.Title)
		}
	}

	return ret
}

func codeToMarkdown(code guolai.Block) []string {
	language := "plaintext"
	if code.Language != nil {
		language = string(*code.Language)
	}

	ret := []string{}
	ret = append(ret, "```"+language)
	ret = append(ret, code.Content[0].Title)
	ret = append(ret, "```")

	return ret
}

func (ctx *PageToMarkdownContext) calloutToMarkdown(callout guolai.Block) []string {
	ret := []string{}

	ret = append(ret, "::: tip "+callout.Icon.Icon)
	ret = append(ret, ctx.richTextToMarkdown(callout.Content))
	ret = append(ret, ":::")

	return ret
}

func (ctx *PageToMarkdownContext) headingToMarkdown(block guolai.Block) []string {
	ret := strings.Repeat("#", int(*block.Level))
	ret += " " + ctx.richTextToMarkdown(block.Content)

	return []string{ret}
}

func (ctx *PageToMarkdownContext) imageToMarkdown(block guolai.Block) string {
	var url string
	if block.Media.Type == "internal" {
		url = *block.Media.DownloadUrl
	} else if block.Media.Type == "external" {
		url = *block.Media.Url
	}

	fileName := fmt.Sprintf("image%d", len(ctx.Images))
	ctx.Images[url] = fileName

	return fmt.Sprintf("<img src=\"[%s]\" width=\"%f\" height=\"%f\">", fileName, *block.Dimensions.Width, *block.Dimensions.Height)
}

func (ctx *PageToMarkdownContext) taskListToMarkdown(block guolai.Block) string {
	checked := " "
	if block.Checked != nil && *block.Checked {
		checked = "x"
	}

	return fmt.Sprintf("- [%s] %s", checked, ctx.richTextToMarkdown(block.Content))
}

func (ctx *PageToMarkdownContext) fetchChildBlock(wolaiClient *guolai.WolaiAPI, blockId string) []BlockContent {
	blockChildren := []BlockContent{}

	fmt.Printf("fetching children of block: %s\n", blockId)

	blocks, err := wolaiClient.GetBlockChildren(blockId)
	if err != nil {
		var wolaiErr guolai.WolaiError
		if errors.As(err, &wolaiErr) {
			ProcessWolaiError(wolaiErr, blockId)
		} else {
			fmt.Printf("failed to get content of block %s: %v\n", blockId, err)
			os.Exit(ExitCodeUnknownError)
		}
	}

	// retry, but dirty
	for err != nil {
		blocks, err = wolaiClient.GetBlockChildren(blockId)
		if err != nil {
			var wolaiErr guolai.WolaiError
			if errors.As(err, &wolaiErr) {
				ProcessWolaiError(wolaiErr, blockId)
			} else {
				fmt.Printf("failed to get content of block %s: %v\n", blockId, err)
				os.Exit(ExitCodeUnknownError)
			}
		}
	}

	for _, block := range blocks {
		blockContent := ctx.blockToMarkdown(block)
		if blockContent.Children != nil {
			blockContent.Children = ctx.fetchChildBlock(wolaiClient, block.ID)
		}

		// should not append line break when a block is list or quote
		if blockContent.Children == nil {
			blockContent.Content = append(blockContent.Content, "\n")
		}

		blockChildren = append(blockChildren, blockContent)
	}

	return blockChildren
}

func PageToMarkdown(wolaiClient *guolai.WolaiAPI, title string, page []guolai.BlockApiResponse) *PageToMarkdownContext {
	ctx := &PageToMarkdownContext{
		ChildPages: map[string]string{},
		Images:     map[string]string{},
		Result:     []BlockContent{},
	}

	ctx.Result = append(ctx.Result, BlockContent{Content: []string{"# " + title}})

	for _, block := range page {
		blockContent := ctx.blockToMarkdown(block)
		if blockContent.Children != nil {
			blockContent.Children = ctx.fetchChildBlock(wolaiClient, block.ID)
		}

		blockContent.Content = append(blockContent.Content, "\n")
		ctx.Result = append(ctx.Result, blockContent)
	}

	for index, footnote := range ctx.FootNotes {
		ctx.Result = append(ctx.Result, BlockContent{Content: []string{fmt.Sprintf("\n[^%d]: %s\n", index+1, footnote)}})
	}

	return ctx
}

func (ctx *PageToMarkdownContext) blockToMarkdown(block guolai.BlockApiResponse) BlockContent {
	ret := BlockContent{}
	switch block.Type {
	case "code":
		ret.Content = codeToMarkdown(block.Block)
	case "heading":
		ret.Content = ctx.headingToMarkdown(block.Block)
	case "text":
		ret.Content = []string{ctx.richTextToMarkdown(block.Content)}
	case "quote":
		ret.Content = []string{fmt.Sprintf("> %s", ctx.richTextToMarkdown(block.Content))}
	case "enum_list":
		ret.Content = []string{fmt.Sprintf("1. %s", ctx.richTextToMarkdown(block.Content))}
		ret.Children = []BlockContent{}
	case "bull_list":
		ret.Content = []string{fmt.Sprintf("- %s", ctx.richTextToMarkdown(block.Content))}
		ret.Children = []BlockContent{}
	case "divider":
		ret.Content = []string{"---"}
	case "image":
		ret.Content = []string{ctx.imageToMarkdown(block.Block)}
	case "todo_list":
		ret.Content = []string{ctx.taskListToMarkdown(block.Block)}
		ret.Children = []BlockContent{}
	case "callout":
		ret.Content = ctx.calloutToMarkdown(block.Block)
	case "block_equation":
		ret.Content = []string{fmt.Sprintf("$$%s$$", block.Content[0].Title)}
	case "embed":
		ret.Content = []string{fmt.Sprintf(`<iframe src="%s" width="100%%" style="border:none;"></iframe>`, *block.EmbedLink)}
	case "page":
		title := ctx.richTextToMarkdown(block.Content)
		if strings.TrimSpace(title) == "" {
			title = "untitled-page-" + block.ID
		}
		ctx.ChildPages[block.ID] = title
		ret.Content = []string{fmt.Sprintf("[%s](./%s/index.md)", title, url.PathEscape(title))}
	}

	if block.Type != "enum_list" && block.Type != "bull_list" && block.Type != "todo_list" {
		retContent := []string{
			fmt.Sprintf("<p id=\"%s\">", block.ID),
			"",
		}
		retContent = append(retContent, ret.Content...)
		retContent = append(retContent, "")
		retContent = append(retContent, "</p>")

		ret.Content = retContent
	}

	return ret
}
