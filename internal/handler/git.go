package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/proxy"
)

// GitHandler handles Git smart HTTP protocol requests.
// Routes:
//   - /:owner/:repo.git/info/refs (GET)
//   - /:owner/:repo.git/git-upload-pack (POST)
//   - /:owner/:repo.git/git-receive-pack (POST)
type GitHandler struct {
	client *proxy.ProxyClient
	token  string // GitHub token for authentication
}

// NewGitHandler creates a new git protocol handler.
func NewGitHandler(client *proxy.ProxyClient, token string) *GitHandler {
	return &GitHandler{
		client: client,
		token:  token,
	}
}

// HandleInfoRefs handles the git info/refs request.
func (h *GitHandler) HandleInfoRefs(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Strip .git suffix if present (Git clients may send repo.git)
	repo = strings.TrimSuffix(repo, ".git")

	// Validate parameters
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	// Get service query parameter
	service := c.Query("service")
	if service == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing service parameter"})
		return
	}

	// Validate service
	if service != "git-upload-pack" && service != "git-receive-pack" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service"})
		return
	}

	// Generate upstream URL
	upstreamURL := fmt.Sprintf("https://github.com/%s/%s.git/info/refs?service=%s", owner, repo, service)

	// Forward the request
	h.forwardRequest(c, upstreamURL, http.MethodGet, nil)
}

// HandleUploadPack handles the git-upload-pack request (fetch/clone).
func (h *GitHandler) HandleUploadPack(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Strip .git suffix if present (Git clients may send repo.git)
	repo = strings.TrimSuffix(repo, ".git")

	// Validate parameters
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	// Generate upstream URL
	upstreamURL := fmt.Sprintf("https://github.com/%s/%s.git/git-upload-pack", owner, repo)

	// Forward the request with body
	h.forwardRequest(c, upstreamURL, http.MethodPost, c.Request.Body)
}

// HandleReceivePack handles the git-receive-pack request (push).
func (h *GitHandler) HandleReceivePack(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	// Strip .git suffix if present (Git clients may send repo.git)
	repo = strings.TrimSuffix(repo, ".git")

	// Validate parameters
	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	// Generate upstream URL
	upstreamURL := fmt.Sprintf("https://github.com/%s/%s.git/git-receive-pack", owner, repo)

	// Forward the request with body
	h.forwardRequest(c, upstreamURL, http.MethodPost, c.Request.Body)
}

// forwardRequest forwards a Git protocol request to GitHub.
func (h *GitHandler) forwardRequest(c *gin.Context, upstreamURL, method string, body io.Reader) {
	// Create request
	req, err := http.NewRequest(method, upstreamURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Copy relevant headers
	h.copyHeaders(c, req)

	// Add authentication if token is provided
	if h.token != "" {
		req.Header.Set("Authorization", "token "+h.token)
	}

	// Execute request
	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to forward request to GitHub"})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Stream response
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// copyHeaders copies relevant headers from the client request to the upstream request.
func (h *GitHandler) copyHeaders(c *gin.Context, req *http.Request) {
	// Git-specific headers
	gitHeaders := []string{
		"Content-Type",
		"Content-Encoding",
		"Accept",
		"Accept-Encoding",
		"Git-Protocol",
	}

	for _, header := range gitHeaders {
		if value := c.GetHeader(header); value != "" {
			req.Header.Set(header, value)
		}
	}

	// Set User-Agent
	if userAgent := c.GetHeader("User-Agent"); userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "git/github-reverse-proxy")
	}

	// Handle Content-Length for POST requests
	if req.Method == http.MethodPost {
		if contentLength := c.Request.ContentLength; contentLength > 0 {
			req.ContentLength = contentLength
		}
	}
}

// ValidateGitPath validates that a path is a valid git repository path.
func ValidateGitPath(path string) bool {
	return strings.HasSuffix(path, ".git") || strings.Contains(path, ".git/")
}
