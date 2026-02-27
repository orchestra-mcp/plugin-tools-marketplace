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

// --- detect_stacks ---

func DetectStacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	})
	return s
}

func DetectStacks(workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		detected := packs.DetectStacks(workspace)

		if len(detected) == 0 {
			return helpers.TextResult("## Stack Detection\n\nNo technology stacks detected in this workspace."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Detected Stacks (%d)\n\n", len(detected))
		fmt.Fprintf(&b, "| Stack | Evidence |\n")
		fmt.Fprintf(&b, "|-------|----------|\n")
		for _, s := range detected {
			fmt.Fprintf(&b, "| **%s** | %s |\n", s.Name, s.Evidence)
		}

		return helpers.TextResult(b.String()), nil
	}
}

// --- recommend_packs ---

func RecommendPacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"stacks": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Override detected stacks (optional)",
			},
		},
	})
	return s
}

func RecommendPacks(ps *storage.PackStorage, workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		stackNames := helpers.GetStringSlice(req.Arguments, "stacks")

		// If no stacks provided, try configured stacks first, then auto-detect.
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

		if len(stackNames) == 0 {
			return helpers.TextResult("## Pack Recommendations\n\nNo stacks detected. Use `set_project_stacks` to configure manually, or use `search_packs` to browse."), nil
		}

		recommended := packs.RecommendPacks(stackNames)

		// Check which are already installed.
		reg, _, _ := ps.ReadRegistry(ctx)
		installed := make(map[string]bool)
		for name := range reg.Packs {
			installed[name] = true
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Recommended Packs\n\n")
		fmt.Fprintf(&b, "**Detected stacks:** %s\n\n", strings.Join(stackNames, ", "))
		fmt.Fprintf(&b, "| Pack | Stacks | Description | Status |\n")
		fmt.Fprintf(&b, "|------|--------|-------------|--------|\n")

		for _, p := range recommended {
			// Derive the pack name from repo.
			parts := strings.SplitN(p.Repo, "/", 3)
			packName := ""
			if len(parts) >= 3 {
				packName = parts[1] + "/" + parts[2]
			}

			status := "available"
			if installed[packName] {
				status = "**installed**"
			}

			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
				strings.TrimPrefix(p.Repo, "github.com/"),
				strings.Join(p.Stacks, ", "),
				p.Description, status)
		}

		fmt.Fprintf(&b, "\nInstall with: `install_pack` tool, passing the full repo path (e.g., `github.com/orchestra-mcp/pack-go-backend`).")

		return helpers.TextResult(b.String()), nil
	}
}
