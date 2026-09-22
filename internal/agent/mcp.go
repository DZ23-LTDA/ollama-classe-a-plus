package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type MCPServerConfig struct {
	ID               string   `json:"id"`
	OrganizationID   string   `json:"organization_id,omitempty"`
	Command          string   `json:"command"`
	Args             []string `json:"args,omitempty"`
	WorkingDirectory string   `json:"working_directory,omitempty"`
	AllowedMethods   []string `json:"allowed_methods,omitempty"`
	EnvironmentVars  []string `json:"environment_vars,omitempty"`
	TimeoutSeconds   int      `json:"timeout_seconds,omitempty"`
	Disabled         bool     `json:"disabled,omitempty"`
}

type MCPManager struct {
	mu      sync.RWMutex
	servers map[string]*MCPServer
}

func NewMCPManager() *MCPManager {
	return &MCPManager{servers: make(map[string]*MCPServer)}
}

func (m *MCPManager) Register(config MCPServerConfig) error {
	if strings.TrimSpace(config.ID) == "" || strings.TrimSpace(config.Command) == "" {
		return errors.New("MCP server id and command are required")
	}
	if config.TimeoutSeconds <= 0 || config.TimeoutSeconds > 300 {
		config.TimeoutSeconds = 30
	}
	config.Command = strings.TrimSpace(config.Command)
	if !filepath.IsAbs(config.Command) {
		return errors.New("MCP command must be an absolute executable path")
	}
	commandInfo, err := os.Lstat(config.Command)
	if err != nil || commandInfo.Mode()&os.ModeSymlink != 0 || commandInfo.IsDir() || commandInfo.Mode()&0o111 == 0 {
		return errors.New("MCP command must be an executable regular file")
	}
	if len(config.Args) > 64 {
		return errors.New("MCP args limit exceeded")
	}
	for _, arg := range config.Args {
		if strings.IndexByte(arg, 0) >= 0 || len(arg) > 4096 {
			return errors.New("MCP argument is invalid or too long")
		}
	}
	config.Args = append([]string(nil), config.Args...)
	if len(config.AllowedMethods) == 0 {
		return errors.New("MCP allowed_methods must contain at least one method")
	}
	allowed := make(map[string]bool, len(config.AllowedMethods))
	for _, method := range config.AllowedMethods {
		method = strings.TrimSpace(method)
		if method == "" {
			return errors.New("MCP allowed_methods cannot contain empty methods")
		}
		allowed[method] = true
	}
	normalizedEnvironmentVars := make([]string, 0, len(config.EnvironmentVars))
	for _, name := range config.EnvironmentVars {
		name = strings.TrimSpace(name)
		if !validEnvName(name) {
			return fmt.Errorf("invalid MCP environment variable %q", name)
		}
		normalizedEnvironmentVars = append(normalizedEnvironmentVars, name)
	}
	config.EnvironmentVars = normalizedEnvironmentVars
	workingDirectory, err := prepareMCPWorkingDirectory(config.ID, config.WorkingDirectory)
	if err != nil {
		return err
	}
	config.WorkingDirectory = workingDirectory.path
	m.mu.Lock()
	defer m.mu.Unlock()
	if old := m.servers[config.ID]; old != nil {
		_ = old.Stop()
	}
	m.servers[config.ID] = &MCPServer{config: config, allowedMethods: allowed, cleanupDirectory: workingDirectory.cleanup}
	return nil
}

func (m *MCPManager) StopAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, server := range m.servers {
		_ = server.Stop()
	}
}

func (m *MCPManager) List() []MCPServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MCPServerConfig, 0, len(m.servers))
	for _, server := range m.servers {
		config := server.config
		config.EnvironmentVars = append([]string(nil), config.EnvironmentVars...)
		result = append(result, config)
	}
	return result
}

func (m *MCPManager) ListForOrganization(organizationID string) []MCPServerConfig {
	organizationID = strings.TrimSpace(organizationID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MCPServerConfig, 0)
	for _, server := range m.servers {
		if server.config.OrganizationID != "" && !pluginOwnedByOrganization(server.config.OrganizationID, organizationID) {
			continue
		}
		config := server.config
		config.Args = append([]string(nil), config.Args...)
		config.EnvironmentVars = append([]string(nil), config.EnvironmentVars...)
		result = append(result, config)
	}
	return result
}

