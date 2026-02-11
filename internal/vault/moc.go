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

var (
	// Matches wikilinks: [[Note Name]] or [[path/to/note|Alias]]
	wikilinkRegex = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	// Matches H1 title: # Title
	h1Regex = regexp.MustCompile(`(?m)^#\s+(.+)$`)
)

// MOC represents a Map of Content
type MOC struct {
	Path        string   `json:"path"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	LinkedNotes []string `json:"linkedNotes"`
}

// ExtractWikilinks extracts all wikilinks from content
func ExtractWikilinks(content string) []string {
	matches := wikilinkRegex.FindAllStringSubmatch(content, -1)
	var links []string
	seen := make(map[string]bool)
	for _, match := range matches {
		link := strings.TrimSpace(match[1])
		if !seen[link] {
			links = append(links, link)
			seen[link] = true
		}
	}
	return links
}

// ExtractH1Title extracts the first H1 title from content
func ExtractH1Title(content string) string {
	match := h1Regex.FindStringSubmatch(content)
	if match != nil {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// IsMOC checks if a note is a MOC (has #moc tag)
func IsMOC(content string) bool {
	tags := ExtractTags(content)
	for _, tag := range tags {
		if strings.ToLower(tag) == "moc" {
			return true
		}
	}
	return false
}

// DiscoverMOCsHandler discovers all MOCs in the vault
func (v *Vault) DiscoverMOCsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var mocs []MOC

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
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

		contentStr := string(content)
		if !IsMOC(contentStr) {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		title := ExtractH1Title(contentStr)
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), ".md")
		}

		mocs = append(mocs, MOC{
			Path:        relPath,
			Title:       title,
			Tags:        ExtractTags(contentStr),
			LinkedNotes: ExtractWikilinks(contentStr),
		})

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Discovery failed: %v", err)), nil
	}

	if len(mocs) == 0 {
		return mcp.NewToolResultText("No MOCs found (notes with #moc tag)"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d MOCs:\n\n", len(mocs)))

	for _, m := range mocs {
		sb.WriteString(fmt.Sprintf("## %s\n", m.Title))
		sb.WriteString(fmt.Sprintf("Path: %s\n", m.Path))
		sb.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(m.Tags, ", ")))
		if len(m.LinkedNotes) > 0 {
			sb.WriteString(fmt.Sprintf("Links (%d): %s\n", len(m.LinkedNotes), strings.Join(m.LinkedNotes, ", ")))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
