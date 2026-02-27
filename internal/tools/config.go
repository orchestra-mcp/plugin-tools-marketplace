package tools

import (
	"context"
	"fmt"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/packs"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"google.golang.org/protobuf/types/known/structpb"
)

// --- set_project_stacks ---

func SetProjectStacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"stacks": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Technology stacks (e.g., go, rust, react, typescript, python)",
			},
		},
		"required": []any{"stacks"},
	})
	return s
}

func SetProjectStacks(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		stacks := helpers.GetStringSlice(req.Arguments, "stacks")
		if len(stacks) == 0 {
			return helpers.ErrorResult("validation_error", "stacks array is required"), nil
		}

		_, version, _ := ps.ReadStacks(ctx)
		if _, err := ps.WriteStacks(ctx, stacks, version); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Project stacks set to: %s", strings.Join(stacks, ", "))), nil
	}
}

// --- get_project_stacks ---

func GetProjectStacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

func GetProjectStacks(ps *storage.PackStorage, workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		// Try configured stacks first.
		configured, _, _ := ps.ReadStacks(ctx)

		var b strings.Builder
		if len(configured) > 0 {
			fmt.Fprintf(&b, "## Project Stacks (configured)\n\n")
			for _, s := range configured {
				fmt.Fprintf(&b, "- **%s**\n", s)
			}
		} else {
			// Fall back to auto-detection.
			detected := packs.DetectStacks(workspace)
			if len(detected) == 0 {
				return helpers.TextResult("## Project Stacks\n\nNo stacks configured or detected. Use `set_project_stacks` to configure."), nil
			}
			fmt.Fprintf(&b, "## Project Stacks (auto-detected)\n\n")
			for _, s := range detected {
				fmt.Fprintf(&b, "- **%s** — %s\n", s.Name, s.Evidence)
			}
			fmt.Fprintf(&b, "\nUse `set_project_stacks` to save these or override.")
		}

		return helpers.TextResult(b.String()), nil
	}
}
