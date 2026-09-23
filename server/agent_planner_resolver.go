package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
	"github.com/ollama/ollama/internal/agent"
	"github.com/ollama/ollama/internal/multillm"
)

type multiProviderPlannerResolver struct {
	registry *multillm.Registry
	client   *api.Client
}

func (r multiProviderPlannerResolver) ResolvePlanner(provider, model string) (agent.Planner, error) {
	provider = strings.TrimSpace(provider)
	model = strings.TrimSpace(model)
	if provider == "" {
		return nil, errors.New("planner provider is required")
	}
	if provider == "ollama-local" {
		if model == "" {
			return nil, errors.New("local provider requires an explicit model")
		}
		client := r.client
		if client == nil {
			client = api.NewClient(envconfig.ConnectableHost(), http.DefaultClient)
		}
		return agent.OllamaPlanner{Client: client, Model: model}, nil
	}
	if r.registry == nil {
		return nil, fmt.Errorf("provider %q is not configured", provider)
	}
	if model == "" {
		return nil, fmt.Errorf("provider %q requires an explicit model", provider)
	}
	modelID := model
	if !strings.Contains(modelID, "/") {
		modelID = provider + "/" + modelID
	}
	resolved, ok := r.registry.Resolve(modelID, multillm.Policy{Path: "/api/chat"})
	if !ok || resolved.Provider != provider {
		return nil, fmt.Errorf("provider model %q is unavailable or does not support /api/chat", modelID)
	}
	client := r.client
	if client == nil {
		client = api.NewClient(envconfig.ConnectableHost(), http.DefaultClient)
	}
	return agent.OllamaPlanner{Client: client, Model: resolved.ID}, nil
}

var _ agent.PlannerResolver = multiProviderPlannerResolver{}
