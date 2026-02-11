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
func (v *Vault) ListNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")
	limit := int(req.GetInt("limit", 0))   // 0 = no limit
	offset := int(req.GetInt("offset", 0)) // 0 = start from beginning

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

	totalCount := len(notes)
	if totalCount == 0 {
		return mcp.NewToolResultText("No notes found"), nil
	}

	// Apply pagination
	if offset > 0 {
		if offset >= totalCount {
			return mcp.NewToolResultText(fmt.Sprintf("Offset %d exceeds total count %d", offset, totalCount)), nil
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

	return mcp.NewToolResultText(sb.String()), nil
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
	if !v.isPathSafe(fullPath) {
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
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
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
	if !v.isPathSafe(fullPath) {
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

// AppendNoteHandler appends content to a note with optional targeted insertion.
// Supports: position (end/start), after (heading/text), before (heading/text), context_lines.
func (v *Vault) AppendNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	insertContent, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	position := req.GetString("position", "end")
	afterTarget := req.GetString("after", "")
	beforeTarget := req.GetString("before", "")
	contextN := int(req.GetInt("context_lines", 0))

	if !strings.HasSuffix(notePath, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, notePath)

	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Create directory if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	// If after/before is specified, override position
	if afterTarget != "" {
		position = "after"
	} else if beforeTarget != "" {
		position = "before"
	}

	// For simple end/start append on new files, use original behavior
	if position == "end" && afterTarget == "" && beforeTarget == "" {
		return v.appendSimple(fullPath, notePath, insertContent, contextN)
	}

	// For targeted insertion, we need to read the file
	existing, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - create with content regardless of position
			if err := os.WriteFile(fullPath, []byte(insertContent), 0o600); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to create note: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Created: %s", notePath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	lines := strings.Split(string(existing), "\n")
	insertLines := strings.Split(insertContent, "\n")

	insertLineIdx, result, errResult := v.computeInsertion(lines, insertLines, position, afterTarget, beforeTarget)
	if errResult != nil {
		return errResult, nil
	}

	newContent := strings.Join(result, "\n")
	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Inserted content in %s at line %d", notePath, insertLineIdx+1)

	if contextN > 0 {
		allLines := strings.Split(newContent, "\n")
		editEnd := insertLineIdx + len(insertLines)
		if editEnd > len(allLines) {
			editEnd = len(allLines)
		}
		ctxStr := buildEditContext(allLines, insertLineIdx, editEnd, contextN, insertLines)
		fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// computeInsertion determines where to insert lines and builds the result slice.
// Returns the insertion line index, the new full lines slice, and an optional error result.
func (v *Vault) computeInsertion(lines, insertLines []string, position, afterTarget, beforeTarget string) (insertIdx int, resultLines []string, errResult *mcp.CallToolResult) {
	switch position {
	case "start":
		result := make([]string, 0, len(insertLines)+len(lines))
		result = append(result, insertLines...)
		result = append(result, lines...)
		return 0, result, nil

	case "after":
		idx, err := findTargetLine(lines, afterTarget)
		if err != nil {
			return -1, nil, mcp.NewToolResultError(err.Error())
		}
		insertAt := idx + 1
		result := make([]string, 0, len(lines)+len(insertLines))
		result = append(result, lines[:insertAt]...)
		result = append(result, insertLines...)
		result = append(result, lines[insertAt:]...)
		return insertAt, result, nil

	case "before":
		idx, err := findTargetLine(lines, beforeTarget)
		if err != nil {
			return -1, nil, mcp.NewToolResultError(err.Error())
		}
		result := make([]string, 0, len(lines)+len(insertLines))
		result = append(result, lines[:idx]...)
		result = append(result, insertLines...)
		result = append(result, lines[idx:]...)
		return idx, result, nil

	default: // "end"
		result := make([]string, 0, len(lines)+len(insertLines))
		result = append(result, lines...)
		result = append(result, insertLines...)
		return len(lines), result, nil
	}
}

// appendSimple handles the original simple append-to-end behavior
func (v *Vault) appendSimple(fullPath, notePath, content string, contextN int) (*mcp.CallToolResult, error) {
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to open note: %v", err)), nil
	}
	defer f.Close()

	info, _ := f.Stat()
	if info.Size() > 0 {
		content = "\n" + content
	}

	if _, err := f.WriteString(content); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to append to note: %v", err)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Appended to: %s", notePath)

	if contextN > 0 {
		// Re-read file to build context
		data, err := os.ReadFile(fullPath)
		if err == nil {
			allLines := strings.Split(string(data), "\n")
			insertLines := strings.Split(content, "\n")
			editStart := len(allLines) - len(insertLines)
			if editStart < 0 {
				editStart = 0
			}
			ctxStr := buildEditContext(allLines, editStart, len(allLines), contextN, insertLines)
			fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// findTargetLine finds a target line (heading or text match).
// Headings are matched first (case-insensitive), then falls back to text match.
// Returns error if not found or ambiguous (multiple matches).
func findTargetLine(lines []string, target string) (int, error) {
	// First, try to match as a heading
	headingMatches := []int{}
	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			text := strings.TrimSpace(matches[2])
			if strings.EqualFold(text, target) {
				headingMatches = append(headingMatches, i)
			}
		}
	}

	if len(headingMatches) == 1 {
		return headingMatches[0], nil
	}
	if len(headingMatches) > 1 {
		return -1, fmt.Errorf("ambiguous target: found %d headings matching '%s'", len(headingMatches), target)
	}

	// Fall back to text match
	textMatches := []int{}
	for i, line := range lines {
		if strings.Contains(line, target) {
			textMatches = append(textMatches, i)
		}
	}

	if len(textMatches) == 0 {
		return -1, fmt.Errorf("target not found: '%s'", target)
	}
	if len(textMatches) > 1 {
		return -1, fmt.Errorf("ambiguous target: found %d lines containing '%s'. Provide more specific text.", len(textMatches), target)
	}

	return textMatches[0], nil
}

// RecentNotesHandler lists recently modified notes
func (v *Vault) RecentNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := int(req.GetInt("limit", 10))
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	type noteInfo struct {
		path    string
		modTime int64
	}

	var notes []noteInfo
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(v.path, path)
			notes = append(notes, noteInfo{path: relPath, modTime: info.ModTime().Unix()})
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list notes: %v", err)), nil
	}

	if len(notes) == 0 {
		return mcp.NewToolResultText("No notes found"), nil
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(notes)-1; i++ {
		for j := i + 1; j < len(notes); j++ {
			if notes[j].modTime > notes[i].modTime {
				notes[i], notes[j] = notes[j], notes[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Recent %d notes:\n\n", len(notes)))
	for _, n := range notes {
		sb.WriteString(n.path + "\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
