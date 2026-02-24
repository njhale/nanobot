package system

import (
	"context"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanobot-ai/nanobot/pkg/mcp"
)

// CreateFileParams are the parameters for the createFile tool.
type CreateFileParams struct {
	Name     string `json:"name"`
	Blob     string `json:"blob"`
	MimeType string `json:"mimeType"`
}

func (s *Server) uploadFile(ctx context.Context, params CreateFileParams) (*mcp.Resource, error) {
	if params.Name == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("name is required")
	}
	if params.Blob == "" {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("blob is required")
	}

	// Security: clean path and reject traversal / absolute paths
	cleanPath := filepath.Clean(params.Name)
	if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside working directory")
	}

	// Decode base64 content
	data, err := base64.StdEncoding.DecodeString(params.Blob)
	if err != nil {
		return nil, mcp.ErrRPCInvalidParams.WithMessage("invalid base64 blob: %v", err)
	}

	// Create parent directories if needed
	dir := filepath.Dir(cleanPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directories: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(cleanPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Determine MIME type
	mimeType := params.MimeType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(cleanPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return &mcp.Resource{
		URI:      "file:///" + cleanPath,
		Name:     cleanPath,
		MimeType: mimeType,
		Size:     info.Size(),
		Meta:     fileResourceMeta(cleanPath, info),
	}, nil
}

// DeleteFileParams are the parameters for the deleteFile tool.
type DeleteFileParams struct {
	URI string `json:"uri"`
}

func (s *Server) deleteFile(ctx context.Context, params DeleteFileParams) (string, error) {
	if params.URI == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("uri is required")
	}

	if !strings.HasPrefix(params.URI, "file:///") {
		return "", mcp.ErrRPCInvalidParams.WithMessage("uri must be a file:/// URI")
	}

	relPath := strings.TrimPrefix(params.URI, "file:///")
	if relPath == "" {
		return "", mcp.ErrRPCInvalidParams.WithMessage("file path is required")
	}

	// Security: clean path and reject traversal / absolute paths
	cleanPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return "", mcp.ErrRPCInvalidParams.WithMessage("invalid file path: cannot access files outside working directory")
	}

	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", mcp.ErrRPCInvalidParams.WithMessage("file not found: %s", params.URI)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		if err := os.RemoveAll(cleanPath); err != nil {
			return "", fmt.Errorf("failed to remove directory: %w", err)
		}
		return fmt.Sprintf("Deleted directory: %s", cleanPath), nil
	}

	if err := os.Remove(cleanPath); err != nil {
		return "", fmt.Errorf("failed to remove file: %w", err)
	}

	return fmt.Sprintf("Deleted file: %s", cleanPath), nil
}
