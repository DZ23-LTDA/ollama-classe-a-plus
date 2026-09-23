package agent

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DeployConfig struct {
	ID             string `json:"id"`
	Provider       string `json:"provider"`
	BaseURL        string `json:"base_url"`
	TokenEnv       string `json:"token_env,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	AccountID      string `json:"account_id,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

type DeploymentRequest struct {
	Name   string
	Root   string
	Target string
}

type DeploymentResult struct {
	Provider     string `json:"provider"`
	DeploymentID string `json:"deployment_id,omitempty"`
	URL          string `json:"url,omitempty"`
	Status       string `json:"status"`
	Files        int    `json:"files"`
}

// DeploymentError preserves the provider state that is known when a deploy
// fails after a remote side effect. Callers must not report the operation as a
// clean failure when the provider may have created a site or deployment.
type DeploymentError struct {
	Result DeploymentResult
	Err    error
}

func (e *DeploymentError) Error() string {
	if e == nil || e.Err == nil {
		return "deployment failed with unknown provider state"
	}
	return e.Err.Error()
}

func (e *DeploymentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type DeploymentManager struct {
	configs map[string]DeployConfig
	client  *http.Client
}

type deploymentLoopbackContextKey struct{}

func NewDeploymentManager() *DeploymentManager {
	return &DeploymentManager{configs: map[string]DeployConfig{}, client: newDeploymentHTTPClient()}
}

func newDeploymentHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = deploymentDialContext
	return &http.Client{Timeout: 120 * time.Second, Transport: transport, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return errors.New("deployment redirects are disabled") }}
}

func deploymentRequestContext(ctx context.Context, rawURL string) (context.Context, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" {
		return nil, errors.New("deployment URL is invalid")
	}
	return context.WithValue(ctx, deploymentLoopbackContextKey{}, isLoopbackHost(parsed.Hostname())), nil
}

func deploymentDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return deploymentDialContextWithResolver(ctx, network, address, net.DefaultResolver.LookupIPAddr)
}

func deploymentDialContextWithResolver(ctx context.Context, network, address string, lookup func(context.Context, string) ([]net.IPAddr, error)) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	if loopback, _ := ctx.Value(deploymentLoopbackContextKey{}).(bool); loopback {
		return dialer.DialContext(ctx, network, address)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil || host == "" || port == "" {
		return nil, errors.New("deployment destination address is invalid")
	}
	var addresses []net.IPAddr
	if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil {
		addresses = []net.IPAddr{{IP: ip}}
	} else {
		if lookup == nil {
			return nil, errors.New("deployment destination resolver is unavailable")
		}
		addresses, err = lookup(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("deployment destination lookup failed: %w", err)
		}
	}
	if len(addresses) == 0 {
		return nil, errors.New("deployment destination has no addresses")
	}
	for _, address := range addresses {
		if deploymentPrivateIP(address.IP) {
			return nil, errors.New("deployment destination resolves to a private address")
		}
	}
	var lastErr error
	for _, address := range addresses {
		if network == "tcp4" && address.IP.To4() == nil {
			continue
		}
		if network == "tcp6" && address.IP.To4() != nil {
			continue
		}
		target := net.JoinHostPort(address.IP.String(), port)
		if address.Zone != "" {
			target = net.JoinHostPort(address.IP.String()+"%"+address.Zone, port)
		}
		conn, dialErr := dialer.DialContext(ctx, network, target)
		if dialErr == nil {
			return conn, nil
		}
		lastErr = dialErr
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("deployment destination has no address for requested network")
}

func deploymentPrivateIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsUnspecified() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast()
}

func (m *DeploymentManager) Register(config DeployConfig) error {
	config.ID = strings.TrimSpace(config.ID)
	config.Provider = strings.ToLower(strings.TrimSpace(config.Provider))
	if config.ID == "" || config.Provider == "" {
		return errors.New("deployment id and provider are required")
	}
	if config.Provider != "vercel" && config.Provider != "netlify" && config.Provider != "generic" {
		return fmt.Errorf("unsupported deployment provider %q", config.Provider)
	}
	base, err := url.Parse(strings.TrimSpace(config.BaseURL))
	if err != nil || base == nil || base.Host == "" || base.User != nil {
		return errors.New("deployment base_url must be HTTPS or loopback HTTP without userinfo")
	}
	loopbackHTTP := base.Scheme == "http" && isLoopbackHost(base.Hostname())
	if base.Scheme != "https" && !loopbackHTTP {
		return errors.New("deployment base_url must be HTTPS or loopback HTTP without userinfo")
	}
	if config.TokenEnv != "" && !validEnvName(config.TokenEnv) {
		return errors.New("invalid deployment token_env")
	}
	if config.TimeoutSeconds <= 0 || config.TimeoutSeconds > 600 {
		config.TimeoutSeconds = 120
	}
	m.configs[config.ID] = config
	return nil
}

