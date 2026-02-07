package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a limiter: 2 requests per second, burst of 3
	rl := NewRateLimiter(2, 3)

	ip := "192.168.1.1"

	// First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		if !rl.Allow(ip) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be denied (exceeded burst)
	if rl.Allow(ip) {
		t.Error("Request 4 should be denied")
	}

	// Wait for token to refill
	time.Sleep(600 * time.Millisecond)

	// Should allow one more request
	if !rl.Allow(ip) {
		t.Error("Request after waiting should be allowed")
	}
}

func TestRateLimiter_MultipleIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1)

	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"

	// Both IPs should get their own limiter
	if !rl.Allow(ip1) {
		t.Error("IP1 first request should be allowed")
	}
	if !rl.Allow(ip2) {
		t.Error("IP2 first request should be allowed")
	}

	// Both should be rate limited independently
	if rl.Allow(ip1) {
		t.Error("IP1 second request should be denied")
	}
	if rl.Allow(ip2) {
		t.Error("IP2 second request should be denied")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, 10)

	var wg sync.WaitGroup
	ip := "192.168.1.1"

	// Concurrent access should not cause race conditions
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.Allow(ip)
		}()
	}

	wg.Wait()
}

func TestRateLimiter_GetSetRate(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	if got := rl.GetRate(); got != 10 {
		t.Errorf("GetRate() = %v, want 10", got)
	}
	if got := rl.GetBurst(); got != 20 {
		t.Errorf("GetBurst() = %v, want 20", got)
	}

	rl.SetRate(5)
	rl.SetBurst(15)

	if got := rl.GetRate(); got != 5 {
		t.Errorf("GetRate() after SetRate = %v, want 5", got)
	}
	if got := rl.GetBurst(); got != 15 {
		t.Errorf("GetBurst() after SetBurst = %v, want 15", got)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Use shorter timeout for testing
	oldCleanupInterval := cleanupInterval
	oldIdleTimeout := idleTimeout
	defer func() {
		// Note: These are constants, so this won't actually change them
		// In a real implementation, you'd make these configurable
		_ = oldCleanupInterval
		_ = oldIdleTimeout
	}()

	rl := NewRateLimiter(10, 10)

	// Add some limiters
	rl.Allow("192.168.1.1")
	rl.Allow("192.168.1.2")
	rl.Allow("192.168.1.3")

	// Count limiters
	count := 0
	rl.limiters.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("Expected 3 limiters, got %d", count)
	}

	// Manually call cleanup (won't remove anything as they're fresh)
	rl.cleanup()

	count = 0
	rl.limiters.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("After cleanup, expected 3 limiters (none idle), got %d", count)
	}
}
