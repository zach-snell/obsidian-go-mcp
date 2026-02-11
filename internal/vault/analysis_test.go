package vault

import (
	"testing"
)

func TestIsAlreadyLinked(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		noteName string
		expected bool
	}{
		{
			name:     "linked with wikilink",
			line:     "See [[My Note]] for details",
			noteName: "My Note",
			expected: true,
		},
		{
			name:     "linked with alias",
			line:     "See [[My Note|the note]] for details",
			noteName: "My Note",
			expected: true,
		},
		{
			name:     "not linked",
			line:     "See My Note for details",
			noteName: "My Note",
			expected: false,
		},
		{
			name:     "case insensitive link",
			line:     "See [[my note]] for details",
			noteName: "My Note",
			expected: true,
		},
		{
			name:     "partial match not linked",
			line:     "See My Note's content",
			noteName: "My Note",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlreadyLinked(tt.line, tt.noteName)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for line %q", tt.expected, result, tt.line)
			}
		})
	}
}
