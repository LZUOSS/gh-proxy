package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter manages per-IP rate limiting using token bucket algorithm
type RateLimiter struct {
	limiters sync.Map // map[string]*limiterEntry
	rate     rate.Limit
	burst    int
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter with specified rate and burst
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		rate:  r,
		burst: b,
	}
	// Start cleanup goroutine
	go rl.cleanupLoop()
	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	limiter := rl.getLimiter(ip)
	return limiter.limiter.Allow()
}

// getLimiter retrieves or creates a limiter for the given IP
func (rl *RateLimiter) getLimiter(ip string) *limiterEntry {
	// Try to load existing limiter
	if entry, ok := rl.limiters.Load(ip); ok {
		limiterEntry := entry.(*limiterEntry)
		limiterEntry.updateLastSeen()
		return limiterEntry
	}

	// Create new limiter
	newEntry := newLimiterEntry(rl.rate, rl.burst)

	// Store it, handling race condition
	actual, loaded := rl.limiters.LoadOrStore(ip, newEntry)
	if loaded {
		// Another goroutine created it first, use that one
		entry := actual.(*limiterEntry)
		entry.updateLastSeen()
		return entry
	}

	return newEntry
}

// GetRate returns the current rate limit
func (rl *RateLimiter) GetRate() rate.Limit {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.rate
}

// GetBurst returns the current burst size
func (rl *RateLimiter) GetBurst() int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return rl.burst
}

// SetRate updates the rate limit for all future limiters
func (rl *RateLimiter) SetRate(r rate.Limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.rate = r
}

// SetBurst updates the burst size for all future limiters
func (rl *RateLimiter) SetBurst(b int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.burst = b
}
