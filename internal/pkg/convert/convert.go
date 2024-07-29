package convert

import (
	"fmt"
	"net/url"
	"strings"

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

type PageToMarkdownContext struct {
	FootNotes  []string
	Result     string
	ChildPages map[string]string
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

func codeToMarkdown(code guolai.Block) string {
	language := "plaintext"
	if code.Language != nil {
		language = string(*code.Language)
	}
	return fmt.Sprintf("```%s\n%s\n```\n", language, code.Content[0].Title)
}

func (ctx *PageToMarkdownContext) headingToMarkdown(block guolai.Block) string {
	ret := strings.Repeat("#", int(*block.Level))
	ret += " " + ctx.richTextToMarkdown(block.Content)

	return ret
}

func imageToMarkdown(block guolai.Block) string {
	var url string
	if block.Media.Type == "internal" {
		url = *block.Media.DownloadUrl
	} else if block.Media.Type == "external" {
		url = *block.Media.Url
	}

	return fmt.Sprintf("<img src=\"%s\" width=\"%f\" height=\"%f\">", url, *block.Dimensions.Width, *block.Dimensions.Height)
}

func (ctx *PageToMarkdownContext) taskListToMarkdown(block guolai.Block) string {
	checked := " "
	if block.Checked != nil && *block.Checked {
		checked = "x"
	}

	return fmt.Sprintf("- [%s] %s", checked, ctx.richTextToMarkdown(block.Content))
}

func PageToMarkdown(page []guolai.BlockApiResponse) *PageToMarkdownContext {
	ctx := &PageToMarkdownContext{
		ChildPages: map[string]string{},
	}

	for _, block := range page {
		ctx.Result += "\n"
		ctx.Result += ctx.blockToMarkdown(block)
		ctx.Result += "\n"
	}

	for index, footnote := range ctx.FootNotes {
		ctx.Result += fmt.Sprintf("\n[^%d]: %s\n", index+1, footnote)
	}

	return ctx
}

func (ctx *PageToMarkdownContext) blockToMarkdown(block guolai.BlockApiResponse) string {
	ret := ""
	switch block.Type {
	case "code":
		ret += codeToMarkdown(block.Block)
	case "heading":
		ret += ctx.headingToMarkdown(block.Block)
	case "text":
		ret += ctx.richTextToMarkdown(block.Content)
	case "quote":
		ret += fmt.Sprintf("> %s", ctx.richTextToMarkdown(block.Content))
	case "enum_list":
		ret += fmt.Sprintf("1. %s", ctx.richTextToMarkdown(block.Content))
	case "bull_list":
		ret += fmt.Sprintf("- %s", ctx.richTextToMarkdown(block.Content))
	case "divider":
		ret += "---"
	case "image":
		ret += imageToMarkdown(block.Block)
	case "todo_list":
		ret += ctx.taskListToMarkdown(block.Block)
	case "callout":
		ret += fmt.Sprintf("::: tip %s\n%s\n:::", block.Icon.Icon, ctx.richTextToMarkdown(block.Content))
	case "block_equation":
		ret += fmt.Sprintf("$$%s$$", block.Content[0].Title)
	case "embed":
		ret += fmt.Sprintf(`<iframe src="%s" width="100%%" style="border:none;"></iframe>`, *block.EmbedLink)
	case "page":
		title := ctx.richTextToMarkdown(block.Content)
		ctx.ChildPages[block.ID] = title
		ret += fmt.Sprintf("[%s](./%s/index.md)", title, url.PathEscape(title))
	}

	if block.Type != "enum_list" && block.Type != "bull_list" && block.Type != "todo_list" {
		return fmt.Sprintf("<p id=\"%s\">\n\n%s\n\n</p>", block.ID, ret)
	}

	return ret
}