func (m *MCPManager) SetEnabled(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	server, ok := m.servers[strings.TrimSpace(id)]
	if !ok {
		return fmt.Errorf("MCP server %q is not registered", id)
	}
	server.config.Disabled = !enabled
	if !enabled {
		_ = server.Stop()
	}
	return nil
}

func (m *MCPManager) SetEnabledForOrganization(organizationID, id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	server, ok := m.servers[strings.TrimSpace(id)]
	if !ok {
		return fmt.Errorf("MCP server %q is not registered", id)
	}
	if !pluginOwnedByOrganization(server.config.OrganizationID, strings.TrimSpace(organizationID)) {
		return ErrPluginOrganizationScope
	}
	server.config.Disabled = !enabled
	if !enabled {
		_ = server.Stop()
	}
	return nil
}

func (m *MCPManager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	server, ok := m.servers[id]
	if !ok {
		return fmt.Errorf("MCP server %q is not registered", id)
	}
	_ = server.Stop()
	delete(m.servers, id)
	return nil
}

func (m *MCPManager) RemoveForOrganization(organizationID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	server, ok := m.servers[id]
	if !ok {
		return fmt.Errorf("MCP server %q is not registered", id)
	}
	if !pluginOwnedByOrganization(server.config.OrganizationID, strings.TrimSpace(organizationID)) {
		return ErrPluginOrganizationScope
	}
	_ = server.Stop()
	delete(m.servers, id)
	return nil
}

func (m *MCPManager) Call(ctx context.Context, serverID, method string, params any) (json.RawMessage, error) {
	m.mu.RLock()
	server := m.servers[serverID]
	m.mu.RUnlock()
	if server == nil {
		return nil, fmt.Errorf("MCP server %q is not registered", serverID)
	}
	server.mu.Lock()
	disabled := server.config.Disabled
	server.mu.Unlock()
	if disabled {
		return nil, errors.New("MCP server is disabled")
	}
	return server.Call(ctx, method, params)
}

type MCPServer struct {
	mu               sync.Mutex
	config           MCPServerConfig
	allowedMethods   map[string]bool
	cmd              *exec.Cmd
	stdin            io.WriteCloser
	stdout           *bufio.Reader
	stderr           *mcpStderrBuffer
	cancel           context.CancelFunc
	nextID           int64
	cleanupDirectory bool
}

const (
	mcpMaxMessageBytes = 4 << 20
	mcpMaxStderrBytes  = 64 << 10
)

type mcpStderrBuffer struct {
	bytes.Buffer
	limit int
}

func (b *mcpStderrBuffer) Write(data []byte) (int, error) {
	if b.limit > b.Len() {
		remaining := b.limit - b.Len()
		if len(data) > remaining {
			_, _ = b.Buffer.Write(data[:remaining])
		} else {
			_, _ = b.Buffer.Write(data)
		}
	}
	return len(data), nil
}

func (s *MCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.startLocked()
}

func (s *MCPServer) startLocked() error {
	if s.cmd != nil {
		return nil
	}
	if s.cleanupDirectory {
		if err := os.MkdirAll(s.config.WorkingDirectory, 0o700); err != nil {
			return fmt.Errorf("recreate MCP working directory: %w", err)
		}
	}
	processContext, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(processContext, s.config.Command, s.config.Args...)
	cmd.Dir = s.config.WorkingDirectory
	configureMCPProcess(cmd)
	cmd.Env = []string{"PATH=/usr/bin:/bin", "HOME=/tmp"}
	for _, name := range s.config.EnvironmentVars {
		name = strings.TrimSpace(name)
		if name == "" || strings.ContainsAny(name, "=\x00\r\n") {
			cancel()
			return fmt.Errorf("invalid MCP environment variable %q", name)
		}
		if value, ok := os.LookupEnv(name); ok {
			cmd.Env = append(cmd.Env, name+"="+value)
		}
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		cancel()
		return err
	}
	stderr := &mcpStderrBuffer{limit: mcpMaxStderrBytes}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		cancel()
		return err
	}
	s.cmd = cmd
	s.stdin = stdin
	s.stdout = bufio.NewReaderSize(stdout, 64<<10)
	s.stderr = stderr
	s.cancel = cancel
	return nil
}

