package util

import (
	"net"
	"net/http"
	"strings"
)

// ExtractRealIP extracts the real client IP address from the request.
// Priority order:
// 1. X-Real-IP header
// 2. X-Forwarded-For header (first IP)
// 3. CF-Connecting-IP header (Cloudflare)
// 4. True-Client-IP header (Akamai/Cloudflare)
// 5. RemoteAddr from request
func ExtractRealIP(r *http.Request) string {
	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// Check X-Forwarded-For header (take first IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			if parsedIP := parseIP(strings.TrimSpace(ips[0])); parsedIP != "" {
				return parsedIP
			}
		}
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// Check True-Client-IP header (Akamai/Cloudflare)
	if ip := r.Header.Get("True-Client-IP"); ip != "" {
		if parsedIP := parseIP(ip); parsedIP != "" {
			return parsedIP
		}
	}

	// Fallback to RemoteAddr
	return parseIP(r.RemoteAddr)
}

// parseIP extracts and validates an IP address from a string.
// Handles both IPv4 and IPv6 addresses, including those with ports.
func parseIP(ipStr string) string {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return ""
	}

	// Try to parse as host:port (handles both IPv4:port and [IPv6]:port)
	if host, _, err := net.SplitHostPort(ipStr); err == nil {
		ipStr = host
	}

	// Remove brackets from IPv6 addresses
	ipStr = strings.Trim(ipStr, "[]")

	// Validate IP address
	if ip := net.ParseIP(ipStr); ip != nil {
		return ip.String()
	}

	return ""
}
