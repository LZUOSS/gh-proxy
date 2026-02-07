package ssh

import (
	"fmt"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

const (
	githubSSHHost = "github.com:22"
	githubSSHUser = "git"
)

// handleGitPassthrough handles bidirectional streaming between client and GitHub.
func handleGitPassthrough(clientChannel ssh.Channel, gitCmd *GitCommand) int {
	// Validate command
	if err := gitCmd.Validate(); err != nil {
		log.Printf("invalid git command: %v", err)
		clientChannel.Write([]byte(fmt.Sprintf("Error: %v\r\n", err)))
		return 1
	}

	// Connect to GitHub's SSH server
	githubConn, err := connectToGitHub(gitCmd)
	if err != nil {
		log.Printf("failed to connect to GitHub: %v", err)
		clientChannel.Write([]byte(fmt.Sprintf("Error connecting to GitHub: %v\r\n", err)))
		return 1
	}
	defer githubConn.Close()

	// Create bidirectional streaming
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Client -> GitHub
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(githubConn, clientChannel)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("client to GitHub copy error: %w", err)
		}
		// Close write side to signal EOF to GitHub
		if closer, ok := githubConn.(interface{ CloseWrite() error }); ok {
			closer.CloseWrite()
		}
	}()

	// GitHub -> Client
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(clientChannel, githubConn)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("GitHub to client copy error: %w", err)
		}
		// Close write side to signal EOF to client
		if closer, ok := clientChannel.(interface{ CloseWrite() error }); ok {
			closer.CloseWrite()
		}
	}()

	// Wait for both goroutines to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		log.Printf("passthrough error: %v", err)
		return 1
	}

	return 0
}

// connectToGitHub establishes an SSH connection to GitHub's SSH server.
func connectToGitHub(gitCmd *GitCommand) (ssh.Channel, error) {
	// Create SSH client config for connecting to GitHub
	config := &ssh.ClientConfig{
		User: githubSSHUser,
		Auth: []ssh.AuthMethod{
			// GitHub doesn't actually check authentication for public repos via SSH protocol
			// The authentication happens at the Git protocol level
			ssh.Password(""),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // In production, verify GitHub's host key
	}

	// Connect to GitHub's SSH server
	conn, err := ssh.Dial("tcp", githubSSHHost, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial GitHub SSH: %w", err)
	}

	// Open a session channel
	session, err := conn.NewSession()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get stdin/stdout pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	// Start the Git command on GitHub's SSH server
	gitCommand := gitCmd.FormatGitHubCommand()
	if err := session.Start(gitCommand); err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to start git command on GitHub: %w", err)
	}

	// Create a wrapper that implements ssh.Channel interface
	wrapper := &githubChannelWrapper{
		session: session,
		conn:    conn,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}

	return wrapper, nil
}

// githubChannelWrapper wraps an SSH session to GitHub to implement ssh.Channel.
type githubChannelWrapper struct {
	session *ssh.Session
	conn    *ssh.Client
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
}

// Read reads from GitHub's stdout.
func (w *githubChannelWrapper) Read(data []byte) (int, error) {
	return w.stdout.Read(data)
}

// Write writes to GitHub's stdin.
func (w *githubChannelWrapper) Write(data []byte) (int, error) {
	return w.stdin.Write(data)
}

// Close closes the session and connection.
func (w *githubChannelWrapper) Close() error {
	w.stdin.Close()
	w.session.Close()
	return w.conn.Close()
}

// CloseWrite closes the write side (stdin).
func (w *githubChannelWrapper) CloseWrite() error {
	return w.stdin.Close()
}

// CloseRead is not implemented but required for ssh.Channel interface.
func (w *githubChannelWrapper) CloseRead() error {
	return nil
}

// SendRequest is not implemented but required for ssh.Channel interface.
func (w *githubChannelWrapper) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, fmt.Errorf("SendRequest not supported")
}

// Stderr returns the stderr reader.
func (w *githubChannelWrapper) Stderr() io.ReadWriter {
	return &stderrWrapper{w.stderr}
}

// stderrWrapper wraps stderr to implement io.ReadWriter.
type stderrWrapper struct {
	io.Reader
}

func (s *stderrWrapper) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("stderr write not supported")
}
