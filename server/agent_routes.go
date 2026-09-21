package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
	"github.com/ollama/ollama/internal/agent"
)

type agentAPI struct {
	runtime      *agent.Runtime
	context      *agent.ContextStore
	auth         *agent.AuthStore
	authRequired bool
}

func newAgentAPI(runtime *agent.Runtime) (*agentAPI, error) {
	storeRoot := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_AUTH_STORE"))
	auth, err := agent.NewAuthStore(storeRoot)
	if err != nil {
		return nil, err
	}
	required, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("OLLAMA_AGENT_AUTH_REQUIRED")))
	return &agentAPI{runtime: runtime, context: runtime.Context(), auth: auth, authRequired: required}, nil
}

func newDefaultAgentRuntime() (*agent.Runtime, error) {
	workspaceRoot := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_ROOT"))
	if workspaceRoot == "" {
		workspaceRoot = filepath.Join(os.TempDir(), "ollama-agent-workspace")
	}
	storeRoot := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_STORE"))
	if storeRoot == "" {
		storeRoot = filepath.Join(filepath.Dir(workspaceRoot), ".ollama-agent-store")
	}
	store, err := agent.NewJSONStore(storeRoot)
	if err != nil {
		return nil, err
	}
	contextStore, err := agent.NewContextStore(filepath.Join(storeRoot, "context"))
	if err != nil {
		return nil, err
	}
	if embedModel := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_EMBED_MODEL")); embedModel != "" {
		contextStore.SetEmbedder(agent.OllamaEmbedder{Client: api.NewClient(envconfig.ConnectableHost(), http.DefaultClient), Model: embedModel})
	}
	connectors, err := loadAgentConnectors()
	if err != nil {
		return nil, err
	}
	mcp, err := loadAgentMCP()
	if err != nil {
		return nil, err
	}
	media, err := loadAgentMedia()
	if err != nil {
		return nil, err
	}
	var planner agent.Planner = agent.RulePlanner{}
	if model := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_MODEL")); model != "" {
		planner = agent.OllamaPlanner{
			Client:   api.NewClient(envconfig.ConnectableHost(), http.DefaultClient),
			Model:    model,
			Fallback: agent.RulePlanner{},
		}
	}
	return agent.NewRuntime(agent.RuntimeConfig{Store: store, Context: contextStore, Planner: planner, WorkspaceRoot: workspaceRoot, Connectors: connectors, MCP: mcp, Media: media})
}

func loadAgentConnectors() (*agent.ConnectorManager, error) {
	configPath := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_CONNECTORS"))
	if configPath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var configs []agent.ConnectorConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	manager := agent.NewConnectorManager()
	for _, config := range configs {
		if err := manager.Register(config); err != nil {
			return nil, err
		}
	}
	return manager, nil
}

func loadAgentMedia() (*agent.MediaManager, error) {
	baseURL := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_MEDIA_BASE_URL"))
	if baseURL == "" {
		return nil, nil
	}
	return agent.NewMediaManager(agent.MediaProvider{
		Name:               "configured",
		BaseURL:            baseURL,
		APIKey:             os.Getenv("OLLAMA_AGENT_MEDIA_API_KEY"),
		ImageModel:         os.Getenv("OLLAMA_AGENT_MEDIA_IMAGE_MODEL"),
		VideoModel:         os.Getenv("OLLAMA_AGENT_MEDIA_VIDEO_MODEL"),
		SpeechModel:        os.Getenv("OLLAMA_AGENT_MEDIA_SPEECH_MODEL"),
		TranscriptionModel: os.Getenv("OLLAMA_AGENT_MEDIA_TRANSCRIPTION_MODEL"),
	})
}

func loadAgentMCP() (*agent.MCPManager, error) {
	configPath := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_MCP"))
	if configPath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var configs []agent.MCPServerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}
	manager := agent.NewMCPManager()
	for _, config := range configs {
		if err := manager.Register(config); err != nil {
			return nil, err
		}
	}
	return manager, nil
}

