package sourcepolicy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// PolicyGateway enforces VALENTINA source policy for every fetch.
// It validates scheme, host, path, redirects, DNS/IP resolution, and SSRF protections.
type PolicyGateway struct {
	registry *Registry
}

// NewPolicyGateway creates a gateway backed by the given registry.
func NewPolicyGateway(registry *Registry) *PolicyGateway {
	return &PolicyGateway{registry: registry}
}

// FetchDecision is the result of evaluating a fetch request against policy.
type FetchDecision struct {
	Allowed        bool   `json:"allowed"`
	SourceID       string `json:"source_id"`
	CanonicalURL   string `json:"canonical_url,omitempty"`
	Reason         string `json:"reason,omitempty"`
	InternalReason string `json:"-"`
}

// Evaluate checks whether a fetch of the given URL under the given source ID is permitted.
// Fail-closed: any uncertainty results in Allowed=false.
func (g *PolicyGateway) Evaluate(sourceID, rawURL string) *FetchDecision {
	if g.registry == nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "policy gateway not initialized",
			InternalReason: "nil registry",
		}
	}

	profile := g.registry.Get(sourceID)
	if profile == nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "source not registered",
			InternalReason: ErrSourceNotFound.Error(),
		}
	}

	if !profile.CanFetch() {
		var reason, internal string
		switch profile.ApprovalStatus {
		case ApprovalPending:
			reason, internal = "source pending review", ErrSourceNotApproved.Error()
		case ApprovalPaused:
			reason, internal = "source paused", ErrSourcePaused.Error()
		case ApprovalBlocked:
			reason, internal = "source blocked", ErrSourceDisabled.Error()
		default:
			reason, internal = "source not fetchable", "unknown status"
		}
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         reason,
			InternalReason: internal,
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "invalid URL",
			InternalReason: err.Error(),
		}
	}

	// Reject URL userinfo (credentials in URL are never allowed).
	if parsed.User != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "userinfo not allowed in URL",
			InternalReason: "URL contains userinfo",
		}
	}

	// Reject malformed ports (non-numeric or out of range).
	if parsed.Port() != "" {
		portStr := parsed.Port()
		for _, c := range portStr {
			if c < '0' || c > '9' {
				return &FetchDecision{
					Allowed:        false,
					SourceID:       sourceID,
					Reason:         "malformed port",
					InternalReason: "port contains non-digit characters",
				}
			}
		}
		portNum := 0
		for _, c := range portStr {
			portNum = portNum*10 + int(c-'0')
		}
		if portNum < 1 || portNum > 65535 {
			return &FetchDecision{
				Allowed:        false,
				SourceID:       sourceID,
				Reason:         "port out of range",
				InternalReason: "port not in 1-65535",
			}
		}
	}

	// Reject empty hostnames.
	if parsed.Hostname() == "" {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "empty hostname",
			InternalReason: "URL has no host",
		}
	}

	// Reject ambiguous hosts: malformed IPv6 literals and leading-dot names.
	hostname := parsed.Hostname()
	if strings.Contains(hostname, ":") && net.ParseIP(hostname) == nil {
		return &FetchDecision{Allowed: false, SourceID: sourceID, Reason: "ambiguous host", InternalReason: "invalid IPv6 literal"}
	}
	if strings.HasPrefix(hostname, ".") {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "ambiguous host",
			InternalReason: "hostname starts with dot",
		}
	}

	// Scheme check.
	schemeOK := false
	for _, s := range profile.AllowedSchemes {
		if strings.EqualFold(parsed.Scheme, s) {
			schemeOK = true
			break
		}
	}
	if !schemeOK {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "scheme not allowed",
			InternalReason: ErrSchemeNotAllowed.Error(),
		}
	}

	// Restrict ports to the standard port for the selected scheme. This
	// prevents a trusted hostname from being used to reach unrelated services.
	if port := parsed.Port(); port != "" {
		portNumber, portErr := strconv.Atoi(port)
		expectedPort := 443
		if strings.EqualFold(parsed.Scheme, "http") {
			expectedPort = 80
		}
		if portErr != nil || portNumber < 1 || portNumber > 65535 || strconv.Itoa(portNumber) != port || portNumber != expectedPort {
			return &FetchDecision{Allowed: false, SourceID: sourceID, Reason: "port not allowed", InternalReason: "non-standard or invalid destination port"}
		}
	}

	// Host check (exact match against allowed hosts).
	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	hostOK := false
	for _, h := range profile.AllowedHosts {
		if host == strings.TrimSuffix(strings.ToLower(strings.TrimSpace(h)), ".") {
			hostOK = true
			break
		}
	}
	if !hostOK {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "host not allowed",
			InternalReason: ErrHostNotAllowed.Error(),
		}
	}

	// Decode once before checking path segments. Reject encoded separators and
	// double-encoded dot sequences because downstream servers may decode again.
	escapedPath := parsed.EscapedPath()
	decodedPath, pathErr := url.PathUnescape(escapedPath)
	if pathErr != nil || strings.ContainsAny(decodedPath, "\\\\") ||
		strings.Contains(strings.ToLower(decodedPath), "%2e") ||
		strings.Contains(strings.ToLower(decodedPath), "%2f") ||
		strings.Contains(strings.ToLower(decodedPath), "%5c") {
		return &FetchDecision{Allowed: false, SourceID: sourceID, Reason: "path not allowed", InternalReason: "malformed or multiply encoded path"}
	}
	for _, segment := range strings.Split(decodedPath, "/") {
		if segment == "." || segment == ".." {
			return &FetchDecision{Allowed: false, SourceID: sourceID, Reason: "path traversal not allowed", InternalReason: "path contains dot-segments"}
		}
	}
	pathVal := path.Clean(decodedPath)
	if pathVal == "." || pathVal == "" {
		pathVal = "/"
	}
	pathOK := false
	for _, prefix := range profile.AllowedPathPrefixes {
		p := path.Clean(prefix)
		if p == "." || p == "" {
			p = "/"
		}
		if strings.HasPrefix(p, "/") != strings.HasPrefix(pathVal, "/") {
			continue
		}
		if strings.HasSuffix(p, "/") {
			if strings.HasPrefix(pathVal, p) {
				pathOK = true
				break
			}
		} else if pathVal == p || strings.HasPrefix(pathVal, p+"/") {
			pathOK = true
			break
		}
	}
	if !pathOK {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "path not allowed",
			InternalReason: ErrPathNotAllowed.Error(),
		}
	}

	// DNS/IP resolution: reject if ANY resolved address is blocked.
	ips, err := resolveHosts(host)
	if err != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "DNS resolution failed",
			InternalReason: err.Error(),
		}
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return &FetchDecision{
				Allowed:        false,
				SourceID:       sourceID,
				Reason:         "destination IP not permitted",
				InternalReason: ErrPrivateIP.Error(),
			}
		}
	}

	canonical := canonicalURL(parsed)

	return &FetchDecision{
		Allowed:        true,
		SourceID:       sourceID,
		CanonicalURL:   canonical,
		Reason:         "allowed",
		InternalReason: "",
	}
}

