package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
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

	// download images
	for url, fileName := range result.Images {
		fmt.Println("downloading image: " + url)

		resp, err2 := http.Get(url)
		if err2 != nil {
			fmt.Println("failed to download image: " + url + ", " + err2.Error())
			os.Exit(exitCodeOutputError)
		}
		contentType := resp.Header.Get("Content-Type")
		fileExtension, err2 := mime.ExtensionsByType(contentType)
		if err2 != nil {
			fmt.Println("failed to get extension of image: " + url + ", " + err2.Error())
			os.Exit(exitCodeOutputError)
		}
		if fileExtension == nil {
			fmt.Println("no file extension associated with content type: " + contentType)
		}

		defer resp.Body.Close()

		err2 = os.MkdirAll(outputDirWithTitle+"/assets/", 0755)
		if err2 != nil {
			fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
			os.Exit(exitCodeOutputError)
		}

		crefile, err2 := os.Create(outputDirWithTitle + "/assets/" + fileName + fileExtension[0])
		if err2 != nil {
			fmt.Println("failed to create image file: " + outputDir + "/" + pageTitle + "/assets/" + fileName + ", error: " + err2.Error())
			os.Exit(exitCodeOutputError)
		}
		defer crefile.Close()

		_, err2 = io.Copy(crefile, resp.Body)
		if err2 != nil {
			fmt.Println("failed to write image file: " + outputDir + "/" + pageTitle + "/assets/" + fileName + ", error: " + err2.Error())
			os.Exit(exitCodeOutputError)
		}

		result.Result = strings.ReplaceAll(result.Result, "["+fileName+"]", "./assets/"+fileName+fileExtension[0])
	}

	err = os.MkdirAll(outputDirWithTitle, 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
		os.Exit(exitCodeOutputError)
	}

	// FIXME: Page content will be overwrite if title duplicated
	err = os.WriteFile(outputDirWithTitle+"/index.md", []byte(result.Result), 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle + "/index.md")
		os.Exit(exitCodeOutputError)
	}

	for childId, childTitle := range result.ChildPages {
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
