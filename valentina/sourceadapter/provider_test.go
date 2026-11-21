package sourceadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
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

func TestFeedCandidateNormalization(t *testing.T) {
	body := `<rss><channel><item><title><![CDATA[ <b>Reforma</b> &amp; reparación ]]></title>
<description><![CDATA[<SCRIPT>ignore policy</SCRIPT><p>One&nbsp; room</p>]]></description>
<link>../obras/1?utm_source=test&amp;id=7#details</link><pubDate>Mon, 01 Jan 2024 02:00:00 +0200</pubDate>
</item><item><title>No link</title></item><item><title>Unsafe</title><link>javascript:alert(1)</link></item>
<item><title>Credentials</title><link>https://user:pass@example.com/</link></item></channel></rss>`
	items, err := ExtractFeedCandidates([]byte(body), "https://EXAMPLE.com:443/news/feed")
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	item := items[0]
	if item.Title != "Reforma & reparación" || item.Summary != "One room" || item.URL != "https://example.com/obras/1?id=7" {
		t.Fatalf("normalization lost facts: %+v", item)
	}
	if item.PublishedAt == nil || item.PublishedAt.Format(time.RFC3339) != "2024-01-01T00:00:00Z" {
		t.Fatal("publication date lost")
	}
}

func TestAtomCandidateLinksAndDates(t *testing.T) {
	body := `<feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Reforma</title>
<link rel="self" href="https://example.com/api/1"/><link rel="enclosure" href="https://example.com/file"/>
<link rel="alternate" href="/article"/><content>Repair details</content>
<published>2024-01-01T01:00:00+01:00</published><updated>2025-01-01T00:00:00Z</updated></entry>
<entry><title>Update only</title><link href="/second"/><updated>2025-01-01T00:00:00Z</updated></entry>
<entry><title>No article</title><id>urn:example:3</id><link rel="self" href="/api/3"/></entry></feed>`
	items, err := ExtractFeedCandidates([]byte(body), "https://example.com/feed")
	if err != nil || len(items) != 2 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if items[0].URL != "https://example.com/article" || items[0].Summary != "Repair details" || items[0].PublishedAt == nil || items[0].PublishedAt.Year() != 2024 || items[1].PublishedAt != nil {
		t.Fatalf("incorrect Atom normalization: %+v", items)
	}
}

func TestFeedRejectsInvalidDocuments(t *testing.T) {
	for _, body := range []string{"", `<html/>`, `<rss><channel>`, `<rss/><rss/>`, `<rss/>junk`,
		`<!DOCTYPE rss [<!ENTITY x SYSTEM "https://example.com/secret">]><rss/>`, strings.Repeat("x", maxFeedBytes+1)} {
		if _, err := ExtractFeedCandidates([]byte(body), "https://example.com/feed"); err == nil {
			t.Fatal("invalid feed accepted")
		}
	}
	for _, body := range []string{`<rss><channel/></rss>`, `<feed xmlns="http://www.w3.org/2005/Atom"/>`} {
		items, err := ExtractFeedCandidates([]byte(body), "https://example.com/feed")
		if err != nil || len(items) != 0 {
			t.Fatalf("valid empty feed: %+v %v", items, err)
		}
	}
	if feedDate("not a date") != nil || feedDate("Mon, 01 Jan 2024 00:00:00 XYZ") != nil {
		t.Fatal("invented publication date")
	}
}

