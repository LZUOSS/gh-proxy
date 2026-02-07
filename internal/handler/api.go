package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kexi/github-reverse-proxy/internal/cache"
	"github.com/kexi/github-reverse-proxy/internal/proxy"
)

// APIHandler handles GitHub API requests.
// Route: /api/*path
type APIHandler struct {
	cache  *cache.Cache
	client *proxy.ProxyClient
	token  string // GitHub API token for authentication
}

// NewAPIHandler creates a new API handler.
func NewAPIHandler(cache *cache.Cache, client *proxy.ProxyClient, token string) *APIHandler {
	return &APIHandler{
		cache:  cache,
		client: client,
		token:  token,
	}
}

// Handle processes API requests.
func (h *APIHandler) Handle(c *gin.Context) {
	// Extract path after /api/
	path := c.Param("path")
	if path == "" {
		path = "/"
	}

	// Remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	// Build upstream URL
	upstreamURL := fmt.Sprintf("https://api.github.com/%s", path)

	// Add query parameters
	if c.Request.URL.RawQuery != "" {
		upstreamURL += "?" + c.Request.URL.RawQuery
	}

	// Determine if this request should be cached
	// Only cache GET requests
	shouldCache := c.Request.Method == http.MethodGet

	// Generate cache key for GET requests
	var cacheKey string
	if shouldCache {
		cacheKey = cache.GenerateKey("api", path, c.Request.URL.RawQuery, "", "", "")

		// Try memory cache first
		if entry, ok := h.cache.Get(cacheKey); ok {
			h.serveFromCache(c, entry)
			return
		}
	}

	// Forward the request
	h.forwardRequest(c, upstreamURL, shouldCache, cacheKey)
}

// serveFromCache serves a response from memory cache.
func (h *APIHandler) serveFromCache(c *gin.Context, entry *cache.CacheEntry) {
	// Set headers
	for key, value := range entry.Headers {
		c.Header(key, value)
	}
	if entry.ETag != "" {
		c.Header("ETag", entry.ETag)
	}
	c.Header("X-Cache", "HIT-MEMORY")

	// Stream the data
	c.Data(http.StatusOK, c.GetHeader("Content-Type"), entry.Data)
}

// forwardRequest forwards an API request to GitHub.
func (h *APIHandler) forwardRequest(c *gin.Context, upstreamURL string, shouldCache bool, cacheKey string) {
	// Create request
	var body io.Reader
	if c.Request.Body != nil {
		body = c.Request.Body
	}

	req, err := http.NewRequest(c.Request.Method, upstreamURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Copy headers
	h.copyHeaders(c, req)

	// Add authentication if token is provided
	if h.token != "" {
		req.Header.Set("Authorization", "token "+h.token)
	}

	// Execute request
	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to forward request to GitHub API"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			value := values[0]
			c.Header(key, value)
			headers[key] = value
		}
	}

	// Check if we should cache this response
	if shouldCache && resp.StatusCode == http.StatusOK {
		c.Header("X-Cache", "MISS")

		// Get ETag
		etag := resp.Header.Get("ETag")

		// Cache API responses with TeeReader
		var buf bytes.Buffer
		teeReader := io.TeeReader(resp.Body, &buf)

		// Stream to client
		c.Status(resp.StatusCode)
		written, err := io.Copy(c.Writer, teeReader)
		if err != nil {
			// Stream was interrupted, don't cache
			return
		}

		// Cache the data
		if written > 0 && written < 5*1024*1024 { // Only cache responses < 5MB
			entry := &cache.CacheEntry{
				Data:    buf.Bytes(),
				Headers: headers,
				ETag:    etag,
			}

			// Determine TTL based on endpoint
			ttl := h.determineTTL(c.Request.URL.Path)
			h.cache.Set(cacheKey, entry, ttl)
		}
	} else {
		// Just stream without caching
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}

// copyHeaders copies relevant headers from the client request to the upstream request.
func (h *APIHandler) copyHeaders(c *gin.Context, req *http.Request) {
	// API-specific headers
	apiHeaders := []string{
		"Accept",
		"Accept-Encoding",
		"Content-Type",
		"If-None-Match",
		"If-Modified-Since",
	}

	for _, header := range apiHeaders {
		if value := c.GetHeader(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	// Set User-Agent
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "github-reverse-proxy/1.0")
	}

	// GitHub API version header
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// Handle Content-Length for POST/PUT/PATCH requests
	if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut || c.Request.Method == http.MethodPatch {
		if contentLength := c.Request.ContentLength; contentLength > 0 {
			req.ContentLength = contentLength
		}
	}
}

// determineTTL determines the cache TTL based on the API endpoint.
func (h *APIHandler) determineTTL(path string) time.Duration {
	// Different endpoints have different caching strategies
	switch {
	case strings.Contains(path, "/releases"):
		// Releases change infrequently
		return 1 * time.Hour
	case strings.Contains(path, "/repos"):
		// Repository info changes occasionally
		return 30 * time.Minute
	case strings.Contains(path, "/users"):
		// User info changes occasionally
		return 30 * time.Minute
	case strings.Contains(path, "/issues") || strings.Contains(path, "/pulls"):
		// Issues and PRs change frequently
		return 5 * time.Minute
	case strings.Contains(path, "/commits"):
		// Commits are immutable
		return 24 * time.Hour
	default:
		// Default cache time
		return 15 * time.Minute
	}
}
