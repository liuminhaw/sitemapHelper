package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/liuminhaw/sitemap/link"
	"github.com/liuminhaw/sitemap/renderer"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url that you want to build a sitemap for")
	outputDir := flag.String("outputDir", "sitemaps", "directory where result sitemap will be stored")
	maxDepth := flag.Int("depth", 3, "the maximum number of links deep to traverse")
	render := flag.Bool("render", false, "indicate if the parsing url is single page application")
	flag.Parse()

	fmt.Println(*urlFlag)
	submitUrl, err := url.Parse(*urlFlag)
	if err != nil {
		log.Fatal(err)
	}
	outputPath := fmt.Sprintf("%s/%s", *outputDir, submitUrl.Hostname())
	if err := os.MkdirAll(outputPath, 0775); err != nil {
		log.Fatal(err)
	}

	// pages := hrefs(resp.Body, base)
	pages := bfs(*urlFlag, *maxDepth, *render)
	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{Value: page})
	}

	outputFile, err := os.Create(fmt.Sprintf("%s/sitemap.xml", outputPath))
	if err != nil {
		log.Fatal(err)
	}
	defer outputFile.Close()
	// enc := xml.NewEncoder(os.Stdout)
	enc := xml.NewEncoder(outputFile)
	enc.Indent("", "  ")
	// fmt.Print(xml.Header)
	outputFile.Write([]byte(xml.Header))
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}
	fmt.Println()
}

func bfs(urlStr string, maxDepth int, render bool) []string {
	seen := make(map[string]struct{})
	var q map[string]struct{}
	nq := map[string]struct{}{
		urlStr: {},
	}

	for i := 0; i <= maxDepth; i++ {
		q, nq = nq, make(map[string]struct{})
		if len(q) == 0 {
			break
		}
		for url := range q {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			for _, link := range get(url, render) {
				if _, ok := seen[link]; !ok {
					nq[link] = struct{}{}
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

func get(urlStr string, render bool) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		// panic(err)
		return []string{}
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
		ret, _ := renderer.RenderPage(reqUrl.String())
		return filter(hrefs(bytes.NewReader(ret), base), withPrefix(base))
	} else {
		return filter(hrefs(resp.Body, base), withPrefix(base))
	}
}

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

func filter(links []string, keepFn func(string) bool) []string {
	var ret []string
	for _, link := range links {
		if keepFn(link) {
			ret = append(ret, link)
		}
	}

	return ret
}

func withPrefix(pfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, pfx)
	}
}