func (m *DeploymentManager) List() []DeployConfig {
	result := make([]DeployConfig, 0, len(m.configs))
	for _, config := range m.configs {
		config.TokenEnv = ""
		result = append(result, config)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (m *DeploymentManager) Deploy(ctx context.Context, providerID string, request DeploymentRequest) (DeploymentResult, error) {
	if m == nil {
		return DeploymentResult{}, errors.New("deployment manager is unavailable")
	}
	config, ok := m.configs[strings.TrimSpace(providerID)]
	if !ok {
		return DeploymentResult{}, fmt.Errorf("deployment provider %q is not registered", providerID)
	}
	files, err := collectDeployFiles(request.Root)
	if err != nil {
		return DeploymentResult{}, err
	}
	if len(files) == 0 {
		return DeploymentResult{}, errors.New("deployment workspace has no files")
	}
	var result DeploymentResult
	switch config.Provider {
	case "vercel":
		result, err = m.deployVercel(ctx, config, request, files)
	case "netlify":
		result, err = m.deployNetlify(ctx, config, request, files)
	default:
		result, err = m.deployGeneric(ctx, config, request, files)
	}
	if err != nil {
		result.Provider = config.Provider
		if result.Status == "" {
			result.Status = "unknown"
		}
		if result.Status == "unknown" && result.Files == 0 {
			result.Files = len(files)
		}
		return result, &DeploymentError{Result: result, Err: err}
	}
	result.Provider = config.Provider
	result.Files = len(files)
	return result, nil
}

type deployFile struct {
	Path string
	Data []byte
}

func collectDeployFiles(root string) ([]deployFile, error) {
	root, err := filepath.Abs(strings.TrimSpace(root))
	if err != nil || root == "." || root == string(filepath.Separator) {
		return nil, errors.New("invalid deployment root")
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, errors.New("deployment root is not a directory")
	}
	rootInfo, err := os.Lstat(root)
	if err != nil || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("deployment root must not be a symlink")
	}
	var files []deployFile
	var total int64
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			if relative == "." && path == root {
				return nil
			}
			return errors.New("deployment path escaped root")
		}
		relative = filepath.ToSlash(relative)
		if info.IsDir() {
			if deploymentPathContainsPrivateDirectory(relative) {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return errors.New("deployment workspace contains a non-regular file")
		}
		if !deploymentPathIsPublic(relative) {
			return nil
		}
		if len(files) >= 2000 || total+info.Size() > 50<<20 {
			return errors.New("deployment workspace exceeds file or size limit")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files = append(files, deployFile{Path: relative, Data: data})
		total += int64(len(data))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

func deploymentPathContainsPrivateDirectory(relative string) bool {
	for _, part := range strings.Split(filepath.ToSlash(relative), "/") {
		switch strings.ToLower(part) {
		case ".git", ".hg", ".svn", ".agent", ".ollama", ".secrets", "node_modules":
			return true
		}
	}
	return false
}

func deploymentPathIsPublic(relative string) bool {
	if deploymentPathContainsPrivateDirectory(relative) {
		return false
	}
	base := strings.ToLower(filepath.Base(relative))
	if base == ".env" || strings.HasPrefix(base, ".env.") || strings.HasSuffix(base, ".pem") || strings.HasSuffix(base, ".key") || strings.HasSuffix(base, ".crt") || strings.HasSuffix(base, ".p12") || strings.HasSuffix(base, ".pfx") {
		return false
	}
	for _, marker := range []string{"secret", "credential", "password", "token", "apikey", "api_key", "backup", "dump", "log"} {
		if strings.Contains(base, marker) {
			return false
		}
	}
	return true
}

func (m *DeploymentManager) request(ctx context.Context, config DeployConfig, method, endpoint string, body []byte, contentType string) (map[string]any, error) {
	base, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, err
	}
	relative, err := url.Parse(endpoint)
	if err != nil || relative.IsAbs() || !validConnectorPath(relative.Path) {
		return nil, errors.New("invalid deployment endpoint")
	}
	base.Path = strings.TrimSuffix(base.Path, "/") + relative.Path
	base.RawQuery = relative.RawQuery
	requestContext, err := deploymentRequestContext(ctx, base.String())
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(requestContext, method, base.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if config.TokenEnv != "" {
		if token := os.Getenv(config.TokenEnv); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	client := *m.client
	client.Timeout = time.Duration(config.TimeoutSeconds) * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("deployment provider returned status %d: %s", resp.StatusCode, limitError(string(data), 800))
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return map[string]any{"raw": string(data)}, nil
	}
	return payload, nil
}

func (m *DeploymentManager) deployGeneric(ctx context.Context, config DeployConfig, request DeploymentRequest, files []deployFile) (DeploymentResult, error) {
	payload := map[string]any{"name": request.Name, "target": request.Target, "files": encodeFiles(files)}
	body, _ := json.Marshal(payload)
	response, err := m.request(ctx, config, http.MethodPost, "/deploy", body, "application/json")
	if err != nil {
		return DeploymentResult{}, err
	}
	return resultFromPayload(response), nil
}

func (m *DeploymentManager) deployVercel(ctx context.Context, config DeployConfig, request DeploymentRequest, files []deployFile) (DeploymentResult, error) {
	payload := map[string]any{"name": request.Name, "target": request.Target, "files": encodeFiles(files)}
	if config.ProjectID != "" {
		payload["project"] = config.ProjectID
	}
	body, _ := json.Marshal(payload)
	endpoint := "/v13/deployments"
	if config.AccountID != "" {
		endpoint += "?teamId=" + url.QueryEscape(config.AccountID)
	}
	response, err := m.request(ctx, config, http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		return DeploymentResult{}, err
	}
	return resultFromPayload(response), nil
}

func (m *DeploymentManager) deployNetlify(ctx context.Context, config DeployConfig, request DeploymentRequest, files []deployFile) (DeploymentResult, error) {
	siteID := config.ProjectID
	var site map[string]any
	var err error
	if siteID == "" {
		createPayload, _ := json.Marshal(map[string]any{"name": request.Name})
		site, err = m.request(ctx, config, http.MethodPost, "/api/v1/sites", createPayload, "application/json")
		if err != nil {
			return DeploymentResult{}, err
		}
		siteID = firstString(site, "id", "site_id")
	}
	if siteID == "" {
		return DeploymentResult{}, errors.New("netlify response has no site id")
	}
	digests := map[string]string{}
	for _, file := range files {
		digest := sha1.Sum(file.Data)
		digests["/"+file.Path] = hex.EncodeToString(digest[:])
	}
	deployPayload, _ := json.Marshal(map[string]any{"files": digests})
	deploy, err := m.request(ctx, config, http.MethodPost, "/api/v1/sites/"+url.PathEscape(siteID)+"/deploys", deployPayload, "application/json")
	if err != nil {
		return DeploymentResult{DeploymentID: siteID, Status: "partial"}, fmt.Errorf("netlify deployment creation failed after site creation: %w", err)
	}
	deployID := firstString(deploy, "id", "deploy_id")
	if deployID == "" {
		return DeploymentResult{DeploymentID: siteID, Status: "partial"}, errors.New("netlify response has no deployment id after site/deploy side effects")
	}
	result := resultFromPayload(deploy)
	result.Status = "partial"
	result.Files = 0
	for _, file := range files {
		endpoint := "/api/v1/deploys/" + url.PathEscape(deployID) + "/files/" + url.PathEscape(file.Path)
		if _, err := m.request(ctx, config, http.MethodPut, endpoint, file.Data, "application/octet-stream"); err != nil {
			return result, fmt.Errorf("netlify file upload failed after %d files: %w", result.Files, err)
		}
		result.Files++
	}
	result.Status = firstString(deploy, "status", "state")
	if result.Status == "" {
		result.Status = "accepted"
	}
	return result, nil
}

func encodeFiles(files []deployFile) []map[string]string {
	encoded := make([]map[string]string, 0, len(files))
	for _, file := range files {
		encoded = append(encoded, map[string]string{"file": file.Path, "data": base64.StdEncoding.EncodeToString(file.Data)})
	}
	return encoded
}

func resultFromPayload(payload map[string]any) DeploymentResult {
	return DeploymentResult{DeploymentID: firstString(payload, "id", "deployment_id", "deploy_id"), URL: firstString(payload, "url", "deploy_url", "ssl_url"), Status: firstString(payload, "status", "state")}
}

func firstString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
