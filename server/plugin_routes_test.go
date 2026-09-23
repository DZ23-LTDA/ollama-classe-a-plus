package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func TestRequirePluginAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name string
		role agent.Role
		want bool
	}{
		{name: "owner", role: agent.RoleOwner, want: true},
		{name: "admin", role: agent.RoleAdmin, want: true},
		{name: "operator", role: agent.RoleOperator, want: false},
		{name: "viewer", role: agent.RoleViewer, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Set("agent.membership", agent.Membership{Role: test.role})
			api := &agentAPI{authRequired: true}
			if got := api.requirePluginAdmin(context); got != test.want {
				t.Fatalf("requirePluginAdmin()=%v, want %v", got, test.want)
			}
			if !test.want && recorder.Code != 403 {
				t.Fatalf("status=%d, want 403", recorder.Code)
			}
		})
	}
}

func TestRequirePluginAdminAllowsExplicitLocalMode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	api := &agentAPI{authRequired: false}
	if !api.requirePluginAdmin(context) {
		t.Fatal("local loopback mode should allow local plugin lifecycle")
	}
}
