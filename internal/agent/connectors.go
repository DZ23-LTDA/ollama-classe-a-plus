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
	"path"
	"strings"
	"sync"
	"time"
)

type ConnectorConfig struct {
	ID             string               `json:"id"`
	OrganizationID string               `json:"organization_id,omitempty"`
	Provider       string               `json:"provider"`
	BaseURL        string               `json:"base_url"`
	TokenEnv       string               `json:"token_env,omitempty"`
	OAuthProvider  string               `json:"oauth_provider,omitempty"`
	AllowedOrigins []string             `json:"allowed_origins,omitempty"`
	Operations     []ConnectorOperation `json:"operations"`
	TimeoutSeconds int                  `json:"timeout_seconds,omitempty"`
	Disabled       bool                 `json:"disabled,omitempty"`
}

type ConnectorOperation struct {
	Name         string   `json:"name"`
	Methods      []string `json:"methods"`
	PathPrefixes []string `json:"path_prefixes"`
}

type ConnectorManager struct {
	mu         sync.RWMutex
	connectors map[string]ConnectorConfig
	client     *http.Client
	auth       *AuthStore
}

var ErrConnectorDisabled = errors.New("connector is disabled")

type connectorLoopbackContextKey struct{}

func NewConnectorManager() *ConnectorManager {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = connectorDialContext
	return &ConnectorManager{connectors: make(map[string]ConnectorConfig), client: &http.Client{Transport: transport, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("connector redirects are disabled") }}}
}

func (m *ConnectorManager) SetOAuthStore(store *AuthStore) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.auth = store
	m.mu.Unlock()
}

func (m *ConnectorManager) Register(config ConnectorConfig) error {
	config.ID = strings.TrimSpace(config.ID)
	config.Provider = strings.TrimSpace(config.Provider)
	if config.ID == "" || config.Provider == "" {
		return errors.New("connector id and provider are required")
	}
	base, err := url.Parse(strings.TrimSpace(config.BaseURL))
	if err != nil || base.Scheme != "https" || base.Host == "" || base.User != nil {
		return errors.New("connector base_url must be an https URL without userinfo")
	}
	if config.TimeoutSeconds <= 0 || config.TimeoutSeconds > 120 {
		config.TimeoutSeconds = 30
	}
	if config.TokenEnv != "" && !validEnvName(config.TokenEnv) {
		return errors.New("invalid connector token_env")
	}
	for i := range config.Operations {
		operation := &config.Operations[i]
		operation.Name = strings.TrimSpace(operation.Name)
		if operation.Name == "" || len(operation.Methods) == 0 || len(operation.PathPrefixes) == 0 {
			return errors.New("connector operations require name, methods and path_prefixes")
		}
		for j := range operation.Methods {
			operation.Methods[j] = strings.ToUpper(strings.TrimSpace(operation.Methods[j]))
		}
		for j := range operation.PathPrefixes {
			if !validConnectorPath(operation.PathPrefixes[j]) {
				return fmt.Errorf("invalid connector path prefix %q", operation.PathPrefixes[j])
			}
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectors[config.ID] = config
	return nil
}

func (m *ConnectorManager) List() []ConnectorConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ConnectorConfig, 0, len(m.connectors))
	for _, connector := range m.connectors {
		copy := connector
		copy.TokenEnv = ""
		result = append(result, copy)
	}
	return result
}

func (m *ConnectorManager) ListForOrganization(organizationID string) []ConnectorConfig {
	organizationID = strings.TrimSpace(organizationID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]ConnectorConfig, 0)
	for _, connector := range m.connectors {
		if connector.OrganizationID != "" && !pluginOwnedByOrganization(connector.OrganizationID, organizationID) {
			continue
		}
		copy := connector
		copy.TokenEnv = ""
		result = append(result, copy)
	}
	return result
}

func (m *ConnectorManager) SetEnabled(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	connector, ok := m.connectors[id]
	if !ok {
		return fmt.Errorf("connector %q is not registered", id)
	}
	connector.Disabled = !enabled
	m.connectors[id] = connector
	return nil
}

