package sitemapHelper

import (
	"encoding/xml"
	"fmt"
	"io"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

// UrlEntry represents a single URL information in the sitemap.
// Matching the <url> tag in <urlset> of the sitemap
type UrlEntry struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// Urlset containes urls of the sitemap, matching the <urlset> tag of the sitemap
type Urlset struct {
	Urls  []UrlEntry `xml:"url"`
	Xmlns string     `xml:"xmlns,attr"`
}

// Write writes Urlset in xml format to provided io writer
func (u *Urlset) Write(w io.Writer) error {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	w.Write([]byte(xml.Header))
	if err := enc.Encode(u); err != nil {
		return fmt.Errorf("failed to write xml: %w", err)
	}

	return nil
}

// Parse reads and decode provided io Reader to self Urlset struct
func (u *Urlset) Parse(r io.Reader) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(u); err != nil {
		return fmt.Errorf("failed to read xml: %w", err)
	}

	return nil
}

// sitemap represent a single sitemap information in the sitemapindex
// Matching the <sitemap> tag in <sitemapindex>
type sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

// SitemapIndex contains sitemaps list, matching the <sitemapindex> tag in the sitemap
type SitemapIndex struct {
	Sitemaps []sitemap `xml:"sitemap"`
	Xmlns    string    `xml:"xmlns,attr"`
}

// Parse reads and decode provided io Reader to self SitemapIndex struct
func (s *SitemapIndex) Parse(r io.Reader) error {
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(s); err != nil {
		return fmt.Errorf("failed to read xml: %w", err)
	}

	return nil
}
