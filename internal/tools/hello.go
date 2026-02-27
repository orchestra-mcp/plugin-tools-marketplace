package tools

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"github.com/orchestra-mcp/plugin-tools-marketplace/internal/storage"
	"google.golang.org/protobuf/types/known/structpb"
)

// HelloSchema returns the JSON Schema for the hello tool.
func HelloSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Name to greet",
			},
		},
		"required": []any{"name"},
	})
	return s
}

// Hello returns a tool handler that greets someone by name.
func Hello(_ *storage.DataStorage) func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "name"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		name := helpers.GetString(req.Arguments, "name")
		return helpers.TextResult(fmt.Sprintf("Hello, %s!", name)), nil
	}
}
