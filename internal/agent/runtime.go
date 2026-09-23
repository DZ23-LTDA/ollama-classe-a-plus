package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Runtime struct {
	store            Store
	planner          Planner
	plannerResolver  PlannerResolver
	tools            *Registry
	capabilityPolicy CapabilityPolicy
	workspaceRoot    string
	dataRoot         string
	context          *ContextStore
	company          *CompanyStore
	remoteMCP        *RemoteMCPManager
	metrics          *RuntimeMetrics
	connectors       *ConnectorManager
	mcp              *MCPManager
	queue            *JobQueue
	redisQueue       *RedisQueue
	traces           *TraceStore
	telemetry        *Telemetry
	media            *MediaManager
	builder          *BuilderService
	collaboration    *CollaborationStore
	orchestrator     *AgentOrchestrator
	research         *ResearchEngine
	devices          *DeviceStore
	ingestion        DocumentIngestor
	push             *PushService
	deployments      *DeploymentManager
	mu               *sync.Mutex
	running          map[string]bool
	activeCancels    map[string]context.CancelFunc
}

type RuntimeConfig struct {
	Store           Store
	Planner         Planner
	PlannerResolver PlannerResolver
	Tools           *Registry
	WorkspaceRoot   string
	DataRoot        string
	Context         *ContextStore
	Company         *CompanyStore
	RemoteMCP       *RemoteMCPManager
	Connectors      *ConnectorManager
	MCP             *MCPManager
	Queue           *JobQueue
	RedisQueue      *RedisQueue
	Traces          *TraceStore
	Telemetry       *Telemetry
	Media           *MediaManager
	Builder         *BuilderService
	Collaboration   *CollaborationStore
	Devices         *DeviceStore
	Push            *PushService
	Deployments     *DeploymentManager
}

var (
	ErrApprovalVersionConflict = errors.New("approval mission version conflict")
	ErrApprovalNonceMismatch   = errors.New("approval nonce mismatch")
	ErrQueueJobForbidden       = errors.New("job is outside the active organization")
)

