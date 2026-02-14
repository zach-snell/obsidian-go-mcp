package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/zach-snell/obsidian-go-mcp/internal/vault"
)

// New creates a new MCP server configured with vault tools
func New(vaultPath string) *mcp.Server {
	v := vault.New(vaultPath)

	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Obsidian Vault MCP",
			Version: "0.3.3",
		},
		nil,
	)

	// Register tools
	registerTools(s, v)

	return s
}

func registerTools(s *mcp.Server, v *vault.Vault) {
	// list-notes
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list-notes",
		Description: "List all notes in the vault or a specific directory",
	}, v.ListNotesHandler)

	// write-note
	mcp.AddTool(s, &mcp.Tool{
		Name:        "write-note",
		Description: "Create or update a note",
	}, v.WriteNoteHandler)

	// Stubbed / Commented out tools for PoC migration
	// In a full migration, we would uncomment these and ensure the handlers
	// in vault.go are updated to match the generic signature.

	/*
		// read-note
		mcp.AddTool(s, &mcp.Tool{
			Name:        "read-note",
			Description: "Read the content of a specific note",
		}, v.ReadNoteHandler)

		// delete-note
		mcp.AddTool(s, &mcp.Tool{
			Name:        "delete-note",
			Description: "Delete a note",
		}, v.DeleteNoteHandler)

		// ... (and so on for all 70 tools)
	*/
}