func (a *agentAPI) register(r *gin.Engine) {
	group := r.Group("/api/agent/v1")
	group.Use(a.authMiddleware)
	group.GET("/health", a.health)
	group.GET("/auth/session", a.authSession)
	group.POST("/auth/dev/token", a.devToken)
	group.GET("/auth/oauth/:provider/start", a.oauthStart)
	group.GET("/auth/oauth/:provider/callback", a.oauthCallback)
	group.GET("/traces", a.allTraces)
	group.POST("/media/image", a.mediaImage)
	group.POST("/media/video", a.mediaVideo)
	group.POST("/media/speech", a.mediaSpeech)
	group.POST("/media/transcribe", a.mediaTranscribe)
	group.POST("/media/tone", a.mediaTone)
	group.GET("/builders", a.builders)
	group.POST("/builders", a.createBuilder)
	group.POST("/builders/:id/preview", a.previewBuilder)
	group.POST("/builders/:id/export", a.exportBuilder)
	group.POST("/builders/:id/publish", a.publishBuilder)
	group.GET("/builders/:id/preview/*path", a.builderPreviewFile)
	group.GET("/metrics/prometheus", a.prometheus)
	group.GET("/tools", a.tools)
	group.GET("/connectors", a.connectors)
	group.GET("/mcp", a.mcp)
	group.GET("/jobs", a.jobs)
	group.POST("/jobs/:id/replay", a.replayJob)
	group.GET("/skills", a.skills)
	group.GET("/schedules", a.schedules)
	group.POST("/schedules", a.createSchedule)
	group.POST("/webhooks/:schedule_id", a.webhook)
	group.POST("/projects", a.createProject)
	group.GET("/projects/:id", a.getProject)
	group.POST("/projects/:id/memories", a.addMemory)
	group.GET("/projects/:id/memories", a.searchMemories)
	group.GET("/collab/:project_id", a.collabSnapshot)
	group.GET("/collab/:project_id/stream", a.collabStream)
	group.POST("/collab/:project_id/comments", a.collabComment)
	group.POST("/collab/:project_id/presence", a.collabPresence)
	group.POST("/missions", a.createMission)
	group.GET("/missions/:id", a.getMission)
	group.GET("/missions/:id/events", a.events)
	group.GET("/missions/:id/events/stream", a.eventStream)
	group.GET("/missions/:id/traces", a.traces)
	group.GET("/missions/:id/artifacts/:artifact_id", a.artifact)
	group.POST("/missions/:id/run", a.runMission)
	group.POST("/missions/:id/cancel", a.cancelMission)
	group.POST("/missions/:id/approvals/:approval_id", a.decideApproval)
}

func (a *agentAPI) authMiddleware(c *gin.Context) {
	if !a.authRequired {
		c.Next()
		return
	}
	if strings.HasSuffix(c.Request.URL.Path, "/auth/dev/token") && strings.EqualFold(strings.TrimSpace(os.Getenv("OLLAMA_AGENT_AUTH_DEV")), "true") {
		c.Next()
		return
	}
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "bearer token is required"})
		return
	}
	user, organization, membership, err := a.auth.Authenticate(strings.TrimSpace(header[len("Bearer "):]))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}
	if requested := strings.TrimSpace(c.GetHeader("X-Ollama-Organization")); requested != "" && requested != organization.ID {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "organization scope mismatch"})
		return
	}
	action := "read"
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		action = "execute"
	}
	if _, err := a.auth.Authorize(user.ID, organization.ID, action); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.Set("agent.user", user)
	c.Set("agent.organization", organization)
	c.Set("agent.membership", membership)
	c.Next()
}

func (a *agentAPI) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "runtime": "agent-v1"})
}

func (a *agentAPI) authSession(c *gin.Context) {
	if !a.authRequired {
		c.JSON(http.StatusOK, gin.H{"authenticated": false, "mode": "local"})
		return
	}
	user, _ := c.Get("agent.user")
	organization, _ := c.Get("agent.organization")
	membership, _ := c.Get("agent.membership")
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user": user, "organization": organization, "membership": membership})
}

