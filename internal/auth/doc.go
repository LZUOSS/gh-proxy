// Package auth provides authentication functionality for the GitHub reverse proxy.
//
// It supports GitHub Personal Access Token (PAT) validation through Basic Authentication,
// where the password field is treated as the GitHub PAT. The package includes:
//
//   - Token validation against GitHub API
//   - Token caching with configurable TTL (default 1 hour)
//   - Thread-safe cache implementation
//   - Rate limit information extraction
//   - OAuth scope extraction
//
// Example usage:
//
//	// Create a cache with 1-hour TTL
//	cache := auth.NewCache(1 * time.Hour)
//
//	// Start background cleanup task
//	stopCleanup := cache.StartCleanupTask(10 * time.Minute)
//	defer close(stopCleanup)
//
//	// Validate credentials (checks cache first, then GitHub API)
//	token, err := cache.GetOrValidate(username, password)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Authenticated as: %s\n", token.Username)
package auth
