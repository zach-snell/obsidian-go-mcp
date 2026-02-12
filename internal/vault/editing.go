package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// maxContextLineChars is the maximum characters per line in context output
	maxContextLineChars = 200
	// maxContextContentPreview is the max lines to show from each edge of inserted/replaced content
	maxContextContentPreview = 2
)

// truncateLine truncates a line to maxChars, adding an indicator if truncated
func truncateLine(line string, maxChars int) string {
	if len(line) <= maxChars {
		return line
	}
	return line[:maxChars] + fmt.Sprintf("... [%d chars total]", len(line))
}

// buildEditContext builds a context snippet around an edit location.
// editStart and editEnd are 0-based line indices of the edited region.
// insertedLines are the lines that were inserted/replaced (may be nil).
// allLines are the full file lines AFTER the edit.
// contextN is the number of surrounding lines to show.
func buildEditContext(allLines []string, editStart, editEnd, contextN int, insertedLines []string) string {
	if contextN <= 0 {
		return ""
	}

	var sb strings.Builder

	// Lines before the edit
	beforeStart := editStart - contextN
	if beforeStart < 0 {
		beforeStart = 0
	}
	for i := beforeStart; i < editStart; i++ {
		fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
	}

	// The edit region itself
	editLen := editEnd - editStart
	switch {
	case editLen <= 0 && len(insertedLines) > 0:
		// Insertion point (no lines replaced, just inserted)
		writeContentPreview(&sb, insertedLines, editStart, "INSERTED")
	case len(insertedLines) > 0:
		// Replacement
		writeContentPreview(&sb, insertedLines, editStart, "CHANGED")
	default:
		// Show the actual edit lines
		for i := editStart; i < editEnd && i < len(allLines); i++ {
			fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
		}
	}

	// Lines after the edit
	afterEnd := editEnd + contextN
	if afterEnd > len(allLines) {
		afterEnd = len(allLines)
	}
	for i := editEnd; i < afterEnd; i++ {
		fmt.Fprintf(&sb, "L%d: %s\n", i+1, truncateLine(allLines[i], maxContextLineChars))
	}

	return sb.String()
}

// writeContentPreview writes a compact preview of inserted/changed content
func writeContentPreview(sb *strings.Builder, contentLines []string, startLine int, label string) {
	if len(contentLines) <= maxContextContentPreview*2+1 {
		// Short enough to show fully
		for i, line := range contentLines {
			marker := ""
			if i == 0 {
				marker = fmt.Sprintf("  ← %s", label)
			}
			fmt.Fprintf(sb, "L%d: %s%s\n", startLine+i+1, truncateLine(line, maxContextLineChars), marker)
		}
	} else {
		// Show first N and last N lines
		for i := 0; i < maxContextContentPreview; i++ {
			marker := ""
			if i == 0 {
				marker = fmt.Sprintf("  ← %s", label)
			}
			fmt.Fprintf(sb, "L%d: %s%s\n", startLine+i+1, truncateLine(contentLines[i], maxContextLineChars), marker)
		}
		fmt.Fprintf(sb, "     [... %d more lines ...]\n", len(contentLines)-maxContextContentPreview*2)
		for i := len(contentLines) - maxContextContentPreview; i < len(contentLines); i++ {
			fmt.Fprintf(sb, "L%d: %s\n", startLine+i+1, truncateLine(contentLines[i], maxContextLineChars))
		}
	}
}

