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
