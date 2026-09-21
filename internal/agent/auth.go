package agent

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Role string

const (
	RoleOwner    Role = "owner"
	RoleAdmin    Role = "admin"
	RoleOperator Role = "operator"
	RoleViewer   Role = "viewer"
	RoleAuditor  Role = "auditor"
)

type Membership struct {
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	Role           Role      `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

type AccessToken struct {
	Hash           string     `json:"hash"`
	UserID         string     `json:"user_id"`
	OrganizationID string     `json:"organization_id"`
	ExpiresAt      time.Time  `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
}

type OAuthState struct {
	Hash         string     `json:"hash"`
	Provider     string     `json:"provider"`
	RedirectURI  string     `json:"redirect_uri"`
	CodeVerifier string     `json:"code_verifier"`
	UserID       string     `json:"user_id,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	ConsumedAt   *time.Time `json:"consumed_at,omitempty"`
}

type OAuthCredential struct {
	ID                     string    `json:"id"`
	UserID                 string    `json:"user_id"`
	OrganizationID         string    `json:"organization_id"`
	Provider               string    `json:"provider"`
	AccessTokenCiphertext  string    `json:"access_token_ciphertext"`
	RefreshTokenCiphertext string    `json:"refresh_token_ciphertext,omitempty"`
	ExpiresAt              time.Time `json:"expires_at,omitempty"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type AuthStore struct {
	mu            sync.RWMutex
	root          string
	users         map[string]User
	organizations map[string]Organization
	memberships   map[string]Membership
	tokens        map[string]AccessToken
	oauthStates   map[string]OAuthState
	credentials   map[string]OAuthCredential
}

func NewAuthStore(root string) (*AuthStore, error) {
	store := &AuthStore{root: strings.TrimSpace(root), users: map[string]User{}, organizations: map[string]Organization{}, memberships: map[string]Membership{}, tokens: map[string]AccessToken{}, oauthStates: map[string]OAuthState{}, credentials: map[string]OAuthCredential{}}
	if store.root == "" {
		return store, nil
	}
	if err := os.MkdirAll(store.root, 0o700); err != nil {
		return nil, err
	}
	for name, target := range map[string]any{"users": &store.users, "organizations": &store.organizations, "memberships": &store.memberships, "tokens": &store.tokens, "oauth-states": &store.oauthStates, "credentials": &store.credentials} {
		path := filepathJoin(store.root, name+".json")
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err := readJSON(path, target); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *AuthStore) CreateUser(email, name string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)
	if email == "" || !strings.Contains(email, "@") {
		return User{}, errors.New("valid email is required")
	}
	if len(email) > 320 || len(name) > 200 {
		return User{}, errors.New("user field is too long")
	}
	now := time.Now().UTC()
	user := User{ID: "usr_" + uuid.NewString(), Email: email, Name: name, CreatedAt: now}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.users {
		if existing.Email == email {
			return existing, nil
		}
	}
	s.users[user.ID] = user
	return user, s.persistLocked()
}

func (s *AuthStore) CreateOrganization(name string, owner User) (Organization, Membership, error) {
	name = strings.TrimSpace(name)
	if name == "" || owner.ID == "" {
		return Organization{}, Membership{}, errors.New("organization name and owner are required")
	}
	now := time.Now().UTC()
	organization := Organization{ID: "org_" + uuid.NewString(), Name: name, CreatedAt: now}
	membership := Membership{UserID: owner.ID, OrganizationID: organization.ID, Role: RoleOwner, CreatedAt: now}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.organizations[organization.ID] = organization
	s.memberships[membershipKey(owner.ID, organization.ID)] = membership
	return organization, membership, s.persistLocked()
}

func (s *AuthStore) AddMembership(userID, organizationID string, role Role) (Membership, error) {
	if !validRole(role) {
		return Membership{}, errors.New("invalid organization role")
	}
	membership := Membership{UserID: strings.TrimSpace(userID), OrganizationID: strings.TrimSpace(organizationID), Role: role, CreatedAt: time.Now().UTC()}
	if membership.UserID == "" || membership.OrganizationID == "" {
		return Membership{}, errors.New("user and organization are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[membership.UserID]; !ok {
		return Membership{}, os.ErrNotExist
	}
	if _, ok := s.organizations[membership.OrganizationID]; !ok {
		return Membership{}, os.ErrNotExist
	}
	s.memberships[membershipKey(membership.UserID, membership.OrganizationID)] = membership
	return membership, s.persistLocked()
}

func (s *AuthStore) IssueToken(userID, organizationID string, ttl time.Duration) (string, AccessToken, error) {
	if ttl <= 0 || ttl > 30*24*time.Hour {
		ttl = 24 * time.Hour
	}
	if _, err := s.Authorize(userID, organizationID, "read"); err != nil {
		return "", AccessToken{}, err
	}
	raw, err := randomSecret(32)
	if err != nil {
		return "", AccessToken{}, err
	}
	now := time.Now().UTC()
	token := AccessToken{Hash: hashSecret(raw), UserID: userID, OrganizationID: organizationID, CreatedAt: now, ExpiresAt: now.Add(ttl)}
	s.mu.Lock()
	s.tokens[token.Hash] = token
	err = s.persistLocked()
	s.mu.Unlock()
	return raw, token, err
}

func (s *AuthStore) Authenticate(raw string) (User, Organization, Membership, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return User{}, Organization{}, Membership{}, errors.New("token is required")
	}
	s.mu.RLock()
	token, ok := s.tokens[hashSecret(raw)]
	user := s.users[token.UserID]
	organization := s.organizations[token.OrganizationID]
	membership := s.memberships[membershipKey(token.UserID, token.OrganizationID)]
	s.mu.RUnlock()
	if !ok || token.RevokedAt != nil || time.Now().UTC().After(token.ExpiresAt) || user.ID == "" || organization.ID == "" || membership.UserID == "" {
		return User{}, Organization{}, Membership{}, errors.New("invalid or expired token")
	}
	return user, organization, membership, nil
}

func (s *AuthStore) RevokeToken(raw string) error {
	hash := hashSecret(strings.TrimSpace(raw))
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.tokens[hash]
	if !ok {
		return os.ErrNotExist
	}
	token.RevokedAt = &now
	s.tokens[hash] = token
	return s.persistLocked()
}

func (s *AuthStore) Authorize(userID, organizationID, action string) (Membership, error) {
	s.mu.RLock()
	membership, ok := s.memberships[membershipKey(userID, organizationID)]
	s.mu.RUnlock()
	if !ok {
		return Membership{}, errors.New("user is not a member of organization")
	}
	if !roleAllows(membership.Role, action) {
		return Membership{}, fmt.Errorf("role %s cannot perform %s", membership.Role, action)
	}
	return membership, nil
}

func (s *AuthStore) CreateOAuthState(provider, redirectURI, codeVerifier, userID string, ttl time.Duration) (string, OAuthState, error) {
	provider = strings.TrimSpace(provider)
	parsed, err := url.Parse(strings.TrimSpace(redirectURI))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil {
		return "", OAuthState{}, errors.New("redirect_uri must be an absolute URL")
	}
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return "", OAuthState{}, errors.New("code_verifier must be PKCE length")
	}
	if ttl <= 0 || ttl > 10*time.Minute {
		ttl = 5 * time.Minute
	}
	raw, err := randomSecret(32)
	if err != nil {
		return "", OAuthState{}, err
	}
	state := OAuthState{Hash: hashSecret(raw), Provider: provider, RedirectURI: redirectURI, CodeVerifier: codeVerifier, UserID: userID, ExpiresAt: time.Now().UTC().Add(ttl)}
	s.mu.Lock()
	s.oauthStates[state.Hash] = state
	err = s.persistLocked()
	s.mu.Unlock()
	return raw, state, err
}