func NewRuntime(config RuntimeConfig) (*Runtime, error) {
	store := config.Store
	if store == nil {
		store = NewMemoryStore()
	}
	planner := config.Planner
	if planner == nil {
		planner = RulePlanner{}
	}
	tools := config.Tools
	if tools == nil {
		tools = NewRegistry()
	}
	capabilityPolicy := DefaultCapabilityPolicy()
	if config.Connectors != nil {
		tools.Register(connectorTool{manager: config.Connectors})
	}
	if config.MCP != nil {
		tools.Register(mcpCallTool{manager: config.MCP})
	}
	if config.RemoteMCP != nil {
		tools.Register(remoteMCPCallTool{manager: config.RemoteMCP})
	}
	for _, descriptor := range tools.Descriptors() {
		if err := capabilityPolicy.ValidateToolDescriptor(descriptor); err != nil {
			return nil, err
		}
	}
	root := config.WorkspaceRoot
	if strings.TrimSpace(root) == "" {
		root, _ = os.Getwd()
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	dataRoot, err := resolveRuntimeDataRoot(root, config.DataRoot)
	if err != nil {
		return nil, err
	}
	contextStore := config.Context
	if contextStore == nil {
		contextStore, err = NewContextStore(filepath.Join(dataRoot, ".agent-context"))
		if err != nil {
			return nil, err
		}
	}
	companyStore := config.Company
	if companyStore == nil {
		companyStore, err = NewCompanyStore(filepath.Join(dataRoot, ".agent-companies"))
		if err != nil {
			return nil, err
		}
	}
	queue := config.Queue
	if queue == nil {
		queue, err = NewJobQueue(filepath.Join(dataRoot, ".agent-queue"))
		if err != nil {
			return nil, err
		}
	}
	traces := config.Traces
	if traces == nil {
		traces, err = NewTraceStore(filepath.Join(dataRoot, ".agent-traces"))
		if err != nil {
			return nil, err
		}
	}
	builder := config.Builder
	if builder == nil {
		builder, err = NewBuilderService(filepath.Join(dataRoot, ".agent-builders"))
		if err != nil {
			return nil, err
		}
	}
	collaboration := config.Collaboration
	if collaboration == nil {
		collaboration, err = NewCollaborationStore(filepath.Join(dataRoot, ".agent-collaboration"))
		if err != nil {
			return nil, err
		}
	}
	telemetry := config.Telemetry
	if telemetry == nil {
		telemetry, err = NewTelemetry(context.Background(), "")
		if err != nil {
			return nil, err
		}
	}
	runtime := &Runtime{store: store, planner: planner, plannerResolver: config.PlannerResolver, tools: tools, capabilityPolicy: capabilityPolicy, workspaceRoot: root, dataRoot: dataRoot, context: contextStore, company: companyStore, remoteMCP: config.RemoteMCP, metrics: &RuntimeMetrics{}, connectors: config.Connectors, mcp: config.MCP, queue: queue, redisQueue: config.RedisQueue, traces: traces, telemetry: telemetry, media: config.Media, builder: builder, collaboration: collaboration, push: config.Push, deployments: config.Deployments, mu: &sync.Mutex{}, running: make(map[string]bool), activeCancels: make(map[string]context.CancelFunc)}
	orchestrator, err := NewAgentOrchestrator(filepath.Join(dataRoot, ".agent-orchestrator"), runtime.SubagentRunner)
	if err != nil {
		return nil, err
	}
	runtime.orchestrator = orchestrator
	runtime.research = NewResearchEngine()
	devices := config.Devices
	if devices == nil {
		devices, err = NewDeviceStore(filepath.Join(dataRoot, ".agent-devices"))
		if err != nil {
			return nil, err
		}
	}
	runtime.devices = devices
	runtime.ingestion = DocumentIngestor{Context: contextStore, Research: runtime.research}
	return runtime, nil
}

// WithOrganization returns a request-scoped runtime view. Shared in-memory stores
// remain compatible, while PostgresStore receives a transaction-local RLS scope.
func (r *Runtime) WithOrganization(organizationID string) *Runtime {
	if r == nil {
		return nil
	}
	view := *r
	if postgres, ok := r.store.(*PostgresStore); ok {
		view.store = postgres.WithOrganization(organizationID)
	}
	return &view
}

func (r *Runtime) SetPlannerResolver(resolver PlannerResolver) {
	if r != nil {
		r.plannerResolver = resolver
	}
}

func (r *Runtime) Context() *ContextStore {
	return r.context
}

func (r *Runtime) CompanyStore() *CompanyStore {
	return r.company
}

func (r *Runtime) Metrics() MetricsSnapshot {
	return r.metrics.Snapshot()
}

func (r *Runtime) Connectors() []ConnectorConfig {
	if r.connectors == nil {
		return nil
	}
	return r.connectors.List()
}

func (r *Runtime) SetAuthStore(store *AuthStore) {
	if r != nil && r.connectors != nil {
		r.connectors.SetOAuthStore(store)
	}
}

func (r *Runtime) SetConnectorEnabled(id string, enabled bool) error {
	if r.connectors == nil {
		return errors.New("connector manager is unavailable")
	}
	return r.connectors.SetEnabled(id, enabled)
}

func (r *Runtime) RemoveConnector(id string) error {
	if r.connectors == nil {
		return errors.New("connector manager is unavailable")
	}
	return r.connectors.Remove(id)
}

func (r *Runtime) SetMCPEnabled(id string, enabled bool) error {
	if r.mcp == nil {
		return errors.New("MCP manager is unavailable")
	}
	return r.mcp.SetEnabled(id, enabled)
}

func (r *Runtime) RemoveMCP(id string) error {
	if r.mcp == nil {
		return errors.New("MCP manager is unavailable")
	}
	return r.mcp.Remove(id)
}

func (r *Runtime) SetRemoteMCPEnabled(id string, enabled bool) error {
	if r.remoteMCP == nil {
		return errors.New("remote MCP manager is unavailable")
	}
	return r.remoteMCP.SetEnabled(id, enabled)
}

func (r *Runtime) RemoveRemoteMCP(id string) error {
	if r.remoteMCP == nil {
		return errors.New("remote MCP manager is unavailable")
	}
	return r.remoteMCP.Remove(id)
}

func (r *Runtime) SetSkillEnabled(id string, enabled bool) error {
	if r.context == nil {
		return errors.New("context store is unavailable")
	}
	return r.context.SetSkillEnabled(id, enabled)
}

func (r *Runtime) RemoveSkill(id string) error {
	if r.context == nil {
		return errors.New("context store is unavailable")
	}
	return r.context.RemoveSkill(id)
}

func (r *Runtime) MCPServers() []MCPServerConfig {
	if r.mcp == nil {
		return nil
	}
	return r.mcp.List()
}

func (r *Runtime) RemoteMCPServers() []RemoteMCPServerConfig {
	if r.remoteMCP == nil {
		return nil
	}
	return r.remoteMCP.List()
}

func (r *Runtime) Traces(traceID string) []TraceSpan {
	return r.traces.List(traceID, 500)
}

func (r *Runtime) TracesForOrganization(organizationID, traceID string) []TraceSpan {
	return r.traces.ListForOrganization(organizationID, traceID, 500)
}

func (r *Runtime) Media() *MediaManager { return r.media }

func (r *Runtime) Builder() *BuilderService { return r.builder }

func (r *Runtime) Collaboration() *CollaborationStore { return r.collaboration }

func (r *Runtime) Orchestrator() *AgentOrchestrator { return r.orchestrator }

func (r *Runtime) Research() *ResearchEngine { return r.research }

func (r *Runtime) Devices() *DeviceStore { return r.devices }

func (r *Runtime) Ingestion() DocumentIngestor { return r.ingestion }

func (r *Runtime) Push() *PushService { return r.push }

func (r *Runtime) Deployments() *DeploymentManager { return r.deployments }

func (r *Runtime) CreateMission(ctx context.Context, request CreateMissionRequest) (Mission, error) {
	objective := strings.TrimSpace(request.Objective)
	if objective == "" {
		return Mission{}, errors.New("objective is required")
	}
	if len(objective) > 8<<10 {
		return Mission{}, errors.New("objective is too long")
	}
	provider := strings.TrimSpace(request.Provider)
	if provider == "" {
		provider = "ollama-local"
	}
	workspace, err := r.resolveWorkspace(request.Workspace)
	if err != nil {
		return Mission{}, err
	}
	capabilities := normalizeMissionCapabilities(request.Capabilities)
	if len(capabilities) == 0 {
		capabilities = []string{"workspace:read"}
	}
	capabilities, err = r.capabilityPolicy.ValidateMissionCapabilities(capabilities)
	if err != nil {
		return Mission{}, err
	}
	now := time.Now().UTC()
	mission := Mission{ID: "mis_" + uuid.NewString(), Version: 1, Objective: objective, Provider: provider, Model: strings.TrimSpace(request.Model), Workspace: workspace, ProjectID: strings.TrimSpace(request.ProjectID), OrganizationID: strings.TrimSpace(request.OrganizationID), Capabilities: capabilities, AutoRun: request.AutoRun, State: MissionPlanning, CreatedAt: now, UpdatedAt: now}
	if err := r.store.PutMission(mission); err != nil {
		return Mission{}, err
	}
	r.metrics.missionsCreated.Add(1)
	_ = r.event(mission, "mission.created", "", map[string]any{"objective": objective})
	planner := r.planner
	if provider != "ollama-local" {
		if r.plannerResolver == nil {
			return Mission{}, fmt.Errorf("provider %q is not configured in this runtime", provider)
		}
		planner, err = r.plannerResolver.ResolvePlanner(provider, mission.Model)
		if err != nil {
			return Mission{}, err
		}
		if planner == nil {
			return Mission{}, fmt.Errorf("provider %q returned no planner", provider)
		}
	}
	plan, err := planner.Plan(ctx, mission)
	if err != nil {
		return r.failMission(mission, err)
	}
	plan, err = normalizeSteps(plan)
	if err != nil {
		return r.failMission(mission, err)
	}
	for index := range plan {
		tool, ok := r.tools.Get(plan[index].Kind)
		if !ok {
			return r.failMission(mission, fmt.Errorf("planner returned unregistered tool %q", plan[index].Kind))
		}
		descriptor := tool.Descriptor()
		if err := r.capabilityPolicy.ValidateToolDescriptor(descriptor); err != nil {
			return r.failMission(mission, err)
		}
		plan[index].RequiresApproval = plan[index].RequiresApproval || descriptor.RequiresApproval
		if riskRank(descriptor.Risk) > riskRank(plan[index].Risk) {
			plan[index].Risk = descriptor.Risk
		}
	}
	mission.Plan = plan
	mission.State = MissionReady
	approvalExpiresAt := now.Add(30 * time.Minute)
	for _, step := range plan {
		if step.RequiresApproval {
			descriptor, _ := r.tools.Get(step.Kind)
			mission.Approvals = append(mission.Approvals, Approval{ID: "apr_" + uuid.NewString(), MissionID: mission.ID, StepID: step.ID, OrganizationID: mission.OrganizationID, Policy: toolApprovalPolicy(descriptor.Descriptor(), step.Risk), Nonce: uuid.NewString(), Status: ApprovalPending, ExpiresAt: &approvalExpiresAt, CreatedAt: now, UpdatedAt: now})
		}
	}
	if len(mission.Approvals) > 0 {
		mission.State = MissionAwaitingApproval
	}
	mission.Version++
	mission.UpdatedAt = time.Now().UTC()
	if err := r.store.PutMission(mission); err != nil {
		return Mission{}, err
	}
	_ = r.event(mission, "mission.planned", "", map[string]any{"steps": len(plan), "approvals": len(mission.Approvals)})
	if request.AutoRun && mission.State == MissionReady {
		_, _ = r.EnqueueMission(mission.ID)
	}
	return mission, nil
}

func normalizeMissionCapabilities(capabilities []string) []string {
	seen := make(map[string]struct{}, len(capabilities))
	result := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		capability = strings.TrimSpace(capability)
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		result = append(result, capability)
	}
	sort.Strings(result)
	return result
}

