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
	Workflows   []string `json:"workflows,omitempty"`
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

	asMap := resp.Metadata.AsMap()
	reg := &PackRegistry{Packs: make(map[string]*PackEntry)}

	// Extract "packs" from the metadata — handle both map and array formats.
	packsRaw, ok := asMap["packs"]
	if !ok {
		return reg, resp.Version, nil
	}

	switch packs := packsRaw.(type) {
	case map[string]any:
		// Expected format: {"pack-name": {fields...}}
		for name, entryRaw := range packs {
			entry, err := parsePackEntry(entryRaw)
			if err != nil {
				continue // skip malformed entries
			}
			reg.Packs[name] = entry
		}
	case []any:
		// Legacy array format: [{name: "...", fields...}]
		for _, item := range packs {
			itemMap, ok := item.(map[string]any)
			if !ok {
				continue
			}
			name, _ := itemMap["name"].(string)
			if name == "" {
				continue
			}
			entry, err := parsePackEntry(item)
			if err != nil {
				continue
			}
			reg.Packs[name] = entry
		}
	default:
		// Unrecognised format — return empty registry rather than crash.
		return reg, resp.Version, nil
	}

	return reg, resp.Version, nil
}

// parsePackEntry converts a raw any value (typically map[string]any) into a PackEntry.
func parsePackEntry(raw any) (*PackEntry, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var entry PackEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
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

// Send delegates to the underlying storage client for direct storage operations.
func (ps *PackStorage) Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	return ps.client.Send(ctx, req)
}

// StorageRead performs a low-level storage read and returns the response.
func (ps *PackStorage) StorageRead(ctx context.Context, path string) (*pluginv1.StorageReadResponse, error) {
	return ps.storageRead(ctx, path)
}

// StorageWrite performs a low-level storage write and returns the new version.
func (ps *PackStorage) StorageWrite(ctx context.Context, path string, metadata *structpb.Struct, content []byte, expectedVersion int64) (int64, error) {
	return ps.storageWrite(ctx, path, metadata, content, expectedVersion)
}

// StorageDelete performs a low-level storage delete.
func (ps *PackStorage) StorageDelete(ctx context.Context, path string) error {
	resp, err := ps.client.Send(ctx, &pluginv1.PluginRequest{
		RequestId: helpers.NewUUID(),
		Request: &pluginv1.PluginRequest_StorageDelete{
			StorageDelete: &pluginv1.StorageDeleteRequest{
				Path:        path,
				StorageType: "markdown",
			},
		},
	})
	if err != nil {
		return err
	}
	sd := resp.GetStorageDelete()
	if sd == nil {
		return fmt.Errorf("unexpected response type for storage delete")
	}
	if !sd.Success {
		return fmt.Errorf("storage delete failed for path: %s", path)
	}
	return nil
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
