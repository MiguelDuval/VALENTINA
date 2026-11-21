package sourceadapter

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"golang.org/x/net/html"
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
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
}

type AtomEntry struct {
	Title     string     `xml:"title"`
	Link      []AtomLink `xml:"link"`
	ID        string     `xml:"id"`
	Updated   string     `xml:"updated"`
	Published string     `xml:"published"`
	Summary   string     `xml:"summary"`
	Content   string     `xml:"content"`
}

// AtomFeed represents an Atom 1.0 feed.
type AtomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Link    []AtomLink  `xml:"link"`
	Updated string      `xml:"updated"`
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
	resp, err := a.doRequest(ctx, "GET", rawURL, nil, sourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := readBodyLimit(resp.Body, maxFeedBytes)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || (mediaType != "application/rss+xml" && mediaType != "application/atom+xml" &&
		mediaType != "application/xml" && mediaType != "text/xml") {
		return nil, errors.New("unsupported feed content type")
	}
	feed, atom, err := parseFeed(body)
	if err != nil {
		return nil, err
	}
	var extracted []byte
	if feed != nil {
		extracted = marshalRSSFeed(*feed)
	} else {
		extracted = marshalAtomFeed(*atom)
	}
	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	return &FetchResult{
		SourceID:     sourceID,
		URL:          rawURL,
		CanonicalURL: finalURL,
		Content:      extracted,
		RawFeed:      body,
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

func atomLinkHref(links []AtomLink) string {
	for _, link := range links {
		if (link.Rel == "" || link.Rel == "alternate") && strings.TrimSpace(link.Href) != "" {
			return strings.TrimSpace(link.Href)
		}
	}
	return ""
}

func marshalAtomFeed(f AtomFeed) []byte {
	out := "Atom Feed: " + f.Title + "\n"
	for _, entry := range f.Entry {
		out += "\n- " + entry.Title + "\n"
		out += "  Link: " + atomLinkHref(entry.Link) + "\n"
		out += "  Updated: " + entry.Updated + "\n"
		out += "  " + entry.Summary + "\n"
	}
	return []byte(out)
}

// FeedCandidate is an observation in a feed, never a verified article or demand.
type FeedCandidate struct {
	Title       string
	Summary     string
	URL         string
	PublishedAt *time.Time
}

const maxFeedBytes = 2 * 1024 * 1024

// parseFeed is shared by Fetch and structured extraction. Reject malformed,
// oversized and non-feed XML instead of treating it as a successful empty feed.
// encoding/xml does not resolve external entities or fetch external resources.
func parseFeed(body []byte) (*RSSFeed, *AtomFeed, error) {
	invalid := errors.New("invalid or unsupported feed XML")
	if len(body) == 0 || len(body) > maxFeedBytes {
		return nil, nil, invalid
	}
	d := xml.NewDecoder(bytes.NewReader(body))
	var rss *RSSFeed
	var atom *AtomFeed
	for {
		token, err := d.Token()
		if err == io.EOF {
			if rss == nil && atom == nil {
				return nil, nil, invalid
			}
			return rss, atom, nil
		}
		if err != nil {
			return nil, nil, invalid
		}
		switch t := token.(type) {
		case xml.Directive:
			return nil, nil, invalid // no DTD/entity declarations
		case xml.CharData:
			if strings.TrimSpace(string(t)) != "" {
				return nil, nil, invalid
			}
		case xml.StartElement:
			if rss != nil || atom != nil {
				return nil, nil, invalid
			}
			switch {
			case t.Name.Local == "rss" && t.Name.Space == "":
				rss = &RSSFeed{}
				err = d.DecodeElement(rss, &t)
			case t.Name.Local == "feed" && (t.Name.Space == "http://www.w3.org/2005/Atom" || t.Name.Space == ""):
				atom = &AtomFeed{}
				err = d.DecodeElement(atom, &t)
			default:
				return nil, nil, invalid
			}
			if err != nil {
				return nil, nil, invalid
			}
		}
	}
}

// ExtractFeedCandidates parses only the supplied bytes. Links are normalized
// observations; this function never fetches them or grants source authorization.
func ExtractFeedCandidates(body []byte, feedURL string) ([]FeedCandidate, error) {
	rss, atom, err := parseFeed(body)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(feedURL)
	if err != nil || base.Hostname() == "" || base.User != nil ||
		(base.Scheme != "https" && base.Scheme != "http") {
		return nil, errors.New("invalid feed base URL")
	}
	candidates := make([]FeedCandidate, 0)
	appendItem := func(title, summary, link, date string) {
		title, summary = feedText(title), feedText(summary)
		link = strings.TrimSpace(link)
		if title == "" || link == "" {
			return
		}
		u, err := url.Parse(link)
		if err != nil {
			return
		}
		u = base.ResolveReference(u)
		if u.Hostname() == "" || u.User != nil || u.Opaque != "" ||
			(u.Scheme != "https" && u.Scheme != "http") {
			return
		}
		u.Host = strings.ToLower(u.Host)
		if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
			u.Host = u.Hostname()
			if strings.Contains(u.Host, ":") {
				u.Host = "[" + u.Host + "]"
			}
		}
		u.Fragment, u.RawFragment = "", ""
		if u.Path == "" {
			u.Path = "/"
		}
		query, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return
		}
		for key := range query {
			lower := strings.ToLower(key)
			if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" {
				query.Del(key)
			}
		}
		u.RawQuery = query.Encode()
		candidates = append(candidates, FeedCandidate{Title: title, Summary: summary, URL: u.String(), PublishedAt: feedDate(date)})
	}
	if rss != nil {
		for _, item := range rss.Channel.Item {
			appendItem(item.Title, item.Description, item.Link, item.PubDate)
		}
	} else {
		for _, entry := range atom.Entry {
			summary := entry.Summary
			if strings.TrimSpace(summary) == "" {
				summary = entry.Content
			}
			// updated is not a publication timestamp; unknown dates stay unknown.
			appendItem(entry.Title, summary, atomLinkHref(entry.Link), entry.Published)
		}
	}
	return candidates, nil
}

