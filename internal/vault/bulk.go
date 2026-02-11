package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// hasFrontmatter checks if content has YAML frontmatter
func hasFrontmatter(content string) bool {
	return strings.HasPrefix(content, "---")
}

// BulkTagHandler adds or removes tags from multiple notes
func (v *Vault) BulkTagHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsStr, err := req.RequireString("paths")
	if err != nil {
		return mcp.NewToolResultError("paths is required"), nil
	}

	tag, err := req.RequireString("tag")
	if err != nil {
		return mcp.NewToolResultError("tag is required"), nil
	}

	action := req.GetString("action", "add") // add, remove

	// Normalize tag (remove # prefix if present)
	tag = strings.TrimPrefix(tag, "#")

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return mcp.NewToolResultError("at least one path is required"), nil
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		fullPath := filepath.Join(v.path, p)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read failed", p))
			continue
		}

		contentStr := string(content)
		var modified bool

		if action == "remove" {
			modified, contentStr = removeTagFromNote(contentStr, tag)
		} else {
			modified, contentStr = addTagToNote(contentStr, tag)
		}

		if !modified {
			results = append(results, fmt.Sprintf("%s: no change", p))
			continue
		}

		//#nosec G306 -- Obsidian notes need to be readable by the user
		if err := os.WriteFile(fullPath, []byte(contentStr), 0o644); err != nil {
			errors = append(errors, fmt.Sprintf("%s: write failed", p))
			continue
		}

		results = append(results, fmt.Sprintf("%s: %sed #%s", p, action, tag))
	}

	var sb strings.Builder
	if action == "add" {
		sb.WriteString(fmt.Sprintf("# Bulk Add Tag: #%s\n\n", tag))
	} else {
		sb.WriteString(fmt.Sprintf("# Bulk Remove Tag: #%s\n\n", tag))
	}

	if len(results) > 0 {
		sb.WriteString("## Results\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// addTagToNote adds a tag to a note's content
func addTagToNote(content, tag string) (modified bool, result string) {
	// Check if tag already exists
	existingTags := ExtractTags(content)
	for _, t := range existingTags {
		if strings.EqualFold(t, tag) {
			return false, content
		}
	}

	// Try to add to frontmatter first
	if hasFrontmatter(content) {
		fm := ParseFrontmatter(content)
		if tagsVal, ok := fm["tags"]; ok {
			// Append to existing tags
			newTags := tagsVal + ", " + tag
			return true, setFrontmatterKey(content, "tags", newTags)
		}
		// Add tags field
		return true, addFrontmatterField(content, "tags", tag)
	}

	// Add inline tag at the end
	return true, content + "\n\n#" + tag
}

// removeTagFromNote removes a tag from a note
func removeTagFromNote(content, tag string) (changed bool, newContent string) {
	changed = false
	lines := strings.Split(content, "\n")
	var resultLines []string

	for _, line := range lines {
		newLine := line
		patterns := []string{
			"#" + tag + " ",
			" #" + tag + " ",
			"#" + tag + "\n",
			" #" + tag + "\n",
			"#" + tag,
		}

		for _, pattern := range patterns {
			if strings.Contains(newLine, pattern) {
				newLine = strings.ReplaceAll(newLine, pattern, " ")
				changed = true
			}
		}

		if strings.HasSuffix(strings.TrimSpace(newLine), "#"+tag) {
			idx := strings.LastIndex(newLine, "#"+tag)
			newLine = strings.TrimSpace(newLine[:idx])
			changed = true
		}

		resultLines = append(resultLines, newLine)
	}

	newContent = strings.Join(resultLines, "\n")

	if hasFrontmatter(newContent) {
		fm := ParseFrontmatter(newContent)
		if tagsVal, ok := fm["tags"]; ok {
			tagList := strings.Split(tagsVal, ",")
			var newTags []string
			for _, t := range tagList {
				t = strings.TrimSpace(t)
				t = strings.Trim(t, "[]")
				if !strings.EqualFold(t, tag) {
					newTags = append(newTags, t)
				} else {
					changed = true
				}
			}
			if len(newTags) > 0 {
				newContent = setFrontmatterKey(newContent, "tags", strings.Join(newTags, ", "))
			} else {
				newContent, _ = removeFrontmatterKey(newContent, "tags")
			}
		}
	}

	return changed, newContent
}

// addFrontmatterField adds a new field to existing frontmatter
func addFrontmatterField(content, key, value string) string {
	lines := strings.Split(content, "\n")
	var result []string
	inFrontmatter := false
	fieldAdded := false

	for i, line := range lines {
		if i == 0 && line == "---" {
			inFrontmatter = true
			result = append(result, line)
			continue
		}

		if inFrontmatter && line == "---" {
			// Add field before closing ---
			if !fieldAdded {
				result = append(result, fmt.Sprintf("%s: %s", key, value))
				fieldAdded = true
			}
			inFrontmatter = false
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// BulkMoveHandler moves multiple notes to a folder
func (v *Vault) BulkMoveHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsStr, err := req.RequireString("paths")
	if err != nil {
		return mcp.NewToolResultError("paths is required"), nil
	}

	destination, err := req.RequireString("destination")
	if err != nil {
		return mcp.NewToolResultError("destination folder is required"), nil
	}

	updateLinks := req.GetBool("update_links", true)

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return mcp.NewToolResultError("at least one path is required"), nil
	}

	// Ensure destination exists
	destFull := filepath.Join(v.path, destination)
	if err := os.MkdirAll(destFull, 0o755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create destination: %v", err)), nil
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		oldPath := filepath.Join(v.path, p)
		filename := filepath.Base(p)
		newPath := filepath.Join(destFull, filename)
		newRelPath := filepath.Join(destination, filename)

		// Check source exists
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("%s: not found", p))
			continue
		}

		// Check destination doesn't exist
		if _, err := os.Stat(newPath); err == nil {
			errors = append(errors, fmt.Sprintf("%s: already exists at destination", filename))
			continue
		}

		// Move the file
		if err := os.Rename(oldPath, newPath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: move failed", p))
			continue
		}

		// Update links if requested
		if updateLinks {
			oldName := strings.TrimSuffix(filename, ".md")
			_ = v.updateLinksInVault(oldName, oldName) // Links stay the same, just path changed
		}

		results = append(results, fmt.Sprintf("%s -> %s", p, newRelPath))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Bulk Move to %s\n\n", destination))

	if len(results) > 0 {
		sb.WriteString("## Moved\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// BulkSetFrontmatterHandler sets a frontmatter property on multiple notes
func (v *Vault) BulkSetFrontmatterHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsStr, err := req.RequireString("paths")
	if err != nil {
		return mcp.NewToolResultError("paths is required"), nil
	}

	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError("key is required"), nil
	}

	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError("value is required"), nil
	}

	paths := parsePaths(pathsStr)
	if len(paths) == 0 {
		return mcp.NewToolResultError("at least one path is required"), nil
	}

	var results []string
	var errors []string

	for _, p := range paths {
		if !strings.HasSuffix(p, ".md") {
			p += ".md"
		}

		fullPath := filepath.Join(v.path, p)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: read failed", p))
			continue
		}

		contentStr := string(content)
		newContent := setFrontmatterKey(contentStr, key, value)

		//#nosec G306 -- Obsidian notes need to be readable by the user
		if err := os.WriteFile(fullPath, []byte(newContent), 0o644); err != nil {
			errors = append(errors, fmt.Sprintf("%s: write failed", p))
			continue
		}

		results = append(results, fmt.Sprintf("%s: set %s=%s", p, key, value))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Bulk Set Frontmatter: %s\n\n", key))

	if len(results) > 0 {
		sb.WriteString("## Updated\n\n")
		for _, r := range results {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(errors) > 0 {
		sb.WriteString("\n## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// updateLinksInVault updates wikilinks from oldName to newName across the vault
func (v *Vault) updateLinksInVault(oldName, newName string) error {
	return filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		newContent := updateWikilinks(contentStr, oldName, newName)

		if newContent != contentStr {
			//#nosec G306 -- Obsidian notes need to be readable by the user
			_ = os.WriteFile(path, []byte(newContent), 0o644)
		}

		return nil
	})
}

// updateWikilinks replaces wikilinks from oldName to newName
func updateWikilinks(content, oldName, newName string) string {
	// Handle [[oldName]] and [[oldName|alias]]
	patterns := []struct {
		old string
		new string
	}{
		{"[[" + oldName + "]]", "[[" + newName + "]]"},
		{"[[" + oldName + "|", "[[" + newName + "|"},
	}

	for _, p := range patterns {
		content = strings.ReplaceAll(content, p.old, p.new)
	}

	return content
}
