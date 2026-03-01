package toolsmarketplace

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"github.com/orchestra-mcp/sdk-go/plugin"
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
}
