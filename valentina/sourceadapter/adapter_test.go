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

type stubHTTPClient struct {
	response        *http.Response
	err             error
	policyValidated bool
}

func (s stubHTTPClient) Do(*http.Request) (*http.Response, error) { return s.response, s.err }
func (s stubHTTPClient) PolicyValidated() bool                    { return s.policyValidated }

func TestAdaptersFailClosedWithoutPolicyHTTPClient(t *testing.T) {
	rss := NewRSSAdapter("rss", nil, time.Second)
	if _, err := rss.Fetch(context.Background(), "source-1", "https://example.com/feed"); !errors.Is(err, ErrFetchBlocked) {
		t.Fatalf("RSS adapter must fail closed without a policy client, got %v", err)
	}
	api := NewOfficialAPIAdapter("api", "", "", nil, time.Second)
	if _, err := api.Fetch(context.Background(), "source-1", "https://example.com/api"); !errors.Is(err, ErrFetchBlocked) {
		t.Fatalf("API adapter must fail closed without a policy client, got %v", err)
	}
}

func TestAtomAdapterExtractsAlternateHref(t *testing.T) {
	feed := `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><title>Tenders</title><entry><title>Bathroom works</title><link rel="alternate" href="https://example.com/tender/1"/><updated>2026-11-18T10:00:00Z</updated><summary>Public works notice</summary></entry></feed>`
	client := stubHTTPClient{policyValidated: true, response: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/atom+xml"}},
		Body:       io.NopCloser(strings.NewReader(feed)),
	}}
	adapter := NewRSSAdapter("atom", client, time.Second)
	result, err := adapter.Fetch(context.Background(), "source-1", "https://example.com/feed")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !strings.Contains(string(result.Content), "https://example.com/tender/1") {
		t.Fatalf("Atom alternate href missing from extracted content: %s", result.Content)
	}
}

func TestAdaptersRejectNonSuccessHTTPStatus(t *testing.T) {
	client := stubHTTPClient{policyValidated: true, response: &http.Response{StatusCode: http.StatusServiceUnavailable, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("unavailable"))}}
	adapter := NewRSSAdapter("rss", client, time.Second)
	if _, err := adapter.Fetch(context.Background(), "source-1", "https://example.com/feed"); err == nil || !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("expected HTTP 503 error, got %v", err)
	}
}
