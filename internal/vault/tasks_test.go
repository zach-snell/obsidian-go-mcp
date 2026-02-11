package vault

import (
	"testing"
)

func TestParseTask(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		lineNum   int
		wantNil   bool
		completed bool
		text      string
		dueDate   string
		priority  string
		tags      []string
	}{
		{
			name:      "simple open task",
			line:      "- [ ] Do something",
			lineNum:   1,
			completed: false,
			text:      "Do something",
		},
		{
			name:      "simple completed task",
			line:      "- [x] Done task",
			lineNum:   5,
			completed: true,
			text:      "Done task",
		},
		{
			name:      "completed task uppercase X",
			line:      "- [X] Also done",
			lineNum:   1,
			completed: true,
			text:      "Also done",
		},
		{
			name:      "task with due date",
			line:      "- [ ] Submit report 📅 2024-01-15",
			lineNum:   1,
			completed: false,
			text:      "Submit report 📅 2024-01-15",
			dueDate:   "2024-01-15",
		},
		{
			name:      "task with high priority",
			line:      "- [ ] Urgent task ⏫",
			lineNum:   1,
			completed: false,
			text:      "Urgent task ⏫",
			priority:  "high",
		},
		{
			name:      "task with medium priority",
			line:      "- [ ] Normal task 🔼",
			lineNum:   1,
			completed: false,
			text:      "Normal task 🔼",
			priority:  "medium",
		},
		{
			name:      "task with low priority",
			line:      "- [ ] Low priority task 🔽",
			lineNum:   1,
			completed: false,
			text:      "Low priority task 🔽",
			priority:  "low",
		},
		{
			name:      "task with single tag",
			line:      "- [ ] Fix bug #bugfix",
			lineNum:   1,
			completed: false,
			text:      "Fix bug #bugfix",
			tags:      []string{"bugfix"},
		},
		{
			name:      "task with multiple tags",
			line:      "- [ ] Review code #review #urgent #backend",
			lineNum:   1,
			completed: false,
			text:      "Review code #review #urgent #backend",
			tags:      []string{"review", "urgent", "backend"},
		},
		{
			name:      "task with everything",
			line:      "- [ ] Complete project #work #deadline 📅 2024-12-31 ⏫",
			lineNum:   42,
			completed: false,
			text:      "Complete project #work #deadline 📅 2024-12-31 ⏫",
			dueDate:   "2024-12-31",
			priority:  "high",
			tags:      []string{"work", "deadline"},
		},
		{
			name:      "indented task",
			line:      "  - [ ] Subtask",
			lineNum:   1,
			completed: false,
			text:      "Subtask",
		},
		{
			name:      "deeply indented task",
			line:      "      - [x] Deep subtask",
			lineNum:   1,
			completed: true,
			text:      "Deep subtask",
		},
		{
			name:    "not a task - regular list",
			line:    "- Regular list item",
			lineNum: 1,
			wantNil: true,
		},
		{
			name:    "not a task - plain text",
			line:    "Just some text",
			lineNum: 1,
			wantNil: true,
		},
		{
			name:    "not a task - empty line",
			line:    "",
			lineNum: 1,
			wantNil: true,
		},
		{
			name:    "not a task - heading",
			line:    "## Tasks",
			lineNum: 1,
			wantNil: true,
		},
		{
			name:    "not a task - malformed checkbox",
			line:    "- [] Missing space",
			lineNum: 1,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := ParseTask(tt.line, tt.lineNum)

			if tt.wantNil {
				if task != nil {
					t.Errorf("ParseTask() = %+v, want nil", task)
				}
				return
			}

			if task == nil {
				t.Fatalf("ParseTask() = nil, want non-nil")
			}

			if task.Line != tt.lineNum {
				t.Errorf("Line = %d, want %d", task.Line, tt.lineNum)
			}

			if task.Completed != tt.completed {
				t.Errorf("Completed = %v, want %v", task.Completed, tt.completed)
			}

			if task.Text != tt.text {
				t.Errorf("Text = %q, want %q", task.Text, tt.text)
			}

			// Check due date
			if tt.dueDate != "" {
				if task.DueDate == nil {
					t.Errorf("DueDate = nil, want %q", tt.dueDate)
				} else if *task.DueDate != tt.dueDate {
					t.Errorf("DueDate = %q, want %q", *task.DueDate, tt.dueDate)
				}
			} else if task.DueDate != nil {
				t.Errorf("DueDate = %q, want nil", *task.DueDate)
			}

			// Check priority
			if tt.priority != "" {
				if task.Priority == nil {
					t.Errorf("Priority = nil, want %q", tt.priority)
				} else if *task.Priority != tt.priority {
					t.Errorf("Priority = %q, want %q", *task.Priority, tt.priority)
				}
			} else if task.Priority != nil {
				t.Errorf("Priority = %q, want nil", *task.Priority)
			}

			// Check tags
			if len(tt.tags) > 0 {
				if len(task.Tags) != len(tt.tags) {
					t.Errorf("Tags count = %d, want %d", len(task.Tags), len(tt.tags))
				} else {
					for i, tag := range tt.tags {
						if task.Tags[i] != tag {
							t.Errorf("Tags[%d] = %q, want %q", i, task.Tags[i], tag)
						}
					}
				}
			} else if len(task.Tags) != 0 {
				t.Errorf("Tags = %v, want empty", task.Tags)
			}
		})
	}
}

func TestParseTask_TagFormats(t *testing.T) {
	// Test various valid tag formats
	validTags := []struct {
		line string
		tag  string
	}{
		{"- [ ] Test #simple", "simple"},
		{"- [ ] Test #with-dash", "with-dash"},
		{"- [ ] Test #with_underscore", "with_underscore"},
		{"- [ ] Test #MixedCase", "MixedCase"},
		{"- [ ] Test #has123numbers", "has123numbers"},
	}

	for _, tt := range validTags {
		t.Run(tt.tag, func(t *testing.T) {
			task := ParseTask(tt.line, 1)
			if task == nil {
				t.Fatal("ParseTask() = nil, want non-nil")
			}
			if len(task.Tags) != 1 || task.Tags[0] != tt.tag {
				t.Errorf("Tags = %v, want [%s]", task.Tags, tt.tag)
			}
		})
	}
}
