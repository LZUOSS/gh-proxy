package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/kexi/github-reverse-proxy/internal/auth"
	"golang.org/x/crypto/ssh"
)

// Server represents an SSH server that proxies Git operations to GitHub.
type Server struct {
	config   *ssh.ServerConfig
	listener net.Listener
	addr     string
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

// Config contains SSH server configuration.
type Config struct {
	Address        string // Address to listen on (e.g., ":2222")
	HostKeyPath    string // Path to host private key
	HostKey        []byte // Raw host private key (alternative to HostKeyPath)
	EnablePassword bool   // Enable password authentication
	EnablePubKey   bool   // Enable public key authentication
}

// NewServer creates a new SSH server.
func NewServer(cfg *Config) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.Address == "" {
		cfg.Address = ":2222"
	}

	// Parse or generate host key
	var hostKey ssh.Signer
	var err error

	if len(cfg.HostKey) > 0 {
		hostKey, err = ssh.ParsePrivateKey(cfg.HostKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse host key: %w", err)
		}
	} else {
		// Generate a temporary host key for testing
		// In production, you should load a persistent key
		hostKey, err = generateHostKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate host key: %w", err)
		}
	}

	// Create SSH server config
	sshConfig := &ssh.ServerConfig{
		// Configure authentication callbacks
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if !cfg.EnablePassword {
				return nil, fmt.Errorf("password authentication disabled")
			}
			return handlePasswordAuth(conn.User(), string(password))
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if !cfg.EnablePubKey {
				return nil, fmt.Errorf("public key authentication disabled")
			}
			// For now, accept any public key (in production, validate against authorized keys)
			return &ssh.Permissions{}, nil
		},
		ServerVersion: "SSH-2.0-github-reverse-proxy",
	}

	sshConfig.AddHostKey(hostKey)

	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		config: sshConfig,
		addr:   cfg.Address,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start starts the SSH server and begins accepting connections.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	s.listener = listener
	log.Printf("SSH server listening on %s", s.addr)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// Stop gracefully stops the SSH server.
func (s *Server) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.wg.Wait()
	log.Println("SSH server stopped")
	return nil
}

// acceptLoop continuously accepts incoming SSH connections.
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				log.Printf("failed to accept connection: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single SSH connection.
func (s *Server) handleConnection(netConn net.Conn) {
	defer s.wg.Done()
	defer netConn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.config)
	if err != nil {
		log.Printf("failed to handshake: %v", err)
		return
	}
	defer sshConn.Close()

	log.Printf("new SSH connection from %s (user: %s)", sshConn.RemoteAddr(), sshConn.User())

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		s.wg.Add(1)
		go s.handleChannel(sshConn, newChannel)
	}
}

// handleChannel handles a single SSH channel.
func (s *Server) handleChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel) {
	defer s.wg.Done()

	// Only accept "session" channels
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("failed to accept channel: %v", err)
		return
	}

	// Handle session
	session := NewSession(channel, conn.User())
	session.Handle(requests)
}

// handlePasswordAuth validates password authentication using GitHub PAT.
func handlePasswordAuth(username, password string) (*ssh.Permissions, error) {
	// Validate credentials against GitHub API
	token, err := auth.ValidateBasicAuth(username, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Return permissions with token metadata
	return &ssh.Permissions{
		Extensions: map[string]string{
			"token":    token.Value,
			"username": token.Username,
			"login":    token.Login,
		},
	}, nil
}

// generateHostKey generates a temporary RSA host key.
// In production, you should generate and persist a key to maintain
// consistent host key across restarts.
func generateHostKey() (ssh.Signer, error) {
	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Convert to SSH signer
	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	// Log the public key for debugging
	publicKey := signer.PublicKey()
	log.Printf("Generated SSH host key: %s", ssh.FingerprintSHA256(publicKey))

	return signer, nil
}

// SaveHostKey saves a host key to PEM format (for future use).
func SaveHostKey(key *rsa.PrivateKey) []byte {
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(privateKeyPEM)
}
