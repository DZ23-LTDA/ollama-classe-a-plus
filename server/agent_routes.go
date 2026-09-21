package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
	"github.com/ollama/ollama/internal/agent"
)

type agentAPI struct {
	runtime *agent.Runtime
	context *agent.ContextStore
}

func newAgentAPI(runtime *agent.Runtime) *agentAPI {
	return &agentAPI{runtime: runtime, context: runtime.Context()}
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
	var planner agent.Planner = agent.RulePlanner{}
	if model := strings.TrimSpace(os.Getenv("OLLAMA_AGENT_MODEL")); model != "" {
		planner = agent.OllamaPlanner{
			Client:   api.NewClient(envconfig.ConnectableHost(), http.DefaultClient),
			Model:    model,
			Fallback: agent.RulePlanner{},
		}
	}
	return agent.NewRuntime(agent.RuntimeConfig{Store: store, Context: contextStore, Planner: planner, WorkspaceRoot: workspaceRoot})
}

func (a *agentAPI) register(r *gin.Engine) {
	group := r.Group("/api/agent/v1")
	group.GET("/health", a.health)
	group.GET("/tools", a.tools)
	group.GET("/skills", a.skills)
	group.GET("/schedules", a.schedules)
	group.POST("/schedules", a.createSchedule)
	group.POST("/webhooks/:schedule_id", a.webhook)
	group.POST("/projects", a.createProject)
	group.GET("/projects/:id", a.getProject)
	group.POST("/projects/:id/memories", a.addMemory)
	group.GET("/projects/:id/memories", a.searchMemories)
	group.POST("/missions", a.createMission)
	group.GET("/missions/:id", a.getMission)
	group.GET("/missions/:id/events", a.events)
	group.GET("/missions/:id/artifacts/:artifact_id", a.artifact)
	group.POST("/missions/:id/run", a.runMission)
	group.POST("/missions/:id/cancel", a.cancelMission)
	group.POST("/missions/:id/approvals/:approval_id", a.decideApproval)
}

func (a *agentAPI) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "runtime": "agent-v1"})
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

func (a *agentAPI) addMemory(c *gin.Context) {
	var memory agent.Memory
	if err := decodeJSON(c, &memory); err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	memory.ProjectID = c.Param("id")
	created, err := a.context.AddMemory(memory)
	if err != nil {
		writeAgentError(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (a *agentAPI) searchMemories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"project_id": c.Param("id"), "memories": a.context.SearchMemories(c.Param("id"), c.Query("q"), 20)})
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
	go func() {
		if err := a.runtime.Run(context.Background(), id); err != nil {
			// The failure is persisted by the runtime. The HTTP request has already
			// returned, so logging is intentionally left to the server middleware.
			_ = err
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"mission_id": id, "state": agent.MissionRunning})
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
