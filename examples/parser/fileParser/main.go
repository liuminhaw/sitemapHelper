package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/liuminhaw/sitemapHelper"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <filepath>\n", os.Args[0])
		os.Exit(1)
	}
	filepath := os.Args[1]

	_, err := os.Stat(filepath)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File %s not found\n", filepath)
		os.Exit(1)
	}

	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Error opening file %s: %s\n", filepath, err)
		os.Exit(1)
	}
	defer file.Close()

	entries, err := sitemapHelper.ParseSitemap(file)
	if err != nil {
		fmt.Printf("Error parsing urls from file %s: %s\n", filepath, err)
		os.Exit(1)
	}

	for i, entry := range entries {
		fmt.Printf("===== Url index %d =====\n", i)
		fmt.Printf("Loc: %s\n", entry.Loc)
		if entry.Lastmod != "" {
			fmt.Printf("Lastmod: %s\n", entry.Lastmod)
		}
		if entry.Changefreq != "" {
			fmt.Printf("Changefreq: %s\n", entry.Changefreq)
		}
		if entry.Priority != "" {
			fmt.Printf("Priority: %s\n", entry.Priority)
		}
	}
}