// EditNoteHandler performs surgical find-and-replace within a note
func (v *Vault) EditNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	oldText, err := req.RequireString("old_text")
	if err != nil {
		return mcp.NewToolResultError("old_text is required"), nil
	}

	replacementText, err := req.RequireString("new_text")
	if err != nil {
		return mcp.NewToolResultError("new_text is required"), nil
	}

	replaceAll := req.GetBool("replace_all", false)
	contextN := int(req.GetInt("context_lines", 0))

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

	// Count occurrences
	count := strings.Count(contentStr, oldText)
	if count == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("old_text not found in %s", notePath)), nil
	}

	if count > 1 && !replaceAll {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Found %d occurrences of old_text in %s. Use replace_all=true to replace all, or provide more context to match uniquely.",
			count, notePath,
		)), nil
	}

	// Perform replacement
	var newContent string
	replaced := 0
	if replaceAll {
		newContent = strings.ReplaceAll(contentStr, oldText, replacementText)
		replaced = count
	} else {
		newContent = strings.Replace(contentStr, oldText, replacementText, 1)
		replaced = 1
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Replaced %d occurrence(s) in %s", replaced, notePath)

	if contextN > 0 {
		// Find the first edit location for context
		newLines := strings.Split(newContent, "\n")
		replacementLines := strings.Split(replacementText, "\n")

		// Find where the replacement starts
		idx := strings.Index(newContent, replacementText)
		if idx >= 0 {
			editStartLine := strings.Count(newContent[:idx], "\n")
			editEndLine := editStartLine + len(replacementLines)

			ctxStr := buildEditContext(newLines, editStartLine, editEndLine, contextN, replacementLines)
			fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// ReplaceSectionHandler replaces content under a heading
func (v *Vault) ReplaceSectionHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	heading, err := req.RequireString("heading")
	if err != nil {
		return mcp.NewToolResultError("heading is required"), nil
	}

	sectionContent, err := req.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError("content is required"), nil
	}

	contextN := int(req.GetInt("context_lines", 0))

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

	lines := strings.Split(string(content), "\n")

	// Find the section boundaries
	sectionStart := -1 // Line index of heading
	sectionEnd := -1   // Line index of next heading (exclusive)
	sectionLevel := 0

	for i, line := range lines {
		matches := headingRegex.FindStringSubmatch(line)
		if matches != nil {
			level := len(matches[1])
			text := strings.TrimSpace(matches[2])

			if sectionStart == -1 && strings.EqualFold(text, heading) {
				sectionStart = i
				sectionLevel = level
				continue
			}

			if sectionStart >= 0 && level <= sectionLevel {
				sectionEnd = i
				break
			}
		}
	}

	if sectionStart == -1 {
		return mcp.NewToolResultError(fmt.Sprintf("Heading '%s' not found in %s", heading, notePath)), nil
	}

	if sectionEnd == -1 {
		sectionEnd = len(lines)
	}

	// Content starts after the heading line
	contentStart := sectionStart + 1
	linesReplaced := sectionEnd - contentStart

	// Normalize new content: ensure blank line after heading, trim trailing newlines
	normalizedContent := strings.TrimRight(sectionContent, "\n")
	newContentLines := strings.Split("\n"+normalizedContent+"\n", "\n")

	// Build new file: before heading + heading + new content + after section
	var result []string
	result = append(result, lines[:contentStart]...) // Up to and including heading
	result = append(result, newContentLines...)      // New section content
	result = append(result, lines[sectionEnd:]...)   // Everything after section

	finalContent := strings.Join(result, "\n")

	if err := os.WriteFile(fullPath, []byte(finalContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Replaced section '%s' in %s (%d lines replaced with %d lines)",
		heading, notePath, linesReplaced, len(newContentLines))

	if contextN > 0 {
		finalLines := strings.Split(finalContent, "\n")
		editEnd := contentStart + len(newContentLines)
		if editEnd > len(finalLines) {
			editEnd = len(finalLines)
		}
		ctxStr := buildEditContext(finalLines, contentStart, editEnd, contextN, newContentLines)
		fmt.Fprintf(&sb, "\n\n--- Context ---\n%s", ctxStr)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// editEntry represents a single find-and-replace operation in a batch
type editEntry struct {
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

// locatedEdit is an editEntry with its byte offset in the file content
type locatedEdit struct {
	editEntry
	offset int // byte offset of old_text in content
}

// BatchEditNoteHandler performs multiple find-and-replace operations atomically
func (v *Vault) BatchEditNoteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	editsRaw, err := req.RequireString("edits")
	if err != nil {
		return mcp.NewToolResultError("edits is required (JSON array of {old_text, new_text} objects)"), nil
	}

	contextN := int(req.GetInt("context_lines", 0))

	if !strings.HasSuffix(notePath, ".md") {
		notePath += ".md"
	}

	fullPath := filepath.Join(v.path, notePath)
	if !v.isPathSafe(fullPath) {
		return mcp.NewToolResultError("path must be within vault"), nil
	}

	// Parse edits JSON
	var edits []editEntry
	if err := json.Unmarshal([]byte(editsRaw), &edits); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid edits JSON: %v. Expected [{\"old_text\": \"...\", \"new_text\": \"...\"}, ...]", err)), nil
	}

	if len(edits) == 0 {
		return mcp.NewToolResultError("edits array is empty"), nil
	}

	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return mcp.NewToolResultError(fmt.Sprintf("Note not found: %s", notePath)), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read note: %v", err)), nil
	}

	contentStr := string(content)

	// Validate all edits before applying any
	located, validationErr := validateBatchEdits(contentStr, edits, notePath)
	if validationErr != nil {
		return validationErr, nil
	}

	// Sort by offset descending so we can apply from end to start without shifting
	sort.Slice(located, func(i, j int) bool {
		return located[i].offset > located[j].offset
	})

	// Apply all edits
	result := contentStr
	for _, le := range located {
		result = result[:le.offset] + le.NewText + result[le.offset+len(le.OldText):]
	}

	if err := os.WriteFile(fullPath, []byte(result), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	// Build response
	var sb strings.Builder
	fmt.Fprintf(&sb, "Applied %d edit(s) to %s", len(edits), notePath)

	if contextN > 0 && len(located) > 0 {
		// Show context around the first edit (by file position, which is last in our sorted slice)
		firstEdit := located[len(located)-1]
		newLines := strings.Split(result, "\n")
		replacementLines := strings.Split(firstEdit.NewText, "\n")

		idx := strings.Index(result, firstEdit.NewText)
		if idx >= 0 {
			editStartLine := strings.Count(result[:idx], "\n")
			editEndLine := editStartLine + len(replacementLines)
			ctxStr := buildEditContext(newLines, editStartLine, editEndLine, contextN, replacementLines)
			fmt.Fprintf(&sb, "\n\n--- Context (first edit) ---\n%s", ctxStr)
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

// validateBatchEdits checks that every old_text exists exactly once and that no edits overlap.
// Returns located edits with byte offsets, or an error result.
func validateBatchEdits(content string, edits []editEntry, notePath string) ([]locatedEdit, *mcp.CallToolResult) {
	var located []locatedEdit
	var errors []string

	for i, e := range edits {
		if e.OldText == "" {
			errors = append(errors, fmt.Sprintf("edit %d: old_text is empty", i+1))
			continue
		}

		count := strings.Count(content, e.OldText)
		switch {
		case count == 0:
			errors = append(errors, fmt.Sprintf("edit %d: old_text not found: %q", i+1, truncateLine(e.OldText, 80)))
		case count > 1:
			errors = append(errors, fmt.Sprintf("edit %d: old_text found %d times (must be unique): %q", i+1, count, truncateLine(e.OldText, 80)))
		default:
			offset := strings.Index(content, e.OldText)
			located = append(located, locatedEdit{editEntry: e, offset: offset})
		}
	}

	if len(errors) > 0 {
		return nil, mcp.NewToolResultError(fmt.Sprintf(
			"Batch edit validation failed for %s:\n- %s",
			notePath, strings.Join(errors, "\n- "),
		))
	}

	// Check for overlapping edits
	sort.Slice(located, func(i, j int) bool {
		return located[i].offset < located[j].offset
	})
	for i := 1; i < len(located); i++ {
		prevEnd := located[i-1].offset + len(located[i-1].OldText)
		if located[i].offset < prevEnd {
			return nil, mcp.NewToolResultError(fmt.Sprintf(
				"Batch edit validation failed for %s: edits %d and %d overlap",
				notePath, i, i+1,
			))
		}
	}

	return located, nil
}
