package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrWebhookReplay = errors.New("webhook idempotency key was already consumed")

type WebhookReplayStore struct {
	mu     sync.Mutex
	root   string
	claims map[string]time.Time
}

func NewWebhookReplayStore(root string) (*WebhookReplayStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("webhook replay store root is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	store := &WebhookReplayStore{root: root, claims: map[string]time.Time{}}
	data, err := os.ReadFile(filepath.Join(root, "claims.json"))
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &store.claims); err != nil {
		return nil, err
	}
	if store.claims == nil {
		store.claims = map[string]time.Time{}
	}
	return store, nil
}

func (s *WebhookReplayStore) Claim(scheduleID, idempotencyKey string) error {
	if s == nil {
		return errors.New("webhook replay store is unavailable")
	}
	scheduleID = strings.TrimSpace(scheduleID)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if scheduleID == "" || idempotencyKey == "" {
		return errors.New("schedule id and idempotency key are required")
	}
	key := scheduleID + "\x00" + idempotencyKey
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.claims[key]; exists {
		return ErrWebhookReplay
	}
	previous := s.claims[key]
	s.claims[key] = time.Now().UTC()
	if err := s.persistLocked(); err != nil {
		if previous.IsZero() {
			delete(s.claims, key)
		} else {
			s.claims[key] = previous
		}
		return fmt.Errorf("persist webhook idempotency key: %w", err)
	}
	return nil
}

func (s *WebhookReplayStore) persistLocked() error {
	return writeJSONAtomic(filepath.Join(s.root, "claims.json"), s.claims)
}
