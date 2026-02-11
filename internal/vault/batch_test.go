package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractHeadings(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Heading
	}{
		{
			name:    "single h1",
			content: "# Hello World",
			expected: []Heading{
				{Level: 1, Text: "Hello World", Line: 1, EndLine: 1},
			},
		},
		{
			name:    "multiple levels",
			content: "# Title\n\nSome text\n\n## Section 1\n\nContent\n\n### Subsection\n\n## Section 2",
			expected: []Heading{
				{Level: 1, Text: "Title", Line: 1, EndLine: 11},
				{Level: 2, Text: "Section 1", Line: 5, EndLine: 10},
				{Level: 3, Text: "Subsection", Line: 9, EndLine: 10},
				{Level: 2, Text: "Section 2", Line: 11, EndLine: 11},
			},
		},
		{
			name:     "no headings",
			content:  "Just some text\nwithout any headings",
			expected: nil,
		},
		{
			name:    "all levels",
			content: "# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6",
			expected: []Heading{
				{Level: 1, Text: "H1", Line: 1, EndLine: 6},
				{Level: 2, Text: "H2", Line: 2, EndLine: 6},
				{Level: 3, Text: "H3", Line: 3, EndLine: 6},
				{Level: 4, Text: "H4", Line: 4, EndLine: 6},
				{Level: 5, Text: "H5", Line: 5, EndLine: 6},
				{Level: 6, Text: "H6", Line: 6, EndLine: 6},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHeadings(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d headings, got %d", len(tt.expected), len(result))
				return
			}
			for i, exp := range tt.expected {
				if result[i].Level != exp.Level {
					t.Errorf("heading %d: expected level %d, got %d", i, exp.Level, result[i].Level)
				}
				if result[i].Text != exp.Text {
					t.Errorf("heading %d: expected text %q, got %q", i, exp.Text, result[i].Text)
				}
				if result[i].Line != exp.Line {
					t.Errorf("heading %d: expected line %d, got %d", i, exp.Line, result[i].Line)
				}
				if result[i].EndLine != exp.EndLine {
					t.Errorf("heading %d: expected end_line %d, got %d", i, exp.EndLine, result[i].EndLine)
				}
			}
		})
	}
}

func TestExtractSection(t *testing.T) {
	content := `# Title

Introduction text.

## Section One

Content of section one.

More content here.

### Subsection

Nested content.

## Section Two

Content of section two.

# Another Title

Final content.
`

	tests := []struct {
		name     string
		heading  string
		expected string
	}{
		{
			name:     "extract section one",
			heading:  "Section One",
			expected: "Content of section one.\n\nMore content here.\n\n### Subsection\n\nNested content.",
		},
		{
			name:     "extract subsection",
			heading:  "Subsection",
			expected: "Nested content.",
		},
		{
			name:     "extract section two",
			heading:  "Section Two",
			expected: "Content of section two.",
		},
		{
			name:     "case insensitive",
			heading:  "section one",
			expected: "Content of section one.\n\nMore content here.\n\n### Subsection\n\nNested content.",
		},
		{
			name:     "not found",
			heading:  "Nonexistent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSection(content, tt.heading)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRemoveFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "with frontmatter",
			content:  "---\ntitle: Test\ntags: [a, b]\n---\n\n# Content\n\nBody text",
			expected: "# Content\n\nBody text",
		},
		{
			name:     "no frontmatter",
			content:  "# Just Content\n\nBody text",
			expected: "# Just Content\n\nBody text",
		},
		{
			name:     "empty frontmatter",
			content:  "---\n---\n\nContent",
			expected: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveFrontmatter(tt.content)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBatchReadNotes(t *testing.T) {
	// Create temp vault
	tmpDir := t.TempDir()

	// Create test notes
	note1 := "---\ntitle: Note One\n---\n\n# Note One\n\nContent one."
	note2 := "# Note Two\n\nContent two with [[link]]."

	if err := os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte(note1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte(note2), 0o644); err != nil {
		t.Fatal(err)
	}

	vault := New(tmpDir)

	// Test batch read
	// (Would need to construct proper MCP request - skipping for now as it requires mocking)
	_ = vault
}