func (r *Runtime) GetMission(id string) (Mission, error) {
	return r.store.GetMission(strings.TrimSpace(id))
}

func (r *Runtime) ListMissions() ([]Mission, error) {
	return r.store.ListMissions()
}

func (r *Runtime) Start(ctx context.Context) {
	worker := func(jobContext context.Context, job QueueJob) error {
		return r.Run(jobContext, job.MissionID)
	}
	if r.redisQueue != nil {
		r.redisQueue.Start(ctx, "agent-runtime", worker)
	} else {
		r.queue.Start(ctx, "agent-runtime", worker)
	}
	go func() {
		r.resumePending(ctx)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.resumePending(ctx)
			}
		}
	}()
}

func (r *Runtime) resumePending(ctx context.Context) {
	for _, schedule := range r.context.ClaimDueSchedules(time.Now().UTC()) {
		if companyID := companyIDFromWorkspace(schedule.Workspace); companyID != "" && r.company != nil {
			company, err := r.company.Get(companyID)
			if err == nil && (company.Status == CompanyPaused || company.Risk.Paused) {
				continue
			}
		}
		_, _ = r.CreateMission(ctx, CreateMissionRequest{Objective: schedule.Objective, Model: schedule.Model, Workspace: schedule.Workspace, ProjectID: schedule.ProjectID, OrganizationID: schedule.OrganizationID, AutoRun: true})
	}
	missions, err := r.store.ListMissions()
	if err != nil {
		return
	}
	for _, mission := range missions {
		resume := mission.State == MissionRunning || mission.State == MissionRecovering || (mission.State == MissionReady && mission.AutoRun)
		if !resume || !r.approvalsReady(mission) {
			continue
		}
		_, _ = r.EnqueueMission(mission.ID)
	}
}

