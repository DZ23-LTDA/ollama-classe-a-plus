package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PushOutboxItem struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organization_id"`
	Title          string         `json:"title"`
	Body           string         `json:"body"`
	Data           map[string]any `json:"data,omitempty"`
	Attempts       int            `json:"attempts"`
	NextAttemptAt  time.Time      `json:"next_attempt_at"`
	LeaseUntil     *time.Time     `json:"lease_until,omitempty"`
	LastError      string         `json:"last_error,omitempty"`
}

type PushOutbox struct {
	mu    sync.Mutex
	path  string
	items map[string]PushOutboxItem
}

func NewPushOutbox(root string) (*PushOutbox, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("push outbox root is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	outbox := &PushOutbox{path: filepath.Join(root, "outbox.json"), items: map[string]PushOutboxItem{}}
	if err := readJSON(outbox.path, &outbox.items); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if outbox.items == nil {
		outbox.items = map[string]PushOutboxItem{}
	}
	return outbox, nil
}

func (o *PushOutbox) Enqueue(organizationID, title, body string, data map[string]any) (PushOutboxItem, error) {
	if o == nil {
		return PushOutboxItem{}, errors.New("push outbox is unavailable")
	}
	organizationID = strings.TrimSpace(organizationID)
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if organizationID == "" || title == "" || body == "" {
		return PushOutboxItem{}, errors.New("push outbox organization, title and body are required")
	}
	item := PushOutboxItem{
		ID:             "out_" + uuid.NewString(),
		OrganizationID: organizationID,
		Title:          title,
		Body:           body,
		Data:           clonePushData(data),
		NextAttemptAt:  time.Now().UTC(),
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	o.items[item.ID] = item
	if err := o.persistLocked(); err != nil {
		delete(o.items, item.ID)
		return PushOutboxItem{}, err
	}
	return item, nil
}

func (o *PushOutbox) ClaimDue(now time.Time) (PushOutboxItem, bool, error) {
	if o == nil {
		return PushOutboxItem{}, false, errors.New("push outbox is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for id, item := range o.items {
		if item.NextAttemptAt.After(now) || (item.LeaseUntil != nil && item.LeaseUntil.After(now)) {
			continue
		}
		previous := item
		leaseUntil := now.Add(time.Minute)
		item.Attempts++
		item.LeaseUntil = &leaseUntil
		o.items[id] = item
		if err := o.persistLocked(); err != nil {
			o.items[id] = previous
			return PushOutboxItem{}, false, err
		}
		return item, true, nil
	}
	return PushOutboxItem{}, false, nil
}

func (o *PushOutbox) Complete(id string) error {
	if o == nil {
		return errors.New("push outbox is unavailable")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	id = strings.TrimSpace(id)
	previous, ok := o.items[id]
	if !ok {
		return nil
	}
	delete(o.items, id)
	if err := o.persistLocked(); err != nil {
		o.items[id] = previous
		return err
	}
	return nil
}

func (o *PushOutbox) Fail(id string, cause error, now time.Time) error {
	if o == nil {
		return errors.New("push outbox is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	id = strings.TrimSpace(id)
	item, ok := o.items[id]
	if !ok {
		return errors.New("push outbox item not found")
	}
	previous := item
	item.LeaseUntil = nil
	item.LastError = RedactDLP(limitError(errorString(cause), 800))
	backoff := time.Duration(1<<minInt(item.Attempts, 6)) * time.Second
	item.NextAttemptAt = now.Add(backoff)
	o.items[id] = item
	if err := o.persistLocked(); err != nil {
		o.items[id] = previous
		return err
	}
	return nil
}

func (o *PushOutbox) List() []PushOutboxItem {
	if o == nil {
		return nil
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	items := make([]PushOutboxItem, 0, len(o.items))
	for _, item := range o.items {
		item.Data = clonePushData(item.Data)
		items = append(items, item)
	}
	return items
}

func (o *PushOutbox) persistLocked() error {
	return writeJSONAtomic(o.path, o.items)
}

func clonePushData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	result := make(map[string]any, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}

func errorString(err error) string {
	if err == nil {
		return "unknown push delivery error"
	}
	return fmt.Sprintf("%v", err)
}