// EvaluateRedirect re-validates a redirect target URL against the same source policy.
// Every redirect hop must independently pass scheme/host/path/SSRF checks.
// Redirects to a different host than the original are always rejected.
func (g *PolicyGateway) EvaluateRedirect(sourceID, fromURL, redirectURL string) *FetchDecision {
	fromParsed, err := url.Parse(fromURL)
	if err != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "redirect not allowed",
			InternalReason: "invalid from URL",
		}
	}
	redirectParsed, err := url.Parse(redirectURL)
	if err != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "redirect not allowed",
			InternalReason: "invalid redirect URL",
		}
	}
	// Redirects to a different host than the original are rejected.
	if !strings.EqualFold(fromParsed.Hostname(), redirectParsed.Hostname()) {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "redirect not allowed",
			InternalReason: "redirect crosses hosts",
		}
	}
	decision := g.Evaluate(sourceID, redirectURL)
	if !decision.Allowed {
		decision.Reason = "redirect not allowed"
		decision.InternalReason = ErrRedirectNotAllowed.Error()
	}
	return decision
}

// resolveHosts resolves a hostname to ALL IP addresses.
func resolveHosts(host string) ([]net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("no IP addresses returned")
	}
	return ips, nil
}

// newValidatedTransport disables environment proxies and pins each connection
// to a validated IP. Callers must use NewValidatedHTTPClient so URL policy is
// checked before every request; this raw transport is intentionally unexported.
func newValidatedTransport(profile *SourceProfile) *http.Transport {
	return &http.Transport{
		Proxy:                 nil,
		DialContext:           makeValidatedDialer(profile),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     true,
	}
}

type policyRoundTripper struct {
	gateway  *PolicyGateway
	sourceID string
	base     http.RoundTripper
}

type maxBytesReadCloser struct {
	body io.ReadCloser
	max  int64
	read int64
}

func (r *maxBytesReadCloser) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.read >= r.max {
		var probe [1]byte
		n, err := r.body.Read(probe[:])
		if n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	remaining := r.max - r.read
	if int64(len(p)) > remaining {
		p = p[:remaining]
	}
	n, err := r.body.Read(p)
	r.read += int64(n)
	return n, err
}

func (r *maxBytesReadCloser) Close() error { return r.body.Close() }

