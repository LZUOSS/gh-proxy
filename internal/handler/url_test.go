package handler

import (
	"testing"
)

func TestParseGitHubURL(t *testing.T) {
	handler := &URLHandler{}

	tests := []struct {
		name     string
		url      string
		wantType string
		wantErr  bool
	}{
		{
			name:     "Full GitHub raw URL with refs/heads",
			url:      "https://github.com/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md",
			wantType: "raw",
			wantErr:  false,
		},
		{
			name:     "Full GitHub raw URL without scheme",
			url:      "github.com/owner/repo/raw/main/README.md",
			wantType: "raw",
			wantErr:  false,
		},
		{
			name:     "GitHub releases URL",
			url:      "https://github.com/owner/repo/releases/download/v1.0.0/binary.tar.gz",
			wantType: "releases",
			wantErr:  false,
		},
		{
			name:     "raw.githubusercontent.com URL",
			url:      "https://raw.githubusercontent.com/owner/repo/main/file.md",
			wantType: "raw",
			wantErr:  false,
		},
		{
			name:     "GitHub archive URL",
			url:      "https://github.com/owner/repo/archive/refs/tags/v1.0.0.tar.gz",
			wantType: "archive",
			wantErr:  false,
		},
		{
			name:     "GitHub API URL",
			url:      "https://api.github.com/repos/owner/repo",
			wantType: "api",
			wantErr:  false,
		},
		{
			name:     "GitHub gist URL",
			url:      "https://gist.github.com/user/gist-id/raw/file.txt",
			wantType: "gist",
			wantErr:  false,
		},
		{
			name:     "GitHub blob URL (should convert to raw)",
			url:      "https://github.com/owner/repo/blob/main/file.md",
			wantType: "raw",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := handler.parseGitHubURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGitHubURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && info.Type != tt.wantType {
				t.Errorf("parseGitHubURL() Type = %v, want %v", info.Type, tt.wantType)
			}
		})
	}
}

func TestParseGitHubComURL_RawWithRefsHeads(t *testing.T) {
	handler := &URLHandler{}

	// Test the specific example from the user
	info, err := handler.parseGitHubURL("https://github.com/ZhiShengYuan/inningbo-go/raw/refs/heads/main/ARCHITECTURE_REFACTORING.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Type != "raw" {
		t.Errorf("Type = %v, want raw", info.Type)
	}
	if info.Owner != "ZhiShengYuan" {
		t.Errorf("Owner = %v, want ZhiShengYuan", info.Owner)
	}
	if info.Repo != "inningbo-go" {
		t.Errorf("Repo = %v, want inningbo-go", info.Repo)
	}
	if info.Ref != "refs/heads" {
		t.Errorf("Ref = %v, want refs/heads", info.Ref)
	}
	if info.Filepath != "/main/ARCHITECTURE_REFACTORING.md" {
		t.Errorf("Filepath = %v, want /main/ARCHITECTURE_REFACTORING.md", info.Filepath)
	}
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "Full GitHub HTTPS URL",
			path: "/https://github.com/owner/repo/raw/main/file.md",
			want: true,
		},
		{
			name: "GitHub URL without scheme",
			path: "/github.com/owner/repo/raw/main/file.md",
			want: true,
		},
		{
			name: "raw.githubusercontent.com URL",
			path: "/https://raw.githubusercontent.com/owner/repo/main/file.md",
			want: true,
		},
		{
			name: "API GitHub URL",
			path: "/https://api.github.com/repos/owner/repo",
			want: true,
		},
		{
			name: "Regular path-based URL",
			path: "/owner/repo/raw/main/file.md",
			want: false,
		},
		{
			name: "Health check endpoint",
			path: "/health",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isGitHubURL(tt.path); got != tt.want {
				t.Errorf("isGitHubURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