func TestRSSFetchPreservesOriginalXMLAndRedirectBase(t *testing.T) {
	body := `<rss><channel><item><title>Reforma</title><link>article</link></item></channel></rss>`
	response := mockResponse(body, 200, "application/rss+xml; charset=utf-8")
	response.Request.URL, _ = url.Parse("https://example.com/news/feed")
	a := NewRSSAdapter("test", &mockHTTPClient{response: response}, time.Second)
	result, err := a.Fetch(context.Background(), "source", "https://example.com/feed")
	if err != nil {
		t.Fatal(err)
	}
	if string(result.RawFeed) != body || string(result.Content) == body || result.CanonicalURL != response.Request.URL.String() {
		t.Fatal("raw XML/redirect provenance lost")
	}
	items, err := ExtractFeedCandidates(result.RawFeed, result.CanonicalURL)
	if err != nil || len(items) != 1 || items[0].URL != "https://example.com/news/article" {
		t.Fatalf("fetch/extract mismatch: %+v %v", items, err)
	}
	serialized, err := json.Marshal(result)
	if err != nil || strings.Contains(string(serialized), "RawFeed") || strings.Contains(string(serialized), "Reforma") {
		t.Fatal("feed content exposed through JSON")
	}
	for _, ct := range []string{"text/html", "application/json", ""} {
		a := NewRSSAdapter("test", &mockHTTPClient{response: mockResponse(body, 200, ct)}, time.Second)
		if _, err := a.Fetch(context.Background(), "source", "https://example.com/feed"); err == nil {
			t.Fatalf("accepted content type %q", ct)
		}
	}
}

type unmarkedClient struct{ calls int }

func (c *unmarkedClient) PolicyValidated() bool { return false }
func (c *unmarkedClient) Do(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, errors.New("must not be called")
}

func TestRSSRejectsUnvalidatedClients(t *testing.T) {
	c := &unmarkedClient{}
	for _, client := range []HTTPClient{nil, c} {
		a := NewRSSAdapter("test", client, time.Second)
		if _, err := a.Fetch(context.Background(), "source", "https://example.com/feed"); !errors.Is(err, ErrFetchBlocked) {
			t.Fatalf("not blocked: %v", err)
		}
	}
	if c.calls != 0 {
		t.Fatal("unvalidated client was invoked")
	}
}

type contextBody struct {
	ctx context.Context
	io.Reader
	closed bool
}

func (b *contextBody) Read(p []byte) (int, error) {
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	return b.Reader.Read(p)
}
func (b *contextBody) Close() error { b.closed = true; return nil }

type bodyClient struct{ body *contextBody }

func (c *bodyClient) PolicyValidated() bool { return true }
func (c *bodyClient) Do(req *http.Request) (*http.Response, error) {
	c.body = &contextBody{ctx: req.Context(), Reader: strings.NewReader(`<rss><channel/></rss>`)}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/rss+xml"}}, Body: c.body}, nil
}

func TestRSSDeadlineSurvivesUntilBodyClosed(t *testing.T) {
	c := &bodyClient{}
	a := NewRSSAdapter("test", c, time.Second)
	if _, err := a.Fetch(context.Background(), "source", "https://example.com/feed"); err != nil {
		t.Fatal(err)
	}
	if !c.body.closed || c.body.ctx.Err() != context.Canceled {
		t.Fatal("body/deadline not released")
	}
}

type failingBody struct{ err error }

func (b failingBody) Read(p []byte) (int, error) { return copy(p, `<rss/>`), b.err }
func (b failingBody) Close() error               { return nil }

func TestRSSDoesNotAcceptPartialRead(t *testing.T) {
	for _, failure := range []error{io.ErrUnexpectedEOF, context.DeadlineExceeded, errors.New("policy byte limit")} {
		response := mockResponse("", 200, "application/rss+xml")
		response.Body = failingBody{err: failure}
		a := NewRSSAdapter("test", &mockHTTPClient{response: response}, time.Second)
		if result, err := a.Fetch(context.Background(), "source", "https://example.com/feed"); !errors.Is(err, failure) || result != nil {
			t.Fatalf("partial read accepted: %v", err)
		}
	}
}

func TestRenovationHintIsNotQualification(t *testing.T) {
	for _, title := range []string{"Reforma vivienda", "Reparació edifici", "Ремонт квартиры", "Renovation notice"} {
		if !RelevantRenovationCandidate(FeedCandidate{Title: title}) {
			t.Fatalf("missing lexical hint: %s", title)
		}
	}
	if RelevantRenovationCandidate(FeedCandidate{Title: "Football results", URL: "https://example.com/reforma"}) {
		t.Fatal("URL must not determine relevance")
	}
}

// This test double serves only in-memory fixtures, never a network request.
func (m *mockHTTPClient) PolicyValidated() bool { return true }