func (m *ConnectorManager) SetEnabledForOrganization(organizationID, id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	connector, ok := m.connectors[strings.TrimSpace(id)]
	if !ok {
		return fmt.Errorf("connector %q is not registered", id)
	}
	if !pluginOwnedByOrganization(connector.OrganizationID, strings.TrimSpace(organizationID)) {
		return ErrPluginOrganizationScope
	}
	connector.Disabled = !enabled
	m.connectors[connector.ID] = connector
	return nil
}

func (m *ConnectorManager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	if _, ok := m.connectors[id]; !ok {
		return fmt.Errorf("connector %q is not registered", id)
	}
	delete(m.connectors, id)
	return nil
}

func (m *ConnectorManager) RemoveForOrganization(organizationID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	id = strings.TrimSpace(id)
	connector, ok := m.connectors[id]
	if !ok {
		return fmt.Errorf("connector %q is not registered", id)
	}
	if !pluginOwnedByOrganization(connector.OrganizationID, strings.TrimSpace(organizationID)) {
		return ErrPluginOrganizationScope
	}
	delete(m.connectors, id)
	return nil
}

func (m *ConnectorManager) Call(ctx context.Context, connectorID, operationName, method, requestPath string, body []byte) (int, string, error) {
	return 0, "", errors.New("connector organization scope is required; use CallForOrganization")
}

func (m *ConnectorManager) CallForOrganization(ctx context.Context, organizationID, connectorID, operationName, method, requestPath string, body []byte) (int, string, error) {
	if strings.TrimSpace(organizationID) == "" {
		return 0, "", errors.New("connector organization scope is required")
	}
	m.mu.RLock()
	config, ok := m.connectors[strings.TrimSpace(connectorID)]
	auth := m.auth
	m.mu.RUnlock()
	if !ok {
		return 0, "", fmt.Errorf("connector %q is not registered", connectorID)
	}
	token := ""
	if auth != nil && strings.TrimSpace(config.OAuthProvider) != "" {
		var err error
		token, _, err = auth.OAuthAccessTokenForOrganization(organizationID, config.OAuthProvider)
		if err != nil {
			return 0, "", fmt.Errorf("resolve OAuth credential for connector %q: %w", connectorID, err)
		}
	}
	return m.call(ctx, connectorID, operationName, method, requestPath, body, token)
}

func (m *ConnectorManager) call(ctx context.Context, connectorID, operationName, method, requestPath string, body []byte, tokenOverride string) (int, string, error) {
	m.mu.RLock()
	config, ok := m.connectors[strings.TrimSpace(connectorID)]
	m.mu.RUnlock()
	if !ok {
		return 0, "", fmt.Errorf("connector %q is not registered", connectorID)
	}
	if config.Disabled {
		return 0, "", ErrConnectorDisabled
	}
	operation, allowed := findConnectorOperation(config.Operations, operationName, method, requestPath)
	if !allowed {
		return 0, "", fmt.Errorf("connector operation %q is not allowlisted", operationName)
	}
	_ = operation
	if len(body) > 1<<20 {
		return 0, "", errors.New("connector request payload exceeds limit")
	}
	base, _ := url.Parse(config.BaseURL)
	relative, err := url.Parse(requestPath)
	if err != nil || relative.IsAbs() || !validConnectorPath(relative.Path) {
		return 0, "", errors.New("invalid connector request path")
	}
	base.Path = path.Join(strings.TrimSuffix(base.Path, "/"), relative.Path)
	base.RawQuery = relative.RawQuery
	requestContext := context.WithValue(ctx, connectorLoopbackContextKey{}, connectorHostIsLoopback(base.Hostname()))
	request, err := http.NewRequestWithContext(requestContext, strings.ToUpper(method), base.String(), bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	request.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		request.Header.Set("Content-Type", "application/json")
	}
	token := tokenOverride
	if token == "" && config.TokenEnv != "" {
		token = os.Getenv(config.TokenEnv)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	client := connectorClientForRequest(m.client, time.Duration(config.TimeoutSeconds)*time.Second)
	response, err := client.Do(request)
	if err != nil {
		return 0, "", err
	}
	defer response.Body.Close()
	data, err := readLimitedConnectorBody(response.Body, 2<<20)
	if err != nil {
		return response.StatusCode, "", err
	}
	return response.StatusCode, string(data), nil
}

func connectorClientForRequest(base *http.Client, timeout time.Duration) *http.Client {
	if base == nil {
		client := NewConnectorManager().client
		client.Timeout = timeout
		return client
	}
	client := *base
	client.Timeout = timeout
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return errors.New("connector redirects are disabled") }
	switch transport := base.Transport.(type) {
	case nil:
		safe := http.DefaultTransport.(*http.Transport).Clone()
		safe.Proxy = nil
		safe.DialContext = connectorDialContext
		client.Transport = safe
	case *http.Transport:
		safe := transport.Clone()
		safe.Proxy = nil
		safe.DialContext = connectorDialContext
		client.Transport = safe
	default:
		safe := http.DefaultTransport.(*http.Transport).Clone()
		safe.Proxy = nil
		safe.DialContext = connectorDialContext
		client.Transport = safe
	}
	return &client
}

