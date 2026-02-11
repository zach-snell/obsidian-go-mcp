package vault

import (
	"path/filepath"
	"testing"
)

func TestIsPathSafe(t *testing.T) {
	v := New("/tmp/test-vault")

	tests := []struct {
		name     string
		path     string
		wantSafe bool
	}{
		{
			name:     "valid path in vault",
			path:     "/tmp/test-vault/notes/test.md",
			wantSafe: true,
		},
		{
			name:     "valid path at vault root",
			path:     "/tmp/test-vault/index.md",
			wantSafe: true,
		},
		{
			name:     "valid nested path",
			path:     "/tmp/test-vault/a/b/c/deep.md",
			wantSafe: true,
		},
		{
			name:     "path traversal attempt",
			path:     "/tmp/test-vault/../etc/passwd",
			wantSafe: false,
		},
		{
			name:     "path traversal in middle",
			path:     "/tmp/test-vault/notes/../../../etc/passwd",
			wantSafe: false,
		},
		{
			name:     "path outside vault",
			path:     "/etc/passwd",
			wantSafe: false,
		},
		{
			name:     "similar prefix but different dir",
			path:     "/tmp/test-vault-other/file.md",
			wantSafe: false,
		},
		{
			name:     "path with dot segments resolved inside",
			path:     "/tmp/test-vault/notes/../docs/file.md",
			wantSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use Clean to normalize paths like the real code does
			got := v.isPathSafe(tt.path)
			if got != tt.wantSafe {
				t.Errorf("isPathSafe(%q) = %v, want %v", tt.path, got, tt.wantSafe)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Test that New normalizes paths
	v := New("./relative/../path")

	// Should be an absolute path
	if !filepath.IsAbs(v.path) {
		t.Errorf("Vault path should be absolute, got: %s", v.path)
	}

	// Should not contain ..
	if filepath.Clean(v.path) != v.path {
		t.Errorf("Vault path should be clean, got: %s", v.path)
	}
}
