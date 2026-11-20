package sourceadapter

import (
	"encoding/xml"
	"errors"
	"html"
	"net/url"
	"strings"
	"time"
)

// FeedCandidate is an unverified notice observed in an RSS/Atom feed.
// The target article has NOT been fetched; preserve the feed as provenance.
type FeedCandidate struct {
	Title       string
	URL         string
	Summary     string
	PublishedAt *time.Time
}

const maxFeedCandidates = 100

// ExtractFeedCandidates parses only genuine RSS/Atom XML (never arbitrary HTML).
// It limits the returned entries to keep one manual run bounded.
func ExtractFeedCandidates(raw []byte, feedURL string) ([]FeedCandidate, error) {
	if len(raw) == 0 || len(raw) > 2*1024*1024 {
		return nil, errors.New("feed is empty or exceeds 2 MB")
	}
	base, err := url.Parse(feedURL)
	if err != nil || base.Hostname() == "" {
		return nil, errors.New("invalid feed URL")
	}
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal(raw, &root); err != nil {
		return nil, errors.New("invalid XML feed")
	}
	out := make([]FeedCandidate, 0)
	appendEntry := func(title, href, summary, date string) {
		if len(out) >= maxFeedCandidates {
			return
		}
		title = compactFeedText(title, 240)
		href = strings.TrimSpace(href)
		if title == "" || href == "" {
			return
		}
		parsed, err := url.Parse(href)
		if err != nil || parsed.User != nil {
			return
		}
		resolved := base.ResolveReference(parsed)
		if resolved.User != nil || resolved.Hostname() == "" ||
			(resolved.Scheme != "https" && resolved.Scheme != "http") {
			return
		}
		resolved.Fragment = ""
		item := FeedCandidate{
			Title:   title,
			URL:     resolved.String(),
			Summary: compactFeedText(summary, 1200),
		}
		for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339, time.RFC822Z, time.RFC822} {
			if t, err := time.Parse(layout, strings.TrimSpace(date)); err == nil {
				t = t.UTC()
				item.PublishedAt = &t
				break
			}
		}
		out = append(out, item)
	}
	switch root.XMLName.Local {
	case "rss":
		var feed RSSFeed
		if err := xml.Unmarshal(raw, &feed); err != nil {
			return nil, errors.New("invalid RSS feed")
		}
		for _, item := range feed.Channel.Item {
			appendEntry(item.Title, item.Link, item.Description, item.PubDate)
		}
	case "feed":
		var feed AtomFeed
		if err := xml.Unmarshal(raw, &feed); err != nil {
			return nil, errors.New("invalid Atom feed")
		}
		for _, entry := range feed.Entry {
			appendEntry(entry.Title, atomLinkHref(entry.Link), entry.Summary, entry.Updated)
		}
	default:
		return nil, errors.New("unsupported XML feed format")
	}
	return out, nil
}

func compactFeedText(value string, max int) string {
	value = html.UnescapeString(value)
	value = extractTextFromHTML(value)
	value = strings.Join(strings.Fields(value), " ")
	value = strings.TrimSpace(value)
	r := []rune(value)
	if len(r) > max {
		value = string(r[:max])
	}
	return value
}

// RelevantRenovationCandidate is an explainable conservative keyword filter.
// No match means not relevant enough for automatic inclusion in the review queue.
func RelevantRenovationCandidate(c FeedCandidate) bool {
	haystack := strings.ToLower(c.Title + " " + c.Summary)
	for _, keyword := range []string{
		"reform", "renova", "rehabilit", "reparaci", "obra", "construcci",
		"albañil", "pintur", "fontaner", "tabiquer", "alicat", "impermeabili",
		"pladur", "mantenimient", "building repair", "renovation", "refurbish",
		"construction", "interior finishing", "remont", "ремонт", "отделк", "стройк", "строитель",
	} {
		if strings.Contains(haystack, keyword) {
			return true
		}
	}
	return false
}