func readLimitedConnectorBody(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("connector response payload exceeds limit")
	}
	return data, nil
}

func findConnectorOperation(operations []ConnectorOperation, name, method, requestPath string) (ConnectorOperation, bool) {
	method = strings.ToUpper(strings.TrimSpace(method))
	for _, operation := range operations {
		if operation.Name != name {
			continue
		}
		methodOK := false
		for _, allowedMethod := range operation.Methods {
			if strings.ToUpper(allowedMethod) == method {
				methodOK = true
				break
			}
		}
		if !methodOK {
			return operation, false
		}
		for _, prefix := range operation.PathPrefixes {
			if connectorPathMatches(requestPath, prefix) {
				return operation, true
			}
		}
	}
	return ConnectorOperation{}, false
}

func connectorPathMatches(requestPath, prefix string) bool {
	requestPath = strings.TrimSpace(requestPath)
	prefix = strings.TrimSpace(prefix)
	if prefix == "/" {
		return strings.HasPrefix(requestPath, "/")
	}
	prefix = strings.TrimSuffix(prefix, "/")
	return requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/")
}

func connectorHostIsLoopback(host string) bool {
	if strings.EqualFold(strings.TrimSpace(host), "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func connectorPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast()
}

func connectorDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	remote, _, splitErr := net.SplitHostPort(conn.RemoteAddr().String())
	if splitErr != nil {
		_ = conn.Close()
		return nil, errors.New("connector destination is invalid")
	}
	if ip := net.ParseIP(remote); ip != nil && connectorPrivateIP(ip) {
		allowedLoopback, _ := ctx.Value(connectorLoopbackContextKey{}).(bool)
		if !(allowedLoopback && ip.IsLoopback()) {
			_ = conn.Close()
			return nil, errors.New("connector destination resolves to a private address")
		}
	}
	return conn, nil
}

func validConnectorPath(value string) bool {
	return strings.HasPrefix(value, "/") && !strings.Contains(value, "..") && !strings.ContainsAny(value, "\x00\r\n")
}

func validEnvName(value string) bool {
	if value == "" {
		return false
	}
	for index, char := range value {
		if !(char == '_' || char >= 'A' && char <= 'Z' || index > 0 && char >= '0' && char <= '9') {
			return false
		}
	}
	return true
}

type connectorTool struct{ manager *ConnectorManager }

func (t connectorTool) Descriptor() ToolDescriptor {
	return ToolDescriptor{Name: "connector.http", Version: "1", Description: "Chamar operação allowlisted de um conector externo", Risk: RiskExternalSideEffect, RequiresApproval: true, Scopes: []string{"connector:external"}}
}

func (t connectorTool) Execute(ctx context.Context, toolContext ToolContext, input map[string]any) (ToolResult, error) {
	if t.manager == nil {
		return ToolResult{}, errors.New("connector manager is unavailable")
	}
	body := []byte(stringInput(input, "body", ""))
	if len(body) > 1<<20 {
		return ToolResult{}, errors.New("connector body limit exceeded")
	}
	status, response, err := t.manager.CallForOrganization(ctx, toolContext.OrganizationID, stringInput(input, "connector_id", ""), stringInput(input, "operation", ""), stringInput(input, "method", "GET"), stringInput(input, "path", "/"), body)
	if err != nil {
		return ToolResult{}, err
	}
	var value any
	if json.Unmarshal([]byte(response), &value) != nil {
		value = response
	}
	return ToolResult{Value: map[string]any{"status": status, "response": value}}, nil
}
