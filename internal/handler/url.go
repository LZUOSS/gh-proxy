package handler

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/LZUOSS/gh-proxy/internal/cache"
	"github.com/LZUOSS/gh-proxy/internal/proxy"
)

// URLHandler handles full GitHub URL requests.
// Supports URLs like: /https://github.com/owner/repo/raw/main/file.md
type URLHandler struct {
	cache           *cache.Cache
	client          *proxy.ProxyClient
	releasesHandler *ReleasesHandler
	rawHandler      *RawHandler
	archiveHandler  *ArchiveHandler
	gitHandler      *GitHandler
	gistHandler     *GistHandler
	apiHandler      *APIHandler
}

// NewURLHandler creates a new URL handler.
func NewURLHandler(cache *cache.Cache, client *proxy.ProxyClient) *URLHandler {
	return &URLHandler{
		cache:           cache,
		client:          client,
		releasesHandler: NewReleasesHandler(cache, client),
		rawHandler:      NewRawHandler(cache, client),
		archiveHandler:  NewArchiveHandler(cache, client),
		gitHandler:      NewGitHandler(client, ""),
		gistHandler:     NewGistHandler(cache, client),
		apiHandler:      NewAPIHandler(cache, client, ""),
	}
}

// GitHubURLInfo represents parsed GitHub URL information
type GitHubURLInfo struct {
	Type     string // "releases", "raw", "archive", "git", "gist", "api"
	Owner    string
	Repo     string
	Tag      string
	Ref      string
	Filename string
	Filepath string
	GistID   string
	User     string
	APIPath  string
}

// Handle processes full GitHub URL requests.
func (h *URLHandler) Handle(c *gin.Context) {
	// Get the full request path
	fullPath := c.Request.URL.Path

	// Remove base path if present
	basePath := c.GetString("base_path")
	if basePath != "" {
		fullPath = strings.TrimPrefix(fullPath, basePath)
	}

	// Remove leading slash
	fullPath = strings.TrimPrefix(fullPath, "/")

	// Parse the GitHub URL
	info, err := h.parseGitHubURL(fullPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Route to appropriate handler based on type
	switch info.Type {
	case "releases":
		h.handleReleases(c, info)
	case "raw":
		h.handleRaw(c, info)
	case "archive":
		h.handleArchive(c, info)
	case "git":
		h.handleGit(c, info)
	case "gist":
		h.handleGist(c, info)
	case "api":
		h.handleAPI(c, info)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported GitHub URL type"})
	}
}

// parseGitHubURL parses a full GitHub URL and extracts relevant information.
func (h *URLHandler) parseGitHubURL(fullURL string) (*GitHubURLInfo, error) {
	// Parse the URL
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, err
	}

	// Handle URLs without scheme (e.g., github.com/owner/repo/...)
	if parsedURL.Scheme == "" {
		if strings.HasPrefix(fullURL, "github.com/") || strings.HasPrefix(fullURL, "raw.githubusercontent.com/") ||
		   strings.HasPrefix(fullURL, "api.github.com/") || strings.HasPrefix(fullURL, "gist.github.com/") {
			fullURL = "https://" + fullURL
			parsedURL, err = url.Parse(fullURL)
			if err != nil {
				return nil, err
			}
		}
	}

	host := parsedURL.Host
	path := strings.TrimPrefix(parsedURL.Path, "/")

	// Parse based on host - check specific hosts first before falling back to github.com
	switch {
	case strings.Contains(host, "raw.githubusercontent.com"):
		return h.parseRawGitHubUserContentURL(path)
	case strings.Contains(host, "api.github.com"):
		return h.parseAPIGitHubURL(path)
	case strings.Contains(host, "gist.github.com"):
		return h.parseGistGitHubURL(path)
	case strings.Contains(host, "github.com"):
		return h.parseGitHubComURL(path)
	default:
		return nil, &url.Error{Op: "parse", URL: fullURL, Err: http.ErrNotSupported}
	}
}

