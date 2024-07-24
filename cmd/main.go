package main

import (
	"fmt"
	"github.com/lemonnekogh/guolai"
	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/b2m"
	"os"
)

func main() {
	pageId := os.Args[2]
	wolaiToken := os.Args[1]

	wolaiClient := guolai.New(wolaiToken)
	children, err := wolaiClient.GetBlockChildren(pageId)
	if err != nil {
		panic(err)
	}

	for _, child := range children {
		fmt.Println(b2m.BlockToMarkdown(child))
	}
}
