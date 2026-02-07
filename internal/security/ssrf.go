package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

var allowedGitHubDomains = []string{
	"github.com",
	"api.github.com",
	"raw.githubusercontent.com",
	"githubusercontent.com",
	"github.io",
	"codeload.github.com",
	"gist.github.com",
	"objects.githubusercontent.com",
	"avatars.githubusercontent.com",
	"cloud.githubusercontent.com",
}

var privateIPRanges = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
}

func init() {
	for i, cidr := range privateIPRanges {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid private IP range at index %d: %s", i, cidr))
		}
	}
}

// ValidateGitHubURL validates that a URL is a GitHub domain and doesn't point to private IPs
func ValidateGitHubURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s (only http and https are allowed)", parsedURL.Scheme)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("URL host is empty")
	}

	// Check if domain is an allowed GitHub domain
	if !isAllowedGitHubDomain(host) {
		return fmt.Errorf("URL host %s is not an allowed GitHub domain", host)
	}

	// Resolve hostname to IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %s: %w", host, err)
	}

	// Check if any resolved IP is in a private range
	for _, ip := range ips {
		if isPrivateIP(ip) {
			return fmt.Errorf("URL resolves to private IP address: %s", ip.String())
		}
	}

	return nil
}

// isAllowedGitHubDomain checks if a hostname is in the allowed GitHub domains list
func isAllowedGitHubDomain(host string) bool {
	host = strings.ToLower(host)

	for _, domain := range allowedGitHubDomains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}

	return false
}

// isPrivateIP checks if an IP address is in a private range
func isPrivateIP(ip net.IP) bool {
	// Check if it's a loopback address
	if ip.IsLoopback() {
		return true
	}

	// Check if it's a link-local address
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check against private IP ranges
	for _, cidr := range privateIPRanges {
		_, ipNet, _ := net.ParseCIDR(cidr)
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}
