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

// Frontmatter represents parsed YAML frontmatter
type Frontmatter map[string]string

// ParseFrontmatter extracts frontmatter from note content
func ParseFrontmatter(content string) Frontmatter {
	fm := make(Frontmatter)

	if !strings.HasPrefix(content, "---") {
		return fm
	}

	endIdx := strings.Index(content[3:], "---")
	if endIdx < 0 {
		return fm
	}

	fmContent := content[3 : endIdx+3]
	lines := strings.Split(fmContent, "\n")

	// Simple YAML parsing (key: value)
	keyValueRegex := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:\s*(.*)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if match := keyValueRegex.FindStringSubmatch(line); match != nil {
			key := strings.ToLower(match[1])
			value := strings.Trim(strings.TrimSpace(match[2]), `"'`)
			fm[key] = value
		}
	}

	return fm
}

// QueryFrontmatterHandler searches notes by frontmatter properties
func (v *Vault) QueryFrontmatterHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required (format: key=value or key:value)"), nil
	}

	dir := req.GetString("directory", "")

	// Parse query (supports key=value or key:value)
	var key, value string
	if idx := strings.Index(query, "="); idx > 0 {
		key = strings.ToLower(strings.TrimSpace(query[:idx]))
		value = strings.ToLower(strings.TrimSpace(query[idx+1:]))
	} else if idx := strings.Index(query, ":"); idx > 0 {
		key = strings.ToLower(strings.TrimSpace(query[:idx]))
		value = strings.ToLower(strings.TrimSpace(query[idx+1:]))
	} else {
		return mcp.NewToolResultError("Invalid query format. Use: key=value or key:value"), nil
	}

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	type result struct {
		path        string
		frontmatter Frontmatter
	}

	var results []result

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		fm := ParseFrontmatter(string(content))
		if len(fm) == 0 {
			return nil
		}

		// Check if frontmatter matches query
		if fmValue, ok := fm[key]; ok {
			// Support partial matching (contains)
			if strings.Contains(strings.ToLower(fmValue), value) {
				relPath, _ := filepath.Rel(v.path, path)
				results = append(results, result{path: relPath, frontmatter: fm})
			}
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Query failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No notes found matching: %s", query)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d notes matching %q:\n\n", len(results), query))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("## %s\n", r.path))
		for k, v := range r.frontmatter {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// GetFrontmatterHandler returns frontmatter for a specific note
func (v *Vault) GetFrontmatterHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, path)

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

	fm := ParseFrontmatter(string(content))

	if len(fm) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No frontmatter found in: %s", path)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Frontmatter for %s:\n\n", path))
	for k, v := range fm {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
