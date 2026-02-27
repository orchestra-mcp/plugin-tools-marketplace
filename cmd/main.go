// Command tools-marketplace is the entry point for the tools.marketplace plugin
// binary. It provides 15 MCP tools and 5 MCP prompts for managing installable
// packs of skills, agents, and hooks from GitHub repositories.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
)

func main() {
	workspace := flag.String("workspace", ".", "Root workspace directory")

	builder := plugin.New("tools.marketplace").
		Version("0.1.0").
		Description("Marketplace for installable packs of skills, agents, and hooks").
		Author("Orchestra").
		Binary("tools-marketplace").
		NeedsStorage("markdown")

	// Create a placeholder storage that will be wired after ParseFlags.
	adapter := &clientAdapter{}
	store := storage.NewPackStorage(adapter)

	mp := &internal.MarketplacePlugin{
		Storage:   store,
		Workspace: *workspace,
	}
	mp.RegisterTools(builder)
	mp.RegisterPrompts(builder)

	p := builder.BuildWithTools()
	p.ParseFlags()

	// Re-read workspace after flag.Parse has been called and re-wire.
	mp.Workspace = *workspace
	adapter.plugin = p

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("tools.marketplace: %v", err)
	}
}

// clientAdapter implements storage.StorageClient by forwarding to the plugin's
// orchestrator client. This allows tool handlers to use storage operations
// through the QUIC connection that is established during Run.
type clientAdapter struct {
	plugin *plugin.Plugin
}

func (a *clientAdapter) Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	client := a.plugin.OrchestratorClient()
	if client == nil {
		return nil, fmt.Errorf("orchestrator client not connected")
	}
	return client.Send(ctx, req)
}
