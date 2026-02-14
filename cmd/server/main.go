package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpserver "github.com/zach-snell/obsidian-go-mcp/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: obsidian-mcp <vault-path>")
	}

	vaultPath := os.Args[1]

	// Verify vault path exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		log.Fatalf("Vault path does not exist: %s", vaultPath)
	}

	fmt.Fprintf(os.Stderr, "Starting Obsidian MCP Server...\n")
	fmt.Fprintf(os.Stderr, "Vault: %s\n", vaultPath)

	// Create and configure MCP server
	s := mcpserver.New(vaultPath)

	// Start stdio server
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
