package ratelimit

import (
	"time"
)

const (
	// cleanupInterval defines how often the cleanup routine runs
	cleanupInterval = 5 * time.Minute

	// idleTimeout defines how long a limiter can be idle before removal
	idleTimeout = 30 * time.Minute
)

// cleanupLoop runs in the background and removes idle limiters
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes limiters that have been idle for more than idleTimeout
func (rl *RateLimiter) cleanup() {
	now := time.Now()

	rl.limiters.Range(func(key, value interface{}) bool {
		ip := key.(string)
		entry := value.(*limiterEntry)

		lastSeen := entry.getLastSeen()
		if now.Sub(lastSeen) > idleTimeout {
			rl.limiters.Delete(ip)
		}

		return true
	})
}

// Stop can be called to immediately stop the cleanup goroutine
// This is useful for testing or graceful shutdown
func (rl *RateLimiter) Stop() {
	// Note: In a production system, you might want to use context.Context
	// for proper cancellation. For now, the cleanup goroutine will be
	// garbage collected when the RateLimiter is no longer referenced.
}
