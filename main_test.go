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
			Content: []string{"Hello World"},
		},
		{
			Content: []string{"1. list item 1"},
			Children: []convert.BlockContent{{
				Content: []string{"1. child of list item 1"},
				Children: []convert.BlockContent{
					{
						Content: []string{"1. child of child of list item 1"},
					},
					{
						Content: []string{"2. child of child of list item 2"},
					},
				},
			}},
		},
		{
			Content: []string{"2. list item 2"},
		},
	}, ""))
}
