package vault

import (
	"testing"
)

// FuzzParseTask tests that ParseTask doesn't panic on arbitrary input
func FuzzParseTask(f *testing.F) {
	// Seed corpus with interesting inputs
	seeds := []string{
		"- [ ] Simple task",
		"- [x] Completed task",
		"- [X] Also completed",
		"- [ ] Task with date 📅 2024-01-15",
		"- [ ] High priority ⏫",
		"- [ ] Medium priority 🔼",
		"- [ ] Low priority 🔽",
		"- [ ] Task #tag1 #tag2",
		"  - [ ] Indented task",
		"      - [x] Deep indent",
		"",
		"# Not a task",
		"- Regular list item",
		"- [] Malformed",
		"- [  ] Double space",
		"- [xx] Double x",
		"[]",
		"-",
		"- [ ]",
		"- [ ] 📅",
		"- [ ] ⏫ 🔼 🔽",
		"- [ ] #",
		"- [ ] ##tag",
		"- [ ] Task with\nnewline",
		"- [ ] Very long task " + string(make([]byte, 10000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		task := ParseTask(input, 1)

		// If we got a task, verify basic properties
		if task != nil {
			if task.Line != 1 {
				t.Errorf("Line should be 1, got %d", task.Line)
			}
			// Text should not be empty for valid tasks
			if task.Text == "" {
				t.Errorf("Text should not be empty for valid task")
			}
		}
	})
}

// FuzzExtractTags tests that ExtractTags doesn't panic on arbitrary input
func FuzzExtractTags(f *testing.F) {
	seeds := []string{
		"#tag",
		"#tag1 #tag2",
		"Text with #inline tag",
		"---\ntags: [tag1, tag2]\n---",
		"---\ntags:\n  - tag1\n  - tag2\n---",
		"",
		"No tags here",
		"##not-a-tag",
		"#",
		"# Heading",
		"```\n#code-tag\n```",
		"---\ntags: []\n---",
		"---\ntags:\n---",
		"#tag-with-dash",
		"#tag_with_underscore",
		"#MixedCase",
		"#123numeric",
		string(make([]byte, 10000)),
		"---" + string(make([]byte, 1000)) + "---",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		tags := ExtractTags(input)

		// Tags may be nil or empty - both are valid
		_ = tags

		// Each tag should be non-empty
		for _, tag := range tags {
			if tag == "" {
				t.Errorf("Tag should not be empty")
			}
		}
	})
}

// FuzzExtractWikilinks tests that ExtractWikilinks doesn't panic
func FuzzExtractWikilinks(f *testing.F) {
	seeds := []string{
		"[[Note]]",
		"[[path/to/note]]",
		"[[Note|Alias]]",
		"[[Note1]] and [[Note2]]",
		"No links here",
		"[[]]",
		"[[ ]]",
		"[[unclosed",
		"]]closed[[",
		"[[nested [[link]]]]",
		string(make([]byte, 10000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		links := ExtractWikilinks(input)

		// Links should contain no empty strings
		for _, link := range links {
			if link == "" {
				t.Errorf("Link should not be empty")
			}
		}
	})
}

// FuzzExtractH1Title tests that ExtractH1Title doesn't panic
func FuzzExtractH1Title(f *testing.F) {
	seeds := []string{
		"# Title",
		"# Title with spaces",
		"## Not H1",
		"No title",
		"",
		"#",
		"# ",
		"#Title without space",
		"Multiple\n# Title\nLines",
		string(make([]byte, 10000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		_ = ExtractH1Title(input)
	})
}

// FuzzTruncate tests that truncate doesn't panic
func FuzzTruncate(f *testing.F) {
	f.Add("short", 10)
	f.Add("exactly ten", 10)
	f.Add("this is longer than ten", 10)
	f.Add("", 0)
	f.Add("", 10)
	f.Add("test", 0)
	f.Add("test", -1)
	f.Add(string(make([]byte, 10000)), 100)

	f.Fuzz(func(t *testing.T, input string, maxLen int) {
		// Should never panic
		result := truncate(input, maxLen)

		// Result should not exceed maxLen + 3 (for "...")
		if maxLen > 0 && len(result) > maxLen+3 {
			t.Errorf("truncate(%q, %d) = %q, len %d > %d",
				input, maxLen, result, len(result), maxLen+3)
		}
	})
}

// FuzzIsMOC tests that IsMOC doesn't panic
func FuzzIsMOC(f *testing.F) {
	seeds := []string{
		"#moc",
		"#MOC",
		"#Moc",
		"#moc #other",
		"Text #moc text",
		"No moc tag",
		"",
		"---\ntags: [moc]\n---",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		_ = IsMOC(input)
	})
}
