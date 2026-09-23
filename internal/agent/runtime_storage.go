package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var errRuntimeDataRootNested = errors.New("runtime data root must be separate from the execution workspace")

var legacyRuntimeDirectories = []string{
	".agent-context",
	".agent-companies",
	".agent-queue",
	".agent-traces",
	".agent-builders",
	".agent-collaboration",
	".agent-orchestrator",
	".agent-devices",
}

func resolveRuntimeDataRoot(workspaceRoot, configured string) (string, error) {
	workspaceRoot, err := filepath.Abs(filepath.Clean(workspaceRoot))
	if err != nil {
		return "", err
	}
	dataRoot := strings.TrimSpace(configured)
	if dataRoot == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("resolve user config directory: %w", err)
		}
		digest := sha256.Sum256([]byte(workspaceRoot))
		dataRoot = filepath.Join(configDir, "ollama-agent", hex.EncodeToString(digest[:8]))
	}
	dataRoot, err = filepath.Abs(filepath.Clean(dataRoot))
	if err != nil {
		return "", err
	}
	if sameOrWithin(workspaceRoot, dataRoot) || sameOrWithin(dataRoot, workspaceRoot) {
		return "", fmt.Errorf("%w: workspace=%q data=%q", errRuntimeDataRootNested, workspaceRoot, dataRoot)
	}
	if err := os.MkdirAll(dataRoot, 0o700); err != nil {
		return "", fmt.Errorf("create runtime data root: %w", err)
	}
	if err := migrateLegacyRuntimeDirectories(workspaceRoot, dataRoot); err != nil {
		return "", err
	}
	return dataRoot, nil
}

func sameOrWithin(path, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	return err == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))))
}

func migrateLegacyRuntimeDirectories(workspaceRoot, dataRoot string) error {
	for _, name := range legacyRuntimeDirectories {
		legacyPath := filepath.Join(workspaceRoot, name)
		if _, err := os.Stat(legacyPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return fmt.Errorf("inspect legacy runtime directory %q: %w", name, err)
		}
		newPath := filepath.Join(dataRoot, name)
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("legacy runtime directory %q and data-root directory both exist; migrate manually", name)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect data-root directory %q: %w", name, err)
		}
		if err := os.Rename(legacyPath, newPath); err != nil {
			return fmt.Errorf("migrate legacy runtime directory %q: %w", name, err)
		}
	}
	return nil
}
