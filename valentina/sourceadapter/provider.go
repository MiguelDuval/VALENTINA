package sourceadapter

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// AdapterType identifies the kind of source adapter.
type AdapterType string

const (
	AdapterTypeOfficialAPI AdapterType = "official_api"
	AdapterTypeRSS         AdapterType = "rss"
	AdapterTypeAtom        AdapterType = "atom"
	AdapterTypeSitemap     AdapterType = "sitemap"
	AdapterTypeHTML        AdapterType = "html"
	AdapterTypeDataset     AdapterType = "dataset"
)

// FetchResult is the output of a source adapter fetch.
type FetchResult struct {
	SourceID         string            `json:"source_id"`
	URL              string            `json:"url"`
	CanonicalURL     string            `json:"canonical_url"`
	Content          []byte            `json:"-"` // not serialized to JSON
	ContentType      string            `json:"content_type"`
	StatusCode       int               `json:"status_code"`
	FetchedAt        time.Time         `json:"fetched_at"`
	ExtractorID      string            `json:"extractor_id,omitempty"`
	ExtractorVersion string            `json:"extractor_version,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// Provider is the interface all source adapters implement.
// Each adapter is bound to one source profile and enforces its limits.
type Provider interface {
	// AdapterType returns the type of adapter.
	AdapterType() AdapterType
	// Fetch retrieves content from the given URL for the given source.
	// It must respect source policy limits and return sanitized errors.
	Fetch(ctx context.Context, sourceID, rawURL string) (*FetchResult, error)
	// Name returns a human-readable adapter name.
	Name() string
}

// HTTPClient is an interface for making HTTP requests, allowing test injection.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	PolicyValidated() bool
}

// ErrNotConfigured is returned when an adapter requires credentials that are not set.
var ErrNotConfigured = errors.New("adapter not configured: missing credentials")

// ErrSourceNotFetchable is returned when a source cannot be fetched.
var ErrSourceNotFetchable = errors.New("source not fetchable")

// ErrFetchBlocked is returned when the policy gateway blocks a fetch.
var ErrFetchBlocked = errors.New("fetch blocked by source policy")

// BaseProvider provides common fields for adapters.
type BaseProvider struct {
	adapterType AdapterType
	name        string
	client      HTTPClient
	timeout     time.Duration
}

// NewBaseProvider creates a base provider.
func NewBaseProvider(adapterType AdapterType, name string, client HTTPClient, timeout time.Duration) *BaseProvider {
	// Only a policy-marked client is accepted. A nil or unmarked client fails
	// closed; never fall back to http.DefaultClient.
	if client == nil || !client.PolicyValidated() {
		client = nil
	}
	return &BaseProvider{
		adapterType: adapterType,
		name:        name,
		client:      client,
		timeout:     timeout,
	}
}

func (bp *BaseProvider) AdapterType() AdapterType { return bp.adapterType }
func (bp *BaseProvider) Name() string             { return bp.name }

// WithTimeout sets the HTTP timeout for the provider.
func (bp *BaseProvider) WithTimeout(timeout time.Duration) {
	bp.timeout = timeout
}

// doRequest performs an HTTP GET with context timeout and returns the response.
func (bp *BaseProvider) doRequest(ctx context.Context, method, urlStr string, header http.Header) (*http.Response, error) {
	if bp.client == nil {
		return nil, ErrFetchBlocked
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "VALENTINA/1.0 (+local research bot)")
	if bp.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, bp.timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	return bp.client.Do(req)
}