func (a *agentAPI) devToken(c *gin.Context) {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("OLLAMA_AGENT_AUTH_DEV")), "true") {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var input struct {
		Email        string `json:"email"`
		Name         string `json:"name"`
		Organization string `json:"organization"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	user, err := a.auth.CreateUser(input.Email, input.Name)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	organization, _, err := a.auth.CreateOrganization(input.Organization, user)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	raw, token, err := a.auth.IssueToken(user.ID, organization.ID, 24*time.Hour)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"access_token": raw, "token": token, "user": user, "organization": organization})
}

func oauthProviderFromEnv(name string) agent.OAuthProvider {
	key := strings.ToUpper(strings.NewReplacer("-", "_", " ", "_").Replace(strings.TrimSpace(name)))
	return agent.OAuthProvider{Name: name, AuthorizeURL: os.Getenv("OLLAMA_AGENT_OAUTH_" + key + "_AUTHORIZE_URL"), TokenURL: os.Getenv("OLLAMA_AGENT_OAUTH_" + key + "_TOKEN_URL"), ClientIDEnv: os.Getenv("OLLAMA_AGENT_OAUTH_" + key + "_CLIENT_ID_ENV"), SecretEnv: os.Getenv("OLLAMA_AGENT_OAUTH_" + key + "_CLIENT_SECRET_ENV")}
}

func (a *agentAPI) oauthStart(c *gin.Context) {
	provider := oauthProviderFromEnv(c.Param("provider"))
	if err := provider.Validate(); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	redirectURI := strings.TrimSpace(c.Query("redirect_uri"))
	verifier := strings.TrimSpace(c.Query("code_verifier"))
	if redirectURI == "" || verifier == "" {
		writeAgentError(c, http.StatusBadRequest, errors.New("redirect_uri and PKCE code_verifier are required"))
		return
	}
	userID := ""
	if value, ok := c.Get("agent.user"); ok {
		if user, ok := value.(agent.User); ok {
			userID = user.ID
		}
	}
	state, _, err := a.auth.CreateOAuthState(provider.Name, redirectURI, verifier, userID, 5*time.Minute)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	scopes := strings.Fields(c.Query("scope"))
	if len(scopes) == 0 {
		scopes = []string{"openid", "email"}
	}
	authorizationURL, err := provider.AuthorizationURL(state, scopes)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"provider": provider.Name, "authorization_url": authorizationURL, "state": state, "expires_in": 300})
}

func (a *agentAPI) oauthCallback(c *gin.Context) {
	provider := oauthProviderFromEnv(c.Param("provider"))
	redirectURI := strings.TrimSpace(c.Query("redirect_uri"))
	code := strings.TrimSpace(c.Query("code"))
	stateValue := strings.TrimSpace(c.Query("state"))
	if code == "" || stateValue == "" || redirectURI == "" {
		writeAgentError(c, http.StatusBadRequest, errors.New("code, state and redirect_uri are required"))
		return
	}
	state, err := a.auth.ConsumeOAuthState(stateValue, provider.Name, redirectURI)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	payload, err := provider.ExchangeCode(c.Request.Context(), http.DefaultClient, code, redirectURI, state.CodeVerifier)
	if err != nil {
		writeAgentError(c, http.StatusBadGateway, err)
		return
	}
	organization, _, err := a.auth.FirstOrganization(state.UserID)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, errors.New("OAuth user has no organization"))
		return
	}
	credential, err := a.auth.StoreOAuthCredential(provider.Name, state.UserID, organization.ID, payload)
	if err != nil {
		writeAgentError(c, http.StatusInternalServerError, err)
		return
	}
	localToken, session, err := a.auth.IssueToken(state.UserID, organization.ID, 24*time.Hour)
	if err != nil {
		writeAgentError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": localToken, "token": session, "credential_id": credential.ID, "provider": provider.Name, "organization": organization, "expires_at": credential.ExpiresAt})
}

func (a *agentAPI) metrics(c *gin.Context) {
	c.JSON(http.StatusOK, a.runtime.Metrics())
}

func (a *agentAPI) prometheus(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; version=0.0.4", []byte(a.runtime.Metrics().Prometheus()))
}

func (a *agentAPI) connectors(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"connectors": a.runtime.Connectors()})
}

func (a *agentAPI) mcp(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"servers": a.runtime.MCPServers()})
}

func (a *agentAPI) jobs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"jobs": a.runtime.QueueJobs(agent.QueueStatus(c.Query("status")))})
}

func (a *agentAPI) replayJob(c *gin.Context) {
	job, err := a.runtime.ReplayJob(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusAccepted, job)
}

func (a *agentAPI) tools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"tools": a.runtime.ListTools()})
}

func (a *agentAPI) skills(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"skills": a.context.Skills()})
}

func (a *agentAPI) schedules(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"schedules": a.context.ListSchedules()})
}

func (a *agentAPI) createSchedule(c *gin.Context) {
	var schedule agent.Schedule
	if err := decodeJSON(c, &schedule); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	created, err := a.context.CreateSchedule(schedule)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (a *agentAPI) webhook(c *gin.Context) {
	schedule, err := a.context.GetSchedule(c.Param("schedule_id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	secret := strings.TrimSpace(schedule.WebhookSecretEnv)
	if secret == "" || !verifyAgentWebhook(os.Getenv(secret), c.GetHeader("X-Ollama-Agent-Secret")) {
		writeAgentError(c, http.StatusUnauthorized, errors.New("webhook secret is invalid"))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	objective := schedule.Objective + "\nWebhook payload:\n" + string(payload)
	mission, err := a.runtime.CreateMission(c.Request.Context(), agent.CreateMissionRequest{Objective: objective, Model: schedule.Model, Workspace: schedule.Workspace, ProjectID: schedule.ProjectID, AutoRun: true})
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusAccepted, mission)
}

func verifyAgentWebhook(expected, provided string) bool {
	if expected == "" || provided == "" {
		return false
	}
	expectedSum := sha256.Sum256([]byte(expected))
	providedSum := sha256.Sum256([]byte(provided))
	return hmac.Equal(expectedSum[:], providedSum[:])
}

func (a *agentAPI) createProject(c *gin.Context) {
	var request struct {
		Name string `json:"name"`
		Root string `json:"root,omitempty"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	project, err := a.context.CreateProject(request.Name, request.Root)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (a *agentAPI) getProject(c *gin.Context) {
	project, err := a.context.GetProject(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func (a *agentAPI) collabActor(c *gin.Context) string {
	if user, ok := c.Get("agent.user"); ok {
		if actor, ok := user.(agent.User); ok && actor.ID != "" {
			return actor.ID
		}
	}
	if value := strings.TrimSpace(c.GetHeader("X-Ollama-User")); value != "" {
		return value
	}
	return "local"
}

func (a *agentAPI) collabSnapshot(c *gin.Context) {
	c.JSON(http.StatusOK, a.runtime.Collaboration().Snapshot(c.Param("project_id")))
}

func (a *agentAPI) collabComment(c *gin.Context) {
	var request struct {
		Body string `json:"body"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	comment, err := a.runtime.Collaboration().AddComment(c.Param("project_id"), a.collabActor(c), request.Body)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (a *agentAPI) collabPresence(c *gin.Context) {
	var request struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	presence, err := a.runtime.Collaboration().SetPresence(c.Param("project_id"), a.collabActor(c), request.Status)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, presence)
}

func (a *agentAPI) collabStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusNotImplemented)
		return
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		payload, _ := json.Marshal(a.runtime.Collaboration().Snapshot(c.Param("project_id")))
		_, _ = fmt.Fprintf(c.Writer, "event: collaboration\ndata: %s\n\n", payload)
		flusher.Flush()
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *agentAPI) addMemory(c *gin.Context) {
	var memory agent.Memory
	if err := decodeJSON(c, &memory); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	memory.ProjectID = c.Param("id")
	created, err := a.context.AddMemoryContext(c.Request.Context(), memory)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (a *agentAPI) searchMemories(c *gin.Context) {
	memories, err := a.context.SearchMemoriesContext(c.Request.Context(), c.Param("id"), c.Query("q"), 20)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project_id": c.Param("id"), "memories": memories})
}

func (a *agentAPI) createMission(c *gin.Context) {
	var request agent.CreateMissionRequest
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	mission, err := a.runtime.CreateMission(c.Request.Context(), request)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, mission)
}

func (a *agentAPI) getMission(c *gin.Context) {
	mission, err := a.runtime.GetMission(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, mission)
}

func (a *agentAPI) events(c *gin.Context) {
	events, err := a.runtime.Events(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"mission_id": c.Param("id"), "events": events})
}

func (a *agentAPI) eventStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.Status(http.StatusNotImplemented)
		return
	}
	lastCount := 0
	ticker := time.NewTicker(750 * time.Millisecond)
	defer ticker.Stop()
	for {
		events, err := a.runtime.Events(c.Param("id"))
		if err != nil {
			return
		}
		if len(events) > lastCount {
			payload, _ := json.Marshal(events[lastCount:])
			_, _ = fmt.Fprintf(c.Writer, "event: mission\ndata: %s\n\n", payload)
			flusher.Flush()
			lastCount = len(events)
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func (a *agentAPI) traces(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mission_id": c.Param("id"), "spans": a.runtime.Traces("tr_" + c.Param("id"))})
}

func (a *agentAPI) allTraces(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"spans": a.runtime.Traces(c.Query("trace_id"))})
}

func (a *agentAPI) mediaWorkspace(c *gin.Context, missionID string) (agent.Mission, string, error) {
	mission, err := a.runtime.GetMission(missionID)
	if err != nil {
		return agent.Mission{}, "", err
	}
	if strings.TrimSpace(mission.Workspace) == "" {
		return agent.Mission{}, "", errors.New("mission workspace is required")
	}
	return mission, mission.Workspace, nil
}

func (a *agentAPI) mediaImage(c *gin.Context) {
	var request struct {
		MissionID string `json:"mission_id"`
		Prompt    string `json:"prompt"`
		Model     string `json:"model"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	if a.runtime.Media() == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "media provider is not configured"})
		return
	}
	mission, workspace, err := a.mediaWorkspace(c, request.MissionID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	result, err := a.runtime.Media().GenerateImage(c.Request.Context(), workspace, request.Prompt, request.Model)
	if err != nil {
		writeAgentError(c, http.StatusBadGateway, err)
		return
	}
	result.Artifact.MissionID = mission.ID
	c.JSON(http.StatusCreated, result)
}

func (a *agentAPI) mediaVideo(c *gin.Context) {
	var request struct {
		MissionID string `json:"mission_id"`
		Prompt    string `json:"prompt"`
		Model     string `json:"model"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	if a.runtime.Media() == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "media provider is not configured"})
		return
	}
	mission, workspace, err := a.mediaWorkspace(c, request.MissionID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	result, err := a.runtime.Media().GenerateVideo(c.Request.Context(), workspace, request.Prompt, request.Model)
	if err != nil {
		writeAgentError(c, http.StatusBadGateway, err)
		return
	}
	result.Artifact.MissionID = mission.ID
	c.JSON(http.StatusCreated, result)
}

func (a *agentAPI) mediaSpeech(c *gin.Context) {
	var request struct {
		MissionID string `json:"mission_id"`
		Text      string `json:"text"`
		Voice     string `json:"voice"`
		Model     string `json:"model"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	if a.runtime.Media() == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "media provider is not configured"})
		return
	}
	mission, workspace, err := a.mediaWorkspace(c, request.MissionID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	result, err := a.runtime.Media().GenerateSpeech(c.Request.Context(), workspace, request.Text, request.Voice, request.Model)
	if err != nil {
		writeAgentError(c, http.StatusBadGateway, err)
		return
	}
	result.Artifact.MissionID = mission.ID
	c.JSON(http.StatusCreated, result)
}

func (a *agentAPI) mediaTranscribe(c *gin.Context) {
	var request struct {
		MissionID string `json:"mission_id"`
		InputPath string `json:"input_path"`
		Model     string `json:"model"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	if a.runtime.Media() == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "media provider is not configured"})
		return
	}
	mission, workspace, err := a.mediaWorkspace(c, request.MissionID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	inputPath, err := containedPath(workspace, request.InputPath)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	result, err := a.runtime.Media().Transcribe(c.Request.Context(), workspace, inputPath, request.Model)
	if err != nil {
		writeAgentError(c, http.StatusBadGateway, err)
		return
	}
	result.Artifact.MissionID = mission.ID
	c.JSON(http.StatusCreated, result)
}

func (a *agentAPI) mediaTone(c *gin.Context) {
	var request struct {
		MissionID  string  `json:"mission_id"`
		Frequency  float64 `json:"frequency"`
		DurationMS int     `json:"duration_ms"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	mission, workspace, err := a.mediaWorkspace(c, request.MissionID)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	result, err := agent.GenerateTone(workspace, request.Frequency, time.Duration(request.DurationMS)*time.Millisecond)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	result.Artifact.MissionID = mission.ID
	c.JSON(http.StatusCreated, result)
}

func (a *agentAPI) builders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"projects": a.runtime.Builder().List()})
}

