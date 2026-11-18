package sourceadapter

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// SitemapURL represents a single URL entry in a sitemap.
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// SitemapIndex represents a sitemap index referencing other sitemaps.
type SitemapIndex struct {
	XMLName xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	} `xml:"sitemap"`
}

// URLSet represents a standard sitemap URL set.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapAdapter fetches and parses XML sitemaps.
type SitemapAdapter struct {
	*BaseProvider
}

// NewSitemapAdapter creates a sitemap adapter.
func NewSitemapAdapter(name string, client HTTPClient, timeout time.Duration) *SitemapAdapter {
	return &SitemapAdapter{
		BaseProvider: NewBaseProvider(AdapterTypeSitemap, name, client, timeout),
	}
}

// Fetch retrieves and parses a sitemap XML document.
func (a *SitemapAdapter) Fetch(ctx context.Context, sourceID, rawURL string) (*FetchResult, error) {
	resp, err := a.doRequest(ctx, "GET", rawURL, nil, sourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := readBodyLimit(resp.Body, 2*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")

	// Try sitemap index first.
	var index SitemapIndex
	if err := xml.Unmarshal(body, &index); err == nil && len(index.Sitemaps) > 0 {
		out := "Sitemap Index:\n"
		for _, s := range index.Sitemaps {
			out += "- " + s.Loc + "\n"
		}
		return &FetchResult{
			SourceID:     sourceID,
			URL:          rawURL,
			Content:      []byte(out),
			ContentType:  contentType,
			StatusCode:   resp.StatusCode,
			FetchedAt:    time.Now().UTC(),
			ExtractorID:  "sitemap_adapter",
			Metadata:     map[string]string{},
		}, nil
	}

	// Try standard URL set.
	var urlset URLSet
	if err := xml.Unmarshal(body, &urlset); err == nil && len(urlset.URLs) > 0 {
		out := "Sitemap URLs:\n"
		for _, u := range urlset.URLs {
			out += "- " + u.Loc
			if u.LastMod != "" {
				out += " (lastmod: " + u.LastMod + ")"
			}
			out += "\n"
		}
		return &FetchResult{
			SourceID:     sourceID,
			URL:          rawURL,
			Content:      []byte(out),
			ContentType:  contentType,
			StatusCode:   resp.StatusCode,
			FetchedAt:    time.Now().UTC(),
			ExtractorID:  "sitemap_adapter",
			Metadata:     map[string]string{},
		}, nil
	}

	// Return raw body as fallback.
	return &FetchResult{
		SourceID:     sourceID,
		URL:          rawURL,
		Content:      body,
		ContentType:  contentType,
		StatusCode:   resp.StatusCode,
		FetchedAt:    time.Now().UTC(),
		ExtractorID:  "sitemap_adapter",
		Metadata:     map[string]string{},
	}, nil
}