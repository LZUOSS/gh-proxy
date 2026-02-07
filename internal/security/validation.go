package security

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// GitHub username/organization validation: alphanumeric, hyphens, max 39 chars
	ownerRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

	// GitHub repository name validation: alphanumeric, hyphens, underscores, dots, max 100 chars
	repoRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

	// Git reference validation: branch, tag, or commit SHA
	refRegex = regexp.MustCompile(`^[a-zA-Z0-9._/-]{1,255}$`)

	// Commit SHA validation: 40 character hex string
	shaRegex = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
)

// ValidateOwner validates a GitHub owner (username or organization) name
func ValidateOwner(owner string) error {
	if owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}

	if len(owner) > 39 {
		return fmt.Errorf("owner name too long (max 39 characters)")
	}

	if !ownerRegex.MatchString(owner) {
		return fmt.Errorf("invalid owner name: must contain only alphanumeric characters and hyphens, cannot start or end with hyphen")
	}

	return nil
}

// ValidateRepo validates a GitHub repository name
func ValidateRepo(repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name cannot be empty")
	}

	if len(repo) > 100 {
		return fmt.Errorf("repository name too long (max 100 characters)")
	}

	if !repoRegex.MatchString(repo) {
		return fmt.Errorf("invalid repository name: must contain only alphanumeric characters, dots, hyphens, and underscores")
	}

	// Additional checks for reserved names
	if strings.HasPrefix(repo, ".") || strings.HasSuffix(repo, ".") {
		return fmt.Errorf("repository name cannot start or end with a dot")
	}

	if strings.HasPrefix(repo, "-") || strings.HasSuffix(repo, "-") {
		return fmt.Errorf("repository name cannot start or end with a hyphen")
	}

	return nil
}

// ValidateRef validates a Git reference (branch, tag, or commit)
func ValidateRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("reference cannot be empty")
	}

	if len(ref) > 255 {
		return fmt.Errorf("reference too long (max 255 characters)")
	}

	// Check if it's a commit SHA
	if shaRegex.MatchString(ref) {
		return nil
	}

	// Otherwise validate as a branch/tag name
	if !refRegex.MatchString(ref) {
		return fmt.Errorf("invalid reference: must contain only alphanumeric characters, dots, hyphens, underscores, and slashes")
	}

	// Check for invalid patterns
	if strings.Contains(ref, "..") {
		return fmt.Errorf("reference cannot contain '..'")
	}

	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") {
		return fmt.Errorf("reference cannot start or end with a slash")
	}

	if strings.Contains(ref, "//") {
		return fmt.Errorf("reference cannot contain consecutive slashes")
	}

	return nil
}

// ValidateGistID validates a GitHub Gist ID
func ValidateGistID(gistID string) error {
	if gistID == "" {
		return fmt.Errorf("gist ID cannot be empty")
	}

	// Gist IDs are 32 character hex strings
	gistRegex := regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	if !gistRegex.MatchString(gistID) {
		return fmt.Errorf("invalid gist ID: must be a 32-character hexadecimal string")
	}

	return nil
}

// ValidateArchiveFormat validates archive format (zip or tar.gz)
func ValidateArchiveFormat(format string) error {
	format = strings.ToLower(format)

	allowedFormats := []string{"zip", "tar.gz", "tarball", "zipball"}
	for _, allowed := range allowedFormats {
		if format == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid archive format: %s (allowed: zip, tar.gz, tarball, zipball)", format)
}

// ValidateReleaseTag validates a release tag name
func ValidateReleaseTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("release tag cannot be empty")
	}

	if len(tag) > 255 {
		return fmt.Errorf("release tag too long (max 255 characters)")
	}

	// Release tags can contain alphanumeric, dots, hyphens, underscores, slashes, and 'v' prefix
	tagRegex := regexp.MustCompile(`^[a-zA-Z0-9v._/-]{1,255}$`)
	if !tagRegex.MatchString(tag) {
		return fmt.Errorf("invalid release tag format")
	}

	return nil
}