func (s *AuthStore) ConsumeOAuthState(raw, provider, redirectURI string) (OAuthState, error) {
	hash := hashSecret(strings.TrimSpace(raw))
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.oauthStates[hash]
	if !ok || state.Provider != provider || state.RedirectURI != redirectURI || state.ConsumedAt != nil || time.Now().UTC().After(state.ExpiresAt) {
		return OAuthState{}, errors.New("invalid, expired, or already consumed oauth state")
	}
	now := time.Now().UTC()
	state.ConsumedAt = &now
	s.oauthStates[hash] = state
	if err := s.persistLocked(); err != nil {
		return OAuthState{}, err
	}
	return state, nil
}

func (s *AuthStore) StoreOAuthCredential(provider, userID, organizationID string, payload map[string]any) (OAuthCredential, error) {
	if _, err := s.Authorize(userID, organizationID, "write"); err != nil {
		return OAuthCredential{}, err
	}
	access, _ := payload["access_token"].(string)
	refresh, _ := payload["refresh_token"].(string)
	if strings.TrimSpace(access) == "" {
		return OAuthCredential{}, errors.New("oauth response has no access_token")
	}
	accessCipher, err := encryptCredential(access)
	if err != nil {
		return OAuthCredential{}, err
	}
	refreshCipher := ""
	if refresh != "" {
		refreshCipher, err = encryptCredential(refresh)
		if err != nil {
			return OAuthCredential{}, err
		}
	}
	expires := time.Time{}
	if seconds, ok := payload["expires_in"].(float64); ok && seconds > 0 {
		expires = time.Now().UTC().Add(time.Duration(seconds) * time.Second)
	}
	credential := OAuthCredential{ID: "cred_" + uuid.NewString(), UserID: userID, OrganizationID: organizationID, Provider: provider, AccessTokenCiphertext: accessCipher, RefreshTokenCiphertext: refreshCipher, ExpiresAt: expires, UpdatedAt: time.Now().UTC()}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.credentials[credential.ID] = credential
	return credential, s.persistLocked()
}

