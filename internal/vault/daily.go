package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// DailyNoteHandler gets or creates a daily note
func (v *Vault) DailyNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get optional date parameter (default: today)
	dateStr := req.GetString("date", "")
	folder := req.GetString("folder", "daily")
	format := req.GetString("format", "2006-01-02") // Go date format
	createIfMissing := req.GetBool("create", true)

	var targetDate time.Time
	var err error

	if dateStr == "" {
		targetDate = time.Now()
	} else {
		// Try common date formats
		formats := []string{
			"2006-01-02",
			"01-02-2006",
			"01/02/2006",
			"2006/01/02",
			"Jan 2, 2006",
			"January 2, 2006",
		}
		for _, f := range formats {
			targetDate, err = time.Parse(f, dateStr)
			if err == nil {
				break
			}
		}
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid date format: %s", dateStr)), nil
		}
	}

	// Build filename
	filename := targetDate.Format(format) + ".md"
	var notePath string
	if folder != "" {
		notePath = filepath.Join(folder, filename)
	} else {
		notePath = filename
	}

	fullPath := filepath.Join(v.path, notePath)

	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Check if note exists
	content, err := os.ReadFile(fullPath)
	if err == nil {
		// Note exists, return it
		return mcp.NewToolResultText(fmt.Sprintf("# Daily Note: %s\nPath: %s\n\n%s",
			targetDate.Format("Monday, January 2, 2006"),
			notePath,
			string(content))), nil
	}

	if !os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	// Note doesn't exist
	if !createIfMissing {
		return mcp.NewToolResultText(fmt.Sprintf("Daily note not found: %s", notePath)), nil
	}

	// Create the note
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	// Default template for daily notes
	template := fmt.Sprintf(`# %s

## Tasks
- [ ] 

## Notes

## Journal

`, targetDate.Format("Monday, January 2, 2006"))

	if err := os.WriteFile(fullPath, []byte(template), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Created daily note: %s\n\n%s", notePath, template)), nil
}

// ListDailyNotesHandler lists daily notes in a date range
func (v *Vault) ListDailyNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folder := req.GetString("folder", "daily")
	limit := int(req.GetInt("limit", 30))

	searchPath := v.path
	if folder != "" {
		searchPath = filepath.Join(v.path, folder)
	}

	type noteInfo struct {
		path    string
		modTime time.Time
	}

	var notes []noteInfo

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			relPath, _ := filepath.Rel(v.path, path)
			notes = append(notes, noteInfo{path: relPath, modTime: info.ModTime()})
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list notes: %v", err)), nil
	}

	if len(notes) == 0 {
		return mcp.NewToolResultText("No daily notes found"), nil
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(notes)-1; i++ {
		for j := i + 1; j < len(notes); j++ {
			if notes[j].modTime.After(notes[i].modTime) {
				notes[i], notes[j] = notes[j], notes[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && limit < len(notes) {
		notes = notes[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d daily notes:\n\n", len(notes)))
	for _, n := range notes {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", n.path, n.modTime.Format("Jan 2, 2006")))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
