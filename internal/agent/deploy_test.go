package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeploymentManagerGenericProvider(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>DZ23</h1>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SYNTHETIC_PRIVATE=1"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DZ23_DEPLOY_TOKEN", "deploy-test-token")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/deploy" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer deploy-test-token" {
			t.Fatalf("authorization=%q", r.Header.Get("Authorization"))
		}
		var payload struct {
			Files []struct {
				File string `json:"file"`
				Data string `json:"data"`
			} `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.Files) != 1 || payload.Files[0].File != "index.html" {
			t.Fatalf("files=%+v", payload.Files)
		}
		decoded, err := base64.StdEncoding.DecodeString(payload.Files[0].Data)
		if err != nil || string(decoded) != "<h1>DZ23</h1>" {
			t.Fatalf("decoded payload=%q err=%v", decoded, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"dep_123","url":"https://example.test/d/123","status":"ready"}`))
	}))
	defer server.Close()
	manager := NewDeploymentManager()
	manager.client = server.Client()
	if err := manager.Register(DeployConfig{ID: "self", Provider: "generic", BaseURL: server.URL, TokenEnv: "DZ23_DEPLOY_TOKEN"}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Deploy(context.Background(), "self", DeploymentRequest{Name: "site", Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if result.DeploymentID != "dep_123" || result.URL == "" || result.Files != 1 {
		t.Fatalf("result=%+v", result)
	}
}

func TestDeploymentManagerRejectsExternalHTTP(t *testing.T) {
	manager := NewDeploymentManager()
	if err := manager.Register(DeployConfig{ID: "unsafe", Provider: "generic", BaseURL: "http://example.com"}); err == nil {
		t.Fatal("external HTTP deployment unexpectedly accepted")
	}
}

func TestDeploymentManagerReportsNetlifyPartialState(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{"a.txt": "first", "b.txt": "second"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var uploaded int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sites":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"site_1"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/sites/site_1/deploys":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"dep_1","url":"https://example.test/site_1","state":"building"}`))
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/a.txt"):
			uploaded++
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/b.txt"):
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`provider upload failed`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	manager := NewDeploymentManager()
	manager.client = server.Client()
	if err := manager.Register(DeployConfig{ID: "netlify", Provider: "netlify", BaseURL: server.URL}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Deploy(context.Background(), "netlify", DeploymentRequest{Name: "site", Root: root})
	if err == nil {
		t.Fatal("partial Netlify deploy unexpectedly succeeded")
	}
	var deploymentErr *DeploymentError
	if !errors.As(err, &deploymentErr) {
		t.Fatalf("error=%T %v", err, err)
	}
	if result.Provider != "netlify" || result.Status != "partial" || result.DeploymentID != "dep_1" || result.Files != uploaded || uploaded != 1 {
		t.Fatalf("partial result=%+v uploaded=%d", result, uploaded)
	}
}

func TestDeploymentRejectsSymlinkRoot(t *testing.T) {
	parent := t.TempDir()
	realRoot := filepath.Join(parent, "real")
	if err := os.Mkdir(realRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	linkRoot := filepath.Join(parent, "link")
	if err := os.Symlink(realRoot, linkRoot); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := collectDeployFiles(linkRoot); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink root rejection, got %v", err)
	}
}

func TestDeploymentDialRejectsPrivateConnectedAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan bool, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil && conn != nil {
			_ = conn.Close()
			accepted <- true
			return
		}
		accepted <- false
	}()
	_, err = deploymentDialContext(context.Background(), "tcp", listener.Addr().String())
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected private deployment dial rejection, got %v", err)
	}
	select {
	case contacted := <-accepted:
		if contacted {
			t.Fatal("private deployment destination received TCP before refusal")
		}
	case <-time.After(200 * time.Millisecond):
	}
}

func TestDeploymentPackageExcludesPrivateFiles(t *testing.T) {
	root := t.TempDir()
	fixtures := map[string]string{
		"index.html":      "<h1>public</h1>",
		".env":            "SYNTHETIC_PRIVATE=1",
		".env.production": "SYNTHETIC_PRIVATE=2",
		"server.key":      "SYNTHETIC_PRIVATE=3",
		"database.backup": "SYNTHETIC_PRIVATE=4",
		"runtime.log":     "SYNTHETIC_PRIVATE=5",
	}
	for name, content := range fixtures {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("synthetic"), 0o600); err != nil {
		t.Fatal(err)
	}
	files, err := collectDeployFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Path != "index.html" {
		t.Fatalf("public deployment files=%+v", files)
	}
}

func TestBuildDeploymentManifestReportsIncludedExcludedAndHash(t *testing.T) {
	root := t.TempDir()
	for name, content := range map[string]string{
		"index.html": "<h1>public</h1>",
		".env":       "SYNTHETIC_PRIVATE=1",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	manifest, err := BuildDeploymentManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.SHA256 == "" || len(manifest.Included) != 1 || manifest.Included[0].Path != "index.html" {
		t.Fatalf("manifest included=%+v hash=%q", manifest.Included, manifest.SHA256)
	}
	if len(manifest.Excluded) != 2 {
		t.Fatalf("manifest excluded=%+v", manifest.Excluded)
	}
	for _, entry := range manifest.Excluded {
		if entry.Reason == "" {
			t.Fatalf("excluded entry has no reason: %+v", entry)
		}
	}
	files, err := collectDeployFiles(root)
	if err != nil || len(files) != 1 {
		t.Fatalf("collect files=%+v err=%v", files, err)
	}
}

func TestDeploymentManagerRejectsChangedManifestBeforeProvider(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "index.html")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := BuildDeploymentManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	var requests int
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"dep_should_not_exist"}`))
	}))
	defer server.Close()
	manager := NewDeploymentManager()
	manager.client = server.Client()
	if err := manager.Register(DeployConfig{ID: "self", Provider: "generic", BaseURL: server.URL}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = manager.Deploy(context.Background(), "self", DeploymentRequest{Name: "site", Root: root, ManifestSHA256: manifest.SHA256})
	if err == nil || !strings.Contains(err.Error(), "changed after approval") {
		t.Fatalf("expected changed-manifest rejection, got %v", err)
	}
	if requests != 0 {
		t.Fatalf("provider requests=%d, want zero", requests)
	}
}

func TestDeploymentDialRejectsPrivateResolvedAddressBeforeTCP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan bool, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil && conn != nil {
			_ = conn.Close()
			accepted <- true
			return
		}
		accepted <- false
	}()
	lookup := func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}, {IP: net.ParseIP("203.0.113.8")}}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = deploymentDialContextWithResolver(ctx, "tcp", "public.example:"+portOf(listener.Addr().String()), lookup)
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected pre-resolution private rejection, got %v", err)
	}
	select {
	case contacted := <-accepted:
		if contacted {
			t.Fatal("resolved private deployment destination received TCP")
		}
	case <-time.After(200 * time.Millisecond):
	}
}

func portOf(address string) string {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		panic(err)
	}
	return port
}

func TestDeploymentAllowsLocalhostHTTPConfiguration(t *testing.T) {
	manager := NewDeploymentManager()
	if err := manager.Register(DeployConfig{ID: "local", Provider: "generic", BaseURL: "http://localhost:43123"}); err != nil {
		t.Fatalf("localhost deployment config rejected: %v", err)
	}
}
