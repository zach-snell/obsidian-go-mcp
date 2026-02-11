package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/zach-snell/obsidian-go-mcp/internal/vault"
)

// New creates a new MCP server configured with vault tools
func New(vaultPath string) *server.MCPServer {
	v := vault.New(vaultPath)

	s := server.NewMCPServer(
		"Obsidian Vault MCP",
		"0.3.2",
		server.WithToolCapabilities(false),
	)

	// Register tools
	registerTools(s, v)

	return s
}

func registerTools(s *server.MCPServer, v *vault.Vault) {
	// list-notes
	s.AddTool(
		mcp.NewTool("list-notes",
			mcp.WithDescription("List all notes in the vault or a specific directory"),
			mcp.WithString("directory",
				mcp.Description("Directory path relative to vault root (optional)"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of notes to return (optional, 0 = no limit)"),
			),
			mcp.WithNumber("offset",
				mcp.Description("Number of notes to skip for pagination (optional, default 0)"),
			),
		),
		v.ListNotesHandler,
	)

	// read-note
	s.AddTool(
		mcp.NewTool("read-note",
			mcp.WithDescription("Read the content of a specific note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note relative to vault root (.md extension required)"),
			),
		),
		v.ReadNoteHandler,
	)

	// write-note
	s.AddTool(
		mcp.NewTool("write-note",
			mcp.WithDescription("Create or update a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note relative to vault root (.md extension required)"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Content of the note"),
			),
		),
		v.WriteNoteHandler,
	)

	// delete-note
	s.AddTool(
		mcp.NewTool("delete-note",
			mcp.WithDescription("Delete a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note relative to vault root (.md extension required)"),
			),
		),
		v.DeleteNoteHandler,
	)

	// search-vault
	s.AddTool(
		mcp.NewTool("search-vault",
			mcp.WithDescription("Search for content in vault notes"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query (case-insensitive substring match)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory (optional)"),
			),
		),
		v.SearchVaultHandler,
	)

	// list-tasks
	s.AddTool(
		mcp.NewTool("list-tasks",
			mcp.WithDescription("List tasks (checkboxes) across the vault"),
			mcp.WithString("status",
				mcp.Description("Filter by status: 'all', 'open', 'completed' (default: all)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit to specific directory (optional)"),
			),
		),
		v.ListTasksHandler,
	)

	// search-by-tags
	s.AddTool(
		mcp.NewTool("search-by-tags",
			mcp.WithDescription("Search for notes by tags"),
			mcp.WithString("tags",
				mcp.Required(),
				mcp.Description("Comma-separated list of tags to search for (AND operation)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory (optional)"),
			),
		),
		v.SearchByTagsHandler,
	)

	// discover-mocs
	s.AddTool(
		mcp.NewTool("discover-mocs",
			mcp.WithDescription("Discover MOCs (Maps of Content) - notes tagged with #moc"),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory (optional)"),
			),
		),
		v.DiscoverMOCsHandler,
	)

	// toggle-task
	s.AddTool(
		mcp.NewTool("toggle-task",
			mcp.WithDescription("Toggle a task's completion status (checked/unchecked)"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note containing the task (.md extension required)"),
			),
			mcp.WithNumber("line",
				mcp.Required(),
				mcp.Description("Line number of the task to toggle (1-based)"),
			),
		),
		v.ToggleTaskHandler,
	)

	// append-note
	s.AddTool(
		mcp.NewTool("append-note",
			mcp.WithDescription("Append content to a note (creates if doesn't exist). Supports targeted insertion with after/before/position params."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note (.md extension required)"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Content to append"),
			),
			mcp.WithString("position",
				mcp.Description("Where to insert: end, start, after, before (default: end)"),
			),
			mcp.WithString("after",
				mcp.Description("Insert after this heading or text (overrides position to 'after')"),
			),
			mcp.WithString("before",
				mcp.Description("Insert before this heading or text (overrides position to 'before')"),
			),
			mcp.WithNumber("context_lines",
				mcp.Description("Number of surrounding lines to return for verification (default: 0)"),
			),
		),
		v.AppendNoteHandler,
	)

	// edit-note
	s.AddTool(
		mcp.NewTool("edit-note",
			mcp.WithDescription("Surgical find-and-replace within a note. Fails if old_text matches multiple times (use replace_all=true to override)."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note (.md extension required)"),
			),
			mcp.WithString("old_text",
				mcp.Required(),
				mcp.Description("Exact text to find and replace"),
			),
			mcp.WithString("new_text",
				mcp.Required(),
				mcp.Description("Replacement text"),
			),
			mcp.WithBoolean("replace_all",
				mcp.Description("Replace all occurrences instead of failing on multiple matches (default: false)"),
			),
			mcp.WithNumber("context_lines",
				mcp.Description("Number of surrounding lines to return for verification (default: 0)"),
			),
		),
		v.EditNoteHandler,
	)

	// replace-section
	s.AddTool(
		mcp.NewTool("replace-section",
			mcp.WithDescription("Replace all content under a heading (heading line is preserved, content below it is replaced until the next same-or-higher level heading)"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note (.md extension required)"),
			),
			mcp.WithString("heading",
				mcp.Required(),
				mcp.Description("Heading text to find (case-insensitive, e.g. 'Installation')"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("New content for the section (heading line itself is preserved)"),
			),
			mcp.WithNumber("context_lines",
				mcp.Description("Number of surrounding lines to return for verification (default: 0)"),
			),
		),
		v.ReplaceSectionHandler,
	)

	// recent-notes
	s.AddTool(
		mcp.NewTool("recent-notes",
			mcp.WithDescription("List recently modified notes"),
			mcp.WithNumber("limit",
				mcp.Description("Maximum number of notes to return (default: 10)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit to specific directory (optional)"),
			),
		),
		v.RecentNotesHandler,
	)

	// backlinks
	s.AddTool(
		mcp.NewTool("backlinks",
			mcp.WithDescription("Find all notes that link to a given note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note to find backlinks for"),
			),
		),
		v.BacklinksHandler,
	)

	// query-frontmatter
	s.AddTool(
		mcp.NewTool("query-frontmatter",
			mcp.WithDescription("Search notes by frontmatter properties"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Query in format: key=value or key:value (e.g., status=draft, type:project)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory (optional)"),
			),
		),
		v.QueryFrontmatterHandler,
	)

	// get-frontmatter
	s.AddTool(
		mcp.NewTool("get-frontmatter",
			mcp.WithDescription("Get frontmatter properties of a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note (.md extension required)"),
			),
		),
		v.GetFrontmatterHandler,
	)

	// rename-note
	s.AddTool(
		mcp.NewTool("rename-note",
			mcp.WithDescription("Rename a note and update all links to it"),
			mcp.WithString("old_path",
				mcp.Required(),
				mcp.Description("Current path of the note"),
			),
			mcp.WithString("new_path",
				mcp.Required(),
				mcp.Description("New path for the note"),
			),
		),
		v.RenameNoteHandler,
	)

	// daily-note
	s.AddTool(
		mcp.NewTool("daily-note",
			mcp.WithDescription("Get or create a daily note"),
			mcp.WithString("date",
				mcp.Description("Date for the note (default: today). Formats: YYYY-MM-DD, MM-DD-YYYY, etc."),
			),
			mcp.WithString("folder",
				mcp.Description("Folder for daily notes (default: 'daily')"),
			),
			mcp.WithString("format",
				mcp.Description("Date format for filename (default: '2006-01-02', Go time format)"),
			),
			mcp.WithBoolean("create",
				mcp.Description("Create if missing (default: true)"),
			),
		),
		v.DailyNoteHandler,
	)

	// list-daily-notes
	s.AddTool(
		mcp.NewTool("list-daily-notes",
			mcp.WithDescription("List daily notes"),
			mcp.WithString("folder",
				mcp.Description("Folder for daily notes (default: 'daily')"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum notes to return (default: 30)"),
			),
		),
		v.ListDailyNotesHandler,
	)

	// list-templates
	s.AddTool(
		mcp.NewTool("list-templates",
			mcp.WithDescription("List available templates"),
			mcp.WithString("folder",
				mcp.Description("Templates folder (default: 'templates')"),
			),
		),
		v.ListTemplatesHandler,
	)

	// get-template
	s.AddTool(
		mcp.NewTool("get-template",
			mcp.WithDescription("Get a template and show its variables"),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Template name (with or without .md)"),
			),
			mcp.WithString("folder",
				mcp.Description("Templates folder (default: 'templates')"),
			),
		),
		v.GetTemplateHandler,
	)

	// apply-template
	s.AddTool(
		mcp.NewTool("apply-template",
			mcp.WithDescription("Create a new note from a template"),
			mcp.WithString("template",
				mcp.Required(),
				mcp.Description("Template name"),
			),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Target path for the new note"),
			),
			mcp.WithString("variables",
				mcp.Description("Variables as 'key1=value1,key2=value2'. Built-ins: date, time, title, filename"),
			),
			mcp.WithString("template_folder",
				mcp.Description("Templates folder (default: 'templates')"),
			),
		),
		v.ApplyTemplateHandler,
	)

	// vault-stats
	s.AddTool(
		mcp.NewTool("vault-stats",
			mcp.WithDescription("Get statistics about the vault"),
			mcp.WithString("directory",
				mcp.Description("Limit stats to specific directory (optional)"),
			),
		),
		v.VaultStatsHandler,
	)

	// weekly-note
	s.AddTool(
		mcp.NewTool("weekly-note",
			mcp.WithDescription("Get or create a weekly note"),
			mcp.WithString("date",
				mcp.Description("Date within the week (default: this week). Any date in the week works."),
			),
			mcp.WithString("folder",
				mcp.Description("Folder for weekly notes (default: 'weekly')"),
			),
			mcp.WithString("format",
				mcp.Description("Filename format (default: '2006-W02' for YYYY-Www)"),
			),
			mcp.WithBoolean("create",
				mcp.Description("Create if missing (default: true)"),
			),
		),
		v.WeeklyNoteHandler,
	)

	// monthly-note
	s.AddTool(
		mcp.NewTool("monthly-note",
			mcp.WithDescription("Get or create a monthly note"),
			mcp.WithString("date",
				mcp.Description("Date within the month (default: this month)"),
			),
			mcp.WithString("folder",
				mcp.Description("Folder for monthly notes (default: 'monthly')"),
			),
			mcp.WithString("format",
				mcp.Description("Filename format (default: '2006-01' for YYYY-MM)"),
			),
			mcp.WithBoolean("create",
				mcp.Description("Create if missing (default: true)"),
			),
		),
		v.MonthlyNoteHandler,
	)

	// quarterly-note
	s.AddTool(
		mcp.NewTool("quarterly-note",
			mcp.WithDescription("Get or create a quarterly note"),
			mcp.WithString("date",
				mcp.Description("Date within the quarter (default: this quarter)"),
			),
			mcp.WithString("folder",
				mcp.Description("Folder for quarterly notes (default: 'quarterly')"),
			),
			mcp.WithBoolean("create",
				mcp.Description("Create if missing (default: true)"),
			),
		),
		v.QuarterlyNoteHandler,
	)

	// yearly-note
	s.AddTool(
		mcp.NewTool("yearly-note",
			mcp.WithDescription("Get or create a yearly note"),
			mcp.WithString("date",
				mcp.Description("Date within the year (default: this year)"),
			),
			mcp.WithString("folder",
				mcp.Description("Folder for yearly notes (default: 'yearly')"),
			),
			mcp.WithBoolean("create",
				mcp.Description("Create if missing (default: true)"),
			),
		),
		v.YearlyNoteHandler,
	)

	// list-periodic-notes
	s.AddTool(
		mcp.NewTool("list-periodic-notes",
			mcp.WithDescription("List periodic notes by type"),
			mcp.WithString("type",
				mcp.Description("Type: daily, weekly, monthly, quarterly, yearly (default: weekly)"),
			),
			mcp.WithString("folder",
				mcp.Description("Override folder location"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum notes to return (default: 20)"),
			),
		),
		v.ListPeriodicNotesHandler,
	)

	// forward-links
	s.AddTool(
		mcp.NewTool("forward-links",
			mcp.WithDescription("Show outgoing links from a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
		),
		v.ForwardLinksHandler,
	)

	// orphan-notes
	s.AddTool(
		mcp.NewTool("orphan-notes",
			mcp.WithDescription("Find notes with no links to/from them"),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithBoolean("include_no_outgoing",
				mcp.Description("Also show notes with incoming links but no outgoing (dead ends)"),
			),
		),
		v.OrphanNotesHandler,
	)

	// broken-links
	s.AddTool(
		mcp.NewTool("broken-links",
			mcp.WithDescription("Find wikilinks pointing to non-existent notes"),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
		),
		v.BrokenLinksHandler,
	)

	// list-folders
	s.AddTool(
		mcp.NewTool("list-folders",
			mcp.WithDescription("List all folders in the vault"),
			mcp.WithString("directory",
				mcp.Description("Start from a specific directory"),
			),
			mcp.WithBoolean("include_empty",
				mcp.Description("Include empty folders (default: true)"),
			),
		),
		v.ListFoldersHandler,
	)

	// create-folder
	s.AddTool(
		mcp.NewTool("create-folder",
			mcp.WithDescription("Create a new folder"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Folder path to create"),
			),
		),
		v.CreateFolderHandler,
	)

	// move-note
	s.AddTool(
		mcp.NewTool("move-note",
			mcp.WithDescription("Move a note to a new location"),
			mcp.WithString("source",
				mcp.Required(),
				mcp.Description("Current path of the note"),
			),
			mcp.WithString("destination",
				mcp.Required(),
				mcp.Description("New path for the note"),
			),
			mcp.WithBoolean("update_links",
				mcp.Description("Update wikilinks in other notes (default: true)"),
			),
		),
		v.MoveNoteHandler,
	)

	// delete-folder
	s.AddTool(
		mcp.NewTool("delete-folder",
			mcp.WithDescription("Delete a folder"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Folder path to delete"),
			),
			mcp.WithBoolean("force",
				mcp.Description("Delete even if not empty (default: false)"),
			),
		),
		v.DeleteFolderHandler,
	)

	// list-canvases
	s.AddTool(
		mcp.NewTool("list-canvases",
			mcp.WithDescription("List all canvas files in the vault"),
			mcp.WithString("directory",
				mcp.Description("Limit to specific directory"),
			),
		),
		v.ListCanvasesHandler,
	)

	// read-canvas
	s.AddTool(
		mcp.NewTool("read-canvas",
			mcp.WithDescription("Read and parse a canvas file"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the canvas file"),
			),
		),
		v.ReadCanvasHandler,
	)

	// create-canvas
	s.AddTool(
		mcp.NewTool("create-canvas",
			mcp.WithDescription("Create a new empty canvas"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path for the new canvas"),
			),
		),
		v.CreateCanvasHandler,
	)

	// add-canvas-node
	s.AddTool(
		mcp.NewTool("add-canvas-node",
			mcp.WithDescription("Add a node to a canvas"),
			mcp.WithString("canvas",
				mcp.Required(),
				mcp.Description("Path to the canvas file"),
			),
			mcp.WithString("type",
				mcp.Description("Node type: text, file, link, group (default: text)"),
			),
			mcp.WithString("content",
				mcp.Description("Node content (text for text nodes, path for file nodes, URL for link nodes)"),
			),
			mcp.WithNumber("x",
				mcp.Description("X position (default: 0)"),
			),
			mcp.WithNumber("y",
				mcp.Description("Y position (default: 0)"),
			),
			mcp.WithNumber("width",
				mcp.Description("Node width (default: 300)"),
			),
			mcp.WithNumber("height",
				mcp.Description("Node height (default: 200)"),
			),
		),
		v.AddCanvasNodeHandler,
	)

	// add-canvas-edge
	s.AddTool(
		mcp.NewTool("add-canvas-edge",
			mcp.WithDescription("Add an edge between nodes in a canvas"),
			mcp.WithString("canvas",
				mcp.Required(),
				mcp.Description("Path to the canvas file"),
			),
			mcp.WithString("from",
				mcp.Required(),
				mcp.Description("Source node ID"),
			),
			mcp.WithString("to",
				mcp.Required(),
				mcp.Description("Target node ID"),
			),
			mcp.WithString("label",
				mcp.Description("Optional edge label"),
			),
		),
		v.AddCanvasEdgeHandler,
	)

	// read-notes (batch)
	s.AddTool(
		mcp.NewTool("read-notes",
			mcp.WithDescription("Read multiple notes in one call"),
			mcp.WithString("paths",
				mcp.Required(),
				mcp.Description("Comma-separated paths or JSON array of paths"),
			),
			mcp.WithBoolean("include_frontmatter",
				mcp.Description("Include frontmatter in output (default: true)"),
			),
		),
		v.ReadNotesHandler,
	)

	// get-note-summary
	s.AddTool(
		mcp.NewTool("get-note-summary",
			mcp.WithDescription("Get lightweight summary of a note (frontmatter, stats, preview)"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithNumber("lines",
				mcp.Description("Number of preview lines (default: 5)"),
			),
		),
		v.GetNoteSummaryHandler,
	)

	// get-section
	s.AddTool(
		mcp.NewTool("get-section",
			mcp.WithDescription("Extract a specific heading section from a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("heading",
				mcp.Required(),
				mcp.Description("Heading text to extract (case-insensitive)"),
			),
		),
		v.GetSectionHandler,
	)

	// get-headings
	s.AddTool(
		mcp.NewTool("get-headings",
			mcp.WithDescription("List all headings in a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
		),
		v.GetHeadingsHandler,
	)

	// search-headings
	s.AddTool(
		mcp.NewTool("search-headings",
			mcp.WithDescription("Search across all heading content in the vault"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query for heading text"),
			),
			mcp.WithNumber("level",
				mcp.Description("Filter by heading level (1-6, 0 for all)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
		),
		v.SearchHeadingsHandler,
	)

	// set-frontmatter
	s.AddTool(
		mcp.NewTool("set-frontmatter",
			mcp.WithDescription("Set or update a frontmatter property"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Property key"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("Property value"),
			),
		),
		v.SetFrontmatterHandler,
	)

	// remove-frontmatter-key
	s.AddTool(
		mcp.NewTool("remove-frontmatter-key",
			mcp.WithDescription("Remove a property from frontmatter"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Property key to remove"),
			),
		),
		v.RemoveFrontmatterKeyHandler,
	)

	// add-alias
	s.AddTool(
		mcp.NewTool("add-alias",
			mcp.WithDescription("Add an alias to a note's frontmatter"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("alias",
				mcp.Required(),
				mcp.Description("Alias to add"),
			),
		),
		v.AddAliasHandler,
	)

	// add-tag-to-frontmatter
	s.AddTool(
		mcp.NewTool("add-tag-to-frontmatter",
			mcp.WithDescription("Add a tag to frontmatter tags array"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("tag",
				mcp.Required(),
				mcp.Description("Tag to add (with or without #)"),
			),
		),
		v.AddTagToFrontmatterHandler,
	)

	// get-inline-fields
	s.AddTool(
		mcp.NewTool("get-inline-fields",
			mcp.WithDescription("Extract Dataview-style inline fields (key:: value) from a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
		),
		v.GetInlineFieldsHandler,
	)

	// set-inline-field
	s.AddTool(
		mcp.NewTool("set-inline-field",
			mcp.WithDescription("Set or update a Dataview-style inline field"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Field key"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("Field value"),
			),
		),
		v.SetInlineFieldHandler,
	)

	// query-inline-fields
	s.AddTool(
		mcp.NewTool("query-inline-fields",
			mcp.WithDescription("Search notes by Dataview-style inline field values"),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Field key to search for"),
			),
			mcp.WithString("value",
				mcp.Description("Value to match (optional, matches all if empty)"),
			),
			mcp.WithString("operator",
				mcp.Description("Match type: contains, equals, exists (default: contains)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
		),
		v.QueryInlineFieldsHandler,
	)

	// find-stubs
	s.AddTool(
		mcp.NewTool("find-stubs",
			mcp.WithDescription("Find notes with few words that may need expansion"),
			mcp.WithNumber("max_words",
				mcp.Description("Maximum word count to be considered a stub (default: 100)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default: 50)"),
			),
		),
		v.FindStubsHandler,
	)

	// find-outdated
	s.AddTool(
		mcp.NewTool("find-outdated",
			mcp.WithDescription("Find notes not modified in a while"),
			mcp.WithNumber("days",
				mcp.Description("Days since last modification (default: 90)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default: 50)"),
			),
		),
		v.FindOutdatedHandler,
	)

	// unlinked-mentions
	s.AddTool(
		mcp.NewTool("unlinked-mentions",
			mcp.WithDescription("Find text matching a note's name that isn't linked"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note to find mentions of"),
			),
		),
		v.UnlinkedMentionsHandler,
	)

	// suggest-links
	s.AddTool(
		mcp.NewTool("suggest-links",
			mcp.WithDescription("Suggest notes that should be linked based on content"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note to get suggestions for"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum suggestions to return (default: 10)"),
			),
		),
		v.SuggestLinksHandler,
	)

	// search-advanced
	s.AddTool(
		mcp.NewTool("search-advanced",
			mcp.WithDescription("Advanced search with multiple terms and logical operators"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query (supports 'quoted phrases' and multiple terms)"),
			),
			mcp.WithString("in",
				mcp.Description("Where to search: content, frontmatter, tags, all (default: content)"),
			),
			mcp.WithString("operator",
				mcp.Description("Logical operator: and, or (default: and)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default: 50)"),
			),
		),
		v.SearchAdvancedHandler,
	)

	// search-by-date
	s.AddTool(
		mcp.NewTool("search-by-date",
			mcp.WithDescription("Find notes by modification date range"),
			mcp.WithString("from",
				mcp.Description("Start date (formats: YYYY-MM-DD, MM-DD-YYYY, Jan 2, 2006)"),
			),
			mcp.WithString("to",
				mcp.Description("End date (formats: YYYY-MM-DD, MM-DD-YYYY, Jan 2, 2006)"),
			),
			mcp.WithString("type",
				mcp.Description("Date type: modified, created (default: modified)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default: 50)"),
			),
		),
		v.SearchByDateHandler,
	)

	// search-regex
	s.AddTool(
		mcp.NewTool("search-regex",
			mcp.WithDescription("Search using regular expression patterns"),
			mcp.WithString("pattern",
				mcp.Required(),
				mcp.Description("Regular expression pattern"),
			),
			mcp.WithBoolean("case_insensitive",
				mcp.Description("Case insensitive matching (default: true)"),
			),
			mcp.WithString("directory",
				mcp.Description("Limit search to specific directory"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Maximum results to return (default: 50)"),
			),
		),
		v.SearchRegexHandler,
	)

	// generate-moc
	s.AddTool(
		mcp.NewTool("generate-moc",
			mcp.WithDescription("Generate a Map of Content from a directory"),
			mcp.WithString("directory",
				mcp.Description("Directory to generate MOC from (default: vault root)"),
			),
			mcp.WithString("title",
				mcp.Description("Title for the MOC (default: directory name + ' MOC')"),
			),
			mcp.WithString("output",
				mcp.Description("Output path for the MOC file (if not set, returns content)"),
			),
			mcp.WithString("group_by",
				mcp.Description("Grouping: none, alpha, tag (default: none)"),
			),
			mcp.WithBoolean("recursive",
				mcp.Description("Include subdirectories (default: false)"),
			),
		),
		v.GenerateMOCHandler,
	)

	// update-moc
	s.AddTool(
		mcp.NewTool("update-moc",
			mcp.WithDescription("Update an existing MOC with new notes"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the MOC file"),
			),
			mcp.WithString("directory",
				mcp.Description("Directory to scan for new notes"),
			),
			mcp.WithBoolean("recursive",
				mcp.Description("Include subdirectories (default: false)"),
			),
		),
		v.UpdateMOCHandler,
	)

	// generate-index
	s.AddTool(
		mcp.NewTool("generate-index",
			mcp.WithDescription("Generate an alphabetical index of all notes"),
			mcp.WithString("directory",
				mcp.Description("Directory to index (default: vault root)"),
			),
			mcp.WithString("output",
				mcp.Description("Output path for the index file (if not set, returns content)"),
			),
			mcp.WithString("title",
				mcp.Description("Title for the index (default: 'Index')"),
			),
			mcp.WithBoolean("include_orphans",
				mcp.Description("Include notes with no links or tags (default: true)"),
			),
		),
		v.GenerateIndexHandler,
	)

	// split-note
	s.AddTool(
		mcp.NewTool("split-note",
			mcp.WithDescription("Split a note into multiple notes at heading boundaries"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note to split"),
			),
			mcp.WithNumber("level",
				mcp.Description("Heading level to split at (1-6, default: 2)"),
			),
			mcp.WithBoolean("keep_original",
				mcp.Description("Keep the original note after splitting (default: false)"),
			),
			mcp.WithString("output_dir",
				mcp.Description("Directory for new notes (default: same as original)"),
			),
		),
		v.SplitNoteHandler,
	)

	// merge-notes
	s.AddTool(
		mcp.NewTool("merge-notes",
			mcp.WithDescription("Merge multiple notes into a single note"),
			mcp.WithString("paths",
				mcp.Required(),
				mcp.Description("Comma-separated paths or JSON array of notes to merge"),
			),
			mcp.WithString("output",
				mcp.Required(),
				mcp.Description("Output path for the merged note"),
			),
			mcp.WithString("separator",
				mcp.Description("Separator between merged content (default: horizontal rule)"),
			),
			mcp.WithBoolean("delete_originals",
				mcp.Description("Delete original notes after merging (default: false)"),
			),
			mcp.WithBoolean("add_headings",
				mcp.Description("Add filename as heading if note lacks one (default: true)"),
			),
		),
		v.MergeNotesHandler,
	)

	// extract-section
	s.AddTool(
		mcp.NewTool("extract-section",
			mcp.WithDescription("Extract a heading section to a new note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the source note"),
			),
			mcp.WithString("heading",
				mcp.Required(),
				mcp.Description("Heading text to extract"),
			),
			mcp.WithString("output",
				mcp.Description("Output path (default: heading name as filename)"),
			),
			mcp.WithBoolean("remove_from_original",
				mcp.Description("Remove section from original (default: true)"),
			),
			mcp.WithBoolean("add_link",
				mcp.Description("Add link to extracted note in original (default: true)"),
			),
		),
		v.ExtractSectionHandler,
	)

	// duplicate-note
	s.AddTool(
		mcp.NewTool("duplicate-note",
			mcp.WithDescription("Create a copy of a note"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note to duplicate"),
			),
			mcp.WithString("output",
				mcp.Description("Output path (default: original name + ' (copy)')"),
			),
		),
		v.DuplicateNoteHandler,
	)

	// bulk-tag
	s.AddTool(
		mcp.NewTool("bulk-tag",
			mcp.WithDescription("Add or remove a tag from multiple notes"),
			mcp.WithString("paths",
				mcp.Required(),
				mcp.Description("Comma-separated paths or JSON array of notes"),
			),
			mcp.WithString("tag",
				mcp.Required(),
				mcp.Description("Tag to add/remove (with or without #)"),
			),
			mcp.WithString("action",
				mcp.Description("Action: add, remove (default: add)"),
			),
		),
		v.BulkTagHandler,
	)

	// bulk-move
	s.AddTool(
		mcp.NewTool("bulk-move",
			mcp.WithDescription("Move multiple notes to a folder"),
			mcp.WithString("paths",
				mcp.Required(),
				mcp.Description("Comma-separated paths or JSON array of notes"),
			),
			mcp.WithString("destination",
				mcp.Required(),
				mcp.Description("Destination folder path"),
			),
			mcp.WithBoolean("update_links",
				mcp.Description("Update wikilinks in other notes (default: true)"),
			),
		),
		v.BulkMoveHandler,
	)

	// bulk-set-frontmatter
	s.AddTool(
		mcp.NewTool("bulk-set-frontmatter",
			mcp.WithDescription("Set a frontmatter property on multiple notes"),
			mcp.WithString("paths",
				mcp.Required(),
				mcp.Description("Comma-separated paths or JSON array of notes"),
			),
			mcp.WithString("key",
				mcp.Required(),
				mcp.Description("Frontmatter property key"),
			),
			mcp.WithString("value",
				mcp.Required(),
				mcp.Description("Property value to set"),
			),
		),
		v.BulkSetFrontmatterHandler,
	)
}
