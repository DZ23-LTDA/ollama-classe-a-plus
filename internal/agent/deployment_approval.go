package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	DeploymentApprovalPending  = "PENDING"
	DeploymentApprovalApproved = "APPROVED"
	DeploymentApprovalRejected = "REJECTED"
	DeploymentApprovalConsumed = "CONSUMED"
)

var (
	ErrDeploymentApprovalNotFound     = errors.New("deployment approval not found")
	ErrDeploymentApprovalNonce        = errors.New("deployment approval nonce mismatch")
	ErrDeploymentApprovalConflict     = errors.New("deployment approval is no longer pending")
	ErrDeploymentApprovalExpired      = errors.New("deployment approval expired")
	ErrDeploymentApprovalOrganization = errors.New("deployment approval is outside the active organization")
)

type DeploymentApproval struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	BuilderID      string    `json:"builder_id"`
	Provider       string    `json:"provider"`
	Target         string    `json:"target,omitempty"`
	RequestedBy    string    `json:"requested_by"`
	DecidedBy      string    `json:"decided_by,omitempty"`
	Reason         string    `json:"reason,omitempty"`
	Nonce          string    `json:"nonce"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	DecidedAt      time.Time `json:"decided_at,omitempty"`
	ConsumedAt     time.Time `json:"consumed_at,omitempty"`
}

type DeploymentApprovalStore struct {
	mu        sync.Mutex
	root      string
	approvals map[string]DeploymentApproval
}

func NewDeploymentApprovalStore(root string) (*DeploymentApprovalStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("deployment approval store root is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	store := &DeploymentApprovalStore{root: root, approvals: map[string]DeploymentApproval{}}
	data, err := os.ReadFile(filepath.Join(root, "approvals.json"))
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&store.approvals); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("deployment approval store contains trailing JSON")
		}
		return nil, err
	}
	if store.approvals == nil {
		store.approvals = map[string]DeploymentApproval{}
	}
	return store, nil
}

func (s *DeploymentApprovalStore) Request(organizationID, builderID, provider, target, actorID string) (DeploymentApproval, error) {
	if s == nil {
		return DeploymentApproval{}, errors.New("deployment approval store is unavailable")
	}
	organizationID = strings.TrimSpace(organizationID)
	builderID = strings.TrimSpace(builderID)
	provider = strings.ToLower(strings.TrimSpace(provider))
	actorID = strings.TrimSpace(actorID)
	if organizationID == "" || builderID == "" || provider == "" || actorID == "" {
		return DeploymentApproval{}, errors.New("organization, builder, provider and actor are required")
	}
	now := time.Now().UTC()
	approval := DeploymentApproval{ID: "dapr_" + uuid.NewString(), OrganizationID: organizationID, BuilderID: builderID, Provider: provider, Target: strings.TrimSpace(target), RequestedBy: actorID, Nonce: uuid.NewString(), Status: DeploymentApprovalPending, CreatedAt: now, ExpiresAt: now.Add(15 * time.Minute)}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[approval.ID] = approval
	if err := s.persistLocked(); err != nil {
		delete(s.approvals, approval.ID)
		return DeploymentApproval{}, err
	}
	return approval, nil
}

func (s *DeploymentApprovalStore) Decide(id, organizationID, builderID, provider, actorID, reason, nonce string, approved bool) (DeploymentApproval, error) {
	if s == nil {
		return DeploymentApproval{}, errors.New("deployment approval store is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	approval, ok := s.approvals[strings.TrimSpace(id)]
	if !ok {
		return DeploymentApproval{}, ErrDeploymentApprovalNotFound
	}
	if approval.OrganizationID != strings.TrimSpace(organizationID) || approval.BuilderID != strings.TrimSpace(builderID) || approval.Provider != strings.ToLower(strings.TrimSpace(provider)) {
		return DeploymentApproval{}, ErrDeploymentApprovalOrganization
	}
	if strings.TrimSpace(nonce) == "" || nonce != approval.Nonce {
		return DeploymentApproval{}, ErrDeploymentApprovalNonce
	}
	if approval.Status != DeploymentApprovalPending {
		return DeploymentApproval{}, ErrDeploymentApprovalConflict
	}
	if time.Now().UTC().After(approval.ExpiresAt) {
		approval.Status = DeploymentApprovalRejected
		approval.Reason = "approval expired"
		approval.DecidedAt = time.Now().UTC()
		if err := s.persistLocked(); err != nil {
			return DeploymentApproval{}, err
		}
		return DeploymentApproval{}, ErrDeploymentApprovalExpired
	}
	previous := approval
	approval.DecidedBy = strings.TrimSpace(actorID)
	approval.Reason = strings.TrimSpace(reason)
	approval.DecidedAt = time.Now().UTC()
	if approved {
		approval.Status = DeploymentApprovalApproved
	} else {
		approval.Status = DeploymentApprovalRejected
	}
	s.approvals[approval.ID] = approval
	if err := s.persistLocked(); err != nil {
		s.approvals[approval.ID] = previous
		return DeploymentApproval{}, err
	}
	return approval, nil
}

func (s *DeploymentApprovalStore) Consume(id, organizationID, builderID, provider, target, nonce string) (DeploymentApproval, error) {
	if s == nil {
		return DeploymentApproval{}, errors.New("deployment approval store is unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	approval, ok := s.approvals[strings.TrimSpace(id)]
	if !ok {
		return DeploymentApproval{}, ErrDeploymentApprovalNotFound
	}
	if approval.OrganizationID != strings.TrimSpace(organizationID) || approval.BuilderID != strings.TrimSpace(builderID) || approval.Provider != strings.ToLower(strings.TrimSpace(provider)) {
		return DeploymentApproval{}, ErrDeploymentApprovalOrganization
	}
	if approval.Target != strings.TrimSpace(target) {
		return DeploymentApproval{}, errors.New("deployment target does not match approval")
	}
	if strings.TrimSpace(nonce) == "" || nonce != approval.Nonce {
		return DeploymentApproval{}, ErrDeploymentApprovalNonce
	}
	if approval.Status != DeploymentApprovalApproved {
		return DeploymentApproval{}, ErrDeploymentApprovalConflict
	}
	if time.Now().UTC().After(approval.ExpiresAt) {
		return DeploymentApproval{}, ErrDeploymentApprovalExpired
	}
	previous := approval
	approval.Status = DeploymentApprovalConsumed
	approval.ConsumedAt = time.Now().UTC()
	s.approvals[approval.ID] = approval
	if err := s.persistLocked(); err != nil {
		s.approvals[approval.ID] = previous
		return DeploymentApproval{}, err
	}
	return approval, nil
}

func (s *DeploymentApprovalStore) ListForOrganization(organizationID string) []DeploymentApproval {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]DeploymentApproval, 0)
	for _, approval := range s.approvals {
		if approval.OrganizationID == strings.TrimSpace(organizationID) {
			result = append(result, approval)
		}
	}
	return result
}

func (s *DeploymentApprovalStore) persistLocked() error {
	if s.root == "" {
		return nil
	}
	if err := writeJSONAtomic(filepath.Join(s.root, "approvals.json"), s.approvals); err != nil {
		return fmt.Errorf("persist deployment approvals: %w", err)
	}
	return nil
}
