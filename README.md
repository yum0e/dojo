# Dojo

A TUI for orchestrating AI agents across jj workspaces.

## Quick Start

```bash
# Run from any jj repository
go run ./cmd/dojo

# Or build first
go build -o dojo ./cmd/dojo
./dojo
```

## Requirements

- Go 1.24+
- [Jujutsu (jj)](https://github.com/martinvonz/jj) installed and in PATH
- Must be run from inside a jj repository

## Development

```bash
# Run tests
go test ./...

# Build
go build ./...
```
