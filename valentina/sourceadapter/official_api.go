package sourceadapter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OfficialAPIAdapter fetches from official public APIs using a bearer/API key.
type OfficialAPIAdapter struct {
	*BaseProvider
	apiKey       string
	apiKeyHeader string
}

// NewOfficialAPIAdapter creates an adapter for official API sources.
// If apiKey is empty, Fetch returns ErrNotConfigured.
func NewOfficialAPIAdapter(name, apiKey, apiKeyHeader string, client HTTPClient, timeout time.Duration) *OfficialAPIAdapter {
	return &OfficialAPIAdapter{
		BaseProvider: NewBaseProvider(AdapterTypeOfficialAPI, name, client, timeout),
		apiKey:       apiKey,
		apiKeyHeader: apiKeyHeader,
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

	resp, err := a.doRequest(ctx, "GET", rawURL, header, sourceID)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := readBodyLimit(resp.Body, 10*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &FetchResult{
		SourceID:    sourceID,
		URL:         rawURL,
		Content:     body,
		ContentType: resp.Header.Get("Content-Type"),
		StatusCode:  resp.StatusCode,
		FetchedAt:   time.Now().UTC(),
		ExtractorID: "official_api_adapter",
		Metadata:    map[string]string{},
	}, nil
}

// readBodyLimit reads at most maxBytes from r.
func readBodyLimit(r io.Reader, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("invalid response limit")
	}
	body, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("response read failed: %w", err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return body, nil
}