func (a *agentAPI) createBuilder(c *gin.Context) {
	var spec agent.BuilderSpec
	if err := decodeJSON(c, &spec); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	project, err := a.runtime.Builder().Create(c.Request.Context(), spec)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (a *agentAPI) previewBuilder(c *gin.Context) {
	project, artifact, err := a.runtime.Builder().Preview(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": project, "artifact": artifact})
}

func (a *agentAPI) exportBuilder(c *gin.Context) {
	project, archivePath, err := a.runtime.Builder().Export(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"project": project, "archive_path": archivePath})
}

func (a *agentAPI) publishBuilder(c *gin.Context) {
	project, publishedPath, err := a.runtime.Builder().PublishLocal(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"project": project, "published_path": publishedPath})
}

func (a *agentAPI) builderPreviewFile(c *gin.Context) {
	project, err := a.runtime.Builder().Get(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	relative := strings.TrimPrefix(c.Param("path"), "/")
	if relative == "" {
		relative = project.Entry
	}
	path, err := containedPath(project.Root, relative)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		writeAgentError(c, http.StatusNotFound, errors.New("preview file not found"))
		return
	}
	c.File(path)
}

func containedPath(root, requested string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return "", errors.New("input_path is required")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	candidate := requested
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootAbs, candidate)
	}
	candidate, err = filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(rootAbs, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", errors.New("input_path escapes mission workspace")
	}
	return candidate, nil
}

