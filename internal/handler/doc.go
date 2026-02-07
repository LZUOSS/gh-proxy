// Package handler provides HTTP request handlers for the GitHub reverse proxy.
//
// The handler package implements all the HTTP endpoints for proxying various
// GitHub services including releases, raw content, archives, Git protocol,
// gists, and API requests.
//
// All handlers follow a streaming approach to minimize memory usage and
// support efficient caching of responses.
//
// Handlers:
//   - ReleasesHandler: Handles release asset downloads from GitHub releases
//   - RawHandler: Proxies raw file content from repositories
//   - ArchiveHandler: Streams repository archives (zip/tar.gz)
//   - GitHandler: Implements Git smart HTTP protocol for clone/fetch/push
//   - GistHandler: Proxies Gist raw file content
//   - APIHandler: Generic proxy for GitHub API requests
//
// All handlers support:
//   - Streaming responses to minimize memory usage
//   - Multi-tier caching (memory and disk)
//   - Proxy support (SOCKS5/HTTP)
//   - ETag-based validation
//   - Proper header forwarding
//
// Example usage:
//
//	// Create handlers
//	cache := cache.NewCache(cacheConfig)
//	client := proxy.NewProxyClient(proxyConfig)
//
//	releasesHandler := handler.NewReleasesHandler(cache, client)
//	rawHandler := handler.NewRawHandler(cache, client)
//	apiHandler := handler.NewAPIHandler(cache, client, token)
//
//	// Register with Gin router
//	router.GET("/:owner/:repo/releases/download/:tag/:filename", releasesHandler.Handle)
//	router.GET("/:owner/:repo/raw/:ref/*filepath", rawHandler.Handle)
//	router.GET("/api/*path", apiHandler.Handle)
package handler
