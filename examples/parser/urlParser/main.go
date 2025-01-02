package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/liuminhaw/sitemapHelper"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <url>\n", os.Args[0])
		os.Exit(1)
	}
	inputUrl := os.Args[1]

	resp, err := http.Get(inputUrl)
	if err != nil {
		fmt.Printf("Error getting URL %s: %s\n", inputUrl, err)
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non OK response status: %s\n", resp.Status)
		os.Exit(1)
	}

	entries, err := sitemapHelper.ParseSitemap(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing urls from url %s: %s\n", inputUrl, err)
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