func companyIDFromWorkspace(workspace string) string {
	workspace = strings.TrimSpace(workspace)
	if !strings.HasPrefix(workspace, "company://") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(workspace, "company://"))
}

func (r *Runtime) EnqueueMission(missionID string) (QueueJob, error) {
	if _, err := r.store.GetMission(strings.TrimSpace(missionID)); err != nil {
		return QueueJob{}, err
	}
	if r.redisQueue != nil {
		return r.redisQueue.Enqueue(missionID, 3)
	}
	return r.queue.Enqueue(missionID, 3)
}

func (r *Runtime) QueueJobs(status QueueStatus) []QueueJob {
	if r.redisQueue != nil {
		return r.redisQueue.List(status)
	}
	return r.queue.List(status)
}

func (r *Runtime) ReplayJob(jobID string) (QueueJob, error) {
	if r.redisQueue != nil {
		return r.redisQueue.Replay(jobID)
	}
	return r.queue.Replay(jobID)
}

func (r *Runtime) QueueJobsForOrganization(organizationID string, status QueueStatus) ([]QueueJob, error) {
	jobs := r.QueueJobs(status)
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return jobs, nil
	}
	filtered := make([]QueueJob, 0, len(jobs))
	for _, job := range jobs {
		mission, err := r.store.GetMission(job.MissionID)
		if err != nil {
			continue
		}
		if mission.OrganizationID == organizationID {
			filtered = append(filtered, job)
		}
	}
	return filtered, nil
}

