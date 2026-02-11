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
		"0.2.0",
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
			mcp.WithDescription("Append content to a note (creates if doesn't exist)"),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Path to the note (.md extension required)"),
			),
			mcp.WithString("content",
				mcp.Required(),
				mcp.Description("Content to append"),
			),
		),
		v.AppendNoteHandler,
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
}
