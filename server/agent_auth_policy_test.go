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

func TestPublicSSORouteDoesNotBypassAuthenticatedOAuthLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{
		"/api/agent/v1/auth/oauth/github/start",
		"/api/agent/v1/auth/oauth/github/callback",
		"/api/agent/v1/auth/saml/enterprise/start",
		"/api/agent/v1/auth/saml/enterprise/metadata",
		"/api/agent/v1/auth/saml/enterprise/acs",
	} {
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest("GET", path, nil)
		if !isPublicSSORoute(context) {
			t.Fatalf("expected public SSO route: %s", path)
		}
	}
	for _, path := range []string{
		"/api/agent/v1/auth/oauth/github/refresh",
		"/api/agent/v1/auth/oauth/github/revoke",
		"/api/agent/v1/auth/saml/enterprise/disable",
	} {
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest("POST", path, nil)
		if isPublicSSORoute(context) {
			t.Fatalf("unexpected public SSO bypass: %s", path)
		}
	}
}

func TestAgentOriginPolicyBlocksCrossSiteMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name   string
		method string
		origin string
		want   bool
	}{
		{name: "same local web app", method: "POST", origin: "http://localhost:3000", want: true},
		{name: "native bearer client", method: "POST", origin: "", want: true},
		{name: "cross site mutation", method: "POST", origin: "https://evil.example", want: false},
		{name: "safe read", method: "GET", origin: "https://evil.example", want: true},
		{name: "preflight", method: "OPTIONS", origin: "https://evil.example", want: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(test.method, "/api/agent/v1/missions", nil)
			if test.origin != "" {
				context.Request.Header.Set("Origin", test.origin)
			}
			if got := agentOriginAllowed(context); got != test.want {
				t.Fatalf("agentOriginAllowed(%s, %s) = %v, want %v", test.method, test.origin, got, test.want)
			}
		})
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

func TestLocalAuthMiddlewareAppliesOriginPolicyToMutations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := &agentAPI{authRequired: false}
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest("POST", "/api/agent/v1/missions", nil)
	context.Request.Header.Set("Origin", "https://evil.example")

	api.authMiddleware(context)

	if recorder.Code != 403 || !context.IsAborted() {
		t.Fatalf("local cross-site mutation status=%d aborted=%v body=%s", recorder.Code, context.IsAborted(), recorder.Body.String())
	}
}
