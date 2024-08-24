package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/lemonnekogh/guolai"
	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/convert"
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

	if err.Code == 17007 {
		fmt.Println("API rate limit exceeded, waiting for 5 seconds...")
		time.Sleep(5 * time.Second)
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

func pageToMarkdown(wolaiClient *guolai.WolaiAPI, pageId string, outputDir string, pageTitle string, root bool) {
	fmt.Printf("fetching content of page: %s, %s\n", pageId, pageTitle)

	children, err := wolaiClient.GetBlockChildren(pageId)
	if err != nil {
		var wolaiErr guolai.WolaiError
		if errors.As(err, &wolaiErr) {
			processWolaiError(wolaiErr, pageId)
			// retry
			pageToMarkdown(wolaiClient, pageId, outputDir, pageTitle, root)
		} else {
			fmt.Printf("failed to get content of block %s: %v\n", pageId, err)
			os.Exit(exitCodeUnknownError)
		}
	}

	result := convert.PageToMarkdown(pageTitle, children)
	outputDirWithTitle := outputDir + "/" + pageTitle
	if root {
		outputDirWithTitle = outputDir
	}

	err = os.MkdirAll(outputDirWithTitle, 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
		os.Exit(exitCodeOutputError)
	}

	err = os.WriteFile(outputDirWithTitle+"/index.md", []byte(result.Result), 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle + "/index.md")
		os.Exit(exitCodeOutputError)
	}

	for childId, childTitle := range result.ChildPages {
		if strings.TrimSpace(childTitle) == "" {
			pageToMarkdown(wolaiClient, childId, outputDirWithTitle, "untitledNewPage", false)
			continue
		}
		pageToMarkdown(wolaiClient, childId, outputDirWithTitle, childTitle, false)
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

	page, err := wolaiClient.GetBlocks(pageId)
	if err != nil {
		var wolaiErr guolai.WolaiError
		if errors.As(err, &wolaiErr) {
			processWolaiError(wolaiErr, pageId)
		} else {
			fmt.Printf("failed to get content of block %s: %v\n", pageId, err)
			os.Exit(exitCodeUnknownError)
		}
	}

	pageToMarkdown(wolaiClient, pageId, outputDir, page.Block.Content[0].Title, true)
}
