package sourcepolicy

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type concurrencyTransport func(*http.Request) (*http.Response, error)

func (f concurrencyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestValidatedClientHoldsSlotUntilBodyClosed(t *testing.T) {
	transport := concurrencyTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header),
			Body: io.NopCloser(strings.NewReader("payload"))}, nil
	})
	c := &ValidatedClient{client: &http.Client{Transport: transport}, concurrency: NewConcurrencyLimiter(1)}
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "https://example.com/public", nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Do(req); !errors.Is(err, ErrConcurrencyLimit) {
		t.Fatalf("slot released before body close: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close() // idempotent release must not corrupt another request
	next, err := c.Do(req)
	if err != nil {
		t.Fatalf("slot not released: %v", err)
	}
	defer next.Body.Close()
}

func TestValidatedClientReleasesSlotWhenTransportFails(t *testing.T) {
	calls := 0
	transport := concurrencyTransport(func(*http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return nil, errors.New("network failed")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header),
			Body: io.NopCloser(strings.NewReader("ok"))}, nil
	})
	c := &ValidatedClient{client: &http.Client{Transport: transport}, concurrency: NewConcurrencyLimiter(1)}
	req, _ := http.NewRequest("GET", "https://example.com/public", nil)
	_, _ = c.Do(req)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("failed request leaked concurrency slot: %v", err)
	}
	defer resp.Body.Close()
}