func (r *Runtime) ReplayJobForOrganization(jobID, organizationID string) (QueueJob, error) {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return r.ReplayJob(jobID)
	}
	for _, job := range r.QueueJobs("") {
		if job.ID != strings.TrimSpace(jobID) {
			continue
		}
		mission, err := r.store.GetMission(job.MissionID)
		if err != nil {
			return QueueJob{}, err
		}
		if mission.OrganizationID != organizationID {
			return QueueJob{}, ErrQueueJobForbidden
		}
		return r.ReplayJob(jobID)
	}
	return QueueJob{}, os.ErrNotExist
}

func (r *Runtime) ListTools() []ToolDescriptor {
	return r.tools.Descriptors()
}

func (r *Runtime) Run(ctx context.Context, id string) (runErr error) {
	ctx, otelSpan := r.telemetry.Start(ctx, "agent.mission.run", map[string]string{"mission.id": id})
	defer otelSpan.End()
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("mission id is required")
	}
	runCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	if r.running[id] {
		r.mu.Unlock()
		cancel()
		return nil
	}
	r.running[id] = true
	r.activeCancels[id] = cancel
	r.mu.Unlock()
	defer func() {
		cancel()
		r.mu.Lock()
		delete(r.running, id)
		delete(r.activeCancels, id)
		r.mu.Unlock()
	}()

	mission, err := r.store.GetMission(id)
	if err != nil {
		return err
	}
	missionSpan := r.traces.StartForOrganization(mission.OrganizationID, "tr_"+id, "", "mission.run", map[string]any{"mission_id": id})
	defer func() { missionSpan.End("ok", runErr) }()
	if mission.State == MissionCompleted || mission.State == MissionCancelled {
		return nil
	}
	if !r.approvalsReady(mission) {
		mission.State = MissionAwaitingApproval
		mission.UpdatedAt = time.Now().UTC()
		_ = r.store.PutMission(mission)
		return nil
	}
	mission.State = MissionRunning
	mission.Version++
	mission.UpdatedAt = time.Now().UTC()
	if err := r.store.PutMission(mission); err != nil {
		return err
	}
	_ = r.event(mission, "mission.running", "", nil)

	for index := 0; index < len(mission.Plan); index++ {
		if r.missionCancelled(id) {
			return nil
		}
		step := &mission.Plan[index]
		if step.State == StepSucceeded {
			continue
		}
		tool, ok := r.tools.Get(step.Kind)
		if !ok {
			return r.failStep(mission, step, fmt.Errorf("tool %q is not registered", step.Kind))
		}
		if err := r.capabilityPolicy.ValidateToolDescriptor(tool.Descriptor()); err != nil {
			return r.failStep(mission, step, err)
		}
		if !r.capabilityPolicy.Allows(tool.Descriptor(), mission.Capabilities) {
			return r.failStep(mission, step, fmt.Errorf("%w: %s", ErrCapabilityDenied, step.Kind))
		}
		if step.RequiresApproval && !r.stepApproved(mission, step.ID) {
			step.State = StepBlocked
			mission.State = MissionAwaitingApproval
			mission.Version++
			mission.UpdatedAt = time.Now().UTC()
			_ = r.store.PutMission(mission)
			_ = r.event(mission, "step.awaiting_approval", step.ID, nil)
			return nil
		}
		step.State = StepRunning
		step.Attempts++
		r.metrics.stepsStarted.Add(1)
		r.metrics.toolCalls.Add(1)
		mission.State = MissionObserving
		mission.Version++
		mission.UpdatedAt = time.Now().UTC()
		if err := r.store.PutMission(mission); err != nil {
			return err
		}
		_ = r.event(mission, "step.started", step.ID, map[string]any{"tool": step.Kind, "attempt": step.Attempts})
		toolSpan := r.traces.StartForOrganization(mission.OrganizationID, "tr_"+mission.ID, missionSpan.ID(), "tool."+step.Kind, map[string]any{"mission_id": mission.ID, "step_id": step.ID, "tool": step.Kind})
		result, executeErr := tool.Execute(runCtx, ToolContext{MissionID: mission.ID, StepID: step.ID, Workspace: mission.Workspace, OrganizationID: mission.OrganizationID}, step.Input)
		toolSpan.End("ok", executeErr)
		if r.missionCancelled(id) {
			return nil
		}
		if executeErr != nil {
			if step.Attempts < 2 {
				r.metrics.retries.Add(1)
				step.State = StepPending
				mission.State = MissionRecovering
				mission.Version++
				mission.UpdatedAt = time.Now().UTC()
				_ = r.store.PutMission(mission)
				_ = r.event(mission, "step.retry_scheduled", step.ID, map[string]any{"error": RedactDLP(executeErr.Error())})
				index--
				continue
			}
			return r.failStep(mission, step, executeErr)
		}
		step.State = StepSucceeded
		r.metrics.stepsSucceeded.Add(1)
		step.Result = RedactValue(result.Value)
		step.Error = ""
		mission.Artifacts = append(mission.Artifacts, result.Artifacts...)
		mission.State = MissionRunning
		mission.Version++
		mission.UpdatedAt = time.Now().UTC()
		if err := r.store.PutMission(mission); err != nil {
			return err
		}
		_ = r.event(mission, "step.succeeded", step.ID, map[string]any{"artifacts": len(result.Artifacts)})
	}
	if r.missionCancelled(id) {
		return nil
	}
	completed := time.Now().UTC()
	mission.State = MissionCompleted
	r.metrics.missionsCompleted.Add(1)
	mission.CompletedAt = &completed
	mission.Version++
	mission.UpdatedAt = completed
	if err := r.store.PutMission(mission); err != nil {
		return err
	}
	_ = r.event(mission, "mission.completed", "", map[string]any{"artifacts": len(mission.Artifacts)})
	return nil
}

