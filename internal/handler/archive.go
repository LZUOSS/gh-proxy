package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/cache"
	"github.com/LZUOSS/gh-proxy/internal/proxy"
)

// ArchiveHandler handles GitHub archive downloads.
// Routes: /:owner/:repo/archive/:ref.zip and /:owner/:repo/archive/:ref.tar.gz
type ArchiveHandler struct {
	cache  *cache.Cache
	client *proxy.ProxyClient
}

// NewArchiveHandler creates a new archive handler.
func NewArchiveHandler(cache *cache.Cache, client *proxy.ProxyClient) *ArchiveHandler {
	return &ArchiveHandler{
		cache:  cache,
		client: client,
	}
}

// Handle processes archive download requests.
func (h *ArchiveHandler) Handle(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	refWithExt := c.Param("ref") // e.g., "main.zip" or "v1.0.0.tar.gz"

	// Validate parameters
	if owner == "" || repo == "" || refWithExt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	// Determine format and extract ref
	var ref, format string
	if strings.HasSuffix(refWithExt, ".tar.gz") {
		ref = strings.TrimSuffix(refWithExt, ".tar.gz")
		format = "tar.gz"
	} else if strings.HasSuffix(refWithExt, ".zip") {
		ref = strings.TrimSuffix(refWithExt, ".zip")
		format = "zip"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported archive format"})
		return
	}

	// Generate upstream URL
	upstreamURL := fmt.Sprintf("https://github.com/%s/%s/archive/%s.%s", owner, repo, ref, format)

	// Generate cache key
	cacheKey := cache.GenerateKey("archive", owner, repo, ref, format, "")

	// Check disk cache metadata (archives are large, skip memory cache)
	if meta, ok := h.cache.GetMetadata(cacheKey); ok {
		// Serve from disk cache
		dataPath := h.cache.GetDataPath(cacheKey)
		h.serveFromDisk(c, dataPath, meta)
		return
	}

	// Cache miss - fetch from GitHub
	h.fetchAndStream(c, upstreamURL, cacheKey)
}

// serveFromDisk serves a response from disk cache.
func (h *ArchiveHandler) serveFromDisk(c *gin.Context, dataPath string, meta *cache.DiskCacheMetadata) {
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

// fetchAndStream fetches from GitHub and streams the archive.
// Archives are typically large, so we don't cache them in memory.
func (h *ArchiveHandler) fetchAndStream(c *gin.Context, upstreamURL, cacheKey string) {
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

	// Handle redirects (GitHub often redirects to AWS S3)
	req.Header.Set("Accept", "application/octet-stream")

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

	// For archives, we just stream directly without caching
	// Archives are typically too large to cache efficiently
	c.Status(resp.StatusCode)

	// Stream directly to client
	written, err := io.Copy(c.Writer, resp.Body)

	// Log the bytes transferred (optional)
	if err == nil && written > 0 {
		c.Set("bytes_transferred", written)
	}
}
