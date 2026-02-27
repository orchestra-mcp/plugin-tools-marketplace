package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/packs"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
)

// PromptHandler is an alias for readability.
type PromptHandler = func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error)

// --- setup-project ---

func SetupProjectArgs() []*pluginv1.PromptArgument {
	return []*pluginv1.PromptArgument{
		{Name: "project_name", Description: "Name of the project to set up", Required: true},
	}
}

func SetupProject(ps *storage.PackStorage, workspace string) PromptHandler {
	return func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error) {
		projectName := req.Arguments["project_name"]
		if projectName == "" {
			projectName = "my-project"
		}

		detected := packs.DetectStacks(workspace)
		recommended := make([]packs.PackInfo, 0)
		stackNames := make([]string, 0, len(detected))
		for _, s := range detected {
			stackNames = append(stackNames, s.Name)
		}
		if len(stackNames) > 0 {
			recommended = packs.RecommendPacks(stackNames)
		}

		// Check installed packs.
		reg, _, _ := ps.ReadRegistry(ctx)
		installed := make(map[string]bool)
		for name := range reg.Packs {
			installed[name] = true
		}

		var b strings.Builder
		fmt.Fprintf(&b, "Set up project '%s'.\n\n", projectName)

		if len(detected) > 0 {
			fmt.Fprintf(&b, "Detected stacks: %s\n\n", strings.Join(stackNames, ", "))
		} else {
			fmt.Fprintf(&b, "No stacks detected. You may want to run `detect_stacks` first.\n\n")
		}

		if len(recommended) > 0 {
			fmt.Fprintf(&b, "Recommended packs:\n")
			for _, p := range recommended {
				name := strings.TrimPrefix(p.Repo, "github.com/")
				status := ""
				if installed[name] {
					status = " (already installed)"
				}
				fmt.Fprintf(&b, "- %s — %s%s\n", name, p.Description, status)
			}
			fmt.Fprintf(&b, "\nPlease install the recommended packs using `install_pack` for each one, ")
			fmt.Fprintf(&b, "then create the project using `create_project`.")
		} else {
			fmt.Fprintf(&b, "No specific packs recommended. Consider installing pack-essentials for core skills.\n")
			fmt.Fprintf(&b, "Then create the project using `create_project`.")
		}

		return &pluginv1.PromptGetResponse{
			Description: "Set up a new project with recommended packs",
			Messages: []*pluginv1.PromptMessage{
				{
					Role:    "user",
					Content: &pluginv1.ContentBlock{Type: "text", Text: b.String()},
				},
			},
		}, nil
	}
}

// --- recommend-packs ---

func RecommendPacksPromptArgs() []*pluginv1.PromptArgument {
	return []*pluginv1.PromptArgument{
		{Name: "stacks", Description: "Comma-separated list of stacks (optional, auto-detects if empty)", Required: false},
	}
}

func RecommendPacksPrompt(ps *storage.PackStorage, workspace string) PromptHandler {
	return func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error) {
		stacksStr := req.Arguments["stacks"]

		var stackNames []string
		if stacksStr != "" {
			for _, s := range strings.Split(stacksStr, ",") {
				s = strings.TrimSpace(s)
				if s != "" {
					stackNames = append(stackNames, s)
				}
			}
		}

		if len(stackNames) == 0 {
			configured, _, _ := ps.ReadStacks(ctx)
			if len(configured) > 0 {
				stackNames = configured
			} else {
				detected := packs.DetectStacks(workspace)
				for _, s := range detected {
					stackNames = append(stackNames, s.Name)
				}
			}
		}

		var b strings.Builder
		if len(stackNames) == 0 {
			fmt.Fprintf(&b, "No technology stacks detected or configured.\n\n")
			fmt.Fprintf(&b, "Please either:\n")
			fmt.Fprintf(&b, "- Use `set_project_stacks` to manually configure stacks\n")
			fmt.Fprintf(&b, "- Use `search_packs` to browse all available packs\n")
		} else {
			recommended := packs.RecommendPacks(stackNames)
			reg, _, _ := ps.ReadRegistry(ctx)
			installed := make(map[string]bool)
			for name := range reg.Packs {
				installed[name] = true
			}

			fmt.Fprintf(&b, "Based on detected stacks (%s), here are the recommended packs:\n\n", strings.Join(stackNames, ", "))
			for _, p := range recommended {
				name := strings.TrimPrefix(p.Repo, "github.com/")
				status := "not installed"
				if installed[name] {
					status = "installed"
				}
				fmt.Fprintf(&b, "- **%s** [%s] — %s (%s)\n", name, strings.Join(p.Stacks, ", "), p.Description, status)
			}
			fmt.Fprintf(&b, "\nInstall any pack using `install_pack` with the full repo path.")
		}

		return &pluginv1.PromptGetResponse{
			Description: "Recommend packs based on detected or specified stacks",
			Messages: []*pluginv1.PromptMessage{
				{
					Role:    "user",
					Content: &pluginv1.ContentBlock{Type: "text", Text: b.String()},
				},
			},
		}, nil
	}
}

