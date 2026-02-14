package vault

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// helper to create a call tool request with params
func makeRequest(params map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: params,
		},
	}
}

func setupTestVault(t *testing.T) (v *Vault, dir string) {
	t.Helper()
	dir = t.TempDir()
	v = New(dir)
	return v, dir
}

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// --- truncateLine tests ---

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		contains string
	}{
		{"short line", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"long line", "abcdefghij", 5, "abcde... [10 chars total]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateLine(tt.input, tt.max)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected result to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

// --- buildEditContext tests ---

func TestBuildEditContext(t *testing.T) {
	lines := []string{"line1", "line2", "line3", "line4", "line5", "line6", "line7"}

	t.Run("zero context", func(t *testing.T) {
		result := buildEditContext(lines, 3, 4, 0, nil)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("basic context around edit", func(t *testing.T) {
		result := buildEditContext(lines, 3, 4, 2, nil)
		// Should show lines 2-3 before, line 4 (the edit), lines 5-6 after
		if !strings.Contains(result, "L2:") || !strings.Contains(result, "L4:") || !strings.Contains(result, "L6:") {
			t.Errorf("expected context lines around edit, got:\n%s", result)
		}
	})

	t.Run("context at start of file", func(t *testing.T) {
		result := buildEditContext(lines, 0, 1, 2, nil)
		if strings.Contains(result, "L-1:") {
			t.Error("should not have negative line numbers")
		}
		if !strings.Contains(result, "L1:") {
			t.Errorf("expected L1, got:\n%s", result)
		}
	})

	t.Run("context at end of file", func(t *testing.T) {
		result := buildEditContext(lines, 6, 7, 2, nil)
		if !strings.Contains(result, "L5:") {
			t.Errorf("expected L5 as before context, got:\n%s", result)
		}
	})

	t.Run("with inserted lines label", func(t *testing.T) {
		inserted := []string{"new content"}
		result := buildEditContext(lines, 3, 3, 2, inserted)
		if !strings.Contains(result, "INSERTED") {
			t.Errorf("expected INSERTED label, got:\n%s", result)
		}
	})

	t.Run("with changed lines label", func(t *testing.T) {
		changed := []string{"replaced"}
		result := buildEditContext(lines, 3, 4, 2, changed)
		if !strings.Contains(result, "CHANGED") {
			t.Errorf("expected CHANGED label, got:\n%s", result)
		}
	})
}

// --- writeContentPreview tests ---

func TestWriteContentPreview(t *testing.T) {
	t.Run("short content shown fully", func(t *testing.T) {
		var sb strings.Builder
		lines := []string{"a", "b", "c"}
		writeContentPreview(&sb, lines, 10, "INSERTED")
		result := sb.String()
		if !strings.Contains(result, "a") || !strings.Contains(result, "b") || !strings.Contains(result, "c") {
			t.Errorf("expected all lines, got:\n%s", result)
		}
		if !strings.Contains(result, "INSERTED") {
			t.Errorf("expected INSERTED label, got:\n%s", result)
		}
	})

	t.Run("long content summarized", func(t *testing.T) {
		var sb strings.Builder
		lines := make([]string, 20)
		for i := range lines {
			lines[i] = "line content"
		}
		writeContentPreview(&sb, lines, 0, "CHANGED")
		result := sb.String()
		if !strings.Contains(result, "more lines") {
			t.Errorf("expected summary for long content, got:\n%s", result)
		}
	})
}

// --- EditNoteHandler tests ---

func TestEditNoteHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("basic find and replace", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "Hello world\nThis is a test\nGoodbye world")

		req := makeRequest(map[string]any{
			"path":     "test.md",
			"old_text": "This is a test",
			"new_text": "This is REPLACED",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		if !strings.Contains(content, "This is REPLACED") {
			t.Errorf("expected replacement in file, got:\n%s", content)
		}
		if strings.Contains(content, "This is a test") {
			t.Error("old text should be gone")
		}
	})

	t.Run("multiple matches without replace_all fails", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "foo bar\nfoo baz\nfoo qux")

		req := makeRequest(map[string]any{
			"path":     "test.md",
			"old_text": "foo",
			"new_text": "replaced",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for multiple matches without replace_all")
		}
	})

	t.Run("multiple matches with replace_all succeeds", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "foo bar\nfoo baz\nfoo qux")

		req := makeRequest(map[string]any{
			"path":        "test.md",
			"old_text":    "foo",
			"new_text":    "replaced",
			"replace_all": true,
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		if strings.Contains(content, "foo") {
			t.Error("all occurrences should be replaced")
		}
		if strings.Count(content, "replaced") != 3 {
			t.Errorf("expected 3 replacements, got %d", strings.Count(content, "replaced"))
		}
	})

	t.Run("old_text not found", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "Hello world")

		req := makeRequest(map[string]any{
			"path":     "test.md",
			"old_text": "nonexistent",
			"new_text": "replaced",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for not found")
		}
	})

	t.Run("note not found", func(t *testing.T) {
		v, _ := setupTestVault(t)

		req := makeRequest(map[string]any{
			"path":     "nonexistent.md",
			"old_text": "hello",
			"new_text": "world",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for missing file")
		}
	})

	t.Run("with context_lines", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "line1\nline2\nline3\nline4\nline5")

		req := makeRequest(map[string]any{
			"path":          "test.md",
			"old_text":      "line3",
			"new_text":      "REPLACED",
			"context_lines": float64(2),
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Context") {
			t.Errorf("expected context output, got:\n%s", text)
		}
	})

	t.Run("path safety", func(t *testing.T) {
		v, _ := setupTestVault(t)

		req := makeRequest(map[string]any{
			"path":     "../../etc/passwd",
			"old_text": "root",
			"new_text": "hacked",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for path traversal")
		}
	})

	t.Run("auto-adds .md extension", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "Hello world")

		req := makeRequest(map[string]any{
			"path":     "test",
			"old_text": "Hello",
			"new_text": "Hi",
		})

		result, err := v.EditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		content := readTestFile(t, dir, "test.md")
		if !strings.Contains(content, "Hi world") {
			t.Errorf("expected 'Hi world', got:\n%s", content)
		}
	})
}

