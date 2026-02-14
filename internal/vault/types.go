package vault

type ListNotesArgs struct {
	Directory string `json:"directory,omitempty" jsonschema:"Directory path relative to vault root (optional)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum number of notes to return (optional, 0 = no limit)"`
	Offset    int    `json:"offset,omitempty" jsonschema:"Number of notes to skip for pagination (optional, default 0)"`
}

type WriteNoteArgs struct {
	Path    string `json:"path" jsonschema:"Path to the note relative to vault root (.md extension required)"`
	Content string `json:"content" jsonschema:"Content of the note"`
}
