package toolsmarketplace

import (
	"context"
	"log"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/packs"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/sdk-go/plugin"
	"google.golang.org/protobuf/types/known/structpb"
)

// Sender is the interface that the in-process router satisfies.
type Sender interface {
	Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error)
}

// Register adds all 15 marketplace tools and 5 prompts to the builder.
func Register(builder *plugin.PluginBuilder, sender Sender, workspace string) {
	store := storage.NewPackStorage(sender)
	mp := &internal.MarketplacePlugin{
		Storage:   store,
		Workspace: workspace,
	}
	mp.RegisterTools(builder)
	mp.RegisterPrompts(builder)

	// On startup, import .claude/skills/ and .claude/agents/ into Orchestra storage
	// so they appear in the web UI and get synced to cloud.
	go importClaudeSkillsAndAgents(store, workspace)
}

// importClaudeSkillsAndAgents scans .claude/skills/ and .claude/agents/ on startup
// and writes them into Orchestra storage (.skills/{slug}.md / .agents/{slug}.md).
// Always overwrites existing entries so content stays in sync with the filesystem.
func importClaudeSkillsAndAgents(ps *storage.PackStorage, workspace string) {
	ctx := context.Background()

	for _, skill := range packs.ScanClaudeSkills(workspace) {
		path := ".skills/" + skill.Slug + ".md"
		// Read existing entry to get current version (for optimistic lock).
		var version int64
		if existing, err := ps.StorageRead(ctx, path); err == nil {
			version = existing.Version
		}
		meta, err := structpb.NewStruct(map[string]any{
			"name":        skill.Name,
			"slug":        skill.Slug,
			"description": skill.Description,
			"scope":       "personal",
			"created_at":  helpers.NowISO(),
		})
		if err != nil {
			continue
		}
		// Strip SKILL.md frontmatter — store only the markdown body as content.
		body := packs.StripFrontmatter(skill.Content)
		if _, err := ps.StorageWrite(ctx, path, meta, []byte(body), version); err != nil {
			log.Printf("import skill %s: %v", skill.Slug, err)
		}
	}

	for _, agent := range packs.ScanClaudeAgents(workspace) {
		path := ".agents/" + agent.Slug + ".md"
		// Read existing entry to get current version (for optimistic lock).
		var version int64
		if existing, err := ps.StorageRead(ctx, path); err == nil {
			version = existing.Version
		}
		meta, err := structpb.NewStruct(map[string]any{
			"name":        agent.Name,
			"slug":        agent.Slug,
			"description": agent.Description,
			"scope":       "personal",
			"created_at":  helpers.NowISO(),
		})
		if err != nil {
			continue
		}
		// Strip agent .md frontmatter — store only the markdown body as content.
		body := packs.StripFrontmatter(agent.Content)
		if _, err := ps.StorageWrite(ctx, path, meta, []byte(body), version); err != nil {
			log.Printf("import agent %s: %v", agent.Slug, err)
		}
	}
}