// --- audit-packs ---

func AuditPacksArgs() []*pluginv1.PromptArgument {
	return nil
}

func AuditPacks(ps *storage.PackStorage, workspace string) PromptHandler {
	return func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error) {
		reg, _, _ := ps.ReadRegistry(ctx)

		var b strings.Builder
		if len(reg.Packs) == 0 {
			fmt.Fprintf(&b, "No packs are currently installed.\n\n")
			fmt.Fprintf(&b, "Use `recommend_packs` to get suggestions based on your project's stacks, ")
			fmt.Fprintf(&b, "or `search_packs` to browse the marketplace.")
		} else {
			fmt.Fprintf(&b, "Audit of %d installed pack(s):\n\n", len(reg.Packs))

			totalSkills := 0
			totalAgents := 0
			totalHooks := 0

			for name, info := range reg.Packs {
				fmt.Fprintf(&b, "### %s (v%s)\n", name, info.Version)
				fmt.Fprintf(&b, "- Repo: %s\n", info.Repo)
				fmt.Fprintf(&b, "- Installed: %s\n", info.InstalledAt)
				fmt.Fprintf(&b, "- Skills: %s\n", strings.Join(info.Skills, ", "))
				fmt.Fprintf(&b, "- Agents: %s\n", strings.Join(info.Agents, ", "))
				if len(info.Hooks) > 0 {
					fmt.Fprintf(&b, "- Hooks: %s\n", strings.Join(info.Hooks, ", "))
				}
				fmt.Fprintf(&b, "\n")

				totalSkills += len(info.Skills)
				totalAgents += len(info.Agents)
				totalHooks += len(info.Hooks)
			}

			fmt.Fprintf(&b, "**Totals:** %d skills, %d agents, %d hooks across %d packs\n\n",
				totalSkills, totalAgents, totalHooks, len(reg.Packs))
			fmt.Fprintf(&b, "Use `update_pack` to check for updates, or `remove_pack` to uninstall.")
		}

		return &pluginv1.PromptGetResponse{
			Description: "Audit installed packs: versions, contents, and totals",
			Messages: []*pluginv1.PromptMessage{
				{
					Role:    "user",
					Content: &pluginv1.ContentBlock{Type: "text", Text: b.String()},
				},
			},
		}, nil
	}
}

// --- search-marketplace ---

func SearchMarketplaceArgs() []*pluginv1.PromptArgument {
	return []*pluginv1.PromptArgument{
		{Name: "query", Description: "Search keyword", Required: true},
		{Name: "stack", Description: "Filter by technology stack (optional)", Required: false},
	}
}

