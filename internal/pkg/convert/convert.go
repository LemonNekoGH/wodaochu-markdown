package convert

import (
	"fmt"
	"github.com/lemonnekogh/guolai"
	"strings"
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
	FootNotes []string
	Result    string
}

func richTextStyleToMarkdown(text guolai.RichText) string {
	ret := strings.TrimSpace(text.Title)
	if text.Bold {
		ret = fmt.Sprintf("**%s**", ret)
	}
	if text.Italic {
		ret = fmt.Sprintf("_%s_", ret)
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
				txt = fmt.Sprintf("[%s](%s)", txt, *t.Link)
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
	return fmt.Sprintf("```%s\n%s\n```\n", *code.Language, code.Content[0].Title)
}

func (ctx *PageToMarkdownContext) headingToMarkdown(block guolai.Block) string {
	ret := strings.Repeat("#", int(*block.Level))
	ret += " " + ctx.richTextToMarkdown(block.Content)

	return ret
}

func PageToMarkdown(page []guolai.BlockApiResponse) *PageToMarkdownContext {
	ctx := &PageToMarkdownContext{}

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
	ret := fmt.Sprintf("<p id=\"%s\">\n\n", block.ID)
	switch block.Type {
	case "code":
		ret += codeToMarkdown(block.Block)
	case "heading":
		ret += ctx.headingToMarkdown(block.Block)
	case "text":
		ret += ctx.richTextToMarkdown(block.Content)
	}

	return ret + "\n\n</p>"
}