func (s *AuthStore) RefreshOAuthCredential(ctx context.Context, provider OAuthProvider, credentialID string, client *http.Client) (OAuthCredential, error) {
	s.mu.RLock()
	credential, ok := s.credentials[credentialID]
	s.mu.RUnlock()
	if !ok {
		return OAuthCredential{}, os.ErrNotExist
	}
	refresh, err := decryptCredential(credential.RefreshTokenCiphertext)
	if err != nil {
		return OAuthCredential{}, err
	}
	if err := provider.Validate(); err != nil {
		return OAuthCredential{}, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	form := url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refresh}, "client_id": {os.Getenv(provider.ClientIDEnv)}, "client_secret": {os.Getenv(provider.SecretEnv)}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthCredential{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return OAuthCredential{}, err
	}
	defer response.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return OAuthCredential{}, err
	}
	if response.StatusCode/100 != 2 {
		return OAuthCredential{}, fmt.Errorf("oauth refresh failed with status %d", response.StatusCode)
	}
	access, _ := payload["access_token"].(string)
	if access == "" {
		return OAuthCredential{}, errors.New("oauth refresh has no access_token")
	}
	credential.AccessTokenCiphertext, err = encryptCredential(access)
	if err != nil {
		return OAuthCredential{}, err
	}
	if next, _ := payload["refresh_token"].(string); next != "" {
		credential.RefreshTokenCiphertext, err = encryptCredential(next)
		if err != nil {
			return OAuthCredential{}, err
		}
	}
	if seconds, ok := payload["expires_in"].(float64); ok && seconds > 0 {
		credential.ExpiresAt = time.Now().UTC().Add(time.Duration(seconds) * time.Second)
	}
	credential.UpdatedAt = time.Now().UTC()
	s.mu.Lock()
	s.credentials[credential.ID] = credential
	err = s.persistLocked()
	s.mu.Unlock()
	return credential, err
}

func (s *AuthStore) Users() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]User, 0, len(s.users))
	for _, user := range s.users {
		result = append(result, user)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Email < result[j].Email })
	return result
}

