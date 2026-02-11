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
	priorityRegex = regexp.MustCompile(`([⏫🔼🔽])`)
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
		Completed: strings.EqualFold(status, "x"),
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

// ToggleTaskHandler toggles a task's completion status
func (v *Vault) ToggleTaskHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	line, err := req.RequireInt("line")
	if err != nil {
		return mcp.NewToolResultError("line is required"), nil
	}

	if !strings.HasSuffix(path, ".md") {
		return mcp.NewToolResultError("path must end with .md"), nil
	}

	fullPath := filepath.Join(v.path, path)

	// Security: ensure path is within vault
	rel, err := filepath.Rel(v.path, filepath.Clean(fullPath))
	if err != nil || strings.HasPrefix(rel, "..") {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", path)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	lines := strings.Split(string(content), "\n")

	if line < 1 || line > len(lines) {
		return mcp.NewToolResultError(fmt.Sprintf("Line %d out of range (1-%d)", line, len(lines))), nil
	}

	targetLine := lines[line-1]
	task := ParseTask(targetLine, line)
	if task == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Line %d is not a task", line)), nil
	}

	// Toggle the checkbox
	var newLine string
	if task.Completed {
		// [x] or [X] -> [ ]
		newLine = strings.Replace(targetLine, "[x]", "[ ]", 1)
		newLine = strings.Replace(newLine, "[X]", "[ ]", 1)
	} else {
		// [ ] -> [x]
		newLine = strings.Replace(targetLine, "[ ]", "[x]", 1)
	}

	lines[line-1] = newLine

	if err := os.WriteFile(fullPath, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	newStatus := "completed"
	if task.Completed {
		newStatus = "open"
	}

	return mcp.NewToolResultText(fmt.Sprintf("Toggled task on line %d to %s: %s", line, newStatus, task.Text)), nil
}
