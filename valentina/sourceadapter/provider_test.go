package sourceadapter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

// mockHTTPClient implements HTTPClient for testing.
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

func mockResponse(body string, statusCode int, contentType string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{contentType}},
		Body:       io.NopCloser(&readerFromString{data: body}),
		Request:    &http.Request{},
	}
}

type readerFromString struct {
	data string
	pos  int
}

func (s *readerFromString) Read(p []byte) (int, error) {
	if s.pos >= len(s.data) {
		return 0, io.EOF
	}
	n := copy(p, s.data[s.pos:])
	s.pos += n
	return n, nil
}

func TestOfficialAPIAdapterMissingKey(t *testing.T) {
	a := NewOfficialAPIAdapter("test", "", "X-API-Key", &mockHTTPClient{}, 10*time.Second)
	_, err := a.Fetch(nil, "src-1", "https://api.example.com/data")
	if err == nil {
		t.Fatal("expected ErrNotConfigured when API key is missing")
	}
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestOfficialAPIAdapterSuccess(t *testing.T) {
	a := NewOfficialAPIAdapter("test", "secret-key", "X-API-Key", &mockHTTPClient{
		response: mockResponse(`{"items":[{"id":1}]}`, 200, "application/json"),
	}, 10*time.Second)
	res, err := a.Fetch(context.Background(), "src-1", "https://api.example.com/data")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if res.SourceID != "src-1" {
		t.Fatalf("unexpected source ID: %s", res.SourceID)
	}
	if res.StatusCode != 200 {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	if res.ContentType != "application/json" {
		t.Fatalf("unexpected content type: %s", res.ContentType)
	}
}

func TestRSSAdapterParsesFeed(t *testing.T) {
	rss := `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
<title>Test Feed</title>
<link>https://example.com</link>
<description>Test</description>
<item>
<title>Item One</title>
<link>https://example.com/item1</link>
<pubDate>Mon, 01 Jan 2024 00:00:00 GMT</pubDate>
<description>First item</description>
</item>
</channel>
</rss>`
	a := NewRSSAdapter("test", &mockHTTPClient{response: mockResponse(rss, 200, "application/rss+xml")}, 10*time.Second)
	res, err := a.Fetch(context.Background(), "src-1", "https://example.com/feed")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	content := string(res.Content)
	if content == "" {
		t.Fatal("expected extracted content")
	}
	if !stringsContains(content, "Item One") {
		t.Fatal("expected item title in extracted content")
	}
}

func TestAtomAdapterParsesFeed(t *testing.T) {
	atom := `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
<title>Atom Feed</title>
<entry>
<title>Atom Entry</title>
<link href="https://example.com/entry1"/>
<id>https://example.com/entry1</id>
<updated>2024-01-01T00:00:00Z</updated>
<summary>Entry summary</summary>
</entry>
</feed>`
	a := NewRSSAdapter("test", &mockHTTPClient{response: mockResponse(atom, 200, "application/atom+xml")}, 10*time.Second)
	res, err := a.Fetch(context.Background(), "src-1", "https://example.com/feed")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	content := string(res.Content)
	if !stringsContains(content, "Atom Entry") {
		t.Fatal("expected atom entry title in extracted content")
	}
}

func TestSitemapAdapterParsesURLSet(t *testing.T) {
	sitemap := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url>
<loc>https://example.com/page1</loc>
<lastmod>2024-01-01</lastmod>
</url>
<url>
<loc>https://example.com/page2</loc>
</url>
</urlset>`
	a := NewSitemapAdapter("test", &mockHTTPClient{response: mockResponse(sitemap, 200, "application/xml")}, 10*time.Second)
	res, err := a.Fetch(context.Background(), "src-1", "https://example.com/sitemap.xml")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	content := string(res.Content)
	if !stringsContains(content, "https://example.com/page1") {
		t.Fatal("expected page1 URL in sitemap output")
	}
	if !stringsContains(content, "https://example.com/page2") {
		t.Fatal("expected page2 URL in sitemap output")
	}
}

func TestHTMLAdapterExtractsText(t *testing.T) {
	html := `<html><head><title>Test</title></head><body><h1>Hello</h1><p>Some <b>bold</b> text.</p><script>var x = 1;</script><style>body{color:red;}</style></body></html>`
	a := NewHTMLAdapter("test", &mockHTTPClient{response: mockResponse(html, 200, "text/html")}, 10*time.Second)
	res, err := a.Fetch(context.Background(), "src-1", "https://example.com/page")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	content := string(res.Content)
	if !stringsContains(content, "Hello") {
		t.Fatal("expected heading text")
	}
	if !stringsContains(content, "bold") {
		t.Fatal("expected bold text")
	}
	if stringsContains(content, "var x = 1") {
		t.Fatal("script content should be stripped")
	}
	if stringsContains(content, "color:red") {
		t.Fatal("style content should be stripped")
	}
}

func TestHTMLAdapterRejectsNonHTML(t *testing.T) {
	a := NewHTMLAdapter("test", &mockHTTPClient{response: mockResponse(`{"json":true}`, 200, "application/json")}, 10*time.Second)
	_, err := a.Fetch(context.Background(), "src-1", "https://example.com/data")
	if err == nil {
		t.Fatal("expected error for non-HTML content type")
	}
}

func TestStringsContainsHelper(t *testing.T) {
	if !stringsContains("hello world", "world") {
		t.Fatal("helper broken")
	}
	if stringsContains("hello world", "universe") {
		t.Fatal("helper should return false for missing substring")
	}
}

func stringsContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestBaseProviderTimeout(t *testing.T) {
	bp := NewBaseProvider(AdapterTypeOfficialAPI, "test", &mockHTTPClient{}, 10*time.Second)
	if bp.timeout != 10*time.Second {
		t.Fatal("timeout not set")
	}
	bp.WithTimeout(5 * time.Second)
	if bp.timeout != 5*time.Second {
		t.Fatal("WithTimeout did not change timeout")
	}
}