package sourcepolicy

import (
	"errors"
	"net"
	"net/url"
	"strings"
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

	// Host check (exact match against allowed hosts).
	host := strings.ToLower(parsed.Hostname())
	hostOK := false
	for _, h := range profile.AllowedHosts {
		if strings.EqualFold(host, strings.ToLower(h)) {
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

	// Path check: must start with one of the allowed prefixes.
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	pathOK := false
	for _, prefix := range profile.AllowedPathPrefixes {
		p := prefix
		if p == "" {
			p = "/"
		}
		if strings.HasPrefix(path, p) {
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

	// DNS/IP resolution and SSRF checks.
	ip, err := resolveHost(host)
	if err != nil {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "DNS resolution failed",
			InternalReason: err.Error(),
		}
	}
	if isBlockedIP(ip) {
		return &FetchDecision{
			Allowed:        false,
			SourceID:       sourceID,
			Reason:         "destination IP not permitted",
			InternalReason: ErrPrivateIP.Error(),
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
func (g *PolicyGateway) EvaluateRedirect(sourceID, fromURL, redirectURL string) *FetchDecision {
	decision := g.Evaluate(sourceID, redirectURL)
	if !decision.Allowed {
		decision.Reason = "redirect not allowed"
		decision.InternalReason = ErrRedirectNotAllowed.Error()
	}
	return decision
}

// resolveHost resolves a hostname to its first IP address.
func resolveHost(host string) (net.IP, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, errors.New("no IP addresses returned")
	}
	return ips[0], nil
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
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
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
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port != "" {
		switch u.Scheme {
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
	if port != "" {
		hostPart = host + ":" + port
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	out := u.Scheme + "://" + hostPart + path
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	return out
}

// IsBlockedIPExported exposes the SSRF check for tests and adapters.
func IsBlockedIPExported(ip net.IP) bool {
	return isBlockedIP(ip)
}