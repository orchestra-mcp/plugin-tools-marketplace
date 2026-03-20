package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------------------------------------------------------------------------
// Skills CRUD
// ---------------------------------------------------------------------------

// --- create_skill ---

func CreateSkillSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":        map[string]any{"type": "string", "description": "Human-readable skill name"},
			"slug":        map[string]any{"type": "string", "description": "Unique slug identifier (e.g., go-backend)"},
			"description": map[string]any{"type": "string", "description": "Short description of the skill"},
			"content":     map[string]any{"type": "string", "description": "Markdown body content of the skill"},
			"scope":       map[string]any{"type": "string", "description": "Scope: global or project (default: project)", "enum": []any{"global", "project"}},
		},
		"required": []any{"name", "slug", "description", "content"},
	})
	return s
}

func CreateSkill(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name", "slug", "description", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		slug := helpers.GetString(req.Arguments, "slug")
		description := helpers.GetString(req.Arguments, "description")
		content := helpers.GetString(req.Arguments, "content")
		scope := helpers.GetString(req.Arguments, "scope")
		if scope == "" {
			scope = "project"
		}

		path := ".skills/" + slug + ".md"

		// Check if skill already exists.
		if _, err := ps.StorageRead(ctx, path); err == nil {
			return helpers.ErrorResult("already_exists", fmt.Sprintf("skill %q already exists", slug)), nil
		}

		metadata, err := structpb.NewStruct(map[string]any{
			"name":        name,
			"slug":        slug,
			"description": description,
			"scope":       scope,
			"created_at":  helpers.NowISO(),
		})
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, []byte(content), 0); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Skill Created\n\n")
		fmt.Fprintf(&b, "- **Name:** %s\n", name)
		fmt.Fprintf(&b, "- **Slug:** %s\n", slug)
		fmt.Fprintf(&b, "- **Scope:** %s\n", scope)
		fmt.Fprintf(&b, "- **Description:** %s\n", description)
		return helpers.TextResult(b.String()), nil
	}
}

// --- update_skill ---

func UpdateSkillSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug":        map[string]any{"type": "string", "description": "Slug of the skill to update"},
			"name":        map[string]any{"type": "string", "description": "New name (optional)"},
			"description": map[string]any{"type": "string", "description": "New description (optional)"},
			"content":     map[string]any{"type": "string", "description": "New markdown body (optional)"},
			"scope":       map[string]any{"type": "string", "description": "New scope (optional)", "enum": []any{"global", "project"}},
		},
		"required": []any{"slug"},
	})
	return s
}

func UpdateSkill(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".skills/" + slug + ".md"

		// Read existing skill.
		existing, err := ps.StorageRead(ctx, path)
		if err != nil {
			return helpers.ErrorResult("not_found", fmt.Sprintf("skill %q not found", slug)), nil
		}

		// Merge metadata fields.
		meta := mergeMetadata(existing.Metadata, req.Arguments, []string{"name", "description", "scope"})
		meta["slug"] = slug
		meta["updated_at"] = helpers.NowISO()

		metadata, err := structpb.NewStruct(meta)
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		// Use new content if provided, otherwise keep existing.
		content := existing.Content
		if newContent := helpers.GetString(req.Arguments, "content"); newContent != "" {
			content = []byte(newContent)
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, content, existing.Version); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Skill Updated\n\nSkill **%s** has been updated.", slug)), nil
	}
}

// --- delete_skill ---

func DeleteSkillSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug": map[string]any{"type": "string", "description": "Slug of the skill to delete"},
		},
		"required": []any{"slug"},
	})
	return s
}

func DeleteSkill(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".skills/" + slug + ".md"

		if err := ps.StorageDelete(ctx, path); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Skill Deleted\n\nSkill **%s** has been removed.", slug)), nil
	}
}

// ---------------------------------------------------------------------------
// Agents CRUD
// ---------------------------------------------------------------------------

// --- create_agent ---

func CreateAgentSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":        map[string]any{"type": "string", "description": "Human-readable agent name"},
			"slug":        map[string]any{"type": "string", "description": "Unique slug identifier (e.g., devops)"},
			"description": map[string]any{"type": "string", "description": "Short description of the agent"},
			"content":     map[string]any{"type": "string", "description": "Markdown body content of the agent"},
			"scope":       map[string]any{"type": "string", "description": "Scope: global or project (default: project)", "enum": []any{"global", "project"}},
		},
		"required": []any{"name", "slug", "description", "content"},
	})
	return s
}

func CreateAgent(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name", "slug", "description", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		slug := helpers.GetString(req.Arguments, "slug")
		description := helpers.GetString(req.Arguments, "description")
		content := helpers.GetString(req.Arguments, "content")
		scope := helpers.GetString(req.Arguments, "scope")
		if scope == "" {
			scope = "project"
		}

		path := ".agents/" + slug + ".md"

		// Check if agent already exists.
		if _, err := ps.StorageRead(ctx, path); err == nil {
			return helpers.ErrorResult("already_exists", fmt.Sprintf("agent %q already exists", slug)), nil
		}

		metadata, err := structpb.NewStruct(map[string]any{
			"name":        name,
			"slug":        slug,
			"description": description,
			"scope":       scope,
			"created_at":  helpers.NowISO(),
		})
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, []byte(content), 0); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Agent Created\n\n")
		fmt.Fprintf(&b, "- **Name:** %s\n", name)
		fmt.Fprintf(&b, "- **Slug:** %s\n", slug)
		fmt.Fprintf(&b, "- **Scope:** %s\n", scope)
		fmt.Fprintf(&b, "- **Description:** %s\n", description)
		return helpers.TextResult(b.String()), nil
	}
}

// --- update_agent ---

func UpdateAgentSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug":        map[string]any{"type": "string", "description": "Slug of the agent to update"},
			"name":        map[string]any{"type": "string", "description": "New name (optional)"},
			"description": map[string]any{"type": "string", "description": "New description (optional)"},
			"content":     map[string]any{"type": "string", "description": "New markdown body (optional)"},
			"scope":       map[string]any{"type": "string", "description": "New scope (optional)", "enum": []any{"global", "project"}},
		},
		"required": []any{"slug"},
	})
	return s
}

func UpdateAgent(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".agents/" + slug + ".md"

		// Read existing agent.
		existing, err := ps.StorageRead(ctx, path)
		if err != nil {
			return helpers.ErrorResult("not_found", fmt.Sprintf("agent %q not found", slug)), nil
		}

		// Merge metadata fields.
		meta := mergeMetadata(existing.Metadata, req.Arguments, []string{"name", "description", "scope"})
		meta["slug"] = slug
		meta["updated_at"] = helpers.NowISO()

		metadata, err := structpb.NewStruct(meta)
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		// Use new content if provided, otherwise keep existing.
		content := existing.Content
		if newContent := helpers.GetString(req.Arguments, "content"); newContent != "" {
			content = []byte(newContent)
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, content, existing.Version); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Agent Updated\n\nAgent **%s** has been updated.", slug)), nil
	}
}

// --- delete_agent ---

func DeleteAgentSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug": map[string]any{"type": "string", "description": "Slug of the agent to delete"},
		},
		"required": []any{"slug"},
	})
	return s
}

func DeleteAgent(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".agents/" + slug + ".md"

		if err := ps.StorageDelete(ctx, path); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Agent Deleted\n\nAgent **%s** has been removed.", slug)), nil
	}
}

// ---------------------------------------------------------------------------
// Hooks CRUD
// ---------------------------------------------------------------------------

// --- create_hook ---

func CreateHookSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":        map[string]any{"type": "string", "description": "Human-readable hook name"},
			"slug":        map[string]any{"type": "string", "description": "Unique slug identifier (e.g., notify)"},
			"description": map[string]any{"type": "string", "description": "Short description of the hook"},
			"script":      map[string]any{"type": "string", "description": "Hook script content (shell script body)"},
			"scope":       map[string]any{"type": "string", "description": "Scope: global or project (default: project)", "enum": []any{"global", "project"}},
			"event_type":  map[string]any{"type": "string", "description": "Event that triggers this hook (e.g., post-commit, pre-push, on-save)"},
		},
		"required": []any{"name", "slug", "description", "script"},
	})
	return s
}