func SearchMarketplace() PromptHandler {
	return func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error) {
		query := req.Arguments["query"]
		stack := req.Arguments["stack"]

		if query == "" {
			query = "*"
		}

		results := packs.SearchPacks(query)

		// Filter by stack if specified.
		if stack != "" {
			filtered := make([]packs.PackInfo, 0)
			for _, p := range results {
				for _, s := range p.Stacks {
					if s == "*" || strings.EqualFold(s, stack) {
						filtered = append(filtered, p)
						break
					}
				}
			}
			results = filtered
		}

		var b strings.Builder
		if len(results) == 0 {
			fmt.Fprintf(&b, "No packs found matching query '%s'", query)
			if stack != "" {
				fmt.Fprintf(&b, " for stack '%s'", stack)
			}
			fmt.Fprintf(&b, ".\n\nTry a broader search or browse all packs with `list_packs`.")
		} else {
			fmt.Fprintf(&b, "Found %d pack(s)", len(results))
			if stack != "" {
				fmt.Fprintf(&b, " for stack '%s'", stack)
			}
			fmt.Fprintf(&b, ":\n\n")

			for _, p := range results {
				name := strings.TrimPrefix(p.Repo, "github.com/")
				fmt.Fprintf(&b, "- **%s** [%s] — %s\n", name, strings.Join(p.Stacks, ", "), p.Description)
			}
			fmt.Fprintf(&b, "\nInstall any pack using `install_pack` with the full repo path.")
		}

		return &pluginv1.PromptGetResponse{
			Description: "Search and display available packs",
			Messages: []*pluginv1.PromptMessage{
				{
					Role:    "user",
					Content: &pluginv1.ContentBlock{Type: "text", Text: b.String()},
				},
			},
		}, nil
	}
}

// --- onboard-project ---

func OnboardProjectArgs() []*pluginv1.PromptArgument {
	return []*pluginv1.PromptArgument{
		{Name: "project_name", Description: "Name of the project", Required: true},
		{Name: "description", Description: "Brief project description (optional)", Required: false},
	}
}

func OnboardProject(ps *storage.PackStorage, workspace string) PromptHandler {
	return func(ctx context.Context, req *pluginv1.PromptGetRequest) (*pluginv1.PromptGetResponse, error) {
		projectName := req.Arguments["project_name"]
		description := req.Arguments["description"]

		if projectName == "" {
			projectName = "my-project"
		}

		detected := packs.DetectStacks(workspace)
		stackNames := make([]string, 0, len(detected))
		for _, s := range detected {
			stackNames = append(stackNames, s.Name)
		}

		recommended := make([]packs.PackInfo, 0)
		if len(stackNames) > 0 {
			recommended = packs.RecommendPacks(stackNames)
		}

		var b strings.Builder
		fmt.Fprintf(&b, "Full onboarding for project '%s'", projectName)
		if description != "" {
			fmt.Fprintf(&b, " (%s)", description)
		}
		fmt.Fprintf(&b, ".\n\n")

		fmt.Fprintf(&b, "Please perform these steps in order:\n\n")
		fmt.Fprintf(&b, "1. **Create project:** Use `create_project` with name '%s'", projectName)
		if description != "" {
			fmt.Fprintf(&b, " and description '%s'", description)
		}
		fmt.Fprintf(&b, "\n\n")

		if len(stackNames) > 0 {
			fmt.Fprintf(&b, "2. **Set stacks:** Use `set_project_stacks` with stacks: %s\n\n", strings.Join(stackNames, ", "))
		} else {
			fmt.Fprintf(&b, "2. **Detect stacks:** Use `detect_stacks` to identify the project's technology stacks, then `set_project_stacks` to save them\n\n")
		}

		if len(recommended) > 0 {
			fmt.Fprintf(&b, "3. **Install packs:** Install these recommended packs:\n")
			for _, p := range recommended {
				fmt.Fprintf(&b, "   - `install_pack` with repo `%s`\n", p.Repo)
			}
			fmt.Fprintf(&b, "\n")
		} else {
			fmt.Fprintf(&b, "3. **Install essentials:** Use `install_pack` with repo `github.com/orchestra-mcp/pack-essentials`\n\n")
		}

		fmt.Fprintf(&b, "4. **Verify:** Use `list_packs` to confirm installed packs, then `list_skills` and `list_agents` to see available capabilities\n\n")
		fmt.Fprintf(&b, "5. **Start working:** Use `create_feature` to create your first feature and begin development")

		return &pluginv1.PromptGetResponse{
			Description: "Full project onboarding: create, detect stacks, install packs, verify",
			Messages: []*pluginv1.PromptMessage{
				{
					Role:    "user",
					Content: &pluginv1.ContentBlock{Type: "text", Text: b.String()},
				},
			},
		}, nil
	}
}
