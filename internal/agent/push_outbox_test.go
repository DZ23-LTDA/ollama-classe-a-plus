package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestPushOutboxPersistsRetryAndCompletion(t *testing.T) {
	root := filepath.Join(t.TempDir(), "outbox")
	outbox, err := NewPushOutbox(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(time.Second)
	item, err := outbox.Enqueue("org_1", "title", "body", map[string]any{"mission_id": "mis_1"})
	if err != nil {
		t.Fatal(err)
	}
	claimed, ok, err := outbox.ClaimDue(now)
	if err != nil || !ok || claimed.ID != item.ID || claimed.Attempts != 1 {
		t.Fatalf("claim=%+v ok=%v err=%v", claimed, ok, err)
	}
	if err := outbox.Fail(item.ID, context.Canceled, now); err != nil {
		t.Fatal(err)
	}
	restarted, err := NewPushOutbox(root)
	if err != nil {
		t.Fatal(err)
	}
	pending := restarted.List()
	if len(pending) != 1 || pending[0].Attempts != 1 || pending[0].LastError == "" {
		t.Fatalf("pending after restart=%+v", pending)
	}
	claimed, ok, err = restarted.ClaimDue(pending[0].NextAttemptAt.Add(time.Second))
	if err != nil || !ok || claimed.Attempts != 2 {
		t.Fatalf("retry claim=%+v ok=%v err=%v", claimed, ok, err)
	}
	if err := restarted.Complete(item.ID); err != nil {
		t.Fatal(err)
	}
	if len(restarted.List()) != 0 {
		t.Fatalf("outbox still contains completed item: %+v", restarted.List())
	}
}

func TestRuntimeFlushPushOutboxDeliversAndRemovesItem(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	push, err := NewPushService(t.TempDir(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	push.client = server.Client()
	if _, err := push.Register("ExponentPushToken[test]", "android", "user_1", "org_1"); err != nil {
		t.Fatal(err)
	}
	outbox, err := NewPushOutbox(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := outbox.Enqueue("org_1", "title", "body", nil); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{push: push, pushOutbox: outbox, metrics: &RuntimeMetrics{}}
	runtime.flushPushOutbox(context.Background())
	if requests.Load() != 1 || len(outbox.List()) != 0 {
		t.Fatalf("requests=%d outbox=%+v", requests.Load(), outbox.List())
	}
	if metrics := runtime.Metrics(); metrics.PushDeliveryFailures != 0 || metrics.PushOutboxFailures != 0 {
		t.Fatalf("metrics=%+v", metrics)
	}
}

func TestRuntimeFlushPushOutboxRetainsFailedDelivery(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("provider unavailable"))
	}))
	defer server.Close()
	push, err := NewPushService(t.TempDir(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	push.client = server.Client()
	if _, err := push.Register("ExponentPushToken[test]", "ios", "user_1", "org_1"); err != nil {
		t.Fatal(err)
	}
	outbox, err := NewPushOutbox(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := outbox.Enqueue("org_1", "title", "body", nil); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{push: push, pushOutbox: outbox, metrics: &RuntimeMetrics{}}
	runtime.flushPushOutbox(context.Background())
	pending := outbox.List()
	if len(pending) != 1 || pending[0].LastError == "" || pending[0].LeaseUntil != nil {
		t.Fatalf("pending=%+v", pending)
	}
	if metrics := runtime.Metrics(); metrics.PushDeliveryFailures != 1 || metrics.PushOutboxFailures != 0 {
		t.Fatalf("metrics=%+v", metrics)
	}
}
