package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// SearchResult represents a search match
type SearchResult struct {
	File    string
	Line    int
	Content string
}

// SearchVaultHandler searches for content in vault notes
func (v *Vault) SearchVaultHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}

	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult

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

		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(v.path, path)

		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				results = append(results, SearchResult{
					File:    relPath,
					Line:    i + 1,
					Content: strings.TrimSpace(line),
				})
			}
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(results) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No matches found for: %s", query)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matches for %q:\n\n", len(results), query))

	currentFile := ""
	for _, r := range results {
		if r.File != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("## %s\n", r.File))
			currentFile = r.File
		}
		sb.WriteString(fmt.Sprintf("  L%d: %s\n", r.Line, truncate(r.Content, 100)))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
