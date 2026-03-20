package packs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PackManifest is the parsed pack.json from a pack repo.
type PackManifest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	License     string   `json:"license"`
	Stacks      []string `json:"stacks"`
	Contents    struct {
		Skills []string `json:"skills"`
		Agents []string `json:"agents"`
		Hooks  []string `json:"hooks"`
	} `json:"contents"`
	Tags []string `json:"tags"`
}

// ResolvePackRepo resolves a short name, org/repo, or full github.com/org/repo
// to the full github.com path used for git clone.
//
// Examples:
//   - "go-backend"                          → "github.com/orchestra-mcp/pack-go-backend"
//   - "orchestra-mcp/pack-go-backend"       → "github.com/orchestra-mcp/pack-go-backend"
//   - "github.com/orchestra-mcp/pack-go"    → "github.com/orchestra-mcp/pack-go" (unchanged)
//   - "github.com/myuser/my-pack"           → "github.com/myuser/my-pack" (unchanged)
func ResolvePackRepo(input string) string {
	// Already a full github.com/... path
	if strings.HasPrefix(input, "github.com/") {
		return input
	}

	// org/repo format (contains exactly one slash, no dots)
	if strings.Count(input, "/") == 1 && !strings.Contains(input, ".") {
		return "github.com/" + input
	}

	// Short name — look up in known packs index
	for _, p := range KnownPacks {
		// Match by short name: extract last segment of repo path (e.g., "pack-go-backend")
		parts := strings.Split(p.Repo, "/")
		repoName := parts[len(parts)-1] // "pack-go-backend"

		// Match against "go-backend" (strip "pack-" prefix) or full "pack-go-backend"
		shortName := strings.TrimPrefix(repoName, "pack-")
		if input == shortName || input == repoName {
			return p.Repo
		}
	}

	// Fallback: assume it's an orchestra-mcp short name
	return "github.com/orchestra-mcp/pack-" + input
}

// InstallPack clones a pack repo and copies its contents to the workspace.
func InstallPack(workspace, repo, version string) (*PackManifest, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, fmt.Errorf("git not found in PATH")
	}

	tmpDir, err := os.MkdirTemp("", "orchestra-pack-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repo.
	cloneURL := "https://" + repo + ".git"
	cloneArgs := []string{"clone", "--depth", "1"}
	if version != "" {
		cloneArgs = append(cloneArgs, "--branch", version)
	}
	cloneArgs = append(cloneArgs, cloneURL, tmpDir)

	cmd := exec.Command("git", cloneArgs...)
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone %s: %w", cloneURL, err)
	}

	// Read and parse pack.json.
	packJSON, err := os.ReadFile(filepath.Join(tmpDir, "pack.json"))
	if err != nil {
		return nil, fmt.Errorf("read pack.json: %w (is this a valid pack repo?)", err)
	}

	var manifest PackManifest
	if err := json.Unmarshal(packJSON, &manifest); err != nil {
		return nil, fmt.Errorf("parse pack.json: %w", err)
	}

	claudeDir := filepath.Join(workspace, ".claude")

	// Copy skills.
	for _, name := range manifest.Contents.Skills {
		src := filepath.Join(tmpDir, "skills", name)
		dst := filepath.Join(claudeDir, "skills", name)
		if err := copyDir(src, dst); err != nil {
			return nil, fmt.Errorf("copy skill %s: %w", name, err)
		}
	}

	// Copy agents.
	for _, name := range manifest.Contents.Agents {
		src := filepath.Join(tmpDir, "agents", name+".md")
		dst := filepath.Join(claudeDir, "agents", name+".md")
		if err := copyFile(src, dst); err != nil {
			return nil, fmt.Errorf("copy agent %s: %w", name, err)
		}
	}

	// Copy hooks.
	for _, name := range manifest.Contents.Hooks {
		src := filepath.Join(tmpDir, "hooks", name+".sh")
		dst := filepath.Join(claudeDir, "hooks", name+".sh")
		if err := copyFile(src, dst); err != nil {
			return nil, fmt.Errorf("copy hook %s: %w", name, err)
		}
		os.Chmod(dst, 0755)
	}

	return &manifest, nil
}

// RemovePack removes installed files for a pack.
func RemovePack(workspace string, skills, agents, hooks []string) error {
	claudeDir := filepath.Join(workspace, ".claude")

	for _, name := range skills {
		os.RemoveAll(filepath.Join(claudeDir, "skills", name))
	}
	for _, name := range agents {
		os.Remove(filepath.Join(claudeDir, "agents", name+".md"))
	}
	for _, name := range hooks {
		os.Remove(filepath.Join(claudeDir, "hooks", name+".sh"))
	}
	return nil
}

