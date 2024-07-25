package main

import (
	"fmt"
	"github.com/lemonnekogh/guolai"
	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/convert"
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

	result := convert.PageToMarkdown(children)

	fmt.Println(result.Result)
}