// --- ReplaceSectionHandler tests ---

func TestReplaceSectionHandler(t *testing.T) {
	ctx := context.Background()

	const testDoc = `# Title

Introduction.

## Installation

Old install instructions.

More old stuff.

## Usage

Use it like this.

### Advanced

Advanced usage.

## FAQ

Questions.
`

	t.Run("replace middle section", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "doc.md",
			"heading": "Installation",
			"content": "New install instructions.\n\n```bash\nnpm install\n```",
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "New install instructions.") {
			t.Error("new content not found")
		}
		if strings.Contains(content, "Old install instructions.") {
			t.Error("old content should be gone")
		}
		// Heading should be preserved
		if !strings.Contains(content, "## Installation") {
			t.Error("heading should be preserved")
		}
		// Next section should be untouched
		if !strings.Contains(content, "## Usage") {
			t.Error("next section should be untouched")
		}
		if !strings.Contains(content, "Use it like this.") {
			t.Error("next section content should be untouched")
		}
	})

	t.Run("replace last section", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "doc.md",
			"heading": "FAQ",
			"content": "No questions yet.",
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "No questions yet.") {
			t.Error("new FAQ content not found")
		}
		if strings.Contains(content, "Questions.") {
			t.Error("old FAQ content should be gone")
		}
	})

	t.Run("case insensitive heading match", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "doc.md",
			"heading": "installation",
			"content": "Replaced.",
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "Replaced.") {
			t.Error("case-insensitive match should work")
		}
	})

	t.Run("heading not found", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "doc.md",
			"heading": "Nonexistent",
			"content": "content",
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for missing heading")
		}
	})

	t.Run("replaces subsection correctly", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "doc.md",
			"heading": "Advanced",
			"content": "New advanced content.",
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "New advanced content.") {
			t.Error("subsection content not replaced")
		}
		if strings.Contains(content, "Advanced usage.") {
			t.Error("old subsection content should be gone")
		}
		// Parent section content should be untouched
		if !strings.Contains(content, "Use it like this.") {
			t.Error("parent section content should be preserved")
		}
	})

	t.Run("with context_lines", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", testDoc)

		req := makeRequest(map[string]any{
			"path":          "doc.md",
			"heading":       "FAQ",
			"content":       "New FAQ.",
			"context_lines": float64(2),
		})

		result, err := v.ReplaceSectionHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Context") {
			t.Errorf("expected context output, got:\n%s", text)
		}
	})
}

