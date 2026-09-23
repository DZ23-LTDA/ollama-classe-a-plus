package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// canonicalProjectRoot validates a project root against the runtime workspace.
// The final directory may not exist yet, but every existing component must be a
// real directory and the resolved path must remain inside the runtime root.
func canonicalProjectRoot(workspaceRoot, projectRoot string) (string, error) {
	projectRoot = strings.TrimSpace(projectRoot)
	if projectRoot == "" {
		return "", nil
	}
	candidate, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	canonicalCandidate, err := canonicalPathAllowMissing(candidate)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(workspaceRoot) == "" {
		if err := rejectProjectRootSymlinks(candidate); err != nil {
			return "", err
		}
		return canonicalCandidate, nil
	}
	workspace, err := canonicalExistingDirectory(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("runtime workspace is invalid: %w", err)
	}
	if !isWithin(workspace, canonicalCandidate) {
		return "", errors.New("project root is outside the runtime workspace")
	}
	if err := rejectSymlinkComponents(workspace, canonicalCandidate); err != nil {
		return "", fmt.Errorf("project root is not safe: %w", err)
	}
	return canonicalCandidate, nil
}

func canonicalExistingDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", errors.New("path is not a directory")
	}
	linkInfo, err := os.Lstat(absolute)
	if err != nil {
		return "", err
	}
	if linkInfo.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("directory must not be a symlink")
	}
	return filepath.EvalSymlinks(absolute)
}

func canonicalPathAllowMissing(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	current := absolute
	missing := make([]string, 0, 4)
	for {
		info, statErr := os.Lstat(current)
		if statErr == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return "", errors.New("path must not be a symlink")
			}
			if !info.IsDir() {
				return "", errors.New("project root is not a directory")
			}
			real, evalErr := filepath.EvalSymlinks(current)
			if evalErr != nil {
				return "", evalErr
			}
			for index := len(missing) - 1; index >= 0; index-- {
				real = filepath.Join(real, missing[index])
			}
			return real, nil
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return "", statErr
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", statErr
		}
		missing = append(missing, filepath.Base(current))
		current = parent
	}
}

func rejectProjectRootSymlinks(candidate string) error {
	parent := filepath.Dir(candidate)
	if _, err := os.Stat(parent); err != nil {
		return err
	}
	if err := rejectSymlinkComponents(parent, candidate); err != nil {
		return err
	}
	return nil
}
