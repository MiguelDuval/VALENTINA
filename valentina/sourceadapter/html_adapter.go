package sourceadapter

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// HTMLAdapter fetches bounded HTML documents with content-type validation.
// It only allows text/html content types and enforces response size limits.
type HTMLAdapter struct {
	*BaseProvider
}

// NewHTMLAdapter creates a bounded HTML document adapter.
func NewHTMLAdapter(name string, client HTTPClient, timeout time.Duration) *HTMLAdapter {
	return &HTMLAdapter{
		BaseProvider: NewBaseProvider(AdapterTypeHTML, name, client, timeout),
	}
}

// Fetch retrieves an HTML document, validating content type and size.
func (a *HTMLAdapter) Fetch(ctx context.Context, sourceID, rawURL string) (*FetchResult, error) {
	resp, err := a.doRequest(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if !isAllowedHTMLContentType(contentType) {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	body, err := readBodyLimit(resp.Body, 1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	// Normalize whitespace but preserve content.
	text := extractTextFromHTML(string(body))

	return &FetchResult{
		SourceID:     sourceID,
		URL:          rawURL,
		Content:      []byte(text),
		ContentType:  contentType,
		StatusCode:   resp.StatusCode,
		FetchedAt:    time.Now().UTC(),
		ExtractorID:  "html_adapter",
		Metadata:     map[string]string{},
	}, nil
}

// isAllowedHTMLContentType checks whether the content type is HTML or XHTML.
func isAllowedHTMLContentType(ct string) bool {
	ct = strings.ToLower(strings.TrimSpace(ct))
	ct = strings.Split(ct, ";")[0]
	return ct == "text/html" || ct == "application/xhtml+xml"
}

// extractTextFromHTML performs a basic strip of HTML tags to return readable text.
// It removes script and style blocks, strips all remaining tags, and collapses whitespace.
func extractTextFromHTML(html string) string {
	// Remove script blocks.
	html = removeBlock(html, "<script", "</script>")
	// Remove style blocks.
	html = removeBlock(html, "<style", "</style>")
	// Strip remaining tags.
	var out strings.Builder
	inTag := false
	for i := 0; i < len(html); i++ {
		if inTag {
			if html[i] == '>' {
				inTag = false
				out.WriteRune(' ')
			}
			continue
		}
		if html[i] == '<' {
			inTag = true
			continue
		}
		out.WriteByte(html[i])
	}
	// Collapse multiple whitespace.
	result := strings.Builder{}
	prevSpace := false
	for _, r := range out.String() {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}
	return strings.TrimSpace(result.String())
}

// removeBlock removes everything from the first occurrence of openTag to the matching closeTag.
func removeBlock(s, openTag, closeTag string) string {
	for {
		start := strings.Index(s, openTag)
		if start < 0 {
			return s
		}
		end := strings.Index(s[start:], closeTag)
		if end < 0 {
			return s[:start]
		}
		end += start + len(closeTag)
		s = s[:start] + s[end:]
	}
}