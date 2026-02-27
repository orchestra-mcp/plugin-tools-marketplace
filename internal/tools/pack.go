package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/packs"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"google.golang.org/protobuf/types/known/structpb"
)

// ToolHandler is an alias for readability.
type ToolHandler = func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error)

// --- install_pack ---

func InstallPackSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"repo":    map[string]any{"type": "string", "description": "GitHub repo (e.g., github.com/orchestra-mcp/pack-go-backend)"},
			"version": map[string]any{"type": "string", "description": "Version tag (optional, defaults to latest)"},
		},
		"required": []any{"repo"},
	})
	return s
}

func InstallPack(ps *storage.PackStorage, workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "repo"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		repo := helpers.GetString(req.Arguments, "repo")
		version := helpers.GetString(req.Arguments, "version")

		manifest, err := packs.InstallPack(workspace, repo, version)
		if err != nil {
			return helpers.ErrorResult("install_error", err.Error()), nil
		}

		// Update registry.
		reg, regVersion, err := ps.ReadRegistry(ctx)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		reg.Packs[manifest.Name] = &storage.PackEntry{
			Version:     manifest.Version,
			Repo:        repo,
			InstalledAt: helpers.NowISO(),
			Stacks:      manifest.Stacks,
			Skills:      manifest.Contents.Skills,
			Agents:      manifest.Contents.Agents,
			Hooks:       manifest.Contents.Hooks,
		}

		if _, err := ps.WriteRegistry(ctx, reg, regVersion); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Installed: %s\n\n", manifest.Name)
		fmt.Fprintf(&b, "- **Version:** %s\n", manifest.Version)
		if len(manifest.Contents.Skills) > 0 {
			fmt.Fprintf(&b, "- **Skills:** %s\n", strings.Join(manifest.Contents.Skills, ", "))
		}
		if len(manifest.Contents.Agents) > 0 {
			fmt.Fprintf(&b, "- **Agents:** %s\n", strings.Join(manifest.Contents.Agents, ", "))
		}
		if len(manifest.Contents.Hooks) > 0 {
			fmt.Fprintf(&b, "- **Hooks:** %s\n", strings.Join(manifest.Contents.Hooks, ", "))
		}
		return helpers.TextResult(b.String()), nil
	}
}

// --- remove_pack ---

func RemovePackSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Pack name (e.g., orchestra-mcp/pack-go-backend)"},
		},
		"required": []any{"name"},
	})
	return s
}

func RemovePack(ps *storage.PackStorage, workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")

		reg, regVersion, err := ps.ReadRegistry(ctx)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		entry, ok := reg.Packs[name]
		if !ok {
			return helpers.ErrorResult("not_found", fmt.Sprintf("pack %q not installed", name)), nil
		}

		if err := packs.RemovePack(workspace, entry.Skills, entry.Agents, entry.Hooks); err != nil {
			return helpers.ErrorResult("remove_error", err.Error()), nil
		}

		delete(reg.Packs, name)
		if _, err := ps.WriteRegistry(ctx, reg, regVersion); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Removed pack: %s", name)), nil
	}
}

// --- update_pack ---

func UpdatePackSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Pack name to update (omit to update all)"},
		},
	})
	return s
}

func UpdatePack(ps *storage.PackStorage, workspace string) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		name := helpers.GetString(req.Arguments, "name")

		reg, regVersion, err := ps.ReadRegistry(ctx)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		var toUpdate map[string]*storage.PackEntry
		if name != "" {
			entry, ok := reg.Packs[name]
			if !ok {
				return helpers.ErrorResult("not_found", fmt.Sprintf("pack %q not installed", name)), nil
			}
			toUpdate = map[string]*storage.PackEntry{name: entry}
		} else {
			toUpdate = reg.Packs
		}

		var updated []string
		for packName, entry := range toUpdate {
			// Remove old files.
			packs.RemovePack(workspace, entry.Skills, entry.Agents, entry.Hooks)

			// Re-install.
			manifest, err := packs.InstallPack(workspace, entry.Repo, "")
			if err != nil {
				return helpers.ErrorResult("update_error", fmt.Sprintf("update %s: %v", packName, err)), nil
			}

			reg.Packs[packName] = &storage.PackEntry{
				Version:     manifest.Version,
				Repo:        entry.Repo,
				InstalledAt: helpers.NowISO(),
				Stacks:      manifest.Stacks,
				Skills:      manifest.Contents.Skills,
				Agents:      manifest.Contents.Agents,
				Hooks:       manifest.Contents.Hooks,
			}
			updated = append(updated, packName)
		}

		if _, err := ps.WriteRegistry(ctx, reg, regVersion); err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Updated %d pack(s): %s", len(updated), strings.Join(updated, ", "))), nil
	}
}