func (s *AuthStore) FirstOrganization(userID string) (Organization, Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, membership := range s.memberships {
		if membership.UserID == userID {
			if organization, ok := s.organizations[membership.OrganizationID]; ok {
				return organization, membership, nil
			}
		}
	}
	return Organization{}, Membership{}, os.ErrNotExist
}

func (s *AuthStore) persistLocked() error {
	if s.root == "" {
		return nil
	}
	for name, value := range map[string]any{
		"users":         s.users,
		"organizations": s.organizations,
		"memberships":   s.memberships,
		"tokens":        s.tokens,
		"oauth-states":  s.oauthStates,
		"credentials":   s.credentials,
	} {
		if err := writeJSONAtomic(filepathJoin(s.root, name+".json"), value); err != nil {
			return err
		}
	}
	return nil
}

func validRole(role Role) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleOperator, RoleViewer, RoleAuditor:
		return true
	default:
		return false
	}
}

func roleAllows(role Role, action string) bool {
	switch strings.TrimSpace(action) {
	case "read":
		return role == RoleOwner || role == RoleAdmin || role == RoleOperator || role == RoleViewer || role == RoleAuditor
	case "execute":
		return role == RoleOwner || role == RoleAdmin || role == RoleOperator
	case "write":
		return role == RoleOwner || role == RoleAdmin || role == RoleOperator
	case "share", "manage":
		return role == RoleOwner || role == RoleAdmin
	default:
		return false
	}
}

func membershipKey(userID, organizationID string) string { return userID + ":" + organizationID }
func hashSecret(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func credentialKey() ([]byte, error) {
	value := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_CREDENTIAL_KEY"))
	if len(value) < 16 {
		return nil, errors.New("OLLAMA_AGENT_CREDENTIAL_KEY must be configured with at least 16 characters")
	}
	sum := sha256.Sum256([]byte(value))
	return sum[:], nil
}

func encryptCredential(value string) (string, error) {
	key, err := credentialKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nonce, nonce, []byte(value), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func decryptCredential(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("oauth credential has no refresh token")
	}
	key, err := credentialKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	if len(sealed) < aead.NonceSize() {
		return "", errors.New("invalid oauth ciphertext")
	}
	plaintext, err := aead.Open(nil, sealed[:aead.NonceSize()], sealed[aead.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
func randomSecret(size int) (string, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
func filepathJoin(root, name string) string {
	return strings.TrimRight(root, string(os.PathSeparator)) + string(os.PathSeparator) + name
}

// OAuthProvider validates the provider configuration and exchanges an authorization code
// through a caller-supplied HTTP client. Secrets are read only from environment variables.
type OAuthProvider struct {
	Name         string `json:"name"`
	AuthorizeURL string `json:"authorize_url"`
	TokenURL     string `json:"token_url"`
	ClientIDEnv  string `json:"client_id_env"`
	SecretEnv    string `json:"secret_env"`
}

func (p OAuthProvider) Validate() error {
	for _, raw := range []string{p.AuthorizeURL, p.TokenURL} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return errors.New("oauth endpoints must use https")
		}
	}
	if os.Getenv(p.ClientIDEnv) == "" || os.Getenv(p.SecretEnv) == "" {
		return errors.New("oauth client credentials are not configured")
	}
	return nil
}

func (p OAuthProvider) AuthorizationURL(state string, scopes []string) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	u, err := url.Parse(p.AuthorizeURL)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("client_id", os.Getenv(p.ClientIDEnv))
	query.Set("response_type", "code")
	query.Set("state", state)
	query.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (p OAuthProvider) ExchangeCode(ctx context.Context, client *http.Client, code, redirectURI, codeVerifier string) (map[string]any, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirectURI}, "client_id": {os.Getenv(p.ClientIDEnv)}, "client_secret": {os.Getenv(p.SecretEnv)}, "code_verifier": {codeVerifier}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var payload map[string]any
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("oauth token exchange failed with status %d", response.StatusCode)
	}
	return payload, nil
}
