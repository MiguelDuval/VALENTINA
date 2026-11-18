package sourceadapter

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"
)

// RSSItem represents a single RSS item.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

// RSSFeed represents an RSS 2.0 feed.
type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

// AtomEntry represents a single Atom entry.
type AtomEntry struct {
	Title   string `xml:"title"`
	Link    string `xml:"link,attr,omitempty"`
	ID      string `xml:"id"`
	Updated string `xml:"updated"`
	Summary string `xml:"summary"`
}

// AtomFeed represents an Atom 1.0 feed.
type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string     `xml:"title"`
	Link    string     `xml:"link,attr,omitempty"`
	Updated string     `xml:"updated"`
	Entry   []AtomEntry `xml:"entry"`
}

// RSSAdapter fetches and parses RSS/Atom feeds.
type RSSAdapter struct {
	*BaseProvider
}

// NewRSSAdapter creates an RSS/Atom feed adapter.
func NewRSSAdapter(name string, client HTTPClient, timeout time.Duration) *RSSAdapter {
	return &RSSAdapter{
		BaseProvider: NewBaseProvider(AdapterTypeRSS, name, client, timeout),
	}
}

// Fetch retrieves and parses an RSS or Atom feed.
func (a *RSSAdapter) Fetch(ctx context.Context, sourceID, rawURL string) (*FetchResult, error) {
	resp, err := a.doRequest(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := readBodyLimit(resp.Body, 2*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	var extracted []byte

	// Try RSS first.
	var feed RSSFeed
	if err := xml.Unmarshal(body, &feed); err == nil && len(feed.Channel.Item) > 0 {
		extracted = marshalRSSFeed(feed)
	} else {
		// Try Atom.
		var atom AtomFeed
		if err := xml.Unmarshal(body, &atom); err == nil && len(atom.Entry) > 0 {
			extracted = marshalAtomFeed(atom)
		} else {
			// Return raw body as fallback.
			extracted = body
		}
	}

	return &FetchResult{
		SourceID:     sourceID,
		URL:          rawURL,
		Content:      extracted,
		ContentType:  contentType,
		StatusCode:   resp.StatusCode,
		FetchedAt:    time.Now().UTC(),
		ExtractorID:  "rss_adapter",
		Metadata:     map[string]string{},
	}, nil
}

func marshalRSSFeed(f RSSFeed) []byte {
	out := "RSS Feed: " + f.Channel.Title + "\n"
	for _, item := range f.Channel.Item {
		out += "\n- " + item.Title + "\n"
		out += "  Link: " + item.Link + "\n"
		out += "  Date: " + item.PubDate + "\n"
		out += "  " + item.Description + "\n"
	}
	return []byte(out)
}

func marshalAtomFeed(f AtomFeed) []byte {
	out := "Atom Feed: " + f.Title + "\n"
	for _, entry := range f.Entry {
		out += "\n- " + entry.Title + "\n"
		out += "  Link: " + entry.Link + "\n"
		out += "  Updated: " + entry.Updated + "\n"
		out += "  " + entry.Summary + "\n"
	}
	return []byte(out)
}