// copyDir copies a directory recursively.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file, creating parent directories.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// ListInstalledSkills scans the workspace for installed skills.
func ListInstalledSkills(workspace string) []string {
	dir := filepath.Join(workspace, ".claude", "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			skillFile := filepath.Join(dir, e.Name(), "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				names = append(names, e.Name())
			}
		}
	}
	return names
}

// ListInstalledAgents scans the workspace for installed agents.
func ListInstalledAgents(workspace string) []string {
	dir := filepath.Join(workspace, ".claude", "agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	return names
}

// ListInstalledHooks scans the workspace for installed hooks.
func ListInstalledHooks(workspace string) []string {
	dir := filepath.Join(workspace, ".claude", "hooks")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sh") {
			names = append(names, strings.TrimSuffix(e.Name(), ".sh"))
		}
	}
	return names
}

// ReadSkillContent reads a skill's SKILL.md file content.
func ReadSkillContent(workspace, name string) (string, error) {
	path := filepath.Join(workspace, ".claude", "skills", name, "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("skill %q not found", name)
	}
	return string(data), nil
}

// ReadAgentContent reads an agent's markdown file content.
func ReadAgentContent(workspace, name string) (string, error) {
	path := filepath.Join(workspace, ".claude", "agents", name+".md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("agent %q not found", name)
	}
	return string(data), nil
}

// ClaudeSkillInfo holds parsed metadata from a .claude/skills/*/SKILL.md file.
type ClaudeSkillInfo struct {
	Slug        string
	Name        string
	Description string
	Content     string
}

// ClaudeAgentInfo holds parsed metadata from a .claude/agents/*.md file.
type ClaudeAgentInfo struct {
	Slug        string
	Name        string
	Description string
	Content     string
}

// ScanClaudeSkills scans .claude/skills/ and returns all parseable skills.
func ScanClaudeSkills(workspace string) []ClaudeSkillInfo {
	dir := filepath.Join(workspace, ".claude", "skills")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var result []ClaudeSkillInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slug := e.Name()
		skillFile := filepath.Join(dir, slug, "SKILL.md")
		data, err := os.ReadFile(skillFile)
		if err != nil {
			continue
		}
		content := string(data)
		name, desc := parseFrontmatter(content, slug)
		result = append(result, ClaudeSkillInfo{
			Slug:        slug,
			Name:        name,
			Description: desc,
			Content:     content,
		})
	}
	return result
}

// ScanClaudeAgents scans .claude/agents/ and returns all parseable agents.
func ScanClaudeAgents(workspace string) []ClaudeAgentInfo {
	dir := filepath.Join(workspace, ".claude", "agents")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var result []ClaudeAgentInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		content := string(data)
		name, desc := parseFrontmatter(content, slug)
		result = append(result, ClaudeAgentInfo{
			Slug:        slug,
			Name:        name,
			Description: desc,
			Content:     content,
		})
	}
	return result
}

// StripFrontmatter removes YAML frontmatter (---...---) from content,
// returning only the markdown body that follows it.
func StripFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return content
	}
	// Find the closing ---
	rest := content[3:]
	// Skip optional \r\n or \n after opening ---
	if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	} else if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return content
	}
	body := rest[idx+4:] // skip \n---
	// Skip optional \r\n or \n after closing ---
	if strings.HasPrefix(body, "\r\n") {
		body = body[2:]
	} else if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}
	// Skip one more blank line if present
	if strings.HasPrefix(body, "\r\n") {
		body = body[2:]
	} else if strings.HasPrefix(body, "\n") {
		body = body[1:]
	}
	return body
}

// parseFrontmatter extracts name and description from YAML frontmatter.
// Falls back to titleCase(slug) and empty string if not found.
func parseFrontmatter(content, slug string) (name, description string) {
	name = toTitleCase(slug)
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}
		if !inFrontmatter {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			v = strings.Trim(v, `"'`)
			if v != "" {
				name = v
			}
		}
		if strings.HasPrefix(line, "description:") {
			v := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			v = strings.Trim(v, `"'>-`)
			if v != "" {
				description = v
			}
		}
	}
	return name, description
}

// toTitleCase converts a slug like "my-skill" to "My Skill".
func toTitleCase(slug string) string {
	parts := strings.Split(strings.ReplaceAll(slug, "-", " "), " ")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
