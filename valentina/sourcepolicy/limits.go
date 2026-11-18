package sourcepolicy

import (
	"sync"
	"time"
)

// RateLimiter is a simple token-bucket rate limiter per source.
type RateLimiter struct {
	mu          sync.Mutex
	tokens      float64
	last        time.Time
	ratePerSec  float64
	burst       float64
}

// NewRateLimiter creates a rate limiter with the given rate (tokens per second) and burst size.
func NewRateLimiter(ratePerSec float64, burst int) *RateLimiter {
	return &RateLimiter{
		tokens:    float64(burst),
		last:      time.Now(),
		ratePerSec: ratePerSec,
		burst:     float64(burst),
	}
}

// Allow attempts to consume one token. Returns true if allowed.
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(rl.last).Seconds()
	rl.last = now
	rl.tokens += elapsed * rl.ratePerSec
	if rl.tokens > rl.burst {
		rl.tokens = rl.burst
	}
	if rl.tokens < 1.0 {
		return false
	}
	rl.tokens--
	return true
}

// ConcurrencyLimiter limits the number of concurrent operations per source.
type ConcurrencyLimiter struct {
	mu       sync.Mutex
	active   int
	limit    int
	waiters  chan struct{}
}

// NewConcurrencyLimiter creates a concurrency limiter with the given limit.
func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		limit:   limit,
		waiters: make(chan struct{}, limit),
	}
}

// Acquire blocks until a slot is available, then returns a release function.
func (cl *ConcurrencyLimiter) Acquire() func() {
	cl.mu.Lock()
	if cl.active < cl.limit {
		cl.active++
		cl.mu.Unlock()
		return func() {
			cl.mu.Lock()
			cl.active--
			cl.mu.Unlock()
			select {
			case cl.waiters <- struct{}{}:
			default:
			}
		}
	}
	cl.mu.Unlock()
	<-cl.waiters
	cl.mu.Lock()
	cl.active++
	cl.mu.Unlock()
	return func() {
		cl.mu.Lock()
		cl.active--
		cl.mu.Unlock()
		select {
		case cl.waiters <- struct{}{}:
		default:
		}
	}
}

// DailyQuotaTracker tracks daily request counts per source.
type DailyQuotaTracker struct {
	mu         sync.Mutex
	limits     map[string]*int  // sourceID -> daily limit
	counts     map[string]int   // sourceID -> count today
	resetDates map[string]time.Time // sourceID -> last reset date
}

// NewDailyQuotaTracker creates a daily quota tracker.
func NewDailyQuotaTracker() *DailyQuotaTracker {
	return &DailyQuotaTracker{
		limits:     make(map[string]*int),
		counts:     make(map[string]int),
		resetDates: make(map[string]time.Time),
	}
}

// SetLimit sets the daily request limit for a source. nil means unlimited.
func (dqt *DailyQuotaTracker) SetLimit(sourceID string, limit *int) {
	dqt.mu.Lock()
	defer dqt.mu.Unlock()
	dqt.limits[sourceID] = limit
	dqt.resetDates[sourceID] = time.Now().UTC()
}

// CheckAndIncrement checks if the source can make another request today and increments the count.
// Returns false if the daily quota is exceeded.
func (dqt *DailyQuotaTracker) CheckAndIncrement(sourceID string) bool {
	dqt.mu.Lock()
	defer dqt.mu.Unlock()
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	lastReset, ok := dqt.resetDates[sourceID]
	if !ok || lastReset.Before(today) {
		dqt.counts[sourceID] = 0
		dqt.resetDates[sourceID] = today
	}

	limit, hasLimit := dqt.limits[sourceID]
	if hasLimit && limit != nil && dqt.counts[sourceID] >= *limit {
		return false
	}
	dqt.counts[sourceID]++
	return true
}

// Remaining returns the remaining requests for today, or -1 if unlimited.
func (dqt *DailyQuotaTracker) Remaining(sourceID string) int {
	dqt.mu.Lock()
	defer dqt.mu.Unlock()
	limit, ok := dqt.limits[sourceID]
	if !ok || limit == nil {
		return -1
	}
	count := dqt.counts[sourceID]
	rem := *limit - count
	if rem < 0 {
		return 0
	}
	return rem
}