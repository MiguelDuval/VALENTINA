package sourcepolicy

import "fmt"

// ErrSourceNotFound is returned when a source ID is not in the registry.
var ErrSourceNotFound = fmt.Errorf("source not found")

// ErrSourceNotApproved is returned when a source is not approved for fetching.
var ErrSourceNotApproved = fmt.Errorf("source not approved")

// ErrSourcePaused is returned when a source is paused.
var ErrSourcePaused = fmt.Errorf("source paused")

// ErrSourceDisabled is returned when a source is disabled.
var ErrSourceDisabled = fmt.Errorf("source disabled")

// ErrSchemeNotAllowed is returned when the URL scheme is not allowed.
var ErrSchemeNotAllowed = fmt.Errorf("scheme not allowed")

// ErrHostNotAllowed is returned when the host is not in the source allowlist.
var ErrHostNotAllowed = fmt.Errorf("host not allowed")

// ErrPathNotAllowed is returned when the URL path is outside the allowed prefixes.
var ErrPathNotAllowed = fmt.Errorf("path not allowed")

// ErrRedirectNotAllowed is returned when a redirect target is not allowed.
var ErrRedirectNotAllowed = fmt.Errorf("redirect not allowed")

// ErrPrivateIP is returned when the resolved IP is private/reserved.
var ErrPrivateIP = fmt.Errorf("private or reserved IP address")

// ErrTimeout is returned when a fetch exceeds the configured timeout.
var ErrTimeout = fmt.Errorf("request timeout")

// ErrResponseTooLarge is returned when the response exceeds the size limit.
var ErrResponseTooLarge = fmt.Errorf("response too large")

// ErrContentTypeNotAllowed is returned when the response content type is not permitted.
var ErrContentTypeNotAllowed = fmt.Errorf("content type not allowed")

// ErrRateLimited is returned when a source exceeds its rate limit.
var ErrRateLimited = fmt.Errorf("rate limited")

// ErrDailyQuotaExceeded is returned when a source exceeds its daily quota.
var ErrDailyQuotaExceeded = fmt.Errorf("daily quota exceeded")

// ErrConcurrencyLimit is returned when a source exceeds its parallel request limit.
var ErrConcurrencyLimit = fmt.Errorf("concurrency limit exceeded")

// SanitizedError wraps an internal error and returns a safe, generic message
// that leaks no URL, host, IP, or source-specific details to callers.
type SanitizedError struct {
	internal error
	safeMsg  string
}

func (e *SanitizedError) Error() string {
	return e.safeMsg
}

func (e *SanitizedError) Unwrap() error {
	return e.internal
}

// Sanitize converts any error into a safe, generic message suitable for logs
// and API responses. It never includes URLs, hostnames, or IP addresses.
func Sanitize(err error) error {
	if err == nil {
		return nil
	}
	if se, ok := err.(*SanitizedError); ok {
		return se
	}
	return &SanitizedError{
		internal: err,
		safeMsg:  "request blocked by source policy",
	}
}

// Forbidden is a sentinel returned by the gateway when a request must be
// rejected. Callers should treat it as a hard denial.
type Forbidden struct {
	reason string
}

func (f *Forbidden) Error() string {
	return "forbidden: " + f.reason
}

// NewForbidden creates a Forbidden error with an internal reason (not leaked).
func NewForbidden(reason string) error {
	return &Forbidden{reason: reason}
}
