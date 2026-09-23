package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeploymentManagerGenericProvider(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>DZ23</h1>"), 0o600); err != nil {
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
	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil && conn != nil {
			_ = conn.Close()
		}
		close(accepted)
	}()
	_, err = deploymentDialContext(context.Background(), "tcp", listener.Addr().String())
	<-accepted
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected private deployment dial rejection, got %v", err)
	}
}

func TestDeploymentAllowsLocalhostHTTPConfiguration(t *testing.T) {
	manager := NewDeploymentManager()
	if err := manager.Register(DeployConfig{ID: "local", Provider: "generic", BaseURL: "http://localhost:43123"}); err != nil {
		t.Fatalf("localhost deployment config rejected: %v", err)
	}
}
