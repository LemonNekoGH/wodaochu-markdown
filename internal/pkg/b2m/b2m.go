package b2m

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
func richTextToMarkdown(text []guolai.RichText) string {
	ret := ""
	for _, t := range text {
		switch t.Type {
		case "text":
			ret += richTextStyleToMarkdown(t)
		case "equation":
			ret += fmt.Sprintf("$%s$", t.Title)
		}
	}
	return ret
}

func codeToMarkdown(code guolai.Block) string {
	caption := ""
	if code.Caption != nil {
		caption = fmt.Sprintf(`<div style="color:#838383;margin:-0.75rem 10px 0;">%s</div>`, *code.Caption)
	}

	// captionAlign := "left" // wolai api not return this value

	return fmt.Sprintf("```%s\n%s\n```\n%s\n", *code.Language, code.Content[0].Title, caption)
}

func headingToMarkdown(block guolai.Block) string {
	ret := strings.Repeat("#", int(*block.Level))
	ret += richTextToMarkdown(block.Content)

	return ret
}

func BlockToMarkdown(block guolai.BlockApiResponse) string {
	switch block.Type {
	case "code":
		return codeToMarkdown(block.Block)
	case "heading":
		return headingToMarkdown(block.Block)
	case "text":
		return richTextToMarkdown(block.Block.Content)
	}

	return ""
}
