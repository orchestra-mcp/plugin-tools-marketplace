# Tools & Prompts Reference

The `tools.marketplace` plugin provides 15 tools across 4 categories and 5 MCP prompts.

All tools accept arguments as a JSON object. Required fields are marked with **(required)**.

---

## Pack Management Tools (6)

### `install_pack`

Install a pack of skills, agents, and hooks from a GitHub repo.

| Param | Type | Required | Description |
|---|---|---|---|
| `repo` | string | yes | GitHub repo path (e.g., `github.com/orchestra-mcp/pack-go-backend`) |
| `version` | string | no | Version tag (defaults to latest) |

Clones the repo, reads `pack.json`, copies skills to `.claude/skills/`, agents to `.claude/agents/`, hooks to `.claude/hooks/`, and updates the pack registry.

### `remove_pack`

Remove an installed pack and its contents.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Pack name (e.g., `orchestra-mcp/pack-go-backend`) |

Removes all skills, agents, and hooks that were installed by the pack, then removes the pack from the registry.

### `update_pack`

Update an installed pack to the latest version.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | no | Pack name to update (omit to update all installed packs) |

Removes old files and re-clones from the original repo. Updates the registry with new version info.

### `list_packs`

List all installed packs.

| Param | Type | Required | Description |
|---|---|---|---|
| `type` | string | no | Filter by content type: `skills`, `agents`, or `hooks` |

Returns a markdown table with pack names, versions, and content counts.

### `get_pack`

Get details of an installed pack.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Pack name |

Returns version, repo URL, install date, stacks, and lists of included skills, agents, and hooks.

### `search_packs`

Search available packs by keyword or stack.

| Param | Type | Required | Description |
|---|---|---|---|
| `query` | string | yes | Search keyword (matches repo name, description, and tags) |
| `stack` | string | no | Filter results by technology stack |

Searches the built-in index of 17 known packs.

---

## Recommendation Tools (2)

### `detect_stacks`

Detect the project's technology stacks. No parameters.

Scans the workspace for stack indicators (e.g., `go.mod` for Go, `Cargo.toml` for Rust, `package.json` dependencies for React). Returns a markdown table of detected stacks with evidence.

Supported stacks: go, rust, react, typescript, python, ruby, java, kotlin, swift, csharp, php, docker.

### `recommend_packs`

Recommend packs based on detected technology stacks.

| Param | Type | Required | Description |
|---|---|---|---|
| `stacks` | string[] | no | Override detected stacks (auto-detects if omitted) |

Resolution order for stacks: (1) explicit `stacks` argument, (2) configured stacks via `set_project_stacks`, (3) auto-detected stacks. Returns a table of recommended packs with install status.

---

## Content Query Tools (5)

### `list_skills`

List all installed skills. No parameters.

Scans `.claude/skills/` for directories containing a `SKILL.md` file.

### `list_agents`

List all installed agents. No parameters.

Scans `.claude/agents/` for `.md` files.

### `list_hooks`

List all installed hooks. No parameters.

Scans `.claude/hooks/` for `.sh` files.

### `get_skill`

Read a skill's full content.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Skill name (directory name under `.claude/skills/`) |

Returns the full text of `.claude/skills/<name>/SKILL.md`.

### `get_agent`

Read an agent's full content.

| Param | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Agent name (filename without `.md` under `.claude/agents/`) |

Returns the full text of `.claude/agents/<name>.md`.

---

## Configuration Tools (2)

### `set_project_stacks`

Manually set the project's technology stacks.

| Param | Type | Required | Description |
|---|---|---|---|
| `stacks` | string[] | yes | Technology stacks (e.g., `["go", "react", "docker"]`) |

Persists to storage. Overrides auto-detection for `recommend_packs` and `get_project_stacks`.

### `get_project_stacks`

Get the project's detected or configured stacks. No parameters.

Returns configured stacks if set via `set_project_stacks`, otherwise falls back to auto-detection.

---

---

## MCP Prompts (5)

Prompts are parameter-driven message templates available via `prompts/list` and `prompts/get`. They return structured role-based messages that guide the AI through multi-step workflows.

### `setup-project`

Guide setting up a new project: detect stacks, recommend packs, install essentials.

| Param | Type | Required | Description |
|---|---|---|---|
| `project_name` | string | yes | Name of the project to set up |

Detects technology stacks in the workspace, recommends matching packs, checks which are already installed, and returns step-by-step guidance for project setup.

### `recommend-packs`

Return pack recommendations based on detected or specified stacks.

| Param | Type | Required | Description |
|---|---|---|---|
| `stacks` | string | no | Comma-separated list of stacks (auto-detects if empty) |

Resolution order: (1) explicit `stacks` argument, (2) configured stacks, (3) auto-detected stacks. Returns recommended packs with install status.

### `audit-packs`

Audit installed packs: versions, contents, and totals. No parameters.

Returns a summary of all installed packs including version, repo, install date, lists of skills/agents/hooks, and aggregate totals.

### `search-marketplace`

Search and display available packs with descriptions.

| Param | Type | Required | Description |
|---|---|---|---|
| `query` | string | yes | Search keyword |
| `stack` | string | no | Filter by technology stack |

Searches the built-in index of 17 known packs and returns matching results with install instructions.

### `onboard-project`

Full onboarding: create project, detect stacks, install packs, configure workspace.

| Param | Type | Required | Description |
|---|---|---|---|
| `project_name` | string | yes | Name of the project |
| `description` | string | no | Brief project description |

Returns a numbered step-by-step guide: create project, set stacks, install recommended packs, verify installation, start working. Adapts recommendations to detected stacks.

---

## Storage

Pack metadata is stored in `.projects/.packs/registry.json` via the `storage.markdown` plugin over QUIC. Content files (skills, agents, hooks) are installed directly to the `.claude/` directory on the filesystem.

## Registry Format

```json
{
  "packs": {
    "orchestra-mcp/pack-go-backend": {
      "version": "0.1.0",
      "repo": "github.com/orchestra-mcp/pack-go-backend",
      "installed_at": "2026-02-27T12:00:00Z",
      "stacks": ["go"],
      "skills": ["go-backend"],
      "agents": ["go-architect"],
      "hooks": []
    }
  }
}
```