// --- list_packs ---

func ListPacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"description": "Filter by content type",
				"enum":        []any{"skills", "agents", "hooks"},
			},
		},
	})
	return s
}

func ListPacks(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		filterType := helpers.GetString(req.Arguments, "type")

		reg, _, err := ps.ReadRegistry(ctx)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		if len(reg.Packs) == 0 {
			return helpers.TextResult("## Installed Packs\n\nNo packs installed. Use `install_pack` to add packs."), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Installed Packs (%d)\n\n", len(reg.Packs))
		fmt.Fprintf(&b, "| Name | Version | Skills | Agents | Hooks |\n")
		fmt.Fprintf(&b, "|------|---------|--------|--------|-------|\n")

		for name, entry := range reg.Packs {
			if filterType != "" {
				switch filterType {
				case "skills":
					if len(entry.Skills) == 0 {
						continue
					}
				case "agents":
					if len(entry.Agents) == 0 {
						continue
					}
				case "hooks":
					if len(entry.Hooks) == 0 {
						continue
					}
				}
			}
			fmt.Fprintf(&b, "| %s | %s | %d | %d | %d |\n",
				filepath.Base(name), entry.Version,
				len(entry.Skills), len(entry.Agents), len(entry.Hooks))
		}

		return helpers.TextResult(b.String()), nil
	}
}

// --- get_pack ---

func GetPackSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string", "description": "Pack name"},
		},
		"required": []any{"name"},
	})
	return s
}

func GetPack(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		name := helpers.GetString(req.Arguments, "name")

		reg, _, err := ps.ReadRegistry(ctx)
		if err != nil {
			return helpers.ErrorResult("storage_error", err.Error()), nil
		}

		entry, ok := reg.Packs[name]
		if !ok {
			return helpers.ErrorResult("not_found", fmt.Sprintf("pack %q not installed", name)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Pack: %s\n\n", name)
		fmt.Fprintf(&b, "- **Version:** %s\n", entry.Version)
		fmt.Fprintf(&b, "- **Repo:** %s\n", entry.Repo)
		fmt.Fprintf(&b, "- **Installed:** %s\n", entry.InstalledAt)
		if len(entry.Stacks) > 0 {
			fmt.Fprintf(&b, "- **Stacks:** %s\n", strings.Join(entry.Stacks, ", "))
		}
		if len(entry.Skills) > 0 {
			fmt.Fprintf(&b, "- **Skills:** %s\n", strings.Join(entry.Skills, ", "))
		}
		if len(entry.Agents) > 0 {
			fmt.Fprintf(&b, "- **Agents:** %s\n", strings.Join(entry.Agents, ", "))
		}
		if len(entry.Hooks) > 0 {
			fmt.Fprintf(&b, "- **Hooks:** %s\n", strings.Join(entry.Hooks, ", "))
		}

		return helpers.TextResult(b.String()), nil
	}
}

// --- search_packs ---

func SearchPacksSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "Search keyword"},
			"stack": map[string]any{"type": "string", "description": "Filter by technology stack"},
		},
		"required": []any{"query"},
	})
	return s
}

func SearchPacks(ps *storage.PackStorage) ToolHandler {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "query"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		query := helpers.GetString(req.Arguments, "query")
		stack := helpers.GetString(req.Arguments, "stack")

		results := packs.SearchPacks(query)

		// Filter by stack if provided.
		if stack != "" {
			var filtered []packs.PackInfo
			for _, p := range results {
				for _, s := range p.Stacks {
					if s == "*" || s == stack {
						filtered = append(filtered, p)
						break
					}
				}
			}
			results = filtered
		}

		if len(results) == 0 {
			return helpers.TextResult(fmt.Sprintf("No packs found for query: %q", query)), nil
		}

		var b strings.Builder
		fmt.Fprintf(&b, "## Search Results (%d)\n\n", len(results))
		fmt.Fprintf(&b, "| Pack | Stacks | Description |\n")
		fmt.Fprintf(&b, "|------|--------|-------------|\n")
		for _, p := range results {
			fmt.Fprintf(&b, "| %s | %s | %s |\n",
				filepath.Base(p.Repo), strings.Join(p.Stacks, ", "), p.Description)
		}
		fmt.Fprintf(&b, "\nInstall with: `install_pack` tool, passing the full repo path.")

		return helpers.TextResult(b.String()), nil
	}
}
