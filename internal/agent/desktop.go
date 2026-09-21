package agent

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type desktopCompanionTool struct{}

func (desktopCompanionTool) Descriptor() ToolDescriptor {
	return ToolDescriptor{Name: "desktop.companion", Version: "1", Description: "Operações controladas de tela, mouse, teclado, clipboard e processos no Desktop local", Risk: RiskExternalSideEffect, RequiresApproval: true, Scopes: []string{"desktop:screen", "desktop:input", "desktop:clipboard", "desktop:process"}}
}

func (desktopCompanionTool) Execute(ctx context.Context, toolContext ToolContext, input map[string]any) (ToolResult, error) {
	action := strings.TrimSpace(stringInput(input, "action", ""))
	if action == "" {
		return ToolResult{}, errors.New("desktop action is required")
	}
	switch action {
	case "screenshot":
		relativePath := stringInput(input, "save_path", "desktop-screenshot.png")
		path, err := safeWorkspacePath(toolContext.Workspace, relativePath)
		if err != nil {
			return ToolResult{}, err
		}
		if err := runDesktop(ctx, "import", "-window", "root", path); err != nil {
			return ToolResult{}, err
		}
		manifest, err := BuildArtifactManifest(toolContext.Workspace, toolContext.MissionID, toolContext.StepID, "desktop-screenshot", relativePath)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "path": path}, Artifacts: []ArtifactManifest{manifest}}, nil
	case "mouse_click":
		x := intInput(input, "x", -1)
		if x < 0 || x > 10000 {
			return ToolResult{}, errors.New("mouse x must be between 0 and 10000")
		}
		y := intInput(input, "y", -1)
		if y < 0 || y > 10000 {
			return ToolResult{}, errors.New("mouse y must be between 0 and 10000")
		}
		if err := runDesktop(ctx, "xdotool", "mousemove", "--sync", strconv.Itoa(x), strconv.Itoa(y), "click", "1"); err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "x": x, "y": y}}, nil
	case "keyboard_type":
		text := stringInput(input, "text", "")
		if len(text) > 4096 || strings.ContainsRune(text, '\x00') {
			return ToolResult{}, errors.New("keyboard text is empty/too long or contains NUL")
		}
		if err := runDesktop(ctx, "xdotool", "type", "--clearmodifiers", "--delay", "1", text); err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "bytes": len(text)}}, nil
	case "clipboard_get":
		output, err := runDesktopOutput(ctx, "xclip", "-selection", "clipboard", "-o")
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "text": limitString(output, 64<<10)}}, nil
	case "clipboard_set":
		text := stringInput(input, "text", "")
		if len(text) > 64<<10 || strings.ContainsRune(text, '\x00') {
			return ToolResult{}, errors.New("clipboard text is empty/too long or contains NUL")
		}
		if err := runDesktopWithInput(ctx, text, "xclip", "-selection", "clipboard"); err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "bytes": len(text)}}, nil
	case "process_list":
		output, err := runDesktopOutput(ctx, "ps", "-eo", "pid=,comm=,args=")
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "processes": strings.FieldsFunc(limitString(output, 128<<10), func(r rune) bool { return r == '\n' })}}, nil
	case "process_terminate":
		pid := intInput(input, "pid", -1)
		if pid <= 1 || pid == os.Getpid() {
			return ToolResult{}, errors.New("refusing to terminate invalid or current process")
		}
		if err := runDesktop(ctx, "kill", "-TERM", strconv.Itoa(pid)); err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Value: map[string]any{"action": action, "pid": pid, "signal": "TERM"}}, nil
	default:
		return ToolResult{}, fmt.Errorf("unsupported desktop action %q", action)
	}
}

func runDesktop(ctx context.Context, name string, args ...string) error {
	_, err := runDesktopOutput(ctx, name, args...)
	return err
}

func runDesktopOutput(ctx context.Context, name string, args ...string) (string, error) {
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(deadline, name, args...)
	command.Env = append(os.Environ(), "LC_ALL=C")
	var stdout, stderr bytes.Buffer
	command.Stdout = &limitedBuffer{Buffer: &stdout, Limit: 128 << 10}
	command.Stderr = &limitedBuffer{Buffer: &stderr, Limit: 32 << 10}
	if err := command.Run(); err != nil {
		return stdout.String(), fmt.Errorf("desktop %s: %w: %s", name, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func runDesktopWithInput(ctx context.Context, input, name string, args ...string) error {
	deadline, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(deadline, name, args...)
	command.Env = append(os.Environ(), "LC_ALL=C")
	command.Stdin = strings.NewReader(input)
	return command.Run()
}

func limitString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
