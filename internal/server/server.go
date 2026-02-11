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
		"0.1.0",
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
}
