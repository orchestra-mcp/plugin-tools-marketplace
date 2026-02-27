package storage

import (
	"context"
	"encoding/json"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// StorageClient sends requests to the orchestrator for storage operations.
type StorageClient interface {
	Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error)
}

// PackRegistry holds all installed packs.
type PackRegistry struct {
	Packs map[string]*PackEntry `json:"packs"`
}

// PackEntry describes a single installed pack.
type PackEntry struct {
	Version     string   `json:"version"`
	Repo        string   `json:"repo"`
	InstalledAt string   `json:"installed_at"`
	Stacks      []string `json:"stacks"`
	Skills      []string `json:"skills"`
	Agents      []string `json:"agents"`
	Hooks       []string `json:"hooks"`
}

// PackStorage provides operations for reading and writing the pack registry.
type PackStorage struct {
	client StorageClient
}

// NewPackStorage creates a new PackStorage backed by the given client.
func NewPackStorage(client StorageClient) *PackStorage {
	return &PackStorage{client: client}
}

const registryPath = ".packs/registry.json"
const stacksPath = ".packs/stacks.json"

// ReadRegistry loads the pack registry from storage.
func (ps *PackStorage) ReadRegistry(ctx context.Context) (*PackRegistry, int64, error) {
	resp, err := ps.storageRead(ctx, registryPath)
	if err != nil {
		return &PackRegistry{Packs: make(map[string]*PackEntry)}, 0, nil
	}
	if resp.Metadata == nil {
		return &PackRegistry{Packs: make(map[string]*PackEntry)}, resp.Version, nil
	}
	raw, err := json.Marshal(resp.Metadata.AsMap())
	if err != nil {
		return nil, 0, fmt.Errorf("marshal registry metadata: %w", err)
	}
	var reg PackRegistry
	if err := json.Unmarshal(raw, &reg); err != nil {
		return nil, 0, fmt.Errorf("parse registry: %w", err)
	}
	if reg.Packs == nil {
		reg.Packs = make(map[string]*PackEntry)
	}
	return &reg, resp.Version, nil
}

// WriteRegistry persists the pack registry to storage.
func (ps *PackStorage) WriteRegistry(ctx context.Context, reg *PackRegistry, expectedVersion int64) (int64, error) {
	raw, err := json.Marshal(reg)
	if err != nil {
		return 0, fmt.Errorf("marshal registry: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, fmt.Errorf("convert registry: %w", err)
	}
	meta, err := structpb.NewStruct(m)
	if err != nil {
		return 0, fmt.Errorf("struct from registry: %w", err)
	}
	return ps.storageWrite(ctx, registryPath, meta, nil, expectedVersion)
}

// ReadStacks reads the manually configured stacks from storage.
func (ps *PackStorage) ReadStacks(ctx context.Context) ([]string, int64, error) {
	resp, err := ps.storageRead(ctx, stacksPath)
	if err != nil {
		return nil, 0, nil
	}
	if resp.Metadata == nil {
		return nil, resp.Version, nil
	}
	raw, err := json.Marshal(resp.Metadata.AsMap())
	if err != nil {
		return nil, 0, err
	}
	var data struct {
		Stacks []string `json:"stacks"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, 0, err
	}
	return data.Stacks, resp.Version, nil
}

// WriteStacks persists the configured stacks to storage.
func (ps *PackStorage) WriteStacks(ctx context.Context, stacks []string, expectedVersion int64) (int64, error) {
	// structpb.NewStruct requires []any, not []string.
	items := make([]any, len(stacks))
	for i, s := range stacks {
		items[i] = s
	}
	m := map[string]any{"stacks": items}
	meta, err := structpb.NewStruct(m)
	if err != nil {
		return 0, fmt.Errorf("build stacks metadata: %w", err)
	}
	return ps.storageWrite(ctx, stacksPath, meta, nil, expectedVersion)
}

// --- Low-level storage protocol ---

func (ps *PackStorage) storageRead(ctx context.Context, path string) (*pluginv1.StorageReadResponse, error) {
	resp, err := ps.client.Send(ctx, &pluginv1.PluginRequest{
		RequestId: helpers.NewUUID(),
		Request: &pluginv1.PluginRequest_StorageRead{
			StorageRead: &pluginv1.StorageReadRequest{
				Path:        path,
				StorageType: "markdown",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	sr := resp.GetStorageRead()
	if sr == nil {
		return nil, fmt.Errorf("unexpected response type for storage read")
	}
	return sr, nil
}

func (ps *PackStorage) storageWrite(ctx context.Context, path string, metadata *structpb.Struct, content []byte, expectedVersion int64) (int64, error) {
	resp, err := ps.client.Send(ctx, &pluginv1.PluginRequest{
		RequestId: helpers.NewUUID(),
		Request: &pluginv1.PluginRequest_StorageWrite{
			StorageWrite: &pluginv1.StorageWriteRequest{
				Path:            path,
				Content:         content,
				Metadata:        metadata,
				ExpectedVersion: expectedVersion,
				StorageType:     "markdown",
			},
		},
	})
	if err != nil {
		return 0, err
	}
	sw := resp.GetStorageWrite()
	if sw == nil {
		return 0, fmt.Errorf("unexpected response type for storage write")
	}
	if !sw.Success {
		return 0, fmt.Errorf("storage write failed: %s", sw.Error)
	}
	return sw.NewVersion, nil
}
