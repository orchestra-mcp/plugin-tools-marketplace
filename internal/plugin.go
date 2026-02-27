package internal

import (
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/tools"
)

// MarketplacePlugin holds the shared dependencies for all tool handlers.
type MarketplacePlugin struct {
	Storage   *storage.PackStorage
	Workspace string
}

// RegisterTools registers all 15 marketplace tools with the plugin builder.
func (mp *MarketplacePlugin) RegisterTools(builder *plugin.PluginBuilder) {
	ps := mp.Storage
	ws := mp.Workspace

	// --- Pack management (6) ---
	builder.RegisterTool("install_pack",
		"Install a pack of skills, agents, and hooks from a GitHub repo",
		tools.InstallPackSchema(), tools.InstallPack(ps, ws))
	builder.RegisterTool("remove_pack",
		"Remove an installed pack and its contents",
		tools.RemovePackSchema(), tools.RemovePack(ps, ws))
	builder.RegisterTool("update_pack",
		"Update an installed pack to the latest version",
		tools.UpdatePackSchema(), tools.UpdatePack(ps, ws))
	builder.RegisterTool("list_packs",
		"List all installed packs",
		tools.ListPacksSchema(), tools.ListPacks(ps))
	builder.RegisterTool("get_pack",
		"Get details of an installed pack",
		tools.GetPackSchema(), tools.GetPack(ps))
	builder.RegisterTool("search_packs",
		"Search available packs by keyword or stack",
		tools.SearchPacksSchema(), tools.SearchPacks(ps))

	// --- Recommendations (2) ---
	builder.RegisterTool("detect_stacks",
		"Detect the project's technology stacks",
		tools.DetectStacksSchema(), tools.DetectStacks(ws))
	builder.RegisterTool("recommend_packs",
		"Recommend packs based on detected technology stacks",
		tools.RecommendPacksSchema(), tools.RecommendPacks(ps, ws))

	// --- Content queries (5) ---
	builder.RegisterTool("list_skills",
		"List all installed skills",
		tools.ListSkillsSchema(), tools.ListSkills(ws))
	builder.RegisterTool("list_agents",
		"List all installed agents",
		tools.ListAgentsSchema(), tools.ListAgents(ws))
	builder.RegisterTool("list_hooks",
		"List all installed hooks",
		tools.ListHooksSchema(), tools.ListHooks(ws))
	builder.RegisterTool("get_skill",
		"Read a skill's full content",
		tools.GetSkillSchema(), tools.GetSkill(ws))
	builder.RegisterTool("get_agent",
		"Read an agent's full content",
		tools.GetAgentSchema(), tools.GetAgent(ws))

	// --- Configuration (2) ---
	builder.RegisterTool("set_project_stacks",
		"Manually set the project's technology stacks",
		tools.SetProjectStacksSchema(), tools.SetProjectStacks(ps))
	builder.RegisterTool("get_project_stacks",
		"Get the project's detected or configured stacks",
		tools.GetProjectStacksSchema(), tools.GetProjectStacks(ps, ws))
}

// RegisterPrompts registers all 5 marketplace prompts with the plugin builder.
func (mp *MarketplacePlugin) RegisterPrompts(builder *plugin.PluginBuilder) {
	ps := mp.Storage
	ws := mp.Workspace

	builder.RegisterPrompt("setup-project",
		"Guide setting up a new project: detect stacks, recommend packs, install essentials",
		tools.SetupProjectArgs(), tools.SetupProject(ps, ws))
	builder.RegisterPrompt("recommend-packs",
		"Return pack recommendations based on detected or specified stacks",
		tools.RecommendPacksPromptArgs(), tools.RecommendPacksPrompt(ps, ws))
	builder.RegisterPrompt("audit-packs",
		"Audit installed packs: versions, updates available, contents summary",
		tools.AuditPacksArgs(), tools.AuditPacks(ps, ws))
	builder.RegisterPrompt("search-marketplace",
		"Search and display available packs with descriptions",
		tools.SearchMarketplaceArgs(), tools.SearchMarketplace())
	builder.RegisterPrompt("onboard-project",
		"Full onboarding: create project, detect stacks, install packs, configure workspace",
		tools.OnboardProjectArgs(), tools.OnboardProject(ps, ws))
}