func feedDate(value string) *time.Time {
	for _, layout := range []string{time.RFC3339, time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822} {
		if parsed, err := time.Parse(layout, strings.TrimSpace(value)); err == nil {
			// Do not invent an offset for an unrecognized timezone abbreviation.
			zone, offset := parsed.Zone()
			if offset == 0 && zone != "" && zone != "UTC" && zone != "GMT" {
				continue
			}
			utc := parsed.UTC()
			return &utc
		}
	}
	return nil
}

func feedText(value string) string {
	z := html.NewTokenizer(strings.NewReader(value))
	var out strings.Builder
	hidden := ""
	for {
		switch z.Next() {
		case html.ErrorToken:
			return strings.Join(strings.Fields(out.String()), " ")
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "script" || t.Data == "style" {
				hidden = t.Data
			}
			out.WriteByte(' ')
		case html.EndTagToken:
			if z.Token().Data == hidden {
				hidden = ""
			}
			out.WriteByte(' ')
		case html.TextToken:
			if hidden == "" {
				out.Write(z.Text())
			}
		}
	}
}

// RelevantRenovationCandidate is a lexical discovery hint matching the existing
// renovation query packs, not owner-profile qualification or proof of demand.
func RelevantRenovationCandidate(c FeedCandidate) bool {
	words := strings.FieldsFunc(strings.ToLower(c.Title+" "+c.Summary), func(r rune) bool { return !unicode.IsLetter(r) })
	for _, word := range words {
		for _, stem := range []string{"reform", "renova", "rehabilita", "restaura", "reparaci", "repair", "ремонт", "реставрац", "fontaner", "albañil"} {
			if strings.HasPrefix(word, stem) {
				return true
			}
		}
	}
	return false
}
