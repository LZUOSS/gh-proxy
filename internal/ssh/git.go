package ssh

import (
	"fmt"
	"strings"
)

// GitCommand represents a parsed Git command.
type GitCommand struct {
	Operation string // "git-upload-pack" or "git-receive-pack"
	Owner     string // GitHub repository owner
	Repo      string // GitHub repository name
	RepoPath  string // Full repository path (e.g., "/owner/repo.git")
}

// IsUpload returns true if this is a git-upload-pack command (fetch/clone).
func (g *GitCommand) IsUpload() bool {
	return g.Operation == "git-upload-pack"
}

// IsReceive returns true if this is a git-receive-pack command (push).
func (g *GitCommand) IsReceive() bool {
	return g.Operation == "git-receive-pack"
}

// GitHubRepoURL returns the GitHub SSH URL for this repository.
func (g *GitCommand) GitHubRepoURL() string {
	return fmt.Sprintf("%s/%s", g.Owner, g.Repo)
}

// String returns a string representation of the Git command.
func (g *GitCommand) String() string {
	return fmt.Sprintf("%s '%s/%s.git'", g.Operation, g.Owner, g.Repo)
}

// Validate performs additional validation on the Git command.
func (g *GitCommand) Validate() error {
	if g.Operation == "" {
		return fmt.Errorf("operation cannot be empty")
	}

	if !g.IsUpload() && !g.IsReceive() {
		return fmt.Errorf("invalid operation: %s", g.Operation)
	}

	if g.Owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}

	if g.Repo == "" {
		return fmt.Errorf("repo cannot be empty")
	}

	// Additional security checks
	if strings.Contains(g.Owner, "..") || strings.Contains(g.Repo, "..") {
		return fmt.Errorf("path traversal detected in owner/repo")
	}

	if strings.Contains(g.Owner, "/") || strings.Contains(g.Repo, "/") {
		return fmt.Errorf("invalid characters in owner/repo")
	}

	return nil
}

// FormatGitHubCommand formats the command for execution against GitHub.
func (g *GitCommand) FormatGitHubCommand() string {
	// Format: git-upload-pack 'owner/repo.git'
	return fmt.Sprintf("%s '%s/%s.git'", g.Operation, g.Owner, g.Repo)
}
