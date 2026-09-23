package agent

import (
	"context"
	"encoding/base32"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthStorePersistsRBACAndRejectsCrossTenantAccess(t *testing.T) {
	root := t.TempDir()
	store, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("owner@example.com", "Owner")
	if err != nil {
		t.Fatal(err)
	}
	organization, _, err := store.CreateOrganization("Org A", user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Authorize(user.ID, organization.ID, "write"); err != nil {
		t.Fatal(err)
	}
	other, err := store.CreateUser("other@example.com", "Other")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Authorize(other.ID, organization.ID, "read"); err == nil {
		t.Fatal("cross-tenant authorization unexpectedly succeeded")
	}
	raw, _, err := store.IssueToken(user.ID, organization.ID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, gotOrg, _, err := store.Authenticate(raw); err != nil || gotOrg.ID != organization.ID {
		t.Fatalf("authenticate = %v, org=%+v", err, gotOrg)
	}
	if err := store.RevokeToken(raw); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := store.Authenticate(raw); err == nil {
		t.Fatal("revoked token authenticated")
	}
	reloaded, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.Users()) != 2 {
		t.Fatalf("users after reload = %d", len(reloaded.Users()))
	}
}

func TestOAuthStateIsSingleUseAndPKCESafe(t *testing.T) {
	store, err := NewAuthStore("")
	if err != nil {
		t.Fatal(err)
	}
	state, _, err := store.CreateOAuthState("github", "https://app.example.test/oauth/callback", "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~", "", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeOAuthState(state, "github", "https://app.example.test/oauth/callback"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ConsumeOAuthState(state, "github", "https://app.example.test/oauth/callback"); err == nil {
		t.Fatal("oauth state was reusable")
	}
}

func TestOAuthProviderRejectsNonHTTPS(t *testing.T) {
	provider := OAuthProvider{Name: "bad", AuthorizeURL: "http://example.test/authorize", TokenURL: "https://example.test/token", ClientIDEnv: "DZ23_TEST_CLIENT", SecretEnv: "DZ23_TEST_SECRET"}
	t.Setenv("DZ23_TEST_CLIENT", "client")
	t.Setenv("DZ23_TEST_SECRET", "secret")
	if err := provider.Validate(); err == nil {
		t.Fatal("non-HTTPS OAuth endpoint accepted")
	}
}

func TestOAuthProviderRedirectAllowlist(t *testing.T) {
	t.Setenv("DZ23_TEST_CLIENT", "client")
	t.Setenv("DZ23_TEST_SECRET", "secret")
	provider := OAuthProvider{Name: "test", AuthorizeURL: "https://idp.example.test/authorize", TokenURL: "https://idp.example.test/token", ClientIDEnv: "DZ23_TEST_CLIENT", SecretEnv: "DZ23_TEST_SECRET", RedirectURIs: []string{"https://app.example.test/oauth/callback"}}
	if canonical, err := provider.NormalizeRedirectURI("https://app.example.test/oauth/callback"); err != nil || canonical != "https://app.example.test/oauth/callback" {
		t.Fatalf("valid redirect canonical=%q err=%v", canonical, err)
	}
	for _, redirect := range []string{"https://evil.example.test/oauth/callback", "http://app.example.test/oauth/callback", "https://app.example.test/oauth/callback#fragment", "https://user:pass@app.example.test/oauth/callback"} {
		if err := provider.ValidateRedirectURI(redirect); err == nil {
			t.Fatalf("redirect accepted: %s", redirect)
		}
	}
	provider.AllowLoopbackRedirect = true
	provider.RedirectURIs = []string{"http://127.0.0.1:43123/callback"}
	if err := provider.ValidateRedirectURI("http://127.0.0.1:43123/callback"); err != nil {
		t.Fatalf("explicit loopback redirect rejected: %v", err)
	}
}

func TestOAuthStateRejectsInsecureRedirect(t *testing.T) {
	store, err := NewAuthStore("")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.CreateOAuthState("test", "http://127.0.0.1/callback", strings.Repeat("a", 43), "", time.Minute); err == nil {
		t.Fatal("insecure OAuth state redirect accepted")
	}
}

func TestOAuthClientBlocksRedirectAndPrivateActualAddress(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	request, err := oauthRequest(context.Background(), http.MethodGet, server.URL+"/start", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := oauthClient(server.Client()).Do(request)
	if response != nil {
		_ = response.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "redirect") {
		t.Fatalf("expected OAuth redirect rejection, got %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	accepted := make(chan struct{})
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()
	_, err = oauthDialContext(context.Background(), "tcp", listener.Addr().String())
	<-accepted
	if err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("expected private OAuth dial rejection, got %v", err)
	}
}

func TestOAuthCredentialIsEncryptedAtRest(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "local-development-key-long-enough")
	store, err := NewAuthStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("oauth@example.com", "OAuth User")
	if err != nil {
		t.Fatal(err)
	}
	organization, _, err := store.CreateOrganization("OAuth Org", user)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := store.StoreOAuthCredential("github", user.ID, organization.ID, map[string]any{"access_token": "access-fixture", "refresh_token": "refresh-fixture", "expires_in": float64(3600)})
	if err != nil {
		t.Fatal(err)
	}
	if credential.AccessTokenCiphertext == "access-fixture" || credential.RefreshTokenCiphertext == "refresh-fixture" {
		t.Fatal("oauth secret stored in plaintext")
	}
	if _, err := decryptCredential(credential.RefreshTokenCiphertext); err != nil {
		t.Fatal(err)
	}
}

func TestAuthStoreMFAUsesTOTPAndPersistsEncryptedSecret(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "mfa-test-key-long-enough")
	store, err := NewAuthStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("mfa@example.com", "MFA User")
	if err != nil {
		t.Fatal(err)
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("0123456789012345"))
	if _, err := store.EnableMFA(user.ID, secret); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	if err := store.VerifyMFA(user.ID, totpCode(secret, now.Unix()/30), now); err != nil {
		t.Fatal(err)
	}
	if err := store.VerifyMFA(user.ID, "000000", now); err == nil {
		t.Fatal("invalid MFA code accepted")
	}
}

func TestAuthStoreMFAThrottlePersistsAndUnlocks(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "mfa-throttle-test-key-long-enough")
	root := t.TempDir()
	store, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("mfa-throttle@example.com", "MFA Throttle")
	if err != nil {
		t.Fatal(err)
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("0123456789012345"))
	if _, err := store.EnableMFA(user.ID, secret); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	for attempt := 1; attempt < mfaFailureLimit; attempt++ {
		if err := store.VerifyMFAWithThrottle(user.ID, "000000", "127.0.0.1:1234", now); err == nil {
			t.Fatalf("invalid MFA accepted on attempt %d", attempt)
		} else {
			var throttle *MFAThrottleError
			if errors.As(err, &throttle) {
				t.Fatalf("MFA throttled too early on attempt %d", attempt)
			}
		}
	}
	err = store.VerifyMFAWithThrottle(user.ID, "000000", "127.0.0.1:1234", now)
	var throttle *MFAThrottleError
	if !errors.As(err, &throttle) || throttle.RetryAfter <= 0 {
		t.Fatalf("fifth invalid MFA error=%v, throttle=%+v", err, throttle)
	}
	reloaded, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	err = reloaded.VerifyMFAWithThrottle(user.ID, "000000", "127.0.0.1:1234", now.Add(time.Minute))
	if !errors.As(err, &throttle) {
		t.Fatalf("persisted lockout error=%v", err)
	}
	unlockTime := now.Add(mfaLockoutPeriod + time.Minute)
	valid := totpCode(secret, unlockTime.Unix()/30)
	if err := reloaded.VerifyMFAWithThrottle(user.ID, valid, "127.0.0.1:1234", unlockTime); err != nil {
		t.Fatalf("valid MFA did not unlock after lockout: %v", err)
	}
}

func TestRecoveryCodesAreOneTimeAndEncrypted(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "recovery-test-key-long-enough")
	root := t.TempDir()
	store, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	user, err := store.CreateUser("recovery@example.com", "Recovery User")
	if err != nil {
		t.Fatal(err)
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte("0123456789012345"))
	if _, err := store.EnableMFA(user.ID, secret); err != nil {
		t.Fatal(err)
	}
	_, codes, err := store.GenerateRecoveryCodes(user.ID)
	if err != nil || len(codes) != 10 {
		t.Fatalf("generate recovery codes = %v, count=%d", err, len(codes))
	}
	if err := store.VerifyRecoveryCode(user.ID, codes[0]); err != nil {
		t.Fatal(err)
	}
	if err := store.VerifyRecoveryCode(user.ID, codes[0]); err == nil {
		t.Fatal("recovery code was reusable")
	}
	reloaded, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := reloaded.VerifyRecoveryCode(user.ID, codes[1]); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(root + "/users.json")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "" || string(data) == codes[0] {
		t.Fatal("recovery data was not persisted as a structured record")
	}
}

func TestOAuthRefreshRotatesTokensAndRevokeIsTenantScoped(t *testing.T) {
	t.Setenv("OLLAMA_AGENT_CREDENTIAL_KEY", "oauth-lifecycle-test-key-long-enough")
	t.Setenv("DZ23_OAUTH_CLIENT", "client-fixture")
	t.Setenv("DZ23_OAUTH_SECRET", "secret-fixture")
	var revocationToken string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"rotated-access","refresh_token":"rotated-refresh","expires_in":3600}`))
		case "/revoke":
			body, _ := io.ReadAll(r.Body)
			revocationToken = string(body)
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	provider := OAuthProvider{Name: "github", AuthorizeURL: "https://idp.example.test/authorize", TokenURL: server.URL + "/token", RevocationURL: server.URL + "/revoke", ClientIDEnv: "DZ23_OAUTH_CLIENT", SecretEnv: "DZ23_OAUTH_SECRET"}
	root := t.TempDir()
	store, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := store.CreateUser("oauth-owner@example.com", "OAuth Owner")
	if err != nil {
		t.Fatal(err)
	}
	orgA, _, err := store.CreateOrganization("OAuth A", owner)
	if err != nil {
		t.Fatal(err)
	}
	other, err := store.CreateUser("oauth-other@example.com", "OAuth Other")
	if err != nil {
		t.Fatal(err)
	}
	orgB, _, err := store.CreateOrganization("OAuth B", other)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := store.StoreOAuthCredential(provider.Name, owner.ID, orgA.ID, map[string]any{"access_token": "initial-access", "refresh_token": "initial-refresh", "expires_in": float64(60)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RefreshOAuthCredentialForOrganization(context.Background(), orgB.ID, provider, credential.ID, server.Client()); err == nil || !strings.Contains(err.Error(), "outside the active organization") {
		t.Fatalf("cross-tenant refresh error = %v", err)
	}
	rotated, err := store.RefreshOAuthCredentialForOrganization(context.Background(), orgA.ID, provider, credential.ID, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RevokedAt != nil || rotated.ID != credential.ID {
		t.Fatalf("rotated credential = %+v", rotated)
	}
	access, _, err := store.OAuthAccessTokenForOrganization(orgA.ID, provider.Name)
	if err != nil || access != "rotated-access" {
		t.Fatalf("rotated access=%q err=%v", access, err)
	}
	if err := store.RevokeOAuthCredentialForOrganization(context.Background(), orgB.ID, credential.ID, provider, server.Client()); err == nil || !strings.Contains(err.Error(), "outside the active organization") {
		t.Fatalf("cross-tenant revoke error = %v", err)
	}
	if err := store.RevokeOAuthCredentialForOrganization(context.Background(), orgA.ID, credential.ID, provider, server.Client()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(revocationToken, "token=rotated-access") {
		t.Fatalf("revocation request did not carry rotated access token: %q", revocationToken)
	}
	if _, _, err := store.OAuthAccessTokenForOrganization(orgA.ID, provider.Name); err == nil {
		t.Fatal("revoked credential remained usable")
	}
	reloaded, err := NewAuthStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := reloaded.OAuthAccessTokenForOrganization(orgA.ID, provider.Name); err == nil {
		t.Fatal("revocation was not persisted")
	}
}
