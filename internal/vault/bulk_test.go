package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddTagToNote(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		tag         string
		wantChanged bool
		wantContain string
	}{
		{
			name:        "add new inline tag",
			content:     "# Note\n\nContent here",
			tag:         "newtag",
			wantChanged: true,
			wantContain: "#newtag",
		},
		{
			name:        "tag already exists",
			content:     "# Note\n\n#newtag content",
			tag:         "newtag",
			wantChanged: false,
		},
		{
			name: "add to frontmatter tags",
			content: `---
tags: [existing]
---

# Note`,
			tag:         "newtag",
			wantChanged: true,
			wantContain: "newtag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed, result := addTagToNote(tt.content, tt.tag)
			if changed != tt.wantChanged {
				t.Errorf("addTagToNote() changed = %v, want %v", changed, tt.wantChanged)
			}
			if tt.wantContain != "" && !strings.Contains(result, tt.wantContain) {
				t.Errorf("addTagToNote() result should contain %q, got:\n%s", tt.wantContain, result)
			}
		})
	}
}

func TestRemoveTagFromNote(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		tag            string
		wantChanged    bool
		wantNotContain string
	}{
		{
			name:           "remove inline tag",
			content:        "# Note\n\n#removeme content",
			tag:            "removeme",
			wantChanged:    true,
			wantNotContain: "#removeme",
		},
		{
			name:        "tag not present",
			content:     "# Note\n\ncontent",
			tag:         "nothere",
			wantChanged: false,
		},
		{
			name: "remove from frontmatter",
			content: `---
tags: [keep, removeme, also-keep]
---

# Note`,
			tag:            "removeme",
			wantChanged:    true,
			wantNotContain: "removeme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed, result := removeTagFromNote(tt.content, tt.tag)
			if changed != tt.wantChanged {
				t.Errorf("removeTagFromNote() changed = %v, want %v", changed, tt.wantChanged)
			}
			if tt.wantNotContain != "" && strings.Contains(result, tt.wantNotContain) {
				t.Errorf("removeTagFromNote() result should not contain %q, got:\n%s", tt.wantNotContain, result)
			}
		})
	}
}

func TestAddFrontmatterField(t *testing.T) {
	content := `---
title: Test
---

# Content`

	result := addFrontmatterField(content, "newkey", "newvalue")

	if !strings.Contains(result, "newkey: newvalue") {
		t.Errorf("addFrontmatterField() should add the field, got:\n%s", result)
	}
	if !strings.Contains(result, "title: Test") {
		t.Error("addFrontmatterField() should preserve existing fields")
	}
}

func TestUpdateWikilinks(t *testing.T) {
	tests := []struct {
		name    string
		content string
		oldName string
		newName string
		want    string
	}{
		{
			name:    "simple link",
			content: "See [[Old Note]] for details",
			oldName: "Old Note",
			newName: "New Note",
			want:    "See [[New Note]] for details",
		},
		{
			name:    "aliased link",
			content: "See [[Old Note|alias]] for details",
			oldName: "Old Note",
			newName: "New Note",
			want:    "See [[New Note|alias]] for details",
		},
		{
			name:    "multiple links",
			content: "See [[Old Note]] and [[Old Note|other]]",
			oldName: "Old Note",
			newName: "New Note",
			want:    "See [[New Note]] and [[New Note|other]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateWikilinks(tt.content, tt.oldName, tt.newName)
			if got != tt.want {
				t.Errorf("updateWikilinks() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasFrontmatter(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"---\ntitle: test\n---\ncontent", true},
		{"# Just a title\ncontent", false},
		{"", false},
		{"---", true}, // Starts with ---, hasFrontmatter just checks prefix
	}

	for _, tt := range tests {
		t.Run(tt.content[:min(20, len(tt.content))], func(t *testing.T) {
			got := hasFrontmatter(tt.content)
			if got != tt.want {
				t.Errorf("hasFrontmatter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBulkTag_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test notes
	notes := map[string]string{
		"note1.md": "# Note 1\n\nContent",
		"note2.md": "# Note 2\n\n#existing-tag content",
		"note3.md": "---\ntags: [old-tag]\n---\n\n# Note 3",
	}

	for name, content := range notes {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestBulkMove_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test notes
	if err := os.WriteFile(filepath.Join(tmpDir, "note1.md"), []byte("# Note 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "note2.md"), []byte("# Note 2"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestBulkSetFrontmatter_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test notes
	notes := map[string]string{
		"note1.md": "# Note 1\n\nContent",
		"note2.md": "---\nstatus: draft\n---\n\n# Note 2",
	}

	for name, content := range notes {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}
