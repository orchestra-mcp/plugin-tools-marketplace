package packs

import (
	"strings"
)

// PackInfo describes a known pack available for installation.
type PackInfo struct {
	Repo        string
	Stacks      []string
	Description string
	Tags        []string
}

// KnownPacks is the built-in index of available packs.
var KnownPacks = []PackInfo{
	{Repo: "github.com/orchestra-mcp/pack-essentials", Stacks: []string{"*"}, Description: "Core project management skills and agents", Tags: []string{"core", "essential"}},
	{Repo: "github.com/orchestra-mcp/pack-go-backend", Stacks: []string{"go"}, Description: "Go backend skills (Fiber, GORM, REST)", Tags: []string{"go", "backend", "fiber", "gorm"}},
	{Repo: "github.com/orchestra-mcp/pack-rust-engine", Stacks: []string{"rust"}, Description: "Rust engine skills (Tonic, Tree-sitter, Tantivy)", Tags: []string{"rust", "engine", "grpc"}},
	{Repo: "github.com/orchestra-mcp/pack-react-frontend", Stacks: []string{"react", "typescript"}, Description: "React frontend skills (Zustand, shadcn/ui)", Tags: []string{"react", "typescript", "frontend"}},
	{Repo: "github.com/orchestra-mcp/pack-database", Stacks: []string{"*"}, Description: "Database skills (PostgreSQL, SQLite, Redis)", Tags: []string{"database", "sql", "redis"}},
	{Repo: "github.com/orchestra-mcp/pack-ai", Stacks: []string{"*"}, Description: "AI/LLM integration skills", Tags: []string{"ai", "llm", "rag", "embeddings"}},
	{Repo: "github.com/orchestra-mcp/pack-mobile", Stacks: []string{"react-native"}, Description: "React Native mobile skills", Tags: []string{"mobile", "ios", "android"}},
	{Repo: "github.com/orchestra-mcp/pack-desktop", Stacks: []string{"go"}, Description: "Desktop app skills (Wails, macOS)", Tags: []string{"desktop", "wails", "macos"}},
	{Repo: "github.com/orchestra-mcp/pack-extensions", Stacks: []string{"*"}, Description: "Extension system skills", Tags: []string{"extensions", "marketplace"}},
	{Repo: "github.com/orchestra-mcp/pack-chrome", Stacks: []string{"typescript"}, Description: "Chrome extension skills", Tags: []string{"chrome", "extension", "browser"}},
	{Repo: "github.com/orchestra-mcp/pack-infra", Stacks: []string{"docker"}, Description: "Infrastructure and DevOps skills", Tags: []string{"docker", "gcp", "ci", "devops"}},
	{Repo: "github.com/orchestra-mcp/pack-proto", Stacks: []string{"go", "rust"}, Description: "Protobuf/gRPC skills", Tags: []string{"proto", "grpc", "buf"}},
	{Repo: "github.com/orchestra-mcp/pack-native-swift", Stacks: []string{"swift"}, Description: "Swift/macOS/iOS plugin skills", Tags: []string{"swift", "macos", "ios"}},
	{Repo: "github.com/orchestra-mcp/pack-native-kotlin", Stacks: []string{"kotlin", "java"}, Description: "Kotlin/Android plugin skills", Tags: []string{"kotlin", "android"}},
	{Repo: "github.com/orchestra-mcp/pack-native-csharp", Stacks: []string{"csharp"}, Description: "C#/Windows plugin skills", Tags: []string{"csharp", "windows"}},
	{Repo: "github.com/orchestra-mcp/pack-native-gtk", Stacks: []string{"c"}, Description: "GTK4/Linux desktop skills", Tags: []string{"gtk", "linux"}},
	{Repo: "github.com/orchestra-mcp/pack-analytics", Stacks: []string{"*"}, Description: "ClickHouse analytics skills", Tags: []string{"analytics", "clickhouse"}},
}

// RecommendPacks returns packs matching the given stacks.
func RecommendPacks(stacks []string) []PackInfo {
	stackSet := make(map[string]bool)
	for _, s := range stacks {
		stackSet[s] = true
	}

	var result []PackInfo
	for _, p := range KnownPacks {
		for _, ps := range p.Stacks {
			if ps == "*" || stackSet[ps] {
				result = append(result, p)
				break
			}
		}
	}
	return result
}

// SearchPacks searches known packs by query string.
func SearchPacks(query string) []PackInfo {
	q := strings.ToLower(query)
	var result []PackInfo
	for _, p := range KnownPacks {
		if matchesPack(p, q) {
			result = append(result, p)
		}
	}
	return result
}

func matchesPack(p PackInfo, query string) bool {
	if strings.Contains(strings.ToLower(p.Repo), query) {
		return true
	}
	if strings.Contains(strings.ToLower(p.Description), query) {
		return true
	}
	for _, tag := range p.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	for _, stack := range p.Stacks {
		if strings.Contains(strings.ToLower(stack), query) {
			return true
		}
	}
	return false
}
