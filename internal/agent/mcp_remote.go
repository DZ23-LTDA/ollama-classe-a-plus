package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type RemoteMCPServerConfig struct {
	ID             string   `json:"id"`
	URL            string   `json:"url"`
	TokenEnv       string   `json:"token_env,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty"`
	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
}

type RemoteMCPManager struct {
	mu      sync.RWMutex
	client  *http.Client
	servers map[string]RemoteMCPServerConfig
}

func NewRemoteMCPManager() *RemoteMCPManager {
	return &RemoteMCPManager{client: &http.Client{}, servers: map[string]RemoteMCPServerConfig{}}
}

func (m *RemoteMCPManager) Register(config RemoteMCPServerConfig) error {
	config.ID = strings.TrimSpace(config.ID)
	config.URL = strings.TrimSpace(config.URL)
	if config.ID == "" || config.URL == "" {
		return errors.New("remote MCP id and url are required")
	}
	parsed, err := url.Parse(config.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return errors.New("remote MCP url must be an absolute URL without credentials or fragment")
	}
	if parsed.Scheme != "https" && !remoteMCPLoopback(parsed.Hostname()) {
		return errors.New("remote MCP requires HTTPS outside loopback")
	}
	if strings.ContainsAny(config.TokenEnv, "=\x00\r\n") {
		return errors.New("remote MCP token_env is invalid")
	}
	if config.TimeoutSeconds <= 0 || config.TimeoutSeconds > 300 {
		config.TimeoutSeconds = 30
	}
	allowed := make([]string, 0, len(config.AllowedMethods))
	for _, method := range config.AllowedMethods {
		if method = strings.TrimSpace(method); method != "" {
			allowed = append(allowed, method)
		}
	}
	config.AllowedMethods = allowed
	m.mu.Lock()
	defer m.mu.Unlock()
	m.servers[config.ID] = config
	return nil
}

func remoteMCPLoopback(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (m *RemoteMCPManager) List() []RemoteMCPServerConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]RemoteMCPServerConfig, 0, len(m.servers))
	for _, config := range m.servers {
		config.AllowedMethods = append([]string(nil), config.AllowedMethods...)
		result = append(result, config)
	}
	return result
}

func (m *RemoteMCPManager) Call(ctx context.Context, serverID, method string, params any) (json.RawMessage, error) {
	m.mu.RLock()
	config, ok := m.servers[strings.TrimSpace(serverID)]
	client := m.client
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("remote MCP server %q is not registered", serverID)
	}
	method = strings.TrimSpace(method)
	if method == "" {
		return nil, errors.New("remote MCP method is required")
	}
	if len(config.AllowedMethods) > 0 && !remoteMCPContains(config.AllowedMethods, method) {
		return nil, fmt.Errorf("remote MCP method %q is not allowlisted", method)
	}
	requestBody, err := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": time.Now().UnixNano(), "method": method, "params": params})
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(config.TimeoutSeconds) * time.Second
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < timeout {
		timeout = time.Until(deadline)
	}
	requestContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestContext, http.MethodPost, config.URL, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if config.TokenEnv != "" {
		if token, ok := os.LookupEnv(config.TokenEnv); ok && strings.TrimSpace(token) != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 16<<10))
		return nil, fmt.Errorf("remote MCP returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	payload := body
	if strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "text/event-stream") {
		payload = lastSSEData(body)
	}
	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *mcpError       `json:"error,omitempty"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, fmt.Errorf("decode remote MCP response: %w", err)
	}
	if envelope.Error != nil {
		return nil, fmt.Errorf("remote MCP error: %s", envelope.Error.Message)
	}
	if len(envelope.Result) == 0 {
		return nil, errors.New("remote MCP response has no result")
	}
	return envelope.Result, nil
}

func remoteMCPContains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func lastSSEData(body []byte) []byte {
	var last []byte
	for _, line := range bytes.Split(body, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, []byte("data:")) {
			value := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(value) > 0 && !bytes.Equal(value, []byte("[DONE]")) {
				last = value
			}
		}
	}
	return last
}

type remoteMCPCallTool struct{ manager *RemoteMCPManager }

func (t remoteMCPCallTool) Descriptor() ToolDescriptor {
	return ToolDescriptor{Name: "mcp.remote.call", Version: "1", Description: "Chamar método allowlisted de um Remote MCP Streamable HTTP", Risk: RiskExternalSideEffect, RequiresApproval: true, Scopes: []string{"mcp:remote:call"}}
}

func (t remoteMCPCallTool) Execute(ctx context.Context, _ ToolContext, input map[string]any) (ToolResult, error) {
	if t.manager == nil {
		return ToolResult{}, errors.New("remote MCP manager is unavailable")
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