func (p *policyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	decision := p.gateway.Evaluate(p.sourceID, req.URL.String())
	if !decision.Allowed {
		return nil, fmt.Errorf("source policy denied request: %s", decision.Reason)
	}
	resp, err := p.base.RoundTrip(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}
	currentProfile := p.gateway.registry.Get(p.sourceID)
	if currentProfile == nil || !currentProfile.CanFetch() {
		_ = resp.Body.Close()
		return nil, ErrSourceNotFound
	}
	maxBytes := currentProfile.MaxResponseBytes
	if maxBytes <= 0 {
		maxBytes = 512 * 1024
	}
	if resp.ContentLength > int64(maxBytes) {
		_ = resp.Body.Close()
		return nil, ErrResponseTooLarge
	}
	resp.Body = &maxBytesReadCloser{body: resp.Body, max: int64(maxBytes)}
	return resp, nil
}

// ValidatedClient is the only HTTP client type intended for source adapters.
// It wraps a policy-enforcing client and exposes an explicit marker so adapters
// cannot accidentally fall back to http.DefaultClient.
// Per-source rate, daily quota and concurrency limits are enforced on every Do call.
type ValidatedClient struct {
	client       *http.Client
	rateLimiter  *RateLimiter
	dailyQuota   *DailyQuotaTracker
	concurrency  *ConcurrencyLimiter
	acquired     bool
}

func (c *ValidatedClient) Do(req *http.Request) (*http.Response, error) {
	if c == nil || c.client == nil {
		return nil, ErrSourceNotFound
	}
	// Enforce minimum interval between requests (rate limit).
	if c.rateLimiter != nil {
		if !c.rateLimiter.Allow() {
			return nil, ErrRateLimited
		}
	}
	// Enforce daily request quota.
	if c.dailyQuota != nil {
		if !c.dailyQuota.CheckAndIncrement(sourceIDFromRequest(req)) {
			return nil, ErrDailyQuotaExceeded
		}
	}
	// Enforce concurrency limit (non-blocking: returns immediately if at capacity).
	if c.concurrency != nil {
		if !c.concurrency.TryAcquire() {
			return nil, ErrConcurrencyLimit
		}
		c.acquired = true
	}
	resp, err := c.client.Do(req)
	if c.acquired && c.concurrency != nil {
		c.concurrency.Release()
		c.acquired = false
	}
	return resp, err
}

func (c *ValidatedClient) PolicyValidated() bool { return c != nil && c.client != nil }

// sourceIDFromRequest extracts the source ID from a request context or URL.
// It is used for quota tracking when the source ID is not directly available.
func sourceIDFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	if sid := req.Context().Value(sourceIDKey{}); sid != nil {
		if s, ok := sid.(string); ok {
			return s
		}
	}
	return ""
}

// sourceIDKey is a context key for passing the source ID through the request.
type sourceIDKey struct{}

// WithSourceIDContext returns a context carrying the given source ID.
func WithSourceIDContext(ctx context.Context, sourceID string) context.Context {
	return context.WithValue(ctx, sourceIDKey{}, sourceID)
}

// WithSourceID returns a request with the source ID in its context.
func WithSourceID(req *http.Request, sourceID string) *http.Request {
	if req == nil {
		return nil
	}
	return req.WithContext(WithSourceIDContext(req.Context(), sourceID))
}

// TryAcquire attempts to acquire a concurrency slot without blocking.
// Returns true if acquired, false if at capacity.
func (cl *ConcurrencyLimiter) TryAcquire() bool {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.active < cl.limit {
		cl.active++
		return true
	}
	return false
}

// Release releases a previously acquired concurrency slot.
func (cl *ConcurrencyLimiter) Release() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if cl.active > 0 {
		cl.active--
	}
}