func (a *agentAPI) artifact(c *gin.Context) {
	manifest, path, err := a.runtime.Artifact(c.Param("id"), c.Param("artifact_id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.Header("X-Artifact-SHA256", manifest.SHA256)
	c.FileAttachment(path, manifest.Name)
}

func (a *agentAPI) runMission(c *gin.Context) {
	id := c.Param("id")
	if _, err := a.runtime.GetMission(id); err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	job, err := a.runtime.EnqueueMission(id)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"mission_id": id, "state": agent.MissionRunning, "job": job})
}

func (a *agentAPI) cancelMission(c *gin.Context) {
	mission, err := a.runtime.Cancel(c.Param("id"))
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, mission)
}

func (a *agentAPI) decideApproval(c *gin.Context) {
	var request struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason,omitempty"`
	}
	if err := decodeJSON(c, &request); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	mission, err := a.runtime.DecideApproval(c.Param("id"), c.Param("approval_id"), request.Approved, request.Reason)
	if err != nil {
		writeAgentError(c, statusForAgentError(err), err)
		return
	}
	c.JSON(http.StatusOK, mission)
}

func decodeJSON(c *gin.Context, value any) error {
	if c.Request.Body == nil {
		return errors.New("request body is required")
	}
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func writeAgentError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": strings.TrimSpace(err.Error())})
}

func statusForAgentError(err error) int {
	if errors.Is(err, os.ErrNotExist) {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
