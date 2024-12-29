package sitemapHelper

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/liuminhaw/renderer"
	"github.com/liuminhaw/sitemapHelper/link"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type Urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func (u Urlset) Write(w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	w.Write([]byte(xml.Header))
	if err := enc.Encode(u); err != nil {
		return err
	}

	return nil
}

func Generate(urlStr string, depth int, render bool) Urlset {
	pages := bfs(urlStr, depth, render)
	toXml := Urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{Value: page})
	}

	return toXml
}

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
				fmt.Printf("failed to get url from %s\n", url)
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
		rendererContext := renderer.RendererContext{
			WindowWidth:  1920,
			WindowHeight: 1080,
			Timeout:      60,
		}
		ctx := renderer.WithRendererContext(context.Background(), &rendererContext)

		fmt.Printf("Rendering: %s\n", reqUrl.String())
		ret, err := renderer.RenderPage(ctx, reqUrl.String())
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