func CreateHook(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name", "slug", "description", "script"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")
		slug := helpers.GetString(req.Arguments, "slug")
		description := helpers.GetString(req.Arguments, "description")
		script := helpers.GetString(req.Arguments, "script")
		scope := helpers.GetString(req.Arguments, "scope")
		if scope == "" {
			scope = "project"
		}
		eventType := helpers.GetString(req.Arguments, "event_type")

		path := ".hooks/" + slug + ".md"

		// Check if hook already exists.
		if _, err := ps.StorageRead(ctx, path); err == nil {
			return helpers.ErrorResult("already_exists", fmt.Sprintf("hook %q already exists", slug)), nil
		}

		metaMap := map[string]any{
			"name":        name,
			"slug":        slug,
			"description": description,
			"scope":       scope,
			"created_at":  helpers.NowISO(),
		}
		if eventType != "" {
			metaMap["event_type"] = eventType
		}

		metadata, err := structpb.NewStruct(metaMap)
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, []byte(script), 0); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Hook Created\n\n")
		fmt.Fprintf(&b, "- **Name:** %s\n", name)
		fmt.Fprintf(&b, "- **Slug:** %s\n", slug)
		fmt.Fprintf(&b, "- **Scope:** %s\n", scope)
		if eventType != "" {
			fmt.Fprintf(&b, "- **Event Type:** %s\n", eventType)
		}
		fmt.Fprintf(&b, "- **Description:** %s\n", description)
		return helpers.TextResult(b.String()), nil
	}
}

// --- update_hook ---

func UpdateHookSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug":        map[string]any{"type": "string", "description": "Slug of the hook to update"},
			"name":        map[string]any{"type": "string", "description": "New name (optional)"},
			"description": map[string]any{"type": "string", "description": "New description (optional)"},
			"script":      map[string]any{"type": "string", "description": "New script content (optional)"},
			"scope":       map[string]any{"type": "string", "description": "New scope (optional)", "enum": []any{"global", "project"}},
			"event_type":  map[string]any{"type": "string", "description": "New event type (optional)"},
		},
		"required": []any{"slug"},
	})
	return s
}

func UpdateHook(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".hooks/" + slug + ".md"

		// Read existing hook.
		existing, err := ps.StorageRead(ctx, path)
		if err != nil {
			return helpers.ErrorResult("not_found", fmt.Sprintf("hook %q not found", slug)), nil
		}

		// Merge metadata fields.
		meta := mergeMetadata(existing.Metadata, req.Arguments, []string{"name", "description", "scope", "event_type"})
		meta["slug"] = slug
		meta["updated_at"] = helpers.NowISO()

		metadata, err := structpb.NewStruct(meta)
		if err != nil {
			return helpers.ErrorResult("internal_error", fmt.Sprintf("build metadata: %v", err)), nil
		}

		// Use new script if provided, otherwise keep existing.
		content := existing.Content
		if newScript := helpers.GetString(req.Arguments, "script"); newScript != "" {
			content = []byte(newScript)
		}

		if _, err := ps.StorageWrite(ctx, path, metadata, content, existing.Version); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Hook Updated\n\nHook **%s** has been updated.", slug)), nil
	}
}

// --- delete_hook ---

func DeleteHookSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"slug": map[string]any{"type": "string", "description": "Slug of the hook to delete"},
		},
		"required": []any{"slug"},
	})
	return s
}

func DeleteHook(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "slug"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		slug := helpers.GetString(req.Arguments, "slug")
		path := ".hooks/" + slug + ".md"

		if err := ps.StorageDelete(ctx, path); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("## Hook Deleted\n\nHook **%s** has been removed.", slug)), nil
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// mergeMetadata takes existing metadata from storage and merges in new values
// from the request arguments. Only non-empty argument values override existing ones.
func mergeMetadata(existing *structpb.Struct, args *structpb.Struct, fields []string) map[string]any {
	meta := make(map[string]any)

	// Copy all existing metadata fields.
	if existing != nil {
		for k, v := range existing.AsMap() {
			meta[k] = v
		}
	}

	// Override with non-empty values from the request.
	for _, field := range fields {
		if val := helpers.GetString(args, field); val != "" {
			meta[field] = val
		}
	}

	return meta
}
