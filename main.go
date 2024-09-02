package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/lemonnekogh/guolai"
	"github.com/lemonnekogh/wodaochu-markdown/internal/pkg/convert"
	"github.com/samber/lo"
)

func checkOutputDir(outputDir string) {
	info, err := os.Stat(outputDir)
	if os.IsNotExist(err) {
		fmt.Println("output directory does not exist")
		os.Exit(convert.ExitCodeParamError)
	}

	if !info.IsDir() {
		fmt.Println("output directory is not a directory")
		os.Exit(convert.ExitCodeParamError)
	}
}

func blockResultToString(results []convert.BlockContent, indent string) string {
	str := lo.Map(results, func(it convert.BlockContent, index int) string {
		content := lo.Map(it.Content, func(line string, _ int) string {
			return indent + line + "\n"
		}) // avoid multiple line break <p> tag
		return strings.Join(content, "") + blockResultToString(it.Children, indent+"\t")
	})

	return strings.Join(str, "")
}

func pageToMarkdown(wolaiClient *guolai.WolaiAPI, pageId string, outputDir string, pageTitle string, root bool) {
	fmt.Printf("fetching content of page: %s, %s\n", pageId, pageTitle)

	children, err := wolaiClient.GetBlockChildren(pageId)
	if err != nil {
		var wolaiErr guolai.WolaiError
		if errors.As(err, &wolaiErr) {
			convert.ProcessWolaiError(wolaiErr, pageId)
			// retry
			pageToMarkdown(wolaiClient, pageId, outputDir, pageTitle, root)
		} else {
			fmt.Printf("failed to get content of block %s: %v\n", pageId, err)
			os.Exit(convert.ExitCodeUnknownError)
		}
	}

	result := convert.PageToMarkdown(wolaiClient, pageTitle, children)
	outputDirWithTitle := outputDir + "/" + pageTitle
	if root {
		outputDirWithTitle = outputDir
	}

	stringResult := blockResultToString(result.Result, "")

	// download images
	for url, fileName := range result.Images {
		fmt.Println("downloading image: " + url)

		resp, err2 := http.Get(url)
		if err2 != nil {
			fmt.Println("failed to download image: " + url + ", " + err2.Error())
			os.Exit(convert.ExitCodeOutputError)
		}
		contentType := resp.Header.Get("Content-Type")
		fileExtension, err2 := mime.ExtensionsByType(contentType)
		if err2 != nil {
			fmt.Println("failed to get extension of image: " + url + ", " + err2.Error())
			os.Exit(convert.ExitCodeOutputError)
		}
		if fileExtension == nil {
			fmt.Println("no file extension associated with content type: " + contentType)
		}

		defer resp.Body.Close()

		err2 = os.MkdirAll(outputDirWithTitle+"/assets/", 0755)
		if err2 != nil {
			fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
			os.Exit(convert.ExitCodeOutputError)
		}

		crefile, err2 := os.Create(outputDirWithTitle + "/assets/" + fileName + fileExtension[0])
		if err2 != nil {
			fmt.Println("failed to create image file: " + outputDir + "/" + pageTitle + "/assets/" + fileName + ", error: " + err2.Error())
			os.Exit(convert.ExitCodeOutputError)
		}
		defer crefile.Close()

		_, err2 = io.Copy(crefile, resp.Body)
		if err2 != nil {
			fmt.Println("failed to write image file: " + outputDir + "/" + pageTitle + "/assets/" + fileName + ", error: " + err2.Error())
			os.Exit(convert.ExitCodeOutputError)
		}

		stringResult = strings.ReplaceAll(stringResult, "["+fileName+"]", "./assets/"+fileName+fileExtension[0])
	}

	err = os.MkdirAll(outputDirWithTitle, 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle)
		os.Exit(convert.ExitCodeOutputError)
	}

	// FIXME: Page content will be overwrite if title duplicated
	err = os.WriteFile(outputDirWithTitle+"/index.md", []byte(stringResult), 0755)
	if err != nil {
		fmt.Println("failed to create convert result to: " + outputDir + "/" + pageTitle + "/index.md")
		os.Exit(convert.ExitCodeOutputError)
	}

	for childId, childTitle := range result.ChildPages {
		pageToMarkdown(wolaiClient, childId, outputDirWithTitle, childTitle, false)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: wodaochu-markdown <wolai-token> <page-id> <output-dir>")
		os.Exit(convert.ExitCodeParamError)
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
			convert.ProcessWolaiError(wolaiErr, pageId)
		} else {
			fmt.Printf("failed to get content of block %s: %v\n", pageId, err)
			os.Exit(convert.ExitCodeUnknownError)
		}
	}

	pageToMarkdown(wolaiClient, pageId, outputDir, page.Block.Content[0].Title, true)
}