func (s *MCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopLocked()
}

func (s *MCPServer) stopLocked() error {
	if s.cancel != nil {
		s.cancel()
	}
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	stderr := ""
	if s.stderr != nil {
		stderr = strings.TrimSpace(RedactDLP(s.stderr.String()))
	}
	var err error
	if s.cmd != nil {
		_ = terminateMCPProcess(s.cmd)
		err = s.cmd.Wait()
	}
	if s.cleanupDirectory && s.config.WorkingDirectory != "" {
		_ = os.RemoveAll(s.config.WorkingDirectory)
	}
	s.cmd, s.stdin, s.stdout, s.stderr, s.cancel = nil, nil, nil, nil, nil
	if err != nil && stderr != "" {
		return fmt.Errorf("%w: %s", err, stderr)
	}
	return err
}

func (s *MCPServer) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	method = strings.TrimSpace(method)
	if method == "" {
		return nil, errors.New("MCP method is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.allowedMethods) == 0 || !s.allowedMethods[method] {
		return nil, fmt.Errorf("MCP method %q is not allowlisted", method)
	}
	s.nextID++
	requestID := s.nextID
	request := map[string]any{"jsonrpc": "2.0", "id": requestID, "method": method, "params": params}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	if len(data) > mcpMaxMessageBytes {
		return nil, errors.New("MCP request payload limit exceeded")
	}
	if err := s.startLocked(); err != nil {
		return nil, err
	}
	if _, err := s.stdin.Write(append(data, '\n')); err != nil {
		_ = s.stopLocked()
		return nil, err
	}
	resultChannel := make(chan mcpResponse, 1)
	go func() {
		line, err := readMCPMessage(s.stdout)
		if err != nil {
			resultChannel <- mcpResponse{err: err}
			return
		}
		var response mcpResponse
		if err := json.Unmarshal(line, &response); err != nil {
			resultChannel <- mcpResponse{err: err}
			return
		}
		resultChannel <- response
	}()
	timeout := time.Duration(s.config.TimeoutSeconds) * time.Second
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < timeout {
		timeout = time.Until(deadline)
	}
	select {
	case <-ctx.Done():
		_ = s.stopLocked()
		return nil, ctx.Err()
	case <-time.After(timeout):
		_ = s.stopLocked()
		return nil, errors.New("MCP call timed out")
	case response := <-resultChannel:
		if response.err != nil {
			_ = s.stopLocked()
			return nil, response.err
		}
		if response.ID != requestID {
			_ = s.stopLocked()
			return nil, errors.New("MCP response id does not match request")
		}
		if response.Error != nil {
			return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
		}
		return response.Result, nil
	}
}

func readMCPMessage(reader *bufio.Reader) ([]byte, error) {
	message := make([]byte, 0, 4096)
	for {
		fragment, err := reader.ReadSlice('\n')
		if len(message)+len(fragment) > mcpMaxMessageBytes {
			return nil, errors.New("MCP response payload limit exceeded")
		}
		message = append(message, fragment...)
		if err == nil {
			return bytes.TrimSpace(message), nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}
		return nil, err
	}
}

type mcpResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *mcpError       `json:"error,omitempty"`
	err    error
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpCallTool struct{ manager *MCPManager }

func (t mcpCallTool) Descriptor() ToolDescriptor {
	return ToolDescriptor{Name: "mcp.call", Version: "1", Description: "Chamar método allowlisted de servidor MCP stdio", Risk: RiskExternalSideEffect, RequiresApproval: true, Scopes: []string{"mcp:call"}}
}

func (t mcpCallTool) Execute(ctx context.Context, _ ToolContext, input map[string]any) (ToolResult, error) {
	if t.manager == nil {
		return ToolResult{}, errors.New("MCP manager is unavailable")
	}
	result, err := t.manager.Call(ctx, stringInput(input, "server_id", ""), stringInput(input, "method", ""), input["params"])
	if err != nil {
		return ToolResult{}, err
	}
	var value any
	if err := json.Unmarshal(result, &value); err != nil {
		return ToolResult{Value: string(result)}, nil
	}
	return ToolResult{Value: value}, nil
}
