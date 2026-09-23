package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

type plannerChatStub struct {
	model    string
	response string
	err      error
}

func (s *plannerChatStub) Chat(_ context.Context, request *api.ChatRequest, callback api.ChatResponseFunc) error {
	s.model = request.Model
	if s.err != nil {
		return s.err
	}
	return callback(api.ChatResponse{Message: api.Message{Content: s.response}})
}

func TestOllamaPlannerUsesMissionSelectedModel(t *testing.T) {
	stub := &plannerChatStub{response: `{"steps":[{"kind":"workspace.read","title":"inspect","risk":"read","input":{"path":"."}}]}`}
	planner := OllamaPlanner{Client: stub, Model: "default-model"}
	steps, err := planner.Plan(context.Background(), Mission{Objective: "inspect", Model: "selected-model"})
	if err != nil {
		t.Fatal(err)
	}
	if stub.model != "selected-model" {
		t.Fatalf("request model = %q, want selected-model", stub.model)
	}
	if len(steps) != 1 || steps[0].Kind != "workspace.read" {
		t.Fatalf("steps = %+v", steps)
	}
}

func TestOllamaPlannerSurfacesProviderFailure(t *testing.T) {
	planner := OllamaPlanner{
		Client: &plannerChatStub{err: errors.New("provider unavailable")},
		Model:  "selected-model",
	}
	steps, err := planner.Plan(context.Background(), Mission{Objective: "inspect"})
	if err == nil || !strings.Contains(err.Error(), "planner provider request failed") {
		t.Fatalf("err = %v, want explicit provider failure", err)
	}
	if steps != nil {
		t.Fatalf("steps = %+v, want nil on provider failure", steps)
	}
}

func TestOllamaPlannerSurfacesInvalidPlan(t *testing.T) {
	planner := OllamaPlanner{
		Client: &plannerChatStub{response: `{}`},
		Model:  "selected-model",
	}
	steps, err := planner.Plan(context.Background(), Mission{Objective: "inspect"})
	if err == nil || !strings.Contains(err.Error(), "planner returned invalid plan") {
		t.Fatalf("err = %v, want explicit invalid plan", err)
	}
	if steps != nil {
		t.Fatalf("steps = %+v, want nil on invalid plan", steps)
	}
}
