package grok

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestResponsesUsesBearerAndDecodesOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("request = %s %s auth=%q", r.Method, r.URL.Path, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_1","model":"grok-4","output_text":"hello","usage":{"total_tokens":7}}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "test-key", "grok-4")
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Responses(context.Background(), ResponseRequest{Input: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if response.ID != "resp_1" || response.OutputText != "hello" || response.Usage.TotalTokens != 7 {
		t.Fatalf("response = %+v", response)
	}
}

func TestStreamResponsesParsesSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprint(w, "data: {\"delta\":\"hello \"}\n\ndata: {\"text\":\"world\"}\n\ndata: [DONE]\n\n")
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "key", "grok-4")
	if err != nil {
		t.Fatal(err)
	}
	var parts []string
	if err := client.StreamResponses(context.Background(), ResponseRequest{Input: "hi"}, func(delta string) error {
		parts = append(parts, delta)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(parts, ""); got != "hello world" {
		t.Fatalf("stream = %q", got)
	}
}

func TestResponsesRetriesTransientFailures(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 2 {
			http.Error(w, "temporary", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_ok","output_text":"ok"}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "key", "grok-4")
	if err != nil {
		t.Fatal(err)
	}
	client.Backoff = 1
	client.MaxRetries = 2
	response, err := client.Responses(context.Background(), ResponseRequest{Input: "hi"})
	if err != nil || response.ID != "resp_ok" || calls.Load() != 2 {
		t.Fatalf("response=%+v err=%v calls=%d", response, err, calls.Load())
	}
}

func TestLiveModeSeparatesMemoryAndLiveEvidence(t *testing.T) {
	report, err := NewLiveMode().Build(LiveQuery{Query: "latest", RequireHTTPS: true, Memory: []Evidence{{URL: "https://memory.example", Title: "old", Excerpt: "memory"}}, Live: []Evidence{{URL: "https://news.example", Title: "live", Excerpt: "new", Kind: "live_web"}}})
	if err != nil {
		t.Fatal(err)
	}
	if report.MemoryEvidence != 1 || report.LiveEvidence != 1 || len(report.Citations) != 2 || report.Confidence != 0.75 {
		t.Fatalf("report = %+v", report)
	}
}

func TestLiveModeRejectsInsecureEvidence(t *testing.T) {
	if _, err := NewLiveMode().Build(LiveQuery{Query: "x", RequireHTTPS: true, Memory: []Evidence{{URL: "http://example.test"}}}); err == nil {
		t.Fatal("expected insecure evidence rejection")
	}
}