// --- Enhanced AppendNoteHandler tests ---

func TestAppendNoteHandler_Enhanced(t *testing.T) {
	ctx := context.Background()

	const testDoc = "# Title\n\nIntro text.\n\n## Section One\n\nContent one.\n\n## Section Two\n\nContent two.\n"

	t.Run("default append to end (backward compat)", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "existing content")

		req := makeRequest(map[string]any{
			"path":    "test.md",
			"content": "appended",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		if !strings.HasSuffix(content, "appended") {
			t.Errorf("expected content appended at end, got:\n%s", content)
		}
	})

	t.Run("append to start", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "existing content")

		req := makeRequest(map[string]any{
			"path":     "test.md",
			"content":  "prepended",
			"position": "start",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		if !strings.HasPrefix(content, "prepended") {
			t.Errorf("expected content at start, got:\n%s", content)
		}
	})

	t.Run("insert after heading", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "test.md",
			"content": "Inserted after heading.",
			"after":   "Section One",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		lines := strings.Split(content, "\n")
		// Find "## Section One" and check next line is our insert
		for i, line := range lines {
			if strings.Contains(line, "## Section One") {
				if i+1 < len(lines) && lines[i+1] != "Inserted after heading." {
					t.Errorf("expected insert after heading, next line is: %q", lines[i+1])
				}
				break
			}
		}
	})

	t.Run("insert before heading", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", testDoc)

		req := makeRequest(map[string]any{
			"path":    "test.md",
			"content": "Inserted before heading.",
			"before":  "Section Two",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "test.md")
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(line, "## Section Two") {
				if i > 0 && lines[i-1] != "Inserted before heading." {
					t.Errorf("expected insert before heading, previous line is: %q", lines[i-1])
				}
				break
			}
		}
	})

	t.Run("ambiguous target returns error", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "foo\nbar\nfoo\n")

		req := makeRequest(map[string]any{
			"path":    "test.md",
			"content": "inserted",
			"after":   "foo",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for ambiguous target")
		}
	})

	t.Run("target not found returns error", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", "some content")

		req := makeRequest(map[string]any{
			"path":    "test.md",
			"content": "inserted",
			"after":   "nonexistent",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Error("expected error for target not found")
		}
	})

	t.Run("creates file if not exists", func(t *testing.T) {
		v, dir := setupTestVault(t)

		req := makeRequest(map[string]any{
			"path":    "new-file.md",
			"content": "brand new content",
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "new-file.md")
		if content != "brand new content" {
			t.Errorf("expected 'brand new content', got: %q", content)
		}
	})

	t.Run("with context_lines", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "test.md", testDoc)

		req := makeRequest(map[string]any{
			"path":          "test.md",
			"content":       "Inserted.",
			"after":         "Section One",
			"context_lines": float64(2),
		})

		result, err := v.AppendNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Context") {
			t.Errorf("expected context output, got:\n%s", text)
		}
	})
}

// --- findTargetLine tests ---

