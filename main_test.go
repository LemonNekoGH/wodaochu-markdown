package main

import (
	"testing"

	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/convert"
	"github.com/stretchr/testify/assert"
)

func TestBlockResultToString(t *testing.T) {
	a := assert.New(t)

	a.Equal(`Hello World
1. list item 1
	1. child of list item 1
		1. child of child of list item 1
		2. child of child of list item 2
2. list item 2
`, blockResultToString([]convert.BlockContent{
		{
			Content: "Hello World",
		},
		{
			Content: "1. list item 1",
			Children: []convert.BlockContent{{
				Content: "1. child of list item 1",
				Children: []convert.BlockContent{
					{
						Content: "1. child of child of list item 1",
					},
					{
						Content: "2. child of child of list item 2",
					},
				},
			}},
		},
		{
			Content: "2. list item 2",
		},
	}, ""))
}
