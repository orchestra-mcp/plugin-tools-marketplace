# Contributing to plugin-tools-marketplace

## Prerequisites

- Go 1.23+
- `gofmt`, `go vet`
- `git` (required for pack installation)

## Development Setup

```bash
git clone https://github.com/orchestra-mcp/plugin-tools-marketplace.git
cd plugin-tools-marketplace
go mod download
go build ./cmd/...
```

## Running Locally

```bash
go build -o tools-marketplace ./cmd/
./tools-marketplace --workspace=. --orchestrator-addr=localhost:50100 --certs-dir=~/.orchestra/certs
```

The plugin connects to the orchestrator as a client and also starts its own QUIC server for incoming requests. The `--workspace` flag sets the root directory for stack detection and content file operations.

## Running Tests

```bash
go test ./... -v
```

Tests cover stack detection, pack index search/recommend, filesystem operations (list/read skills/agents/hooks, copy directories, remove packs). No running orchestrator is required for tests.

## Code Organization

```
plugin-tools-marketplace/
  cmd/main.go                    # Entry point with --workspace flag
  internal/
    plugin.go                    # MarketplacePlugin: RegisterTools wires all 15 tools
    storage/
      client.go                  # PackStorage: QUIC-backed registry and stacks storage
    packs/
      installer.go               # Git clone + file copy, list/read installed content
      stacks.go                  # Stack detection (12 rules)
      index.go                   # Known packs index (17 packs), recommend, search
      packs_test.go              # 21 unit tests
    tools/
      pack.go                    # install_pack, remove_pack, update_pack, list_packs, get_pack, search_packs
      recommend.go               # detect_stacks, recommend_packs
      content.go                 # list_skills, list_agents, list_hooks, get_skill, get_agent
      config.go                  # set_project_stacks, get_project_stacks
```

## Adding a New Tool

1. Create the schema function and handler function in the appropriate file under `internal/tools/`.
2. Register the tool in `internal/plugin.go` via `builder.RegisterTool(...)`.
3. Add a test case in `internal/packs/packs_test.go` or create a new test file.
4. Update `docs/TOOLS_REFERENCE.md`.

## Adding a New Stack Detection Rule

1. Add a new entry to the `checks` slice in `internal/packs/stacks.go`.
2. Use `checkFileExists`, `checkAnyFileExists`, `checkPackageJSONDep`, or write a custom function.
3. Add a matching entry in `libs/cli/internal/detect.go` for CLI consistency.
4. Add a test in `internal/packs/packs_test.go`.

## Adding a New Known Pack

1. Add an entry to the `KnownPacks` slice in `internal/packs/index.go`.
2. Update the matching list in `libs/cli/internal/pack.go` (search and recommend functions).
3. Update the Known Packs table in `README.md`.

## Code Style

- Run `gofmt` on all files.
- Run `go vet ./...` before committing.
- All exported functions and types must have doc comments.
- Use `helpers.ValidateRequired` for argument validation.
- Use `helpers.TextResult` / `helpers.ErrorResult` for building responses.
- Never return a Go error from a tool handler for expected failures (validation errors, not-found). Use `helpers.ErrorResult` instead. Reserve Go errors for unexpected infrastructure failures.

## Pull Request Process

1. Fork the repository and create a feature branch from `main`.
2. Write or update tests for your changes.
3. Run `go test ./...` and `go vet ./...`.
4. Update `docs/TOOLS_REFERENCE.md` if applicable.

## Related Repositories

- [orchestra-mcp/proto](https://github.com/orchestra-mcp/proto) -- Protobuf schema
- [orchestra-mcp/sdk-go](https://github.com/orchestra-mcp/sdk-go) -- Go Plugin SDK
- [orchestra-mcp/orchestrator](https://github.com/orchestra-mcp/orchestrator) -- Central hub
- [orchestra-mcp/plugin-storage-markdown](https://github.com/orchestra-mcp/plugin-storage-markdown) -- Storage backend
- [orchestra-mcp/plugin-tools-features](https://github.com/orchestra-mcp/plugin-tools-features) -- Feature workflow plugin
- [orchestra-mcp](https://github.com/orchestra-mcp) -- Organization home
