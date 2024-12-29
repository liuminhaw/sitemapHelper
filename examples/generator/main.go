package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/liuminhaw/sitemapHelper"
)

func main() {
	outputDir := flag.String(
		"outputDir",
		"sitemaps",
		"directory where result sitemap will be stored",
	)
	maxDepth := flag.Int("depth", 3, "the maximum number of links deep to traverse")
	render := flag.Bool("render", false, "indicate if the parsing url is single page application")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %s [options] url\n", os.Args[0])
		os.Exit(1)
	}
	inputUrl := flag.Arg(0)

	submitUrl, err := url.Parse(inputUrl)
	if err != nil {
		fmt.Printf("Error parsing URL %s: %s\n", inputUrl, err)
		os.Exit(1)
	}
	outputPath := fmt.Sprintf("%s/%s", *outputDir, submitUrl.Hostname())
	if err := os.MkdirAll(outputPath, 0775); err != nil {
		fmt.Printf("Failed create output directory %s: %s\n", outputPath, err)
		os.Exit(1)
	}

	xmlUrls := sitemapHelper.Generate(inputUrl, *maxDepth, *render)

	outputFile, err := os.Create(fmt.Sprintf("%s/sitemap.xml", outputPath))
	if err != nil {
		fmt.Printf("Failed to create output file %s: %s\n", outputPath, err)
		os.Exit(1)
	}
	defer outputFile.Close()

	if err := xmlUrls.Write(outputFile); err != nil {
		fmt.Printf("Failed to write sitemap to file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sitemap generated at %s/sitemap.xml\n", outputPath)
}
