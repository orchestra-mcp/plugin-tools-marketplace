package packs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Stack detection tests ---

func TestDetectStacksGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\n"), 0644)

	stacks := DetectStacks(dir)
	if len(stacks) == 0 {
		t.Fatal("expected at least one stack detected")
	}
	found := false
	for _, s := range stacks {
		if s.Name == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'go' stack to be detected")
	}
}

func TestDetectStacksRust(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"foo\"\n"), 0644)

	stacks := DetectStacks(dir)
	found := false
	for _, s := range stacks {
		if s.Name == "rust" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'rust' stack to be detected")
	}
}

func TestDetectStacksReact(t *testing.T) {
	dir := t.TempDir()
	pkg := map[string]any{
		"name":         "my-app",
		"dependencies": map[string]any{"react": "^18.0.0"},
	}
	data, _ := json.Marshal(pkg)
	os.WriteFile(filepath.Join(dir, "package.json"), data, 0644)

	stacks := DetectStacks(dir)
	found := false
	for _, s := range stacks {
		if s.Name == "react" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'react' stack to be detected")
	}
}

func TestDetectStacksPython(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==2.0\n"), 0644)

	stacks := DetectStacks(dir)
	found := false
	for _, s := range stacks {
		if s.Name == "python" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'python' stack to be detected")
	}
}

func TestDetectStacksDocker(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang:1.21\n"), 0644)

	stacks := DetectStacks(dir)
	found := false
	for _, s := range stacks {
		if s.Name == "docker" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'docker' stack to be detected")
	}
}

func TestDetectStacksEmpty(t *testing.T) {
	dir := t.TempDir()
	stacks := DetectStacks(dir)
	if len(stacks) != 0 {
		t.Errorf("expected no stacks, got %d", len(stacks))
	}
}

func TestDetectStacksMultiple(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module foo\n"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang\n"), 0644)
	os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte("{}"), 0644)

	stacks := DetectStacks(dir)
	if len(stacks) < 3 {
		t.Errorf("expected at least 3 stacks, got %d", len(stacks))
	}
}

// --- Pack index tests ---

func TestRecommendPacksGo(t *testing.T) {
	packs := RecommendPacks([]string{"go"})
	if len(packs) == 0 {
		t.Fatal("expected recommendations for Go stack")
	}
	// Should include essentials (wildcard) and go-backend.
	foundEssentials := false
	foundGo := false
	for _, p := range packs {
		if p.Repo == "github.com/orchestra-mcp/pack-essentials" {
			foundEssentials = true
		}
		if p.Repo == "github.com/orchestra-mcp/pack-go-backend" {
			foundGo = true
		}
	}
	if !foundEssentials {
		t.Error("expected pack-essentials in recommendations")
	}
	if !foundGo {
		t.Error("expected pack-go-backend in recommendations")
	}
}

func TestRecommendPacksUnknownStack(t *testing.T) {
	packs := RecommendPacks([]string{"cobol"})
	// Should still get wildcard packs.
	for _, p := range packs {
		hasWildcard := false
		for _, s := range p.Stacks {
			if s == "*" {
				hasWildcard = true
			}
		}
		if !hasWildcard {
			t.Errorf("pack %s should only be wildcard for unknown stack", p.Repo)
		}
	}
}

func TestSearchPacksGo(t *testing.T) {
	results := SearchPacks("go")
	if len(results) == 0 {
		t.Fatal("expected search results for 'go'")
	}
}

func TestSearchPacksNoResults(t *testing.T) {
	results := SearchPacks("nonexistent-stack-xyz")
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestSearchPacksByTag(t *testing.T) {
	results := SearchPacks("fiber")
	found := false
	for _, p := range results {
		if p.Repo == "github.com/orchestra-mcp/pack-go-backend" {
			found = true
		}
	}
	if !found {
		t.Error("expected pack-go-backend when searching by tag 'fiber'")
	}
}

// --- Installer filesystem tests ---

func TestListInstalledSkillsEmpty(t *testing.T) {
	dir := t.TempDir()
	skills := ListInstalledSkills(dir)
	if len(skills) != 0 {
		t.Errorf("expected no skills, got %d", len(skills))
	}
}

func TestListInstalledSkills(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "my-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# My Skill\n"), 0644)

	skills := ListInstalledSkills(dir)
	if len(skills) != 1 || skills[0] != "my-skill" {
		t.Errorf("expected [my-skill], got %v", skills)
	}
}

func TestListInstalledAgents(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "my-agent.md"), []byte("# My Agent\n"), 0644)

	agents := ListInstalledAgents(dir)
	if len(agents) != 1 || agents[0] != "my-agent" {
		t.Errorf("expected [my-agent], got %v", agents)
	}
}

func TestListInstalledHooks(t *testing.T) {
	dir := t.TempDir()
	hookDir := filepath.Join(dir, ".claude", "hooks")
	os.MkdirAll(hookDir, 0755)
	os.WriteFile(filepath.Join(hookDir, "my-hook.sh"), []byte("#!/bin/bash\n"), 0644)

	hooks := ListInstalledHooks(dir)
	if len(hooks) != 1 || hooks[0] != "my-hook" {
		t.Errorf("expected [my-hook], got %v", hooks)
	}
}

func TestReadSkillContent(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "test-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill\nContent here."), 0644)

	content, err := ReadSkillContent(dir, "test-skill")
	if err != nil {
		t.Fatal(err)
	}
	if content != "# Test Skill\nContent here." {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestReadSkillContentNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadSkillContent(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing skill")
	}
}

func TestReadAgentContent(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".claude", "agents")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "test-agent.md"), []byte("# Test Agent"), 0644)

	content, err := ReadAgentContent(dir, "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if content != "# Test Agent" {
		t.Errorf("unexpected content: %q", content)
	}
}

func TestRemovePack(t *testing.T) {
	dir := t.TempDir()

	// Create skill, agent, hook files.
	skillDir := filepath.Join(dir, ".claude", "skills", "rm-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("skill"), 0644)

	agentDir := filepath.Join(dir, ".claude", "agents")
	os.MkdirAll(agentDir, 0755)
	os.WriteFile(filepath.Join(agentDir, "rm-agent.md"), []byte("agent"), 0644)

	hookDir := filepath.Join(dir, ".claude", "hooks")
	os.MkdirAll(hookDir, 0755)
	os.WriteFile(filepath.Join(hookDir, "rm-hook.sh"), []byte("hook"), 0644)

	err := RemovePack(dir, []string{"rm-skill"}, []string{"rm-agent"}, []string{"rm-hook"})
	if err != nil {
		t.Fatal(err)
	}

	// Verify all removed.
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("skill directory should be removed")
	}
	if _, err := os.Stat(filepath.Join(agentDir, "rm-agent.md")); !os.IsNotExist(err) {
		t.Error("agent file should be removed")
	}
	if _, err := os.Stat(filepath.Join(hookDir, "rm-hook.sh")); !os.IsNotExist(err) {
		t.Error("hook file should be removed")
	}
}

func TestCopyDirRecursive(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "dest")

	// Create nested files.
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(src, "sub", "file2.txt"), []byte("world"), 0644)

	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "file1.txt"))
	if err != nil || string(data) != "hello" {
		t.Error("file1.txt not copied correctly")
	}
	data, err = os.ReadFile(filepath.Join(dst, "sub", "file2.txt"))
	if err != nil || string(data) != "world" {
		t.Error("sub/file2.txt not copied correctly")
	}
}