// parseGitHubComURL parses github.com URLs
func (h *URLHandler) parseGitHubComURL(path string) (*GitHubURLInfo, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, &url.Error{Op: "parse", URL: path, Err: http.ErrNotSupported}
	}

	info := &GitHubURLInfo{
		Owner: parts[0],
		Repo:  parts[1],
	}

	// Remove .git extension if present
	info.Repo = strings.TrimSuffix(info.Repo, ".git")

	if len(parts) < 3 {
		return info, nil
	}

	// Determine type based on path structure
	switch parts[2] {
	case "releases":
		// /owner/repo/releases/download/tag/filename
		if len(parts) >= 5 && parts[3] == "download" {
			info.Type = "releases"
			info.Tag = parts[4]
			if len(parts) > 5 {
				info.Filename = strings.Join(parts[5:], "/")
			}
		}

	case "raw":
		// /owner/repo/raw/ref/filepath
		info.Type = "raw"
		if len(parts) >= 4 {
			// Handle refs/heads/branch or refs/tags/tag
			if parts[3] == "refs" && len(parts) >= 5 {
				info.Ref = strings.Join(parts[3:5], "/")
				if len(parts) > 5 {
					info.Filepath = "/" + strings.Join(parts[5:], "/")
				}
			} else {
				info.Ref = parts[3]
				if len(parts) > 4 {
					info.Filepath = "/" + strings.Join(parts[4:], "/")
				}
			}
		}

	case "archive":
		// /owner/repo/archive/ref.tar.gz or /owner/repo/archive/refs/heads/main.tar.gz
		info.Type = "archive"
		if len(parts) >= 4 {
			info.Ref = strings.Join(parts[3:], "/")
		}

	case "blob", "tree":
		// /owner/repo/blob/ref/filepath -> convert to raw
		info.Type = "raw"
		if len(parts) >= 4 {
			if parts[3] == "refs" && len(parts) >= 5 {
				info.Ref = strings.Join(parts[3:5], "/")
				if len(parts) > 5 {
					info.Filepath = "/" + strings.Join(parts[5:], "/")
				}
			} else {
				info.Ref = parts[3]
				if len(parts) > 4 {
					info.Filepath = "/" + strings.Join(parts[4:], "/")
				}
			}
		}

	default:
		// Check if it's a git operation (ends with .git or has git-* paths)
		if strings.HasSuffix(parts[1], ".git") || strings.Contains(path, "/info/refs") ||
		   strings.Contains(path, "/git-upload-pack") || strings.Contains(path, "/git-receive-pack") {
			info.Type = "git"
		}
	}

	return info, nil
}

// parseRawGitHubUserContentURL parses raw.githubusercontent.com URLs
func (h *URLHandler) parseRawGitHubUserContentURL(path string) (*GitHubURLInfo, error) {
	// raw.githubusercontent.com/owner/repo/ref/filepath
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil, &url.Error{Op: "parse", URL: path, Err: http.ErrNotSupported}
	}

	info := &GitHubURLInfo{
		Type:  "raw",
		Owner: parts[0],
		Repo:  parts[1],
		Ref:   parts[2],
	}

	if len(parts) > 3 {
		info.Filepath = "/" + strings.Join(parts[3:], "/")
	}

	return info, nil
}

// parseAPIGitHubURL parses api.github.com URLs
func (h *URLHandler) parseAPIGitHubURL(path string) (*GitHubURLInfo, error) {
	return &GitHubURLInfo{
		Type:    "api",
		APIPath: "/" + path,
	}, nil
}

