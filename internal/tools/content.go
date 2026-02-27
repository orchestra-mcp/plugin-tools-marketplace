package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/packs"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- list_skills ---

func ListSkillsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

func ListSkills(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		names := packs.ListInstalledSkills(workspace)

		if len(names) == 0 {
			return helpers.TextResult("## Installed Skills\n\nNo skills found. Use `install_pack` to add skill packs."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Installed Skills (%d)\n\n", len(names))
		for _, name := range names {
			fmt.Fprintf(&b, "- `%s`\n", name)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// --- list_agents ---

func ListAgentsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

func ListAgents(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		names := packs.ListInstalledAgents(workspace)

		if len(names) == 0 {
			return helpers.TextResult("## Installed Agents\n\nNo agents found. Use `install_pack` to add agent packs."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Installed Agents (%d)\n\n", len(names))
		for _, name := range names {
			fmt.Fprintf(&b, "- `%s`\n", name)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// --- list_hooks ---

func ListHooksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

func ListHooks(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		names := packs.ListInstalledHooks(workspace)

		if len(names) == 0 {
			return helpers.TextResult("## Installed Hooks\n\nNo hooks found. Use `install_pack` to add hook packs."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Installed Hooks (%d)\n\n", len(names))
		for _, name := range names {
			fmt.Fprintf(&b, "- `%s`\n", name)
		}
		return helpers.TextResult(b.String()), nil
	}
}

// --- get_skill ---

func GetSkillSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Skill name"},
		},
		"required": []any{"name"},
	})
	return s
}

func GetSkill(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		content, err := packs.ReadSkillContent(workspace, name)
		if err != nil {
			return helpers.ErrorResult("not_found", err.Error()), nil
		}

		return helpers.TextResult(content), nil
	}
}

// --- get_agent ---

func GetAgentSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Agent name"},
		},
		"required": []any{"name"},
	})
	return s
}

func GetAgent(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		content, err := packs.ReadAgentContent(workspace, name)
		if err != nil {
			return helpers.ErrorResult("not_found", err.Error()), nil
		}

		return helpers.TextResult(content), nil
	}
}
