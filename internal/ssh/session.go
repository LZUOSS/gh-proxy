package ssh

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Session represents an SSH session handling Git operations.
type Session struct {
	channel  ssh.Channel
	username string
}

// NewSession creates a new SSH session.
func NewSession(channel ssh.Channel, username string) *Session {
	return &Session{
		channel:  channel,
		username: username,
	}
}

// Handle processes SSH session requests.
func (s *Session) Handle(requests <-chan *ssh.Request) {
	defer s.channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			// Git commands come as "exec" requests
			s.handleExec(req)
		case "shell":
			// Reject shell requests - we only support Git operations
			req.Reply(false, nil)
			s.channel.Write([]byte("Shell access is not allowed. This server only supports Git operations.\r\n"))
		case "pty-req":
			// Reject PTY requests
			req.Reply(false, nil)
		case "env":
			// Accept but ignore environment variables
			req.Reply(true, nil)
		default:
			// Reject unknown request types
			log.Printf("unknown request type: %s", req.Type)
			req.Reply(false, nil)
		}
	}
}

// handleExec handles the "exec" request for Git commands.
func (s *Session) handleExec(req *ssh.Request) {
	// Parse the command from the request payload
	command := string(req.Payload[4:]) // Skip the first 4 bytes (length prefix)

	log.Printf("exec command from %s: %s", s.username, command)

	// Parse Git command
	gitCmd, err := parseGitCommand(command)
	if err != nil {
		req.Reply(false, nil)
		s.channel.Write([]byte(fmt.Sprintf("Error: %v\r\n", err)))
		s.channel.SendRequest("exit-status", false, ssh.Marshal(exitStatus{Status: 1}))
		return
	}

	// Reply to the exec request
	req.Reply(true, nil)

	// Execute the Git command through passthrough
	exitCode := handleGitPassthrough(s.channel, gitCmd)

	// Send exit status
	s.channel.SendRequest("exit-status", false, ssh.Marshal(exitStatus{Status: uint32(exitCode)}))
}

// parseGitCommand parses a Git command and extracts owner and repo.
func parseGitCommand(command string) (*GitCommand, error) {
	// Trim whitespace
	command = strings.TrimSpace(command)

	// Expected format: "git-upload-pack '/owner/repo.git'" or "git-receive-pack '/owner/repo.git'"
	parts := strings.Fields(command)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid git command format: %s", command)
	}

	gitOp := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")

	// Validate Git operation
	if gitOp != "git-upload-pack" && gitOp != "git-receive-pack" {
		return nil, fmt.Errorf("unsupported git operation: %s (only git-upload-pack and git-receive-pack are supported)", gitOp)
	}

	// Parse repository path
	owner, repo, err := parseRepoPath(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid repository path: %w", err)
	}

	return &GitCommand{
		Operation: gitOp,
		Owner:     owner,
		Repo:      repo,
		RepoPath:  repoPath,
	}, nil
}

// parseRepoPath extracts owner and repo from a repository path.
// Accepts formats: "/owner/repo.git", "owner/repo.git", "/owner/repo"
func parseRepoPath(path string) (owner, repo string, err error) {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	// Remove .git suffix if present
	path = strings.TrimSuffix(path, ".git")

	// Split by slash
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repo path format: %s (expected: owner/repo)", path)
	}

	owner = parts[0]
	repo = parts[1]

	// Validate owner and repo names
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("owner and repo cannot be empty")
	}

	// Basic validation - GitHub allows alphanumeric, hyphens, underscores, and dots
	if !isValidGitHubName(owner) {
		return "", "", fmt.Errorf("invalid owner name: %s", owner)
	}
	if !isValidGitHubName(repo) {
		return "", "", fmt.Errorf("invalid repo name: %s", repo)
	}

	return owner, repo, nil
}

// isValidGitHubName checks if a name is valid for GitHub (owner/repo).
func isValidGitHubName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}

	// GitHub names can contain alphanumeric characters, hyphens, underscores, and dots
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.') {
			return false
		}
	}

	return true
}

// exitStatus is used for sending exit status over SSH.
type exitStatus struct {
	Status uint32
}
