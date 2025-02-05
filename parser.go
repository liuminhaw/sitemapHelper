package sitemapHelper

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ParseSitemap parses the content of a sitemap from the provided io.Reader and returns a slice of UrlEntry objects.
// The function supports parsing both `<urlset>` and `<sitemapindex>` XML root elements.
//
// For a `<urlset>` root element, it extracts the URLs listed in the sitemap and returns them as UrlEntry objects.
// For a `<sitemapindex>` root element, it recursively fetches and parses the linked sitemaps, aggregating their URLs.
//
// Notes:
//   - The function expects valid XML input adhering to the sitemap protocol specification.
//   - For `<sitemapindex>`, each linked sitemap is fetched via an HTTP GET request, so network access is required.
//   - The function will return an error if the root XML element is unknown or the parsing process fails.
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
			// fmt.Printf("Parsing sitemap: %s\n", sitemap.Loc)
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
