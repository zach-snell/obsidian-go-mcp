package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Vault represents an Obsidian vault
type Vault struct {
	path string
}

// New creates a new Vault instance
func New(path string) *Vault {
	return &Vault{path: path}
}

// ListNotesHandler lists all notes in the vault
func (v *Vault) ListNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")

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
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list notes: %v", err)), nil
	}

	if len(notes) == 0 {
		return mcp.NewToolResultText("No notes found"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Found %d notes:\n\n%s", len(notes), strings.Join(notes, "\n"))), nil
}

// ReadNoteHandler reads a note's content
func (v *Vault) ReadNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	if !strings.HasPrefix(fullPath, v.path) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", path)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	return mcp.NewToolResultText(string(content)), nil
}

// WriteNoteHandler creates or updates a note
func (v *Vault) WriteNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	content, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	if !strings.HasPrefix(fullPath, v.path) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote: %s", path)), nil
}

// DeleteNoteHandler deletes a note
func (v *Vault) DeleteNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	if !strings.HasPrefix(fullPath, v.path) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", path)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted: %s", path)), nil
}
