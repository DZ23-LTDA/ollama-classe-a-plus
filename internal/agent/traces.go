package agent

import (
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TraceSpan struct {
	TraceID        string         `json:"trace_id"`
	OrganizationID string         `json:"organization_id,omitempty"`
	SpanID         string         `json:"span_id"`
	ParentID       string         `json:"parent_id,omitempty"`
	Name           string         `json:"name"`
	Status         string         `json:"status"`
	StartAt        time.Time      `json:"start_at"`
	EndAt          *time.Time     `json:"end_at,omitempty"`
	Duration       time.Duration  `json:"duration_ns,omitempty"`
	Attributes     map[string]any `json:"attributes,omitempty"`
	Error          string         `json:"error,omitempty"`
}

type TraceStore struct {
	mu    sync.RWMutex
	root  string
	spans []TraceSpan
}

func NewTraceStore(root string) (*TraceStore, error) {
	store := &TraceStore{root: root, spans: []TraceSpan{}}
	if root == "" {
		return store, nil
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	path := filepath.Join(root, "spans.json")
	if _, err := os.Stat(path); err == nil {
		if err := readJSON(path, &store.spans); err != nil {
			return nil, err
		}
	}
	return store, nil
}

type SpanHandle struct {
	store *TraceStore
	index int
	ended bool
}

func (h *SpanHandle) ID() string {
	if h == nil || h.store == nil {
		return ""
	}
	h.store.mu.RLock()
	defer h.store.mu.RUnlock()
	if h.index < 0 || h.index >= len(h.store.spans) {
		return ""
	}
	return h.store.spans[h.index].SpanID
}

func (s *TraceStore) Start(traceID, parentID, name string, attributes map[string]any) *SpanHandle {
	return s.StartForOrganization("local", traceID, parentID, name, attributes)
}

func (s *TraceStore) StartForOrganization(organizationID, traceID, parentID, name string, attributes map[string]any) *SpanHandle {
	if traceID == "" {
		traceID = "tr_" + uuid.NewString()
	}
	redactedAttributes, _ := RedactValue(attributes).(map[string]any)
	span := TraceSpan{TraceID: traceID, OrganizationID: normalizedOrganizationID(organizationID), SpanID: "sp_" + uuid.NewString(), ParentID: parentID, Name: RedactDLP(name), Status: "running", StartAt: time.Now().UTC(), Attributes: redactedAttributes}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spans = append(s.spans, span)
	return &SpanHandle{store: s, index: len(s.spans) - 1}
}

func (h *SpanHandle) End(status string, runErr error) {
	if h == nil || h.store == nil || h.ended {
		return
	}
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	if h.ended || h.index < 0 || h.index >= len(h.store.spans) {
		return
	}
	now := time.Now().UTC()
	span := &h.store.spans[h.index]
	previous := *span
	span.Status = status
	span.EndAt = &now
	span.Duration = now.Sub(span.StartAt)
	if runErr != nil {
		span.Status = "error"
		span.Error = RedactDLP(limitError(runErr.Error(), 2000))
	}
	if err := h.store.persistLocked(); err != nil {
		*span = previous
		slog.Error("agent trace persistence failed", "trace_id", previous.TraceID, "span_id", previous.SpanID, "error", err)
		return
	}
	h.ended = true
}

func (s *TraceStore) List(traceID string, limit int) []TraceSpan {
	return s.ListForOrganization("local", traceID, limit)
}

func (s *TraceStore) ListForOrganization(organizationID, traceID string, limit int) []TraceSpan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	result := make([]TraceSpan, 0, len(s.spans))
	for _, span := range s.spans {
		if normalizedOrganizationID(span.OrganizationID) == normalizedOrganizationID(organizationID) && (traceID == "" || span.TraceID == traceID) {
			result = append(result, span)
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].StartAt.Before(result[j].StartAt) })
	if len(result) > limit {
		result = result[len(result)-limit:]
	}
	return result
}

func (s *TraceStore) persistLocked() error {
	if s.root == "" {
		return nil
	}
	return writeJSONAtomic(filepath.Join(s.root, "spans.json"), s.spans)
}
