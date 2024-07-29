package main

import (
	"errors"
	"fmt"
	"github.com/lemonnekogh/guolai"
	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/convert"
	"os"
)

const (
	exitCodeParamError = iota + 1
	exitCodeTokenError
	exitCodePermissionError
	exitCodeOutputError
	exitCodeUnknownError
)

func processWolaiError(err guolai.WolaiError, blockId string) {
	if err.Code == 17003 {
		fmt.Println("token is invalid")
		os.Exit(exitCodeTokenError)
	}

	if err.Code == 17011 {
		fmt.Println("failed to get content of block " + blockId + ": permission denied")
		os.Exit(exitCodePermissionError)
	}
}

func checkOutputDir(outputDir string) {
	info, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		fmt.Println("output directory does not exist")
		os.Exit(exitCodeParamError)
	}

	if !info.IsDir() {
		fmt.Println("output directory is not a directory")
		os.Exit(exitCodeParamError)
	}
}

func pageToMarkdown(wolaiClient *guolai.WolaiAPI, pageId string, outputDir string, pageTitle string) {
	children, err := wolaiClient.GetBlockChildren(pageId)
	if err != nil {
		var wolaiErr guolai.WolaiError
		if errors.As(err, &wolaiErr) {
			processWolaiError(wolaiErr, pageId)
		}
		fmt.Printf("failed to get content of block %s: %v\n", pageId, err)
		os.Exit(exitCodeUnknownError)
	}

	result := convert.PageToMarkdown(children)

	err = os.MkdirAll(outputDir+"/"+pageTitle, 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
		os.Exit(exitCodeOutputError)
	}

	err = os.WriteFile(outputDir+"/"+pageTitle+"/index.md", []byte(result.Result), 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle + "/index.md")
		os.Exit(exitCodeOutputError)
	}

	for childId, childTitle := range result.ChildPages {
		pageToMarkdown(wolaiClient, childId, outputDir+"/"+pageTitle, childTitle)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: wodaochu-markdown <wolai-token> <page-id> <output-dir>")
		os.Exit(exitCodeParamError)
	}

	wolaiToken := os.Args[1]
	pageId := os.Args[2]
	outputDir := os.Args[3]

	checkOutputDir(outputDir)

	wolaiClient := guolai.New(wolaiToken)
	pageToMarkdown(wolaiClient, pageId, outputDir, "")
}
