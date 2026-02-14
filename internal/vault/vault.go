package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Vault represents an Obsidian vault
type Vault struct {
	path string
}

// New creates a new Vault instance
func New(path string) *Vault {
	cleanPath, _ := filepath.Abs(filepath.Clean(path))
	return &Vault{path: cleanPath}
}

// isPathSafe checks if the given path is within the vault (prevents path traversal)
func (v *Vault) isPathSafe(fullPath string) bool {
	cleanPath := filepath.Clean(fullPath)
	rel, err := filepath.Rel(v.path, cleanPath)
	if err != nil {
		return false
	}
	// Path is unsafe if it tries to escape via ".."
	return !strings.HasPrefix(rel, "..") && rel != ".."
}

// ListNotesHandler lists all notes in the vault with optional pagination
func (v *Vault) ListNotesHandler(ctx context.Context, req *mcp.CallToolRequest, args ListNotesArgs) (*mcp.CallToolResult, any, error) {
	dir := args.Directory
	limit := args.Limit
	offset := args.Offset

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var notes []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(v.path, path)
			notes = append(notes, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to list notes: %v", err)
	}

	totalCount := len(notes)
	if totalCount == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "No notes found"},
			},
		}, nil, nil
	}

	// Apply pagination
	if offset > 0 {
		if offset >= totalCount {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Offset %d exceeds total count %d", offset, totalCount)},
				},
			}, nil, nil
		}
		notes = notes[offset:]
	}

	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes", totalCount))
	if offset > 0 || limit > 0 {
		sb.WriteString(fmt.Sprintf(" (showing %d-%d)", offset+1, offset+len(notes)))
	}
	sb.WriteString(":\n\n")
	sb.WriteString(strings.Join(notes, "\n"))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}

// ReadNoteHandler reads a note's content
func (v *Vault) ReadNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args any) (*mcp.CallToolResult, any, error) {
	// Stubbed for now as we are focusing on ListNotes and WriteNote
	return nil, nil, fmt.Errorf("not implemented yet")
}

// WriteNoteHandler creates or updates a note
func (v *Vault) WriteNoteHandler(ctx context.Context, req *mcp.CallToolRequest, args WriteNoteArgs) (*mcp.CallToolResult, any, error) {
	path := args.Path
	content := args.Content

	if !strings.HasSuffix(path, ".md") {
		return nil, nil, fmt.Errorf("path must end with .md")
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	if !v.isPathSafe(fullPath) {
		return nil, nil, fmt.Errorf("path must be within vault")
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
		return nil, nil, fmt.Errorf("failed to write note: %v", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Successfully wrote: %s", path)},
		},
	}, nil, nil
}

// The rest of the handlers are commented out for the PoC to avoid compilation errors
// due to the SDK switch. In a full migration, they would all be updated.

/*
// DeleteNoteHandler deletes a note
func (v *Vault) DeleteNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// ... (content omitted)
}

// ... (other handlers)
*/
