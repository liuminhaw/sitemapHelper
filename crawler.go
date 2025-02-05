package sitemapHelper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/liuminhaw/renderer"
	"github.com/liuminhaw/sitemapHelper/link"
)

// Generate traverses the given URL to the specified depth and fetch for
// all links within the domain to create Urlset struct with loc field as each link.
// Returns the fetched result as an Urlset struct.
func Generate(urlStr string, depth int, render bool) Urlset {
	pages := bfs(urlStr, depth, render)
	toXml := Urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, UrlEntry{Loc: page})
	}

	return toXml
}

// bfs performs a breadth-first search on the given URL to the specified depth
// and returns a slice of all URLs found within the domain.
func bfs(urlStr string, maxDepth int, render bool) []string {
	seen := make(map[string]struct{})
	var queue map[string]struct{}
	newQueue := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i <= maxDepth; i++ {
		queue, newQueue = newQueue, make(map[string]struct{})
		if len(queue) == 0 {
			break
		}
		for url := range queue {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			links, err := get(url, render)
			if err != nil {
				// fmt.Printf("failed to get url from %s\n", url)
				continue
			}
			for _, link := range links {
				if _, ok := seen[link]; !ok {
					newQueue[link] = struct{}{}
				}
			}
		}
	}

	ret := make([]string, 0, len(seen))
	for url := range seen {
		ret = append(ret, url)
	}
	return ret
}

// get fetches the specified URL and returns a slice of strings containing
// all links found within the same domain.
// If the render parameter is true, an automated browser is used to render the page before fetching the links.
// This is particularly useful for client-rendered sites, such as Single Page Applications (SPAs).
// The rendering functionality relies on the github.com/liuminhaw/renderer module, which uses chromedp
// under the hood. As a result, a Chrome browser is required when rendering is enabled.
func get(urlStr string, render bool) ([]string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL
	/*
		/somepath
		https://domain/some-path
		http://domain/some-path
		#fragment
		mailto:username@example.com
	*/

	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}
	base := baseUrl.String()

	if render {
		r := renderer.NewRenderer()

		// fmt.Printf("Rendering: %s\n", reqUrl.String())
		ret, err := r.RenderPage(reqUrl.String(), nil)
		if err != nil {
			return filter(
					hrefs(resp.Body, base),
					withPrefix(base),
				), fmt.Errorf(
					"failed to render page: %w",
					err,
				)
		}

		return filter(hrefs(bytes.NewReader(ret), base), withPrefix(base)), nil
	} else {
		return filter(hrefs(resp.Body, base), withPrefix(base)), nil
	}
}

// hrefs extracts all links from the provided io.Reader and returns them as a slice of strings.
// Based on the extracted links, it combines the base URL with each relative path to form absolute URLs.
func hrefs(r io.Reader, base string) []string {
	links, _ := link.Parse(r)
	var ret []string
	for _, l := range links {
		switch {
		case strings.HasPrefix(l.Href, "/"):
			ret = append(ret, fmt.Sprintf("%s%s", base, l.Href))
		case strings.HasPrefix(l.Href, "http"):
			ret = append(ret, l.Href)
		}
	}

	return ret
}

// filter use the provided keepFn to filter the links slice and return a new slice
// containing only the links that satisfy the keepFn condition.
func filter(links []string, keepFn func(string) bool) []string {
	var ret []string
	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}

	return ret
}

// withPrefix returns a function that checks if the provided string has the specified prefix.
func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
