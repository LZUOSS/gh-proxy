package auth

import (
	"time"
)

// Token represents a validated GitHub Personal Access Token.
type Token struct {
	// Value is the original token string
	Value string

	// Username is the GitHub username associated with this token
	Username string

	// Login is the GitHub login (usually same as Username)
	Login string

	// Email is the email address associated with the GitHub account
	Email string

	// Scopes contains the OAuth scopes granted to this token
	Scopes []string

	// RateLimit contains the current rate limit information
	RateLimit *RateLimitInfo

	// ValidatedAt is when this token was last validated
	ValidatedAt time.Time

	// ExpiresAt is when this token validation expires (cache TTL)
	ExpiresAt time.Time
}

// RateLimitInfo contains GitHub API rate limit information.
type RateLimitInfo struct {
	// Limit is the maximum number of requests per hour
	Limit int

	// Remaining is the number of requests remaining in the current window
	Remaining int

	// Reset is when the rate limit window resets
	Reset time.Time
}

// IsExpired checks if the token validation has expired.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the token is still valid (not expired).
func (t *Token) IsValid() bool {
	return !t.IsExpired()
}
