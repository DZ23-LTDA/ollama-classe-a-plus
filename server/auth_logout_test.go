package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/internal/agent"
)

func TestAgentAuthLogoutRevokesBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, err := agent.NewAuthStore("")
	if err != nil {
		t.Fatal(err)
	}
	user, err := auth.CreateUser("logout@example.com", "Logout User")
	if err != nil {
		t.Fatal(err)
	}
	organization, _, err := auth.CreateOrganization("Logout Org", user)
	if err != nil {
		t.Fatal(err)
	}
	raw, _, err := auth.IssueToken(user.ID, organization.ID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	api := &agentAPI{auth: auth, authRequired: true}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/agent/v1/auth/logout", nil)
	ctx.Request.Header.Set("Authorization", "Bearer "+raw)

	api.authLogout(ctx)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("logout status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if _, _, _, err := auth.Authenticate(raw); err == nil {
		t.Fatal("revoked token authenticated")
	}
}
