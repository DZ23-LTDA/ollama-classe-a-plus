package server

import (
	"testing"
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