// parseGistGitHubURL parses gist.github.com URLs
func (h *URLHandler) parseGistGitHubURL(path string) (*GitHubURLInfo, error) {
	// gist.github.com/username/gist-id/raw/filename
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return nil, &url.Error{Op: "parse", URL: path, Err: http.ErrNotSupported}
	}

	info := &GitHubURLInfo{
		Type:   "gist",
		User:   parts[0],
		GistID: parts[1],
	}

	// Find "raw" in the path
	for i, part := range parts {
		if part == "raw" && i+1 < len(parts) {
			info.Filename = parts[i+1]
			break
		}
	}

	return info, nil
}

// handleReleases routes to the releases handler
func (h *URLHandler) handleReleases(c *gin.Context, info *GitHubURLInfo) {
	c.Params = gin.Params{
		{Key: "owner", Value: info.Owner},
		{Key: "repo", Value: info.Repo},
		{Key: "tag", Value: info.Tag},
		{Key: "filename", Value: info.Filename},
	}
	h.releasesHandler.Handle(c)
}

// handleRaw routes to the raw handler
func (h *URLHandler) handleRaw(c *gin.Context, info *GitHubURLInfo) {
	c.Params = gin.Params{
		{Key: "owner", Value: info.Owner},
		{Key: "repo", Value: info.Repo},
		{Key: "ref", Value: info.Ref},
		{Key: "filepath", Value: info.Filepath},
	}
	h.rawHandler.Handle(c)
}

// handleArchive routes to the archive handler
func (h *URLHandler) handleArchive(c *gin.Context, info *GitHubURLInfo) {
	c.Params = gin.Params{
		{Key: "owner", Value: info.Owner},
		{Key: "repo", Value: info.Repo},
		{Key: "ref", Value: info.Ref},
	}
	h.archiveHandler.Handle(c)
}

// handleGit routes to the git handler
func (h *URLHandler) handleGit(c *gin.Context, info *GitHubURLInfo) {
	// Determine which git operation based on path
	path := c.Request.URL.Path

	if strings.Contains(path, "/info/refs") {
		c.Params = gin.Params{
			{Key: "owner", Value: info.Owner},
			{Key: "repo", Value: info.Repo},
		}
		h.gitHandler.HandleInfoRefs(c)
	} else if strings.Contains(path, "/git-upload-pack") {
		c.Params = gin.Params{
			{Key: "owner", Value: info.Owner},
			{Key: "repo", Value: info.Repo},
		}
		h.gitHandler.HandleUploadPack(c)
	} else if strings.Contains(path, "/git-receive-pack") {
		c.Params = gin.Params{
			{Key: "owner", Value: info.Owner},
			{Key: "repo", Value: info.Repo},
		}
		h.gitHandler.HandleReceivePack(c)
	}
}

// handleGist routes to the gist handler
func (h *URLHandler) handleGist(c *gin.Context, info *GitHubURLInfo) {
	c.Params = gin.Params{
		{Key: "user", Value: info.User},
		{Key: "gist_id", Value: info.GistID},
		{Key: "file", Value: info.Filename},
	}
	h.gistHandler.Handle(c)
}

// handleAPI routes to the API handler
func (h *URLHandler) handleAPI(c *gin.Context, info *GitHubURLInfo) {
	c.Params = gin.Params{
		{Key: "path", Value: info.APIPath},
	}
	h.apiHandler.Handle(c)
}

// isGitHubURL checks if a path looks like a GitHub URL
func isGitHubURL(path string) bool {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Match patterns like:
	// - https://github.com/...
	// - http://github.com/...
	// - github.com/...
	// - https://raw.githubusercontent.com/...
	// - https://api.github.com/...
	// - https://gist.github.com/...
	patterns := []string{
		`^https?://github\.com/`,
		`^github\.com/`,
		`^https?://raw\.githubusercontent\.com/`,
		`^raw\.githubusercontent\.com/`,
		`^https?://api\.github\.com/`,
		`^api\.github\.com/`,
		`^https?://gist\.github\.com/`,
		`^gist\.github\.com/`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, path)
		if matched {
			return true
		}
	}

	return false
}
