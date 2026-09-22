package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type mcpWorkingDirectory struct {
	path    string
	cleanup bool
}

func prepareMCPWorkingDirectory(serverID, requested string) (mcpWorkingDirectory, error) {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		path, err := filepath.Abs(requested)
		if err != nil {
			return mcpWorkingDirectory{}, fmt.Errorf("MCP working directory is invalid: %w", err)
		}
		if err := os.MkdirAll(path, 0o700); err != nil {
			return mcpWorkingDirectory{}, fmt.Errorf("create MCP working directory: %w", err)
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return mcpWorkingDirectory{}, errors.New("MCP working directory must be a directory")
		}
		return mcpWorkingDirectory{path: path}, nil
	}
	prefix := "ollama-mcp-" + sanitizeMCPID(serverID) + "-"
	path, err := os.MkdirTemp("", prefix)
	if err != nil {
		return mcpWorkingDirectory{}, fmt.Errorf("create MCP temporary workspace: %w", err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		_ = os.RemoveAll(path)
		return mcpWorkingDirectory{}, err
	}
	return mcpWorkingDirectory{path: path, cleanup: true}, nil
}

func sanitizeMCPID(value string) string {
	var builder strings.Builder
	for _, char := range strings.TrimSpace(value) {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			builder.WriteRune(char)
		}
	}
	if builder.Len() == 0 {
		return "server"
	}
	return builder.String()
}
