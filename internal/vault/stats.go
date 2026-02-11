package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/mark3labs/mcp-go/mcp"
)

// vaultStats holds all collected statistics
type vaultStats struct {
	noteCount      int
	totalWords     int
	totalChars     int
	totalLines     int
	totalTasks     int
	completedTasks int
	totalTags      map[string]int
	totalLinks     int
	folders        map[string]bool
}

// processNoteContent extracts statistics from a note's content
func (s *vaultStats) processNoteContent(content string) {
	lines := strings.Split(content, "\n")
	s.totalLines += len(lines)
	s.totalChars += utf8.RuneCountInString(content)
	s.totalWords += len(strings.Fields(content))

	// Count tasks
	for _, line := range lines {
		if task := ParseTask(line, 0); task != nil {
			s.totalTasks++
			if task.Completed {
				s.completedTasks++
			}
		}
	}

	// Count tags
	for _, tag := range ExtractTags(content) {
		s.totalTags[tag]++
	}

	// Count wikilinks
	s.totalLinks += len(ExtractWikilinks(content))
}

// formatStats builds the markdown output for vault statistics
func (s *vaultStats) formatStats(dir string) string {
	var sb strings.Builder
	sb.WriteString("# Vault Statistics\n\n")

	if dir != "" {
		fmt.Fprintf(&sb, "**Directory:** %s\n\n", dir)
	}

	sb.WriteString("## Overview\n")
	fmt.Fprintf(&sb, "- **Notes:** %d\n", s.noteCount)
	fmt.Fprintf(&sb, "- **Folders:** %d\n", len(s.folders))
	fmt.Fprintf(&sb, "- **Words:** %d\n", s.totalWords)
	fmt.Fprintf(&sb, "- **Characters:** %d\n", s.totalChars)
	fmt.Fprintf(&sb, "- **Lines:** %d\n", s.totalLines)
	fmt.Fprintf(&sb, "- **Internal Links:** %d\n", s.totalLinks)

	s.formatTasks(&sb)
	s.formatTags(&sb)
	s.formatAverages(&sb)

	return sb.String()
}

// formatTasks writes task statistics to the builder
func (s *vaultStats) formatTasks(sb *strings.Builder) {
	if s.totalTasks == 0 {
		return
	}
	sb.WriteString("\n## Tasks\n")
	fmt.Fprintf(sb, "- **Total:** %d\n", s.totalTasks)
	fmt.Fprintf(sb, "- **Completed:** %d\n", s.completedTasks)
	fmt.Fprintf(sb, "- **Open:** %d\n", s.totalTasks-s.completedTasks)
	pct := float64(s.completedTasks) / float64(s.totalTasks) * 100
	fmt.Fprintf(sb, "- **Completion:** %.1f%%\n", pct)
}

// formatTags writes tag statistics to the builder
func (s *vaultStats) formatTags(sb *strings.Builder) {
	if len(s.totalTags) == 0 {
		return
	}
	fmt.Fprintf(sb, "\n## Top Tags (%d unique)\n", len(s.totalTags))

	// Sort tags by count descending
	type tagCount struct {
		tag   string
		count int
	}
	sortedTags := make([]tagCount, 0, len(s.totalTags))
	for tag, count := range s.totalTags {
		sortedTags = append(sortedTags, tagCount{tag, count})
	}
	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].count > sortedTags[j].count
	})

	limit := 10
	if len(sortedTags) < limit {
		limit = len(sortedTags)
	}
	for i := 0; i < limit; i++ {
		fmt.Fprintf(sb, "- #%s (%d)\n", sortedTags[i].tag, sortedTags[i].count)
	}
}

// formatAverages writes average statistics to the builder
func (s *vaultStats) formatAverages(sb *strings.Builder) {
	if s.noteCount == 0 {
		return
	}
	sb.WriteString("\n## Averages\n")
	fmt.Fprintf(sb, "- **Words/note:** %d\n", s.totalWords/s.noteCount)
	fmt.Fprintf(sb, "- **Links/note:** %.1f\n", float64(s.totalLinks)/float64(s.noteCount))
}

// VaultStatsHandler returns statistics about the vault
func (v *Vault) VaultStatsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	stats := &vaultStats{
		totalTags: make(map[string]int),
		folders:   make(map[string]bool),
	}

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if path != searchPath {
				relPath, _ := filepath.Rel(v.path, path)
				stats.folders[relPath] = true
			}
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		stats.noteCount++

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		stats.processNoteContent(string(content))
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to gather stats: %v", err)), nil
	}

	return mcp.NewToolResultText(stats.formatStats(dir)), nil
}
