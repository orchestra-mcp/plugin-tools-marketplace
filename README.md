# Orchestra Tools Marketplace Plugin

Marketplace plugin providing 15 tools for managing installable packs of skills, agents, and hooks from GitHub repositories.

## Install

```bash
go get github.com/orchestra-mcp/plugin-tools-marketplace
```

## Usage

```bash
# Build
go build -o bin/tools-marketplace ./cmd/

# Run (started automatically by the orchestrator)
bin/tools-marketplace --workspace=. --orchestrator-addr localhost:9100
```

Add to your `plugins.yaml`:

```yaml
- id: tools.marketplace
  binary: ./bin/tools-marketplace
  enabled: true
  args:
    - --workspace=.
```

## Tools (15)

Organized into 4 categories:

| Category | Tools |
|----------|-------|
| **Pack Management** | `install_pack`, `remove_pack`, `update_pack`, `list_packs`, `get_pack`, `search_packs` |
| **Recommendations** | `detect_stacks`, `recommend_packs` |
| **Content Queries** | `list_skills`, `list_agents`, `list_hooks`, `get_skill`, `get_agent` |
| **Configuration** | `set_project_stacks`, `get_project_stacks` |

## Pack Format

Packs are GitHub repos with the following structure:

```
pack-go-backend/
  pack.json              # Pack manifest (required)
  skills/
    go-backend/
      SKILL.md           # Skill definition
  agents/
    go-architect.md      # Agent definition
  hooks/
    my-hook.sh           # Hook script (optional)
```

### pack.json

```json
{
  "name": "orchestra-mcp/pack-go-backend",
  "description": "Go backend skills and agents",
  "version": "0.1.0",
  "stacks": ["go"],
  "contents": {
    "skills": ["go-backend"],
    "agents": ["go-architect"],
    "hooks": []
  },
  "tags": ["go", "fiber", "gorm", "backend"]
}
```

## Stack Detection

The plugin auto-detects technology stacks from workspace files:

| Stack | Detection |
|-------|-----------|
| go | `go.mod` or `go.work` |
| rust | `Cargo.toml` |
| react | `react` in package.json dependencies |
| typescript | `tsconfig.json` |
| python | `pyproject.toml`, `requirements.txt`, or `setup.py` |
| ruby | `Gemfile` |
| java | `pom.xml` or `build.gradle` |
| kotlin | `build.gradle.kts` |
| swift | `Package.swift` or `*.xcodeproj` |
| csharp | `*.csproj` or `*.sln` |
| php | `composer.json` |
| docker | `Dockerfile` or `docker-compose.yml` |

## CLI Commands

The `orchestra` CLI includes pack management:

```bash
orchestra pack install <repo>[@version]   # Install a pack from GitHub
orchestra pack remove <name>              # Remove an installed pack
orchestra pack update [name]              # Update one or all packs
orchestra pack list                       # List installed packs
orchestra pack search <query>             # Search available packs
orchestra pack recommend                  # Detect stacks & recommend packs
```

## Known Packs

17 official packs are indexed for recommendation:

| Pack | Stacks | Description |
|------|--------|-------------|
| [pack-essentials](https://github.com/orchestra-mcp/pack-essentials) | * | Core project management skills and agents |
| [pack-go-backend](https://github.com/orchestra-mcp/pack-go-backend) | go | Go backend skills (Fiber, GORM, REST) |
| [pack-rust-engine](https://github.com/orchestra-mcp/pack-rust-engine) | rust | Rust engine skills (Tonic, Tree-sitter, Tantivy) |
| [pack-react-frontend](https://github.com/orchestra-mcp/pack-react-frontend) | react, typescript | React frontend skills (Zustand, shadcn/ui) |
| [pack-database](https://github.com/orchestra-mcp/pack-database) | * | Database skills (PostgreSQL, SQLite, Redis) |
| [pack-ai](https://github.com/orchestra-mcp/pack-ai) | * | AI/LLM integration skills |
| [pack-mobile](https://github.com/orchestra-mcp/pack-mobile) | react-native | React Native mobile skills |
| [pack-desktop](https://github.com/orchestra-mcp/pack-desktop) | go | Desktop app skills (Wails, macOS) |
| [pack-extensions](https://github.com/orchestra-mcp/pack-extensions) | * | Extension system skills |
| [pack-chrome](https://github.com/orchestra-mcp/pack-chrome) | typescript | Chrome extension skills |
| [pack-infra](https://github.com/orchestra-mcp/pack-infra) | docker | Infrastructure and DevOps skills |
| [pack-proto](https://github.com/orchestra-mcp/pack-proto) | go, rust | Protobuf/gRPC skills |
| [pack-native-swift](https://github.com/orchestra-mcp/pack-native-swift) | swift | Swift/macOS/iOS plugin skills |
| [pack-native-kotlin](https://github.com/orchestra-mcp/pack-native-kotlin) | kotlin, java | Kotlin/Android plugin skills |
| [pack-native-csharp](https://github.com/orchestra-mcp/pack-native-csharp) | csharp | C#/Windows plugin skills |
| [pack-native-gtk](https://github.com/orchestra-mcp/pack-native-gtk) | c | GTK4/Linux desktop skills |
| [pack-analytics](https://github.com/orchestra-mcp/pack-analytics) | * | ClickHouse analytics skills |

## Related Packages

| Package | Description |
|---------|-------------|
| [sdk-go](https://github.com/orchestra-mcp/sdk-go) | Plugin SDK this plugin is built on |
| [orchestrator](https://github.com/orchestra-mcp/orchestrator) | Central hub that loads this plugin |
| [plugin-storage-markdown](https://github.com/orchestra-mcp/plugin-storage-markdown) | Storage backend for pack registry |

## License

[MIT](LICENSE)
