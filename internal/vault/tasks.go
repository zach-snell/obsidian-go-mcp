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

// Task represents a parsed task from markdown
type Task struct {
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Completed bool     `json:"completed"`
	Text      string   `json:"text"`
	DueDate   *string  `json:"dueDate,omitempty"`
	Priority  *string  `json:"priority,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

var (
	// Matches: - [ ] or - [x] or - [X]
	taskRegex = regexp.MustCompile(`^(\s*)-\s*\[([ xX])\]\s*(.+)$`)
	// Matches: 📅 2024-01-15
	dueDateRegex = regexp.MustCompile(`📅\s*(\d{4}-\d{2}-\d{2})`)
	// Matches: ⏫ (high), 🔼 (medium), 🔽 (low)
	priorityRegex = regexp.MustCompile(`(⏫|🔼|🔽)`)
	// Matches: #tag
	tagRegex = regexp.MustCompile(`#([a-zA-Z0-9_\-]+)`)
)

// ParseTask parses a single line into a Task if it matches
func ParseTask(line string, lineNum int) *Task {
	match := taskRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	status := match[2]
	text := strings.TrimSpace(match[3])

	task := &Task{
		Line:      lineNum,
		Completed: strings.ToLower(status) == "x",
		Text:      text,
	}

	// Extract due date
	if dateMatch := dueDateRegex.FindStringSubmatch(text); dateMatch != nil {
		task.DueDate = &dateMatch[1]
	}

	// Extract priority
	if prioMatch := priorityRegex.FindStringSubmatch(text); prioMatch != nil {
		var prio string
		switch prioMatch[1] {
		case "⏫":
			prio = "high"
		case "🔼":
			prio = "medium"
		case "🔽":
			prio = "low"
		}
		task.Priority = &prio
	}

	// Extract tags
	tagMatches := tagRegex.FindAllStringSubmatch(text, -1)
	for _, tm := range tagMatches {
		task.Tags = append(task.Tags, tm[1])
	}

	return task
}

// ListTasksHandler lists all tasks across the vault
func (v *Vault) ListTasksHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := req.GetString("status", "all")
	dir := req.GetString("directory", "")

	searchPath := v.path
	if dir != "" {
		searchPath = filepath.Join(v.path, dir)
	}

	var tasks []Task

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

		lines := strings.Split(string(content), "\n")
		relPath, _ := filepath.Rel(v.path, path)

		for i, line := range lines {
			if task := ParseTask(line, i+1); task != nil {
				task.File = relPath

				// Filter by status
				switch status {
				case "open":
					if task.Completed {
						continue
					}
				case "completed":
					if !task.Completed {
						continue
					}
				}

				tasks = append(tasks, *task)
			}
		}
		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks: %v", err)), nil
	}

	if len(tasks) == 0 {
		return mcp.NewToolResultText("No tasks found"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d tasks:\n\n", len(tasks)))

	currentFile := ""
	for _, t := range tasks {
		if t.File != currentFile {
			if currentFile != "" {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("## %s\n", t.File))
			currentFile = t.File
		}

		checkbox := "[ ]"
		if t.Completed {
			checkbox = "[x]"
		}

		sb.WriteString(fmt.Sprintf("  L%d: - %s %s", t.Line, checkbox, t.Text))

		if t.Priority != nil {
			sb.WriteString(fmt.Sprintf(" [%s]", *t.Priority))
		}
		if t.DueDate != nil {
			sb.WriteString(fmt.Sprintf(" (due: %s)", *t.DueDate))
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
