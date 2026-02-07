package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidatePath validates that a path doesn't contain path traversal attempts
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Normalize the path
	cleanPath := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: path contains '..'")
	}

	// Check for absolute paths (should be relative)
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths are not allowed")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// Check for common dangerous patterns
	dangerousPatterns := []string{
		"../",
		"..\\",
		"/..",
		"\\..",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return fmt.Errorf("path traversal detected: path contains '%s'", pattern)
		}
	}

	// Additional check: ensure cleaned path doesn't start with ../
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("path traversal detected: cleaned path starts with '..'")
	}

	return nil
}
