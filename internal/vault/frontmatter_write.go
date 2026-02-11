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

// frontmatterRegex matches the frontmatter block
var frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n?`)

// SetFrontmatterHandler sets or updates a frontmatter property
func (v *Vault) SetFrontmatterHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError("key is required"), nil
	}

	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError("value is required"), nil
	}

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

	newContent := setFrontmatterKey(string(content), key, value)

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Set %s: %s in %s", key, value, notePath)), nil
}

// RemoveFrontmatterKeyHandler removes a frontmatter property
func (v *Vault) RemoveFrontmatterKeyHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError("key is required"), nil
	}

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

	newContent, removed := removeFrontmatterKey(string(content), key)
	if !removed {
		return mcp.NewToolResultText(fmt.Sprintf("Key '%s' not found in frontmatter", key)), nil
	}

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Removed %s from %s", key, notePath)), nil
}

// AddAliasHandler adds an alias to a note's frontmatter
func (v *Vault) AddAliasHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	alias, err := req.RequireString("alias")
	if err != nil {
		return mcp.NewToolResultError("alias is required"), nil
	}

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

	newContent := addToFrontmatterArray(string(content), "aliases", alias)

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Added alias '%s' to %s", alias, notePath)), nil
}

// AddTagToFrontmatterHandler adds a tag to frontmatter
func (v *Vault) AddTagToFrontmatterHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	notePath, err := req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError("path is required"), nil
	}

	tag, err := req.RequireString("tag")
	if err != nil {
		return mcp.NewToolResultError("tag is required"), nil
	}

	// Normalize tag (remove # if present)
	tag = strings.TrimPrefix(tag, "#")

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

	newContent := addToFrontmatterArray(string(content), "tags", tag)

	if err := os.WriteFile(fullPath, []byte(newContent), 0o600); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write note: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Added tag '%s' to frontmatter in %s", tag, notePath)), nil
}

// setFrontmatterKey sets a key in frontmatter, creating frontmatter if needed
func setFrontmatterKey(content, key, value string) string {
	// Normalize key to lowercase
	key = strings.ToLower(key)

	if !strings.HasPrefix(content, "---") {
		// No frontmatter, create it
		return fmt.Sprintf("---\n%s: %s\n---\n\n%s", key, value, content)
	}

	match := frontmatterRegex.FindStringSubmatch(content)
	if match == nil {
		// Invalid frontmatter, add new
		return fmt.Sprintf("---\n%s: %s\n---\n\n%s", key, value, content)
	}

	fmContent := match[1]
	body := content[len(match[0]):]

	// Check if key exists
	keyPattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `\s*:.*$`)
	if keyPattern.MatchString(fmContent) {
		// Update existing key
		fmContent = keyPattern.ReplaceAllString(fmContent, fmt.Sprintf("%s: %s", key, value))
	} else {
		// Add new key
		fmContent = fmContent + "\n" + key + ": " + value
	}

	return fmt.Sprintf("---\n%s\n---\n%s", strings.TrimSpace(fmContent), body)
}

// removeFrontmatterKey removes a key from frontmatter
func removeFrontmatterKey(content, key string) (string, bool) {
	key = strings.ToLower(key)

	if !strings.HasPrefix(content, "---") {
		return content, false
	}

	match := frontmatterRegex.FindStringSubmatch(content)
	if match == nil {
		return content, false
	}

	fmContent := match[1]
	body := content[len(match[0]):]

	// Find and remove the key
	keyPattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `\s*:.*\n?`)
	if !keyPattern.MatchString(fmContent) {
		return content, false
	}

	fmContent = keyPattern.ReplaceAllString(fmContent, "")
	fmContent = strings.TrimSpace(fmContent)

	if fmContent == "" {
		// No more frontmatter, remove it entirely
		return strings.TrimPrefix(body, "\n"), true
	}

	return fmt.Sprintf("---\n%s\n---\n%s", fmContent, body), true
}

// addToFrontmatterArray adds a value to an array property in frontmatter
func addToFrontmatterArray(content, key, value string) string {
	key = strings.ToLower(key)

	if !strings.HasPrefix(content, "---") {
		// No frontmatter, create it with the array
		return fmt.Sprintf("---\n%s:\n  - %s\n---\n\n%s", key, value, content)
	}

	match := frontmatterRegex.FindStringSubmatch(content)
	if match == nil {
		return fmt.Sprintf("---\n%s:\n  - %s\n---\n\n%s", key, value, content)
	}

	fmContent := match[1]
	body := content[len(match[0]):]

	// Parse existing array
	lines := strings.Split(fmContent, "\n")
	var newLines []string
	keyFound := false
	inArray := false
	arrayEnded := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is the key we're looking for
		if strings.HasPrefix(strings.ToLower(trimmed), key+":") {
			keyFound = true
			newLines = append(newLines, line)

			// Check if inline array format: key: [a, b, c]
			if strings.Contains(line, "[") {
				// Convert inline to multi-line and add new value
				inlineMatch := regexp.MustCompile(`\[([^\]]*)\]`).FindStringSubmatch(line)
				if inlineMatch != nil {
					// Remove the inline array from this line
					newLines[len(newLines)-1] = key + ":"
					// Add existing values
					existing := strings.Split(inlineMatch[1], ",")
					for _, v := range existing {
						v = strings.TrimSpace(v)
						if v != "" {
							newLines = append(newLines, "  - "+v)
						}
					}
					// Add new value
					newLines = append(newLines, "  - "+value)
					arrayEnded = true
				}
			} else {
				inArray = true
			}
			continue
		}

		// If we're in the array section
		if inArray && !arrayEnded {
			if strings.HasPrefix(trimmed, "- ") {
				newLines = append(newLines, line)
				continue
			}
			// Array ended, add new value before this line
			newLines = append(newLines, "  - "+value)
			arrayEnded = true
			inArray = false
		}

		newLines = append(newLines, line)
	}

	// If we were in array at end of frontmatter
	if inArray && !arrayEnded {
		newLines = append(newLines, "  - "+value)
	}

	// If key not found, add it
	if !keyFound {
		newLines = append(newLines, key+":", "  - "+value)
	}

	return fmt.Sprintf("---\n%s\n---\n%s", strings.Join(newLines, "\n"), body)
}
