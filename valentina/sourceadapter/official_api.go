package sourceadapter

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// OfficialAPIAdapter fetches from official public APIs using a bearer/API key.
type OfficialAPIAdapter struct {
	*BaseProvider
	apiKey     string
	apiKeyHeader string
}

// NewOfficialAPIAdapter creates an adapter for official API sources.
// If apiKey is empty, Fetch returns ErrNotConfigured.
func NewOfficialAPIAdapter(name, apiKey, apiKeyHeader string, client HTTPClient, timeout time.Duration) *OfficialAPIAdapter {
	return &OfficialAPIAdapter{
		BaseProvider:   NewBaseProvider(AdapterTypeOfficialAPI, name, client, timeout),
		apiKey:         apiKey,
		apiKeyHeader:   apiKeyHeader,
	}
}

// Fetch retrieves content from the official API endpoint.
func (a *OfficialAPIAdapter) Fetch(ctx context.Context, sourceID, rawURL string) (*FetchResult, error) {
	if a.apiKey == "" && a.apiKeyHeader != "" {
		return nil, ErrNotConfigured
	}

	header := http.Header{}
	if a.apiKey != "" && a.apiKeyHeader != "" {
		header.Set(a.apiKeyHeader, a.apiKey)
	}

	resp, err := a.doRequest(ctx, "GET", rawURL, header)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := readBodyLimit(resp.Body, 10*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &FetchResult{
		SourceID:     sourceID,
		URL:          rawURL,
		Content:      body,
		ContentType:  resp.Header.Get("Content-Type"),
		StatusCode:   resp.StatusCode,
		FetchedAt:    time.Now().UTC(),
		ExtractorID:  "official_api_adapter",
		Metadata:     map[string]string{},
	}, nil
}

// readBodyLimit reads at most maxBytes from r.
func readBodyLimit(r interface{ Read([]byte) (int, error) }, maxBytes int64) ([]byte, error) {
	// Simple bounded reader using http.MaxBytesReader-like wrapping
	// We use a wrapper since resp.Body may not be *http.Response body directly.
	buf := make([]byte, 0, maxBytes)
	tmp := make([]byte, 4096)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			if int64(len(buf))+int64(n) > maxBytes {
				return buf, fmt.Errorf("response exceeds %d bytes", maxBytes)
			}
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	return buf, nil
}