func (r *Runtime) Cancel(id string) (Mission, error) {
	id = strings.TrimSpace(id)
	mission, err := r.store.GetMission(id)
	if err != nil {
		return Mission{}, err
	}
	if mission.State == MissionCompleted {
		return Mission{}, errors.New("completed mission cannot be cancelled")
	}
	r.mu.Lock()
	cancel := r.activeCancels[id]
	r.mu.Unlock()
	if mission.State == MissionCancelled {
		if cancel != nil {
			cancel()
		}
		return mission, nil
	}
	for index := range mission.Plan {
		if mission.Plan[index].State == StepPending || mission.Plan[index].State == StepRunning || mission.Plan[index].State == StepBlocked {
			mission.Plan[index].State = StepBlocked
			mission.Plan[index].Error = "mission cancelled"
		}
	}
	mission.State = MissionCancelled
	mission.Version++
	mission.UpdatedAt = time.Now().UTC()
	if err := r.store.PutMission(mission); err != nil {
		return Mission{}, err
	}
	if cancel != nil {
		cancel()
	}
	_ = r.event(mission, "mission.cancelled", "", nil)
	return mission, nil
}

func (r *Runtime) DecideApproval(missionID, approvalID string, approved bool, reason string) (Mission, error) {
	mission, err := r.store.GetMission(strings.TrimSpace(missionID))
	if err != nil {
		return Mission{}, err
	}
	return r.DecideApprovalForActor(missionID, approvalID, approved, reason, "local", mission.OrganizationID)
}

func (r *Runtime) DecideApprovalForActor(missionID, approvalID string, approved bool, reason, actorID, organizationID string) (Mission, error) {
	return r.decideApprovalForActor(missionID, approvalID, approved, reason, actorID, organizationID, 0, "", false)
}

func (r *Runtime) DecideApprovalForActorCAS(missionID, approvalID string, approved bool, reason, actorID, organizationID string, expectedVersion int64, nonce string) (Mission, error) {
	return r.decideApprovalForActor(missionID, approvalID, approved, reason, actorID, organizationID, expectedVersion, nonce, true)
}

