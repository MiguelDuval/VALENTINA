package sourcepolicy

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func makeTestRegistry() *Registry {
	r := NewRegistry()
	reviewedAt := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	_ = r.Register(&SourceProfile{
		SourceID:            "allowed-src",
		Name:                "Allowed Source",
		Publisher:           "Test Pub",
		ApprovalStatus:      ApprovalApproved,
		ApprovalReason:      "Terms and public API scope reviewed for test.",
		TermsReviewedAt:     &reviewedAt,
		AllowedHosts:        []string{"example.com", "api.valencia.es"},
		AllowedPathPrefixes: []string{"/api/", "/public/"},
		AllowedSchemes:      []string{"https"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	})
	_ = r.Register(&SourceProfile{
		SourceID:            "pending-src",
		Name:                "Pending Source",
		Publisher:           "Test Pub",
		ApprovalStatus:      ApprovalPending,
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	})
	_ = r.Register(&SourceProfile{
		SourceID:            "paused-src",
		Name:                "Paused Source",
		Publisher:           "Test Pub",
		ApprovalStatus:      ApprovalPaused,
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	})
	return r
}

func TestGatewayEvaluateApprovedAllowed(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "https://example.com/api/v1/tenders")
	if !d.Allowed {
		t.Fatalf("expected allowed, got: %s (internal: %s)", d.Reason, d.InternalReason)
	}
	if d.CanonicalURL == "" {
		t.Fatal("expected canonical URL")
	}
}

func TestGatewayEvaluatePendingRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("pending-src", "https://example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("pending source should be rejected")
	}
	if d.Reason != "source pending review" {
		t.Fatalf("unexpected reason: %s", d.Reason)
	}
}

func TestGatewayEvaluatePausedRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("paused-src", "https://example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("paused source should be rejected")
	}
}

func TestGatewayEvaluateBlockedRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	_ = r.Register(&SourceProfile{
		SourceID:            "blocked-src",
		Name:                "Blocked Source",
		Publisher:           "Test Pub",
		ApprovalStatus:      ApprovalBlocked,
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	})
	d := g.Evaluate("blocked-src", "https://example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("blocked source should be rejected")
	}
}

func TestGatewayEvaluateUnknownSourceRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("unknown-src", "https://example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("unknown source should be rejected")
	}
	if d.Reason != "source not registered" {
		t.Fatalf("unexpected reason: %s", d.Reason)
	}
}

func TestGatewayEvaluateDisallowedHostRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "https://evil.example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("disallowed host should be rejected")
	}
	if d.Reason != "host not allowed" {
		t.Fatalf("unexpected reason: %s", d.Reason)
	}
	d = g.Evaluate("allowed-src", "https://other-site.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("unregistered host should be rejected")
	}
}

func TestGatewayEvaluateDisallowedSchemeRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "http://example.com/api/v1/tenders")
	if d.Allowed {
		t.Fatal("non-https scheme should be rejected for https-only source")
	}
}

func TestGatewayEvaluateHTTPSchemeAllowedForHTTPSource(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	reviewedAt := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	_ = r.Register(&SourceProfile{
		SourceID:            "http-src",
		Name:                "HTTP Scheme Source",
		Publisher:           "Test Pub",
		ApprovalStatus:      ApprovalApproved,
		ApprovalReason:      "Terms reviewed for test",
		TermsReviewedAt:     &reviewedAt,
		AllowedHosts:        []string{"example.com"},
		AllowedPathPrefixes: []string{"/"},
		AllowedSchemes:      []string{"http"},
		TimeoutSeconds:      10,
		MaxResponseBytes:    1024,
	})
	d := g.Evaluate("http-src", "http://example.com/api/v1/tenders")
	if !d.Allowed {
		t.Fatalf("http scheme should be allowed for http-src: %s", d.Reason)
	}
}

func TestGatewayEvaluateDisallowedPathRejected(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "https://example.com/private/data")
	if d.Allowed {
		t.Fatal("disallowed path should be rejected")
	}
	if d.Reason != "path not allowed" {
		t.Fatalf("unexpected reason: %s", d.Reason)
	}
}

func TestGatewayEvaluateInvalidURL(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "not a url")
	if d.Allowed {
		t.Fatal("invalid URL should be rejected")
	}
}

func TestGatewayEvaluateEmptyURL(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "")
	if d.Allowed {
		t.Fatal("empty URL should be rejected")
	}
}

func TestGatewayEvaluateNilRegistry(t *testing.T) {
	g := NewPolicyGateway(nil)
	d := g.Evaluate("any-src", "https://example.com/")
	if d.Allowed {
		t.Fatal("nil registry should reject all")
	}
}

