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

type Sitemap interface {
	Parse(r io.Reader) error
}

type UrlEntry struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

type Urlset struct {
	Urls  []UrlEntry `xml:"url"`
	Xmlns string     `xml:"xmlns,attr"`
}

// Write writes the XML to provided writer.
func (u *Urlset) Write(w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	w.Write([]byte(xml.Header))
	if err := enc.Encode(u); err != nil {
		return fmt.Errorf("failed to write xml: %w", err)
	}

	return nil
}

func (u *Urlset) Parse(r io.Reader) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(u); err != nil {
		return fmt.Errorf("failed to read xml: %w", err)
	}

	return nil
}

type sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

type SitemapIndex struct {
	Sitemaps []sitemap `xml:"sitemap"`
	Xmlns    string    `xml:"xmlns,attr"`
}

func (s *SitemapIndex) Parse(r io.Reader) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(s); err != nil {
		return fmt.Errorf("failed to read xml: %w", err)
	}

	return nil
}

// Generate traverses the given URL to the specified depth and fetch for
// all links within the domain to create Urlset struct with loc field as each link.
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

func ParseSitemap(r io.Reader) ([]UrlEntry, error) {
	type rootElement struct {
		XMLName xml.Name
	}

	var seeker io.Seeker
	if _, ok := r.(io.Seeker); ok {
		seeker = r.(io.Seeker)
	} else {
		content, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("failed to get reader bytes: %w", err)
		}
		r = bytes.NewReader(content)
		seeker = r.(io.Seeker)
	}

	var root rootElement
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("error decoding xml: %w", err)
	}

	// Reset the reader to the beginning of the file
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset reader: %w", err)
	}

	switch strings.ToLower(root.XMLName.Local) {
	case "urlset":
		var urls []UrlEntry
		var u Urlset
		if err := u.Parse(r); err != nil {
			return nil, fmt.Errorf("failed to parse urlset: %w", err)
		}
		for _, entry := range u.Urls {
			urls = append(urls, entry)
		}
		return urls, nil
	case "sitemapindex":
		var s SitemapIndex
		if err := s.Parse(r); err != nil {
			return nil, fmt.Errorf("failed to parse sitemapindex: %w", err)
		}

		var respUrls []UrlEntry
		for _, sitemap := range s.Sitemaps {
			fmt.Printf("Parsing sitemap: %s\n", sitemap.Loc)
			resp, err := http.Get(sitemap.Loc)
			if err != nil {
				return nil, fmt.Errorf("failed to get sitemap: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("sitemap response status error: %s", resp.Status)
			}

			urls, err := ParseSitemap(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to parse sitemap: %w", err)
			}
			respUrls = append(respUrls, urls...)

		}
		return respUrls, nil
	}

	return nil, fmt.Errorf("unknown root element: %s", root.XMLName.Local)
}
