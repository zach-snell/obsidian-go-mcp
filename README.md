# obsidian-go-mcp

A Go-based MCP (Model Context Protocol) server for Obsidian vaults.

## Features

- **CRUD Operations**: List, read, write, delete notes
- **Search**: Grep-style content search
- **Task Parsing**: Extract and filter checkboxes with metadata
- **Tag Search**: Find notes by tags (AND operation)
- **MOC Discovery**: Find Maps of Content structure

## Quick Start

```bash
# Install dependencies
mise install

# Build
mise run build

# Run with vault
./bin/obsidian-mcp /path/to/vault
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `list-notes` | List markdown files |
| `read-note` | Read note content |
| `write-note` | Create/update note |
| `delete-note` | Delete note |
| `search-vault` | Content search |
| `list-tasks` | Parse checkboxes |
| `search-by-tags` | Tag-based search |
| `discover-mocs` | Find MOC structure |

## Task Format

Compatible with Obsidian Tasks plugin:

```markdown
- [ ] Open task
- [x] Completed task
- [ ] Due date 📅 2024-01-15
- [ ] High priority ⏫
- [ ] With tags #project #urgent
```

## Development

```bash
mise run lint    # Linting
mise run test    # Tests
mise run check   # All checks
```

## License

MIT
