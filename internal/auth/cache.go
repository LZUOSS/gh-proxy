package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Cache is a thread-safe cache for validated GitHub tokens with TTL support.
type Cache struct {
	mu    sync.RWMutex
	store map[string]*Token
	ttl   time.Duration
}

// NewCache creates a new token cache with the specified TTL.
// If ttl is 0, it defaults to 1 hour.
func NewCache(ttl time.Duration) *Cache {
	if ttl == 0 {
		ttl = 1 * time.Hour
	}
	return &Cache{
		store: make(map[string]*Token),
		ttl:   ttl,
	}
}

// Get retrieves a token from the cache by username and password.
// Returns nil if the token is not found or has expired.
func (c *Cache) Get(username, password string) *Token {
	key := c.makeKey(username, password)

	c.mu.RLock()
	token, exists := c.store[key]
	c.mu.RUnlock()

	if !exists {
		return nil
	}

	// Check if token has expired
	if token.IsExpired() {
		// Clean up expired token
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil
	}

	return token
}

// Set stores a token in the cache.
// The token's ExpiresAt field should already be set.
func (c *Cache) Set(username, password string, token *Token) {
	key := c.makeKey(username, password)

	c.mu.Lock()
	c.store[key] = token
	c.mu.Unlock()
}

// GetOrValidate retrieves a token from cache or validates it if not cached/expired.
// This is a convenience method that combines Get and ValidateBasicAuth.
func (c *Cache) GetOrValidate(username, password string) (*Token, error) {
	return c.GetOrValidateWithContext(context.Background(), username, password)
}

// GetOrValidateWithContext retrieves a token from cache or validates it with context.
func (c *Cache) GetOrValidateWithContext(ctx context.Context, username, password string) (*Token, error) {
	// Try to get from cache first
	if token := c.Get(username, password); token != nil {
		return token, nil
	}

	// Not in cache or expired, validate with GitHub API
	token, err := ValidateBasicAuthWithContext(ctx, username, password)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.Set(username, password, token)

	return token, nil
}

// Delete removes a token from the cache.
func (c *Cache) Delete(username, password string) {
	key := c.makeKey(username, password)

	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
}

// Clear removes all tokens from the cache.
func (c *Cache) Clear() {
	c.mu.Lock()
	c.store = make(map[string]*Token)
	c.mu.Unlock()
}

// Cleanup removes all expired tokens from the cache.
// This should be called periodically to prevent memory leaks.
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, token := range c.store {
		if now.After(token.ExpiresAt) {
			delete(c.store, key)
		}
	}
}

// StartCleanupTask starts a background goroutine that periodically cleans up expired tokens.
// The cleanup runs at the specified interval.
// Returns a channel that can be closed to stop the cleanup task.
func (c *Cache) StartCleanupTask(interval time.Duration) chan<- struct{} {
	if interval == 0 {
		interval = 10 * time.Minute
	}

	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.Cleanup()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// Size returns the current number of tokens in the cache (including expired ones).
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// makeKey creates a cache key from username and password.
// Uses SHA256 hash to avoid storing plain passwords as keys.
func (c *Cache) makeKey(username, password string) string {
	data := username + ":" + password
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
