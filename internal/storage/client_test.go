package storage

import (
	"context"
	"fmt"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// mockClient implements StorageClient for testing.
type mockClient struct {
	response *pluginv1.PluginResponse
	err      error
}

func (m *mockClient) Send(_ context.Context, _ *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	return m.response, m.err
}

func makeStorageReadResponse(metadata map[string]any, version int64) *pluginv1.PluginResponse {
	var meta *structpb.Struct
	if metadata != nil {
		meta, _ = structpb.NewStruct(metadata)
	}
	return &pluginv1.PluginResponse{
		Response: &pluginv1.PluginResponse_StorageRead{
			StorageRead: &pluginv1.StorageReadResponse{
				Metadata: meta,
				Version:  version,
			},
		},
	}
}

func TestReadRegistry_MapFormat(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": map[string]any{
				"orchestra-mcp/pack-essentials": map[string]any{
					"version":      "0.1.0",
					"repo":         "github.com/orchestra-mcp/pack-essentials",
					"installed_at": "2026-02-27T09:21:36Z",
					"stacks":       []any{"*"},
					"skills":       []any{"project-manager", "docs"},
					"agents":       []any{"scrum-master"},
					"hooks":        []any{"notify"},
				},
			},
		}, 1),
	}
	ps := NewPackStorage(client)

	reg, version, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
	if len(reg.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(reg.Packs))
	}
	entry := reg.Packs["orchestra-mcp/pack-essentials"]
	if entry == nil {
		t.Fatal("expected pack entry for orchestra-mcp/pack-essentials")
	}
	if entry.Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", entry.Version)
	}
	if entry.Repo != "github.com/orchestra-mcp/pack-essentials" {
		t.Errorf("expected repo github.com/orchestra-mcp/pack-essentials, got %s", entry.Repo)
	}
	if len(entry.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(entry.Skills))
	}
	if len(entry.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(entry.Agents))
	}
	if len(entry.Hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(entry.Hooks))
	}
}

func TestReadRegistry_ArrayFormat(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": []any{
				map[string]any{
					"name":         "orchestra-mcp/pack-go-backend",
					"version":      "1.0.0",
					"repo":         "github.com/orchestra-mcp/pack-go-backend",
					"installed_at": "2026-03-01T10:00:00Z",
					"stacks":       []any{"go"},
					"skills":       []any{"go-backend"},
					"agents":       []any{},
					"hooks":        []any{},
				},
			},
		}, 2),
	}
	ps := NewPackStorage(client)

	reg, version, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != 2 {
		t.Errorf("expected version 2, got %d", version)
	}
	if len(reg.Packs) != 1 {
		t.Fatalf("expected 1 pack, got %d", len(reg.Packs))
	}
	entry := reg.Packs["orchestra-mcp/pack-go-backend"]
	if entry == nil {
		t.Fatal("expected pack entry for orchestra-mcp/pack-go-backend")
	}
	if entry.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", entry.Version)
	}
}

func TestReadRegistry_ArrayFormat_SkipsMissingName(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": []any{
				map[string]any{
					"version": "1.0.0",
					// no "name" field — should be skipped
				},
			},
		}, 1),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs (no name), got %d", len(reg.Packs))
	}
}

func TestReadRegistry_NilMetadata(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(nil, 0),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Packs == nil {
		t.Fatal("expected non-nil Packs map")
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(reg.Packs))
	}
}

func TestReadRegistry_EmptyPacks(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": map[string]any{},
		}, 1),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(reg.Packs))
	}
}

func TestReadRegistry_NoPacksKey(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"other_field": "value",
		}, 1),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(reg.Packs))
	}
}

func TestReadRegistry_UnknownPacksFormat(t *testing.T) {
	// "packs" is a string instead of map or array — should not crash.
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": "invalid-string-value",
		}, 1),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(reg.Packs))
	}
}

func TestReadRegistry_StorageError(t *testing.T) {
	client := &mockClient{
		err: fmt.Errorf("connection refused"),
	}
	ps := NewPackStorage(client)

	reg, version, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("expected no error (graceful fallback), got: %v", err)
	}
	if version != 0 {
		t.Errorf("expected version 0, got %d", version)
	}
	if len(reg.Packs) != 0 {
		t.Errorf("expected 0 packs, got %d", len(reg.Packs))
	}
}

func TestReadRegistry_MultiplePacks(t *testing.T) {
	client := &mockClient{
		response: makeStorageReadResponse(map[string]any{
			"packs": map[string]any{
				"pack-a": map[string]any{
					"version": "1.0.0",
					"repo":    "github.com/org/pack-a",
					"skills":  []any{"skill-1"},
				},
				"pack-b": map[string]any{
					"version": "2.0.0",
					"repo":    "github.com/org/pack-b",
					"agents":  []any{"agent-1", "agent-2"},
				},
			},
		}, 3),
	}
	ps := NewPackStorage(client)

	reg, _, err := ps.ReadRegistry(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Packs) != 2 {
		t.Fatalf("expected 2 packs, got %d", len(reg.Packs))
	}
	if reg.Packs["pack-a"].Version != "1.0.0" {
		t.Errorf("pack-a version: expected 1.0.0, got %s", reg.Packs["pack-a"].Version)
	}
	if reg.Packs["pack-b"].Version != "2.0.0" {
		t.Errorf("pack-b version: expected 2.0.0, got %s", reg.Packs["pack-b"].Version)
	}
}

func TestParsePackEntry(t *testing.T) {
	raw := map[string]any{
		"version":      "1.0.0",
		"repo":         "github.com/org/pack",
		"installed_at": "2026-01-01T00:00:00Z",
		"stacks":       []any{"go", "rust"},
		"skills":       []any{"skill-1"},
		"agents":       []any{"agent-1"},
		"hooks":        []any{"hook-1"},
	}

	entry, err := parsePackEntry(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", entry.Version)
	}
	if len(entry.Stacks) != 2 {
		t.Errorf("expected 2 stacks, got %d", len(entry.Stacks))
	}
}

func TestParsePackEntry_MalformedInput(t *testing.T) {
	// A non-marshalable type should return error.
	_, err := parsePackEntry(make(chan int))
	if err == nil {
		t.Error("expected error for non-marshalable input")
	}
}
