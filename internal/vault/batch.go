package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// headingRegex matches markdown headings (# through ######)
var headingRegex = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)

// Heading represents a markdown heading
type Heading struct {
	Level   int    `json:"level"`
	Text    string `json:"text"`
	Line    int    `json:"line"`
	EndLine int    `json:"end_line"` // Last line of section content (0 = extends to EOF or not computed)
}

// NoteSummary provides a lightweight view of a note
type NoteSummary struct {
	Path        string      `json:"path"`
	Frontmatter Frontmatter `json:"frontmatter,omitempty"`
	Preview     string      `json:"preview"`
	WordCount   int         `json:"word_count"`
	LinkCount   int         `json:"link_count"`
	TagCount    int         `json:"tag_count"`
	Headings    []Heading   `json:"headings,omitempty"`
}

// ReadNotesHandler reads multiple notes in one call
func (v *Vault) ReadNotesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsRaw := req.GetString("paths", "")
	includeFrontmatter := req.GetBool("include_frontmatter", true)

	if pathsRaw == "" {
		return mcp.NewToolResultError("paths is required (comma-separated or JSON array)"), nil
	}

	// Parse paths - support both comma-separated and JSON array
	var paths []string
	if strings.HasPrefix(pathsRaw, "[") {
		if err := json.Unmarshal([]byte(pathsRaw), &paths); err != nil {
			// Fall back to comma-separated
			paths = strings.Split(pathsRaw, ",")
		}
	} else {
		paths = strings.Split(pathsRaw, ",")
	}

	// Clean up paths
	for i := range paths {
		paths[i] = strings.TrimSpace(paths[i])
		if !strings.HasSuffix(paths[i], ".md") {
			paths[i] += ".md"
		}
	}

	var sb strings.Builder
	successCount := 0

	for _, notePath := range paths {
		fullPath := filepath.Join(v.path, notePath)
		if !v.isPathSafe(fullPath) {
			fmt.Fprintf(&sb, "## %s\n\n**Error:** path must be within vault\n\n---\n\n", notePath)
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(&sb, "## %s\n\n**Error:** note not found\n\n---\n\n", notePath)
			} else {
				fmt.Fprintf(&sb, "## %s\n\n**Error:** %v\n\n---\n\n", notePath, err)
			}
			continue
		}

		contentStr := string(content)
		successCount++

		fmt.Fprintf(&sb, "## %s\n\n", notePath)

		if includeFrontmatter {
			fm := ParseFrontmatter(contentStr)
			if len(fm) > 0 {
				sb.WriteString("**Frontmatter:**\n```yaml\n")
				for k, val := range fm {
					fmt.Fprintf(&sb, "%s: %v\n", k, val)
				}
				sb.WriteString("```\n\n")
			}
		}

		// Remove frontmatter for content display
		body := RemoveFrontmatter(contentStr)
		sb.WriteString(body)
		sb.WriteString("\n\n---\n\n")
	}

	if successCount == 0 {
		return mcp.NewToolResultError("no notes could be read"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Read %d/%d notes:\n\n%s", successCount, len(paths), sb.String())), nil
}

// GetNoteSummaryHandler returns a lightweight summary of a note
func (v *Vault) GetNoteSummaryHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	previewLines := int(req.GetInt("lines", 5))

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", notePath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	contentStr := string(content)
	body := RemoveFrontmatter(contentStr)

	// Build summary
	summary := NoteSummary{
		Path:        notePath,
		Frontmatter: ParseFrontmatter(contentStr),
		WordCount:   len(strings.Fields(body)),
		LinkCount:   len(ExtractWikilinks(body)),
		TagCount:    len(ExtractTags(contentStr)),
		Headings:    extractHeadings(body),
	}

	// Get preview lines
	lines := strings.Split(body, "\n")
	if previewLines > len(lines) {
		previewLines = len(lines)
	}
	summary.Preview = strings.Join(lines[:previewLines], "\n")

	// Format output
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Summary: %s\n\n", notePath)

	if len(summary.Frontmatter) > 0 {
		sb.WriteString("## Frontmatter\n")
		for k, val := range summary.Frontmatter {
			fmt.Fprintf(&sb, "- **%s:** %v\n", k, val)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Stats\n")
	fmt.Fprintf(&sb, "- Words: %d\n", summary.WordCount)
	fmt.Fprintf(&sb, "- Links: %d\n", summary.LinkCount)
	fmt.Fprintf(&sb, "- Tags: %d\n", summary.TagCount)
	fmt.Fprintf(&sb, "- Headings: %d\n", len(summary.Headings))
	sb.WriteString("\n")

	if len(summary.Headings) > 0 {
		sb.WriteString("## Structure\n")
		for _, h := range summary.Headings {
			indent := strings.Repeat("  ", h.Level-1)
			fmt.Fprintf(&sb, "%s- %s\n", indent, h.Text)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Preview\n")
	sb.WriteString(summary.Preview)

	return mcp.NewToolResultText(sb.String()), nil
}

// GetSectionHandler extracts a specific heading section from a note
func (v *Vault) GetSectionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	heading, err := req.RequireString("heading")
	if err != nil {
		return mcp.NewToolResultError("heading is required"), nil
	}

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", notePath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	body := RemoveFrontmatter(string(content))
	section := extractSection(body, heading)

	if section == "" {
		return mcp.NewToolResultError(fmt.Sprintf("Section not found: %s", heading)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("## %s\n\n%s", heading, section)), nil
}

// GetHeadingsHandler lists all headings in a note
func (v *Vault) GetHeadingsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", notePath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	body := RemoveFrontmatter(string(content))
	headings := extractHeadings(body)

	if len(headings) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No headings found in: %s", notePath)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Headings in %s\n\n", notePath)

	for _, h := range headings {
		prefix := strings.Repeat("#", h.Level)
		fmt.Fprintf(&sb, "- L%d–%d: `%s %s`\n", h.Line, h.EndLine, prefix, h.Text)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// SearchHeadingsHandler searches across all heading content
func (v *Vault) SearchHeadingsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query is required"), nil
	}

	level := int(req.GetInt("level", 0)) // 0 means all levels
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	queryLower := strings.ToLower(query)

	type headingMatch struct {
		path    string
		heading Heading
	}

	var matches []headingMatch

	err = filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)
		body := RemoveFrontmatter(string(content))
		headings := extractHeadings(body)

		for _, h := range headings {
			if level > 0 && h.Level != level {
				continue
			}
			if strings.Contains(strings.ToLower(h.Text), queryLower) {
				matches = append(matches, headingMatch{path: relPath, heading: h})
			}
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No headings matching '%s' found", query)), nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Headings matching '%s'\n\n", query)
	fmt.Fprintf(&sb, "Found %d matches:\n\n", len(matches))

	for _, m := range matches {
		prefix := strings.Repeat("#", m.heading.Level)
		fmt.Fprintf(&sb, "- **%s** L%d: `%s %s`\n", m.path, m.heading.Line, prefix, m.heading.Text)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// extractHeadings finds all headings in markdown content and computes section boundaries.
// Each heading's EndLine is the last line of its content (before the next same-or-higher level heading).
func extractHeadings(content string) []Heading {
	var headings []Heading
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			headings = append(headings, Heading{
				Level: len(matches[1]),
				Text:  strings.TrimSpace(matches[2]),
				Line:  i + 1,
			})
		}
	}

	// Compute EndLine for each heading: the line before the next same-or-higher level heading
	for i := range headings {
		endLine := totalLines // default: extends to end of file
		for j := i + 1; j < len(headings); j++ {
			if headings[j].Level <= headings[i].Level {
				endLine = headings[j].Line - 1
				break
			}
		}
		headings[i].EndLine = endLine
	}

	return headings
}

// extractSection extracts content under a heading until the next same-or-higher level heading
func extractSection(content, heading string) string {
	lines := strings.Split(content, "\n")

	var sectionStart, sectionEnd int
	var sectionLevel int
	inSection := false

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			level := len(matches[1])
			text := strings.TrimSpace(matches[2])

			if !inSection && strings.EqualFold(text, heading) {
				inSection = true
				sectionLevel = level
				sectionStart = i + 1
				continue
			}

			if inSection && level <= sectionLevel {
				sectionEnd = i
				break
			}
		}
	}

	if !inSection {
		return ""
	}

	if sectionEnd == 0 {
		sectionEnd = len(lines)
	}

	if sectionStart >= sectionEnd {
		return ""
	}

	return strings.TrimSpace(strings.Join(lines[sectionStart:sectionEnd], "\n"))
}

// RemoveFrontmatter strips YAML frontmatter from content
func RemoveFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return content
	}

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		}
	}

	return content
}
