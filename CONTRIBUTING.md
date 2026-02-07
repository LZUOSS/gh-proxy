# Contributing to GitHub Reverse Proxy

Thank you for your interest in contributing to GitHub Reverse Proxy! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Bugs

Before creating a bug report:
- Check the existing issues to avoid duplicates
- Collect information about the bug (logs, configuration, environment)

When creating a bug report, include:
- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- System information (OS, Go version, proxy version)
- Relevant logs and configuration (redact sensitive data)
- Screenshots if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:
- Use a clear, descriptive title
- Provide detailed description of the proposed functionality
- Explain why this enhancement would be useful
- Include examples of how it would be used

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the coding standards
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Ensure tests pass** with `make test`
6. **Run linters** with `make lint`
7. **Create a pull request** with a clear description

## Development Setup

### Prerequisites

- Go 1.24 or higher
- Make
- Git
- Docker (optional, for testing)

### Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/github-reverse-proxy.git
cd github-reverse-proxy

# Add upstream remote
git remote add upstream https://github.com/kexi/github-reverse-proxy.git

# Install dependencies
make install

# Build the project
make build

# Run tests
make test
```

### Project Structure

```
github-reverse-proxy/
├── cmd/server/          # Main application entry point
├── internal/            # Internal packages
│   ├── auth/           # Authentication
│   ├── cache/          # Caching system
│   ├── config/         # Configuration management
│   ├── handler/        # HTTP handlers
│   ├── metrics/        # Prometheus metrics
│   ├── middleware/     # HTTP middleware
│   ├── proxy/          # Proxy client
│   ├── ratelimit/      # Rate limiting
│   ├── security/       # Security features
│   ├── server/         # HTTP server
│   ├── ssh/            # SSH proxy
│   └── util/           # Utilities
├── configs/            # Configuration files
└── deploy/             # Deployment files
```

## Coding Standards

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines and:

- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Handle errors properly

### Code Organization

- Place new features in appropriate packages
- Keep packages focused on a single responsibility
- Use internal packages for implementation details
- Export only what's necessary

### Testing

- Write unit tests for new functionality
- Aim for high test coverage
- Use table-driven tests where appropriate
- Mock external dependencies
- Test error cases

Example test structure:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "result", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Feature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Feature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Feature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Add godoc comments for exported types and functions
- Update CHANGELOG.md for notable changes
- Include examples in documentation

## Development Workflow

### Creating a Feature Branch

```bash
# Update your local main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-feature
```

### Making Changes

```bash
# Make your changes
vim internal/package/file.go

# Format code
make fmt

# Run tests
make test

# Run linters
make lint

# Commit changes
git add .
git commit -m "Add feature: description"
```

### Commit Messages

Write clear commit messages:
- Use present tense ("Add feature" not "Added feature")
- First line should be a short summary (50 chars or less)
- Add detailed description if needed
- Reference issues: "Fixes #123" or "Related to #456"

Example:
```
Add cache eviction policy

Implement LRU cache eviction to manage memory usage.
This improves performance for long-running instances.

Fixes #123
```

### Submitting Pull Request

```bash
# Push to your fork
git push origin feature/my-feature

# Create pull request on GitHub
```

Pull request description should include:
- Summary of changes
- Motivation and context
- How to test the changes
- Related issues

### Code Review

- Address review comments promptly
- Be open to feedback
- Make requested changes in new commits
- Update PR description if scope changes

## Testing Guidelines

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/cache/...

# Run with coverage
make test-coverage

# Run with race detection
go test -race ./...
```

### Writing Tests

- Test happy paths and edge cases
- Test error conditions
- Use meaningful test names
- Clean up resources (defer cleanup)
- Avoid test interdependencies

## Performance Considerations

- Benchmark performance-critical code
- Profile before optimizing
- Use proper synchronization primitives
- Avoid unnecessary allocations
- Consider memory usage for caching

## Security Considerations

- Validate all inputs
- Avoid hardcoding secrets
- Use secure defaults
- Follow OWASP guidelines
- Handle authentication properly

## Documentation

### Code Comments

```go
// Cache provides an interface for caching GitHub responses.
// It supports both memory and disk-based caching with configurable
// eviction policies.
type Cache interface {
    // Get retrieves a value from the cache.
    // Returns ErrCacheMiss if the key is not found.
    Get(key string) ([]byte, error)

    // Set stores a value in the cache with the given TTL.
    Set(key string, value []byte, ttl time.Duration) error
}
```

### README Updates

Update README.md when adding:
- New features
- Configuration options
- API endpoints
- Usage examples

## Release Process

Maintainers handle releases:
1. Update version number
2. Update CHANGELOG.md
3. Create git tag
4. Build release binaries
5. Publish GitHub release

## Getting Help

- GitHub Issues: For bugs and feature requests
- GitHub Discussions: For questions and discussions
- Documentation: Check README.md and wiki

## Recognition

Contributors are recognized in:
- CHANGELOG.md for significant changes
- README.md contributors section
- Release notes

Thank you for contributing to GitHub Reverse Proxy!