func TestFindTargetLine(t *testing.T) {
	lines := []string{
		"# Title",
		"",
		"Some content",
		"",
		"## Section One",
		"",
		"Content of section one.",
		"",
		"## Section Two",
		"",
		"Unique text here",
	}

	t.Run("finds heading by name", func(t *testing.T) {
		idx, err := findTargetLine(lines, "Section One")
		if err != nil {
			t.Fatal(err)
		}
		if idx != 4 {
			t.Errorf("expected index 4, got %d", idx)
		}
	})

	t.Run("heading match is case-insensitive", func(t *testing.T) {
		idx, err := findTargetLine(lines, "section one")
		if err != nil {
			t.Fatal(err)
		}
		if idx != 4 {
			t.Errorf("expected index 4, got %d", idx)
		}
	})

	t.Run("falls back to text match", func(t *testing.T) {
		idx, err := findTargetLine(lines, "Unique text here")
		if err != nil {
			t.Fatal(err)
		}
		if idx != 10 {
			t.Errorf("expected index 10, got %d", idx)
		}
	})

	t.Run("ambiguous text match errors", func(t *testing.T) {
		// Empty string appears in every line via Contains
		ambiguousLines := []string{"alpha foo", "beta foo", "gamma"}
		_, err := findTargetLine(ambiguousLines, "foo")
		if err == nil {
			t.Fatal("expected error for ambiguous match")
		}
		if !strings.Contains(err.Error(), "ambiguous") {
			t.Errorf("expected ambiguous error, got: %v", err)
		}
	})

	t.Run("not found errors", func(t *testing.T) {
		_, err := findTargetLine(lines, "Nonexistent")
		if err == nil {
			t.Fatal("expected error for not found")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("expected not found error, got: %v", err)
		}
	})
}

// --- BatchEditNoteHandler tests ---

