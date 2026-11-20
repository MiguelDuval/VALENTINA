package sourceadapter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type hardeningHTTPClient struct{ response *http.Response }

func (h hardeningHTTPClient) Do(*http.Request) (*http.Response, error) { return h.response, nil }
func (hardeningHTTPClient) PolicyValidated() bool                      { return true }

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	copy(p, "a")
	return 1, errors.New("truncated connection")
}

func TestReadBodyLimitRejectsTruncationAndOversize(t *testing.T) {
	if _, err := readBodyLimit(failingReader{}, 10); err == nil {
		t.Fatal("truncated source response accepted")
	}
	if _, err := readBodyLimit(strings.NewReader("01234567890"), 10); err == nil {
		t.Fatal("oversized source response accepted")
	}
	body, err := readBodyLimit(strings.NewReader("0123456789"), 10)
	if err != nil || string(body) != "0123456789" {
		t.Fatalf("exact limit rejected: %q %v", body, err)
	}
}

func TestHTMLAndSitemapRejectNon2xx(t *testing.T) {
	for _, which := range []string{"html", "sitemap"} {
		client := hardeningHTTPClient{response: &http.Response{
			StatusCode: 503, Header: http.Header{"Content-Type": []string{"text/html"}},
			Body: io.NopCloser(strings.NewReader("<html>Service unavailable</html>")),
		}}
		var err error
		if which == "html" {
			_, err = NewHTMLAdapter("test", client, time.Second).Fetch(context.Background(), "a", "https://example.com/feed")
		} else {
			_, err = NewSitemapAdapter("test", client, time.Second).Fetch(context.Background(), "a", "https://example.com/sitemap")
		}
		if err == nil {
			t.Fatalf("%s accepted 503 response", which)
		}
	}
}