// NewValidatedHTTPClient validates the initial request and every redirect hop
// against source approval, scheme, host, path and DNS/IP policy. Blocked
// redirects are never followed; the underlying dialer pins validated IPs.
// Per-source rate, daily quota and concurrency limits are configured from the
// registered profile and enforced on every Do call.
func NewValidatedHTTPClient(g *PolicyGateway, sourceID string, profile *SourceProfile) (*ValidatedClient, error) {
	if g == nil || g.registry == nil || profile == nil || sourceID == "" || profile.SourceID != sourceID {
		return nil, errors.New("matching source policy gateway, source ID and profile are required")
	}
	// Use the registry's current profile, never caller-supplied limits or hosts.
	registeredProfile := g.registry.Get(sourceID)
	if registeredProfile == nil || !registeredProfile.CanFetch() {
		return nil, ErrSourceNotFound
	}
	profile = registeredProfile
	timeout := time.Duration(profile.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	base := newValidatedTransport(profile)

	// Configure per-source rate limiter from minimum interval.
	var rateLimiter *RateLimiter
	if profile.MinimumIntervalSeconds > 0 {
		ratePerSec := 1.0 / float64(profile.MinimumIntervalSeconds)
		rateLimiter = NewRateLimiter(ratePerSec, 1)
	}

	// Configure daily quota tracker.
	var dailyQuota *DailyQuotaTracker
	if profile.DailyRequestLimit != nil {
		dailyQuota = NewDailyQuotaTracker()
		dailyQuota.SetLimit(sourceID, profile.DailyRequestLimit)
	}

	// Configure concurrency limiter.
	var concurrency *ConcurrencyLimiter
	if profile.MaxParallelRequests > 0 {
		concurrency = NewConcurrencyLimiter(profile.MaxParallelRequests)
	}

	return &ValidatedClient{
		client:       &http.Client{
			Transport: &policyRoundTripper{gateway: g, sourceID: sourceID, base: base},
			Timeout:   timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) == 0 || len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				from := via[len(via)-1].URL.String()
				decision := g.EvaluateRedirect(sourceID, from, req.URL.String())
				if !decision.Allowed {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		rateLimiter:  rateLimiter,
		dailyQuota:   dailyQuota,
		concurrency:  concurrency,
	}, nil
}

// makeValidatedDialer resolves the hostname once, rejects the whole answer set
// if any address is blocked, and dials one of those validated IPs directly.
// This prevents DNS rebinding between validation and connection establishment.
func makeValidatedDialer(profile *SourceProfile) func(ctx context.Context, network, addr string) (net.Conn, error) {
	var dialer net.Dialer
	return makeValidatedDialerWith(profile, resolveHosts, dialer.DialContext)
}

func makeValidatedDialerWith(profile *SourceProfile, resolver func(string) ([]net.IP, error), dial func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if profile == nil || resolver == nil || dial == nil {
			return nil, ErrSourceNotFound
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("invalid destination address")
		}
		host = strings.TrimSuffix(strings.ToLower(host), ".")
		allowed := false
		for _, allowedHost := range profile.AllowedHosts {
			if host == strings.TrimSuffix(strings.ToLower(allowedHost), ".") {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, ErrHostNotAllowed
		}
		ips, err := resolver(host)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, errors.New("no IP addresses returned")
		}
		for _, ip := range ips {
			if isBlockedIP(ip) {
				return nil, ErrPrivateIP
			}
		}
		var lastErr error
		for _, ip := range ips {
			conn, dialErr := dial(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		if lastErr == nil {
			lastErr = errors.New("no validated destination IPs")
		}
		return nil, lastErr
	}
}

// isBlockedIP returns true for loopback, private, link-local, multicast, reserved,
// and metadata-range addresses.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() {
		return true
	}
	if ip.IsUnspecified() {
		return true
	}
	if ip.IsMulticast() {
		return true
	}
	if ip.IsLinkLocalUnicast() {
		return true
	}
	if ip.IsLinkLocalMulticast() {
		return true
	}
	if !ip.IsGlobalUnicast() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 0:
			return true
		case ip4[0] == 10:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
			return true
		case ip4[0] == 192 && ip4[1] == 0 && (ip4[2] == 0 || ip4[2] == 2):
			return true
		case ip4[0] == 192 && ip4[1] == 88 && ip4[2] == 99:
			return true
		case ip4[0] == 198 && ip4[1] == 51 && ip4[2] == 100:
			return true
		case ip4[0] == 203 && ip4[1] == 0 && ip4[2] == 113:
			return true
		case ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127:
			return true
		case ip4[0] == 198 && ip4[1] >= 18 && ip4[1] <= 19:
			return true
		case ip4[0] >= 240:
			return true
		}
	} else {
		if len(ip) >= 1 && ip[0]&0xfe == 0xfc {
			return true
		}
		if len(ip) >= 6 && ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8 {
			return true
		}
	}
	return false
}

// canonicalURL normalizes a URL for deduplication and evidence storage.
func canonicalURL(u *url.URL) string {
	scheme := strings.ToLower(u.Scheme)
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	port := u.Port()
	if port != "" {
		switch scheme {
		case "http":
			if port == "80" {
				port = ""
			}
		case "https":
			if port == "443" {
				port = ""
			}
		}
	}
	hostPart := host
	if strings.Contains(host, ":") {
		hostPart = "[" + host + "]"
	}
	if port != "" {
		hostPart = net.JoinHostPort(host, port)
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	out := scheme + "://" + hostPart + path
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	return out
}

// IsBlockedIPExported exposes the SSRF check for tests and adapters.
func IsBlockedIPExported(ip net.IP) bool {
	return isBlockedIP(ip)
}
