package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/cache"
	"github.com/LZUOSS/gh-proxy/internal/proxy"
)

// RawHandler handles GitHub raw content requests.
// Route: /:owner/:repo/raw/:ref/*filepath
type RawHandler struct {
	cache  *cache.Cache
	client *proxy.ProxyClient
}

// NewRawHandler creates a new raw content handler.
func NewRawHandler(cache *cache.Cache, client *proxy.ProxyClient) *RawHandler {
	return &RawHandler{
		cache:  cache,
		client: client,
	}
}

// Handle processes raw content requests.
func (h *RawHandler) Handle(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ref := c.Param("ref")
	filepath := c.Param("filepath")

	// Validate parameters
	if owner == "" || repo == "" || ref == "" || filepath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	// Generate upstream URL (raw.githubusercontent.com)
	upstreamURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s%s", owner, repo, ref, filepath)

	// Generate cache key
	cacheKey := cache.GenerateKey("raw", owner, repo, ref, filepath, "")

	// Try memory cache first
	if entry, ok := h.cache.Get(cacheKey); ok {
		h.serveFromCache(c, entry)
		return
	}

	// Check disk cache metadata
	if meta, ok := h.cache.GetMetadata(cacheKey); ok {
		// Serve from disk cache
		dataPath := h.cache.GetDataPath(cacheKey)
		h.serveFromDisk(c, dataPath, meta)
		return
	}

	// Cache miss - fetch from GitHub
	h.fetchAndStream(c, upstreamURL, cacheKey)
}

// serveFromCache serves a response from memory cache.
func (h *RawHandler) serveFromCache(c *gin.Context, entry *cache.CacheEntry) {
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

// serveFromDisk serves a response from disk cache.
func (h *RawHandler) serveFromDisk(c *gin.Context, dataPath string, meta *cache.DiskCacheMetadata) {
	// Set headers
	for key, value := range meta.Headers {
		c.Header(key, value)
	}
	if meta.ETag != "" {
		c.Header("ETag", meta.ETag)
	}
	c.Header("X-Cache", "HIT-DISK")

	// Stream file directly from disk
	c.File(dataPath)
}

// fetchAndStream fetches from GitHub and streams while caching.
func (h *RawHandler) fetchAndStream(c *gin.Context, upstreamURL, cacheKey string) {
	// Create request
	req, err := http.NewRequest(http.MethodGet, upstreamURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Set headers
	req.Header.Set("User-Agent", "github-reverse-proxy/1.0")
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}

	// Execute request
	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch from GitHub"})
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
		return
	}

	// Copy response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			value := values[0]
			c.Header(key, value)
			headers[key] = value
		}
	}
	c.Header("X-Cache", "MISS")

	// Get ETag
	etag := resp.Header.Get("ETag")

	// Determine if we should cache based on content length
	contentLength := resp.ContentLength
	shouldCache := contentLength > 0 && contentLength < 100*1024*1024 // Cache files < 100MB

	if shouldCache {
		// Use TeeReader to cache while streaming
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
		if written > 0 {
			entry := &cache.CacheEntry{
				Data:    buf.Bytes(),
				Headers: headers,
				ETag:    etag,
			}

			// Cache for 1 hour (raw files change more frequently)
			ttl := 1 * time.Hour
			h.cache.Set(cacheKey, entry, ttl)
		}
	} else {
		// Just stream without caching
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}