func TestCanonicalURL(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.Evaluate("allowed-src", "https://Example.COM:443/api/v1/Tenders?x=1#section")
	if !d.Allowed {
		t.Fatalf("expected allowed: %s", d.Reason)
	}
	if d.CanonicalURL != "https://example.com/api/v1/Tenders?x=1" {
		t.Fatalf("unexpected canonical URL: %s", d.CanonicalURL)
	}
	d = g.Evaluate("allowed-src", "https://example.com:80/api/")
	if d.Allowed || d.Reason != "port not allowed" {
		t.Fatalf("non-standard HTTPS port must be rejected, got allowed=%v reason=%q", d.Allowed, d.Reason)
	}
	d = g.Evaluate("allowed-src", "https://example.com/api/")
	if !d.Allowed {
		t.Fatalf("expected allowed: %s", d.Reason)
	}
	if d.CanonicalURL != "https://example.com/api/" {
		t.Fatalf("unexpected canonical URL: %s", d.CanonicalURL)
	}
}

func TestGatewayEvaluateRedirectAllowed(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.EvaluateRedirect("allowed-src", "https://example.com/api/", "https://example.com/api/redirected")
	if !d.Allowed {
		t.Fatalf("redirect to same host/path should be allowed: %s", d.Reason)
	}
}

func TestGatewayEvaluateRedirectDisallowedHost(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.EvaluateRedirect("allowed-src", "https://example.com/api/", "https://evil.example.com/api/")
	if d.Allowed {
		t.Fatal("redirect to disallowed host should be rejected")
	}
	if d.Reason != "redirect not allowed" {
		t.Fatalf("unexpected redirect reason: %s", d.Reason)
	}
}

func TestGatewayEvaluateRedirectDisallowedPath(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.EvaluateRedirect("allowed-src", "https://example.com/api/", "https://example.com/private/")
	if d.Allowed {
		t.Fatal("redirect to disallowed path should be rejected")
	}
}

func TestGatewayEvaluateRedirectPendingSource(t *testing.T) {
	r := makeTestRegistry()
	g := NewPolicyGateway(r)
	d := g.EvaluateRedirect("pending-src", "https://example.com/api/", "https://example.com/api/")
	if d.Allowed {
		t.Fatal("redirect for pending source should be rejected")
	}
}

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{"loopback IPv4", "127.0.0.1", true},
		{"loopback IPv6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"private 172.31.x", "172.31.255.255", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"link-local 169.254.x", "169.254.169.254", true},
		{"link-local IPv6 fe80", "fe80::1", true},
		{"unique local IPv6 fc00", "fc00::1", true},
		{"multicast IPv4", "224.0.0.1", true},
		{"multicast IPv6", "ff02::1", true},
		{"unspecified 0.0.0.0", "0.0.0.0", true},
		{"documentation 192.0.2", "192.0.2.1", true},
		{"documentation 198.51.100", "198.51.100.1", true},
		{"documentation 203.0.113", "203.0.113.1", true},
		{"reserved 192.0.0", "192.0.0.1", true},
		{"public IP", "9.9.9.9", false},
		{"public IP 2", "1.1.1.1", false},
		{"documentation IPv6 2001:db8", "2001:db8::1", true},
		{"shared address space 100.64", "100.64.0.1", true},
		{"benchmarking 198.18", "198.18.0.1", true},
		{"reserved 240.x", "240.0.0.1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("invalid test IP: %s", tt.ip)
			}
			got := isBlockedIP(ip)
			if got != tt.want {
				t.Fatalf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestSanitizedError(t *testing.T) {
	err := &SanitizedError{internal: net.ErrClosed, safeMsg: "blocked"}
	if err.Error() != "blocked" {
		t.Fatalf("SanitizedError should return safe message, got %q", err.Error())
	}
	if err.Unwrap() != net.ErrClosed {
		t.Fatal("Unwrap should return internal error")
	}
	sanitized := Sanitize(net.ErrClosed)
	if sanitized == nil {
		t.Fatal("Sanitize should not return nil for non-nil error")
	}
	msg := sanitized.Error()
	if msg == "net: errclosed" || msg == "use of closed network connection" {
		t.Fatal("Sanitize should not leak internal error message")
	}
}

func TestForbidden(t *testing.T) {
	err := NewForbidden("internal reason")
	if err.Error() != "forbidden: internal reason" {
		t.Fatalf("unexpected Forbidden error: %v", err)
	}
}

func TestGatewayRejectsURLUserInfoAndEncodedTraversal(t *testing.T) {
	g := NewPolicyGateway(makeTestRegistry())
	for _, raw := range []string{
		"https://user:pass@example.com/api/v1/tenders",
		"https://example.com/api/%2e%2e/private",
		"https://example.com/api/%252e%252e/private",
		"https://example.com:8443/api/",
		"https://example.com:0443/api/",
		"https://example.com/apix/private",
	} {
		d := g.Evaluate("allowed-src", raw)
		if d.Allowed {
			t.Fatalf("expected URL to be rejected: %s", raw)
		}
	}
}

func TestValidatedDialerRejectsMixedPublicAndPrivateDNSAnswers(t *testing.T) {
	profile := makeTestRegistry().Get("allowed-src")
	dialCalled := false
	dialer := makeValidatedDialerWith(profile, func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("9.9.9.9"), net.ParseIP("127.0.0.1")}, nil
	}, func(context.Context, string, string) (net.Conn, error) {
		dialCalled = true
		return nil, nil
	})
	_, err := dialer(context.Background(), "tcp", "example.com:443")
	if err != ErrPrivateIP {
		t.Fatalf("expected mixed DNS answer to be blocked, got %v", err)
	}
	if dialCalled {
		t.Fatal("dial must not be called when any DNS answer is blocked")
	}
}

