package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry stores a rate limiter with its last access time
type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
	mu       sync.Mutex
}

// newLimiterEntry creates a new limiter entry with the current timestamp
func newLimiterEntry(r rate.Limit, b int) *limiterEntry {
	return &limiterEntry{
		limiter:  rate.NewLimiter(r, b),
		lastSeen: time.Now(),
	}
}

// updateLastSeen updates the last seen timestamp
func (e *limiterEntry) updateLastSeen() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastSeen = time.Now()
}

// getLastSeen returns the last seen timestamp
func (e *limiterEntry) getLastSeen() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastSeen
}
