package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// BacklinksHandler finds all notes that link to a given note
func (v *Vault) BacklinksHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	// Normalize target: remove .md extension for matching
	targetName := strings.TrimSuffix(target, ".md")
	targetBase := filepath.Base(targetName)

	// Build regex patterns for wikilinks
	// Matches: [[target]], [[target|alias]], [[path/to/target]], [[path/to/target|alias]]
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\[\[` + regexp.QuoteMeta(targetName) + `(\|[^\]]+)?\]\]`),
		regexp.MustCompile(`\[\[` + regexp.QuoteMeta(targetBase) + `(\|[^\]]+)?\]\]`),
	}

	type backlink struct {
		path    string
		count   int
		context []string
	}

	var backlinks []backlink

	err = filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		// Skip the target note itself
		if relPath == target || strings.TrimSuffix(relPath, ".md") == targetName {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		var matches []string
		matchCount := 0

		for _, pattern := range patterns {
			for i, line := range lines {
				if pattern.MatchString(line) {
					matchCount++
					// Add context (truncated line)
					ctx := strings.TrimSpace(line)
					if len(ctx) > 100 {
						ctx = ctx[:100] + "..."
					}
					matches = append(matches, fmt.Sprintf("L%d: %s", i+1, ctx))
				}
			}
		}

		if matchCount > 0 {
			backlinks = append(backlinks, backlink{
				path:    relPath,
				count:   matchCount,
				context: matches,
			})
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search backlinks: %v", err)), nil
	}

	if len(backlinks) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No backlinks found for: %s", target)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes linking to %s:\n\n", len(backlinks), target))

	for _, bl := range backlinks {
		sb.WriteString(fmt.Sprintf("## %s (%d links)\n", bl.path, bl.count))
		for _, ctx := range bl.context {
			sb.WriteString(fmt.Sprintf("  %s\n", ctx))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// RenameNoteHandler renames a note and updates all links to it
func (v *Vault) RenameNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	oldPath, err := req.RequireString("old_path")
	if err != nil {
		return mcp.NewToolResultError("old_path is required"), nil
	}

	newPath, err := req.RequireString("new_path")
	if err != nil {
		return mcp.NewToolResultError("new_path is required"), nil
	}

	if !strings.HasSuffix(oldPath, ".md") || !strings.HasSuffix(newPath, ".md") {
		return mcp.NewToolResultError("paths must end with .md"), nil
	}

	oldFullPath := filepath.Join(v.path, oldPath)
	newFullPath := filepath.Join(v.path, newPath)

	if !v.isPathSafe(oldFullPath) || !v.isPathSafe(newFullPath) {
		return mcp.NewToolResultError("paths must be within vault"), nil
	}

	// Check source exists
	if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", oldPath)), nil
	}

	// Check destination doesn't exist
	if _, err := os.Stat(newFullPath); err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Destination already exists: %s", newPath)), nil
	}

	// Prepare link patterns
	oldName := strings.TrimSuffix(oldPath, ".md")
	oldBase := strings.TrimSuffix(filepath.Base(oldPath), ".md")
	newName := strings.TrimSuffix(newPath, ".md")
	newBase := strings.TrimSuffix(filepath.Base(newPath), ".md")

	// Patterns to find and replace
	replacements := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\]\]`),
			replace: "[[" + newName + "]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldName) + `\|([^\]]+)\]\]`),
			replace: "[[" + newName + "|$1]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldBase) + `\]\]`),
			replace: "[[" + newBase + "]]",
		},
		{
			pattern: regexp.MustCompile(`\[\[` + regexp.QuoteMeta(oldBase) + `\|([^\]]+)\]\]`),
			replace: "[[" + newBase + "|$1]]",
		},
	}

	// Update all notes that link to the old path
	updatedFiles := 0
	err = filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if path == oldFullPath {
			return nil // Skip the file we're renaming
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		newContent := contentStr

		for _, r := range replacements {
			newContent = r.pattern.ReplaceAllString(newContent, r.replace)
		}

		if newContent != contentStr {
			if err := os.WriteFile(path, []byte(newContent), 0o600); err != nil {
				return nil // Continue on error
			}
			updatedFiles++
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update links: %v", err)), nil
	}

	// Create destination directory if needed
	newDir := filepath.Dir(newFullPath)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	// Rename the file
	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to rename note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Renamed %s -> %s\nUpdated links in %d files", oldPath, newPath, updatedFiles)), nil
}
