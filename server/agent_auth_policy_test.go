package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func TestAgentAuthRequiredDefaultsToLoopbackOnly(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "127.0.0.1:11434")
	t.Setenv("OLLAMA_AGENT_AUTH_REQUIRED", "")
	required, err := agentAuthRequired()
	if err != nil {
		t.Fatal(err)
	}
	if required {
		t.Fatal("loopback default should allow local-first mode without auth")
	}

	t.Setenv("OLLAMA_HOST", "0.0.0.0:11434")
	required, err = agentAuthRequired()
	if err != nil {
		t.Fatal(err)
	}
	if !required {
		t.Fatal("non-loopback bind must require auth by default")
	}
}

func TestAgentAuthCannotBeDisabledOnNonLoopbackBind(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "0.0.0.0:11434")
	t.Setenv("OLLAMA_AGENT_AUTH_REQUIRED", "false")
	required, err := agentAuthRequired()
	if err != nil {
		t.Fatal(err)
	}
	if !required {
		t.Fatal("non-loopback bind must not accept auth=false")
	}
}

func TestOnlyCompanionConnectRouteUsesDeviceHandshake(t *testing.T) {
	if !isCompanionConnectRoute("/api/agent/v1/devices/:id/connect") {
		t.Fatal("device connect route should use the device handshake")
	}
	for _, path := range []string{
		"/api/agent/v1/missions/:id/connect",
		"/api/agent/v1/connect",
		"/api/agent/v1/devices/:id/connect/extra",
	} {
		if isCompanionConnectRoute(path) {
			t.Fatalf("unexpected auth bypass for %q", path)
		}
	}
}

func TestDevTokenOnlyAllowsActualLoopback(t *testing.T) {
	for _, test := range []struct {
		name string
		addr string
		want bool
	}{
		{name: "ipv4 loopback", addr: "127.0.0.1:1234", want: true},
		{name: "ipv6 loopback", addr: "[::1]:1234", want: true},
		{name: "private network is not loopback", addr: "192.168.1.10:1234", want: false},
		{name: "missing peer is denied", addr: "", want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := isLoopbackRemoteAddr(test.addr); got != test.want {
				t.Fatalf("isLoopbackRemoteAddr(%q) = %v, want %v", test.addr, got, test.want)
			}
		})
	}
}

func TestApprovalApproverRequiresOwnerOrAdminWhenAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &agentAPI{authRequired: true}
	for _, test := range []struct {
		role agent.Role
		want bool
	}{
		{role: agent.RoleViewer, want: false},
		{role: agent.RoleOperator, want: false},
		{role: agent.RoleAdmin, want: true},
		{role: agent.RoleOwner, want: true},
	} {
		t.Run(string(test.role), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Set("agent.membership", agent.Membership{Role: test.role})
			if got := api.requireApprovalApprover(context); got != test.want {
				t.Fatalf("role %s approval access=%v want=%v", test.role, got, test.want)
			}
		})
	}
}