func TestBatchEditNoteHandler(t *testing.T) {
	ctx := context.Background()

	t.Run("basic batch edit", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "## 1. Alpha\n\nContent.\n\n## 2. Beta\n\nContent.\n\n## 3. Gamma\n")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[{"old_text": "## 2. Beta", "new_text": "## 3. Beta"}, {"old_text": "## 3. Gamma", "new_text": "## 4. Gamma"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "## 3. Beta") {
			t.Error("expected '## 3. Beta'")
		}
		if !strings.Contains(content, "## 4. Gamma") {
			t.Error("expected '## 4. Gamma'")
		}
		if strings.Contains(content, "## 2. Beta") {
			t.Error("old '## 2. Beta' should be gone")
		}
		if !strings.Contains(content, "## 1. Alpha") {
			t.Error("untouched heading should remain")
		}
	})

	t.Run("all or nothing - one not found fails entire batch", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "Hello world\nGoodbye world\n")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[{"old_text": "Hello", "new_text": "Hi"}, {"old_text": "NONEXISTENT", "new_text": "nope"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error when edit not found")
		}

		// Verify nothing was changed
		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "Hello world") {
			t.Error("file should be unchanged on validation failure")
		}
	})

	t.Run("duplicate old_text fails", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "foo bar\nfoo baz\n")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[{"old_text": "foo", "new_text": "replaced"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for ambiguous old_text")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "found 2 times") {
			t.Errorf("expected match count in error, got: %s", text)
		}
	})

	t.Run("empty edits array fails", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "content")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for empty edits")
		}
	})

	t.Run("invalid JSON fails", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "content")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `not json`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("overlapping edits fail", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "abcdefgh\n")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[{"old_text": "abcdef", "new_text": "X"}, {"old_text": "defgh", "new_text": "Y"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for overlapping edits")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "overlap") {
			t.Errorf("expected overlap error, got: %s", text)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		v, _ := setupTestVault(t)

		req := makeRequest(map[string]any{
			"path":  "nonexistent.md",
			"edits": `[{"old_text": "a", "new_text": "b"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("path safety", func(t *testing.T) {
		v, _ := setupTestVault(t)

		req := makeRequest(map[string]any{
			"path":  "../../etc/passwd",
			"edits": `[{"old_text": "root", "new_text": "hacked"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for path traversal")
		}
	})

	t.Run("with context_lines", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "line1\nline2\nline3\nline4\nline5\n")

		req := makeRequest(map[string]any{
			"path":          "doc.md",
			"edits":         `[{"old_text": "line2", "new_text": "CHANGED2"}, {"old_text": "line4", "new_text": "CHANGED4"}]`,
			"context_lines": float64(1),
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "Context") {
			t.Errorf("expected context output, got:\n%s", text)
		}
		if !strings.Contains(text, "Applied 2") {
			t.Errorf("expected 'Applied 2' in response, got:\n%s", text)
		}
	})

	t.Run("heading number cascade use case", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "guide.md", `# Guide

## 1. Introduction

Intro text.

## 2. Setup

Setup text.

## 3. Configuration

Config text.

## 4. Deployment

Deploy text.
`)

		// Simulate: inserted a new "## 2. Prerequisites" section, now need to bump 2→3, 3→4, 4→5
		req := makeRequest(map[string]any{
			"path": "guide.md",
			"edits": `[
				{"old_text": "## 2. Setup", "new_text": "## 3. Setup"},
				{"old_text": "## 3. Configuration", "new_text": "## 4. Configuration"},
				{"old_text": "## 4. Deployment", "new_text": "## 5. Deployment"}
			]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}

		content := readTestFile(t, dir, "guide.md")
		if !strings.Contains(content, "## 1. Introduction") {
			t.Error("unchanged heading should remain")
		}
		if !strings.Contains(content, "## 3. Setup") {
			t.Error("expected renumbered Setup")
		}
		if !strings.Contains(content, "## 4. Configuration") {
			t.Error("expected renumbered Configuration")
		}
		if !strings.Contains(content, "## 5. Deployment") {
			t.Error("expected renumbered Deployment")
		}
		// Content under headings should be preserved
		if !strings.Contains(content, "Setup text.") {
			t.Error("section content should be preserved")
		}
	})

	t.Run("empty old_text in one edit fails", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "content")

		req := makeRequest(map[string]any{
			"path":  "doc.md",
			"edits": `[{"old_text": "", "new_text": "something"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsError {
			t.Fatal("expected error for empty old_text")
		}
	})

	t.Run("auto-adds .md extension", func(t *testing.T) {
		v, dir := setupTestVault(t)
		writeTestFile(t, dir, "doc.md", "Hello world")

		req := makeRequest(map[string]any{
			"path":  "doc",
			"edits": `[{"old_text": "Hello", "new_text": "Hi"}]`,
		})

		result, err := v.BatchEditNoteHandler(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		content := readTestFile(t, dir, "doc.md")
		if !strings.Contains(content, "Hi world") {
			t.Errorf("expected 'Hi world', got: %s", content)
		}
	})
}

// --- validateBatchEdits tests ---

func TestValidateBatchEdits(t *testing.T) {
	t.Run("valid edits pass", func(t *testing.T) {
		located, errResult := validateBatchEdits("alpha beta gamma", []editEntry{
			{OldText: "alpha", NewText: "A"},
			{OldText: "gamma", NewText: "G"},
		}, "test.md")
		if errResult != nil {
			t.Fatalf("expected success, got error: %v", errResult)
		}
		if len(located) != 2 {
			t.Fatalf("expected 2 located edits, got %d", len(located))
		}
	})

	t.Run("reports multiple errors at once", func(t *testing.T) {
		_, errResult := validateBatchEdits("hello world", []editEntry{
			{OldText: "MISSING1", NewText: "A"},
			{OldText: "MISSING2", NewText: "B"},
		}, "test.md")
		if errResult == nil {
			t.Fatal("expected error")
		}
		text := errResult.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "MISSING1") || !strings.Contains(text, "MISSING2") {
			t.Errorf("expected both errors reported, got: %s", text)
		}
	})
}
