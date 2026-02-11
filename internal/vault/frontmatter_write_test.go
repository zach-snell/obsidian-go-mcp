package vault

import (
	"strings"
	"testing"
)

func TestSetFrontmatterKey(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		key      string
		value    string
		expected string
	}{
		{
			name:     "add to existing frontmatter",
			content:  "---\ntitle: Test\n---\n\n# Content",
			key:      "status",
			value:    "draft",
			expected: "---\ntitle: Test\nstatus: draft\n---\n\n# Content",
		},
		{
			name:     "update existing key",
			content:  "---\ntitle: Old Title\n---\n\n# Content",
			key:      "title",
			value:    "New Title",
			expected: "---\ntitle: New Title\n---\n\n# Content",
		},
		{
			name:     "create frontmatter if none",
			content:  "# Just Content\n\nBody text",
			key:      "title",
			value:    "My Note",
			expected: "---\ntitle: My Note\n---\n\n# Just Content\n\nBody text",
		},
		{
			name:     "key is lowercased",
			content:  "---\ntitle: Test\n---\n\n# Content",
			key:      "Status",
			value:    "done",
			expected: "---\ntitle: Test\nstatus: done\n---\n\n# Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := setFrontmatterKey(tt.content, tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestRemoveFrontmatterKey(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		key             string
		expectedContent string
		expectedRemoved bool
	}{
		{
			name:            "remove existing key",
			content:         "---\ntitle: Test\nstatus: draft\n---\n\n# Content",
			key:             "status",
			expectedContent: "---\ntitle: Test\n---\n\n# Content",
			expectedRemoved: true,
		},
		{
			name:            "key not found",
			content:         "---\ntitle: Test\n---\n\n# Content",
			key:             "nonexistent",
			expectedContent: "---\ntitle: Test\n---\n\n# Content",
			expectedRemoved: false,
		},
		{
			name:            "remove last key removes frontmatter",
			content:         "---\ntitle: Test\n---\n\n# Content",
			key:             "title",
			expectedContent: "# Content",
			expectedRemoved: true,
		},
		{
			name:            "no frontmatter",
			content:         "# Just Content",
			key:             "title",
			expectedContent: "# Just Content",
			expectedRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, removed := removeFrontmatterKey(tt.content, tt.key)
			if removed != tt.expectedRemoved {
				t.Errorf("expected removed=%v, got %v", tt.expectedRemoved, removed)
			}
			if result != tt.expectedContent {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expectedContent, result)
			}
		})
	}
}

func TestAddToFrontmatterArray(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		key      string
		value    string
		contains []string
	}{
		{
			name:     "add to existing array",
			content:  "---\ntags:\n  - one\n  - two\n---\n\n# Content",
			key:      "tags",
			value:    "three",
			contains: []string{"- one", "- two", "- three"},
		},
		{
			name:     "create array if key missing",
			content:  "---\ntitle: Test\n---\n\n# Content",
			key:      "tags",
			value:    "newtag",
			contains: []string{"tags:", "- newtag"},
		},
		{
			name:     "create frontmatter and array",
			content:  "# No frontmatter",
			key:      "aliases",
			value:    "MyAlias",
			contains: []string{"---", "aliases:", "- MyAlias"},
		},
		{
			name:     "convert inline array",
			content:  "---\ntags: [one, two]\n---\n\n# Content",
			key:      "tags",
			value:    "three",
			contains: []string{"- one", "- two", "- three"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addToFrontmatterArray(tt.content, tt.key, tt.value)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q\nresult:\n%s", expected, result)
				}
			}
		})
	}
}
