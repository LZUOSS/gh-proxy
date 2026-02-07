package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/user"
	userAgent    = "github-reverse-proxy/1.0"
)

// GitHubUser represents the response from GitHub's /user API endpoint.
type GitHubUser struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ValidateBasicAuth validates basic authentication credentials by treating
// the password as a GitHub Personal Access Token (PAT).
// It makes a test request to GitHub API to validate the token and extract user information.
//
// Parameters:
//   - username: The username provided in basic auth (can be empty or arbitrary for PATs)
//   - password: The GitHub Personal Access Token
//
// Returns:
//   - *Token: A validated token with user information and rate limit data
//   - error: An error if validation fails
func ValidateBasicAuth(username, password string) (*Token, error) {
	return ValidateBasicAuthWithContext(context.Background(), username, password)
}

// ValidateBasicAuthWithContext validates basic authentication with a context for cancellation.
func ValidateBasicAuthWithContext(ctx context.Context, username, password string) (*Token, error) {
	if password == "" {
		return nil, fmt.Errorf("password (token) cannot be empty")
	}

	// Create request to GitHub API
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header with the token
	req.Header.Set("Authorization", "token "+password)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Execute the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token with GitHub API: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("invalid token: unauthorized")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("token forbidden: may be expired or have insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse user information
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Extract rate limit information from headers
	rateLimit := extractRateLimitInfo(resp.Header)

	// Extract scopes from response headers
	scopes := extractScopes(resp.Header)

	// Create token object
	now := time.Now()
	token := &Token{
		Value:       password,
		Username:    user.Login,
		Login:       user.Login,
		Email:       user.Email,
		Scopes:      scopes,
		RateLimit:   rateLimit,
		ValidatedAt: now,
		ExpiresAt:   now.Add(1 * time.Hour), // 1 hour TTL
	}

	return token, nil
}

// extractRateLimitInfo extracts rate limit information from GitHub API response headers.
func extractRateLimitInfo(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}

	// Parse X-RateLimit-Limit
	if limitStr := headers.Get("X-RateLimit-Limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			info.Limit = limit
		}
	}

	// Parse X-RateLimit-Remaining
	if remainingStr := headers.Get("X-RateLimit-Remaining"); remainingStr != "" {
		if remaining, err := strconv.Atoi(remainingStr); err == nil {
			info.Remaining = remaining
		}
	}

	// Parse X-RateLimit-Reset (Unix timestamp)
	if resetStr := headers.Get("X-RateLimit-Reset"); resetStr != "" {
		if resetUnix, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
			info.Reset = time.Unix(resetUnix, 0)
		}
	}

	return info
}

// extractScopes extracts OAuth scopes from the X-OAuth-Scopes header.
func extractScopes(headers http.Header) []string {
	scopesHeader := headers.Get("X-OAuth-Scopes")
	if scopesHeader == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	scopeParts := strings.Split(scopesHeader, ",")
	scopes := make([]string, 0, len(scopeParts))
	for _, scope := range scopeParts {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}

	return scopes
}