func TestValidatedDialerPinsConnectionToValidatedIP(t *testing.T) {
	profile := makeTestRegistry().Get("allowed-src")
	var dialAddress string
	dialer := makeValidatedDialerWith(profile, func(string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("9.9.9.9")}, nil
	}, func(_ context.Context, network, address string) (net.Conn, error) {
		dialAddress = address
		client, server := net.Pipe()
		_ = server.Close()
		return client, nil
	})
	conn, err := dialer(context.Background(), "tcp", "example.com:443")
	if err != nil {
		t.Fatalf("validated dial failed: %v", err)
	}
	_ = conn.Close()
	if dialAddress != "9.9.9.9:443" {
		t.Fatalf("expected dial to pinned IP, got %q", dialAddress)
	}
}

func TestValidatedTransportDoesNotUseEnvironmentProxy(t *testing.T) {
	transport := newValidatedTransport(makeTestRegistry().Get("allowed-src"))
	if transport.Proxy != nil {
		t.Fatal("validated transport must not use environment proxies")
	}
}

func TestValidatedHTTPClientEnforcesInitialPathPolicy(t *testing.T) {
	gateway := NewPolicyGateway(makeTestRegistry())
	client, err := NewValidatedHTTPClient(gateway, "allowed-src", makeTestRegistry().Get("allowed-src"))
	if err != nil {
		t.Fatalf("NewValidatedHTTPClient: %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.com/private/data", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	_, err = client.Do(req)
	if err == nil || !strings.Contains(err.Error(), "path not allowed") {
		t.Fatalf("expected initial request to be rejected by path policy, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestPolicyRoundTripperEnforcesResponseSizeLimit(t *testing.T) {
	registry := NewRegistry()
	reviewedAt := time.Now().UTC()
	profile := &SourceProfile{SourceID: "allowed-src", ApprovalStatus: ApprovalApproved, ApprovalReason: "terms reviewed for test", TermsReviewedAt: &reviewedAt, AllowedHosts: []string{"example.com"}, AllowedPathPrefixes: []string{"/api/"}, AllowedSchemes: []string{"https"}, MaxResponseBytes: 4}
	if err := registry.Register(profile); err != nil {
		t.Fatalf("Register: %v", err)
	}
	body := io.NopCloser(strings.NewReader("12345"))
	rt := &policyRoundTripper{
		gateway: NewPolicyGateway(registry), sourceID: "allowed-src",
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body, ContentLength: -1, Request: req}, nil
		}),
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.com/api/test", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected response size limit error, got %v", err)
	}
}

func TestApprovedSourceWithoutTermsReviewCannotFetch(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register(&SourceProfile{SourceID: "unreviewed", ApprovalStatus: ApprovalApproved, ApprovalReason: "owner clicked approve", AllowedHosts: []string{"example.com"}, AllowedPathPrefixes: []string{"/"}, AllowedSchemes: []string{"https"}}); err != nil {
		t.Fatalf("Register: %v", err)
	}
	decision := NewPolicyGateway(registry).Evaluate("unreviewed", "https://example.com/")
	if decision.Allowed {
		t.Fatal("source must not fetch until terms review timestamp is recorded")
	}
}

// Closing a response body twice must never release another request's slot.
func TestConcurrencyReleaseBodyOnce(t *testing.T) {
	calls := 0
	body := &concurrencyReleaseBody{
		ReadCloser: io.NopCloser(strings.NewReader("test")),
		release:    func() { calls++ },
	}
	if err := body.Close(); err != nil {
		t.Fatal(err)
	}
	if err := body.Close(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected one release, got %d", calls)
	}
}
