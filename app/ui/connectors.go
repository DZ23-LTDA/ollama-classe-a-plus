//go:build windows || darwin

package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ollama/ollama/app/secrets"
	"github.com/ollama/ollama/envconfig"
	"github.com/ollama/ollama/internal/agent"
)

// errAgentLoginRequired mirrors the agent API's answer when the server
// listens beyond loopback and demands a bearer token.
var errAgentLoginRequired = errors.New(`o Ollama está exposto na rede, então conectar exige login no Workspace; desative "Expose Ollama to the network" em Configurações para conectar localmente`)

func quickConnectEntry(id string) (agent.ConnectorCatalogEntry, bool) {
	for _, entry := range agent.ConnectorCatalog() {
		if entry.ID == id && entry.APIBaseURL != "" {
			return entry, true
		}
	}
	return agent.ConnectorCatalogEntry{}, false
}

// callAgentAPI sends a request to the local Ollama server's agent API.
func callAgentAPI(ctx context.Context, method, path string, body any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(envconfig.ConnectableHost().String(), "/")+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return errAgentLoginRequired
	}
	if resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		var parsed struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(detail, &parsed) == nil && parsed.Error != "" {
			return fmt.Errorf("agent API %d: %s", resp.StatusCode, parsed.Error)
		}
		return fmt.Errorf("agent API %d: %s", resp.StatusCode, strings.TrimSpace(string(detail)))
	}
	return nil
}

// connectConnector saves a pasted API key for a catalog service and
// registers the connector in the agent runtime in one step.
func (s *Server) connectConnector(w http.ResponseWriter, r *http.Request) error {
	entry, ok := quickConnectEntry(r.PathValue("id"))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("connector %q cannot be connected with an API key", r.PathValue("id"))
	}
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, secrets.MaxKeyBytes+1024)).Decode(&body); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	dir, err := providerSecretsDir()
	if err != nil {
		return err
	}
	envName := agent.ConnectorTokenEnv(entry.ID)
	if _, err := secrets.Save(dir, envName, body.Key); err != nil {
		if errors.Is(err, secrets.ErrInvalid) {
			w.WriteHeader(http.StatusBadRequest)
		}
		return err
	}
	err = callAgentAPI(r.Context(), http.MethodPost, "/api/agent/v1/connectors", agent.ConnectorConfig{
		ID:       entry.ID,
		Provider: entry.ID,
		BaseURL:  entry.APIBaseURL,
		TokenEnv: envName,
		Operations: []agent.ConnectorOperation{
			{Name: "read", Methods: []string{http.MethodGet}, PathPrefixes: []string{"/"}},
			{Name: "write", Methods: []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}, PathPrefixes: []string{"/"}},
		},
	})
	if err != nil {
		if errors.Is(err, errAgentLoginRequired) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusBadGateway)
		}
		return err
	}
	s.log().Info("connector connected", "connector", entry.ID)
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// disconnectConnector removes the connector registration and its key.
func (s *Server) disconnectConnector(w http.ResponseWriter, r *http.Request) error {
	entry, ok := quickConnectEntry(r.PathValue("id"))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return fmt.Errorf("unknown connector %q", r.PathValue("id"))
	}
	if err := callAgentAPI(r.Context(), http.MethodDelete, "/api/agent/v1/connectors/"+entry.ID, nil); err != nil && !strings.Contains(err.Error(), "404") {
		if errors.Is(err, errAgentLoginRequired) {
			w.WriteHeader(http.StatusForbidden)
		}
		return err
	}
	dir, err := providerSecretsDir()
	if err != nil {
		return err
	}
	if err := secrets.Remove(dir, agent.ConnectorTokenEnv(entry.ID)); err != nil {
		return err
	}
	s.log().Info("connector disconnected", "connector", entry.ID)
	w.WriteHeader(http.StatusNoContent)
	return nil
}