func (r *Runtime) decideApprovalForActor(missionID, approvalID string, approved bool, reason, actorID, organizationID string, expectedVersion int64, nonce string, requireNonce bool) (Mission, error) {
	mission, err := r.store.GetMission(strings.TrimSpace(missionID))
	if err != nil {
		return Mission{}, err
	}
	if expectedVersion > 0 && mission.Version != expectedVersion {
		return Mission{}, ErrApprovalVersionConflict
	}
	for index := range mission.Approvals {
		if mission.Approvals[index].ID != approvalID {
			continue
		}
		if mission.Approvals[index].Status != ApprovalPending {
			continue
		}
		if mission.Approvals[index].ExpiresAt != nil && time.Now().UTC().After(*mission.Approvals[index].ExpiresAt) {
			return Mission{}, errors.New("approval has expired")
		}
		if mission.OrganizationID != "" && strings.TrimSpace(organizationID) != mission.OrganizationID {
			return Mission{}, errors.New("approval organization mismatch")
		}
		if strings.TrimSpace(actorID) == "" {
			return Mission{}, errors.New("approval actor is required")
		}
		if strings.TrimSpace(reason) == "" {
			return Mission{}, errors.New("approval reason is required")
		}
		if requireNonce && (strings.TrimSpace(nonce) == "" || nonce != mission.Approvals[index].Nonce) {
			return Mission{}, ErrApprovalNonceMismatch
		}
		mission.Approvals[index].ActorID = strings.TrimSpace(actorID)
		mission.Approvals[index].OrganizationID = mission.OrganizationID
		if approved {
			mission.Approvals[index].Status = ApprovalApproved
		} else {
			mission.Approvals[index].Status = ApprovalRejected
		}
		mission.Approvals[index].Reason = strings.TrimSpace(reason)
		mission.Approvals[index].UpdatedAt = time.Now().UTC()
		if approved && r.approvalsReady(mission) {
			mission.State = MissionReady
		} else if !approved {
			mission.State = MissionFailed
			mission.LastError = "approval rejected"
		}
		mission.Version++
		mission.UpdatedAt = time.Now().UTC()
		var saveErr error
		if requireNonce {
			saveErr = r.store.PutMissionIfVersion(mission, expectedVersion)
			if errors.Is(saveErr, ErrMissionVersionConflict) {
				saveErr = ErrApprovalVersionConflict
			}
		} else {
			saveErr = r.store.PutMission(mission)
		}
		if saveErr != nil {
			return Mission{}, saveErr
		}
		r.metrics.approvals.Add(1)
		_ = r.event(mission, "approval.decided", mission.Approvals[index].StepID, map[string]any{"approved": approved, "reason": reason})
		return mission, nil
	}
	return Mission{}, errors.New("approval not found or already decided")
}

func (r *Runtime) missionCancelled(id string) bool {
	mission, err := r.store.GetMission(strings.TrimSpace(id))
	return err == nil && mission.State == MissionCancelled
}

func (r *Runtime) Events(id string) ([]Event, error) {
	return r.store.ListEvents(strings.TrimSpace(id))
}

func (r *Runtime) Artifact(missionID, artifactID string) (ArtifactManifest, string, error) {
	mission, err := r.store.GetMission(strings.TrimSpace(missionID))
	if err != nil {
		return ArtifactManifest{}, "", err
	}
	for _, artifact := range mission.Artifacts {
		if artifact.ID != artifactID {
			continue
		}
		path, err := safeWorkspacePath(mission.Workspace, artifact.Path)
		if err != nil {
			return ArtifactManifest{}, "", err
		}
		return artifact, path, nil
	}
	return ArtifactManifest{}, "", os.ErrNotExist
}

func (r *Runtime) approvalsReady(mission Mission) bool {
	for _, approval := range mission.Approvals {
		if approval.Status != ApprovalApproved {
			return false
		}
	}
	return true
}

func (r *Runtime) stepApproved(mission Mission, stepID string) bool {
	for _, approval := range mission.Approvals {
		if approval.StepID == stepID {
			return approval.Status == ApprovalApproved
		}
	}
	return true
}

func (r *Runtime) failMission(mission Mission, err error) (Mission, error) {
	mission.State = MissionFailed
	r.metrics.missionsFailed.Add(1)
	mission.LastError = RedactDLP(err.Error())
	mission.Version++
	mission.UpdatedAt = time.Now().UTC()
	if saveErr := r.store.PutMission(mission); saveErr != nil {
		return Mission{}, saveErr
	}
	_ = r.event(mission, "mission.failed", "", map[string]any{"error": err.Error()})
	return mission, err
}

func (r *Runtime) failStep(mission Mission, step *Step, err error) error {
	step.State = StepFailed
	r.metrics.stepsFailed.Add(1)
	step.Error = RedactDLP(err.Error())
	mission.State = MissionFailed
	mission.LastError = RedactDLP(err.Error())
	mission.Version++
	mission.UpdatedAt = time.Now().UTC()
	if saveErr := r.store.PutMission(mission); saveErr != nil {
		return saveErr
	}
	_ = r.event(mission, "step.failed", step.ID, map[string]any{"error": RedactDLP(err.Error()), "attempts": step.Attempts})
	return err
}

