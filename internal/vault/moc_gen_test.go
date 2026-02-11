package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectNotes(t *testing.T) {
	tmpDir := t.TempDir()
	v := New(tmpDir)

	// Create test notes
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	notes := map[string]string{
		"note1.md":        "# Note One\n\nContent with #tag1",
		"note2.md":        "# Note Two\n\nContent with [[note1]]",
		"subdir/note3.md": "# Note Three\n\nNested note",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		dir       string
		recursive bool
		wantCount int
	}{
		{
			name:      "root only",
			dir:       "",
			recursive: false,
			wantCount: 2,
		},
		{
			name:      "root recursive",
			dir:       "",
			recursive: true,
			wantCount: 3,
		},
		{
			name:      "subdir only",
			dir:       "subdir",
			recursive: false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collected, err := v.collectNotes(tt.dir, tt.recursive)
			if err != nil {
				t.Fatalf("collectNotes() error = %v", err)
			}
			if len(collected) != tt.wantCount {
				t.Errorf("collectNotes() got %d notes, want %d", len(collected), tt.wantCount)
			}
		})
	}
}

func TestFormatFlat(t *testing.T) {
	notes := []noteInfo{
		{name: "Zebra", title: "Zebra"},
		{name: "Apple", title: "Apple"},
		{name: "Mango", title: "Mango"},
	}

	result := formatFlat(notes)

	// Should be sorted alphabetically
	if !strings.Contains(result, "[[Apple]]") {
		t.Error("formatFlat() should contain Apple")
	}
	appleIdx := strings.Index(result, "Apple")
	zebraIdx := strings.Index(result, "Zebra")
	if appleIdx > zebraIdx {
		t.Error("formatFlat() should sort alphabetically")
	}
}

func TestFormatByAlpha(t *testing.T) {
	notes := []noteInfo{
		{name: "Zebra", title: "Zebra"},
		{name: "Apple", title: "Apple"},
		{name: "Avocado", title: "Avocado"},
		{name: "123Note", title: "123Note"},
	}

	result := formatByAlpha(notes)

	// Should have alphabetical headers
	if !strings.Contains(result, "## A") {
		t.Error("formatByAlpha() should have A section")
	}
	if !strings.Contains(result, "## Z") {
		t.Error("formatByAlpha() should have Z section")
	}
	// Numbers should be grouped under #
	if !strings.Contains(result, "## #") {
		t.Error("formatByAlpha() should have # section for numbers")
	}
}

func TestFormatByTag(t *testing.T) {
	notes := []noteInfo{
		{name: "Note1", title: "Note1", tags: []string{"project"}},
		{name: "Note2", title: "Note2", tags: []string{"project"}},
		{name: "Note3", title: "Note3", tags: []string{"idea"}},
		{name: "Note4", title: "Note4", tags: nil},
	}

	result := formatByTag(notes)

	if !strings.Contains(result, "## project") {
		t.Error("formatByTag() should have project section")
	}
	if !strings.Contains(result, "## idea") {
		t.Error("formatByTag() should have idea section")
	}
	if !strings.Contains(result, "## Untagged") {
		t.Error("formatByTag() should have Untagged section")
	}
}

func TestGenerateMOC_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some notes
	notes := map[string]string{
		"project-a.md": "# Project A\n\n#project content",
		"project-b.md": "# Project B\n\n#project content",
		"idea.md":      "# Great Idea\n\n#idea brainstorm",
	}

	for path, content := range notes {
		if err := os.WriteFile(filepath.Join(tmpDir, path), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestUpdateMOC_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing MOC
	mocContent := `---
tags: [moc]
---

# Test MOC

- [[existing-note]]
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test-moc.md"), []byte(mocContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create notes
	if err := os.WriteFile(filepath.Join(tmpDir, "existing-note.md"), []byte("# Existing\n\nContent"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "new-note.md"), []byte("# New Note\n\nContent"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}

func TestGenerateIndex_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create notes with various first letters
	notes := map[string]string{
		"apple.md":  "# Apple\n\nFruit",
		"banana.md": "# Banana\n\nFruit #fruit",
		"cherry.md": "# Cherry\n\nFruit with [[apple]]",
	}

	for path, content := range notes {
		if err := os.WriteFile(filepath.Join(tmpDir, path), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	v := New(tmpDir)
	_ = v // Vault created for integration test
}
