package internal

import (
	"github.com/orchestra-mcp/sdk-go/plugin"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/tools"
)

// ToolsPlugin holds the storage reference and registers all tools.
type ToolsPlugin struct {
	Storage *storage.DataStorage
}

// RegisterTools registers all tools with the plugin builder.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	s := tp.Storage

	builder.RegisterTool("hello",
		"Say hello to someone",
		tools.HelloSchema(), tools.Hello(s))
}