func (r *Runtime) event(mission Mission, eventType, stepID string, payload any) error {
	err := r.store.AppendEvent(Event{ID: "evt_" + uuid.NewString(), MissionID: mission.ID, OrganizationID: mission.OrganizationID, Type: eventType, StepID: stepID, Payload: RedactValue(payload), CreatedAt: time.Now().UTC()})
	if r.push != nil && mission.OrganizationID != "" && (eventType == "mission.completed" || eventType == "mission.failed" || eventType == "step.awaiting_approval") {
		title := "DZ23 Agentic"
		body := "A missão " + mission.ID + " mudou de estado"
		if eventType == "mission.completed" {
			body = "A missão " + mission.ID + " foi concluída"
		}
		go func() {
			_ = r.push.NotifyOrganization(context.Background(), mission.OrganizationID, title, body, map[string]any{"mission_id": mission.ID, "event": eventType})
		}()
	}
	return err
}

func riskRank(risk RiskClass) int {
	switch risk {
	case RiskRead:
		return 0
	case RiskWrite:
		return 1
	case RiskExternalSideEffect:
		return 2
	case RiskDestructive:
		return 3
	default:
		return 2
	}
}

func (r *Runtime) resolveWorkspace(requested string) (string, error) {
	root, err := filepath.Abs(r.workspaceRoot)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(requested) == "" {
		return root, nil
	}
	candidate, err := filepath.Abs(requested)
	if err != nil {
		return "", err
	}
	if !isWithin(root, candidate) {
		return "", errors.New("workspace must be inside the configured agent root")
	}
	if err := os.MkdirAll(candidate, 0o700); err != nil {
		return "", err
	}
	return candidate, nil
}

func (r *Runtime) ConnectorsForOrganization(organizationID string) []ConnectorConfig {
	if r.connectors == nil {
		return nil
	}
	return r.connectors.ListForOrganization(organizationID)
}

func (r *Runtime) SetConnectorEnabledForOrganization(organizationID, id string, enabled bool) error {
	if r.connectors == nil {
		return errors.New("connector manager is unavailable")
	}
	return r.connectors.SetEnabledForOrganization(organizationID, id, enabled)
}

func (r *Runtime) RemoveConnectorForOrganization(organizationID, id string) error {
	if r.connectors == nil {
		return errors.New("connector manager is unavailable")
	}
	return r.connectors.RemoveForOrganization(organizationID, id)
}

func (r *Runtime) MCPServersForOrganization(organizationID string) []MCPServerConfig {
	if r.mcp == nil {
		return nil
	}
	return r.mcp.ListForOrganization(organizationID)
}

func (r *Runtime) SetMCPEnabledForOrganization(organizationID, id string, enabled bool) error {
	if r.mcp == nil {
		return errors.New("MCP manager is unavailable")
	}
	return r.mcp.SetEnabledForOrganization(organizationID, id, enabled)
}

func (r *Runtime) RemoveMCPForOrganization(organizationID, id string) error {
	if r.mcp == nil {
		return errors.New("MCP manager is unavailable")
	}
	return r.mcp.RemoveForOrganization(organizationID, id)
}

func (r *Runtime) RemoteMCPServersForOrganization(organizationID string) []RemoteMCPServerConfig {
	if r.remoteMCP == nil {
		return nil
	}
	return r.remoteMCP.ListForOrganization(organizationID)
}

func (r *Runtime) SetRemoteMCPEnabledForOrganization(organizationID, id string, enabled bool) error {
	if r.remoteMCP == nil {
		return errors.New("remote MCP manager is unavailable")
	}
	return r.remoteMCP.SetEnabledForOrganization(organizationID, id, enabled)
}

func (r *Runtime) RemoveRemoteMCPForOrganization(organizationID, id string) error {
	if r.remoteMCP == nil {
		return errors.New("remote MCP manager is unavailable")
	}
	return r.remoteMCP.RemoveForOrganization(organizationID, id)
}

func (r *Runtime) SkillsForOrganization(organizationID string) []SkillManifest {
	if r.context == nil {
		return nil
	}
	return r.context.SkillsForOrganization(organizationID)
}

func (r *Runtime) SetSkillEnabledForOrganization(organizationID, id string, enabled bool) error {
	if r.context == nil {
		return errors.New("context store is unavailable")
	}
	return r.context.SetSkillEnabledForOrganization(organizationID, id, enabled)
}

func (r *Runtime) RemoveSkillForOrganization(organizationID, id string) error {
	if r.context == nil {
		return errors.New("context store is unavailable")
	}
	return r.context.RemoveSkillForOrganization(organizationID, id)
}
