//go:build windows || darwin

package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ollama/ollama/app/secrets"
	"github.com/ollama/ollama/internal/multillm"
)

// ProviderStatus is what the UI may know about a configured AI provider.
// It never contains a credential value.
type ProviderStatus struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	BaseURL    string `json:"base_url"`
	APIKeyEnv  string `json:"api_key_env,omitempty"`
	Models     int    `json:"models"`
	Enabled    bool   `json:"enabled"`
	Configured bool   `json:"configured"`
	NeedsKey   bool   `json:"needs_key"`
}

func providerConfigPath() string {
	return strings.TrimSpace(os.Getenv("OLLAMA_DZ23_CONFIG"))
}

func providerSecretsDir() (string, error) {
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "Ollama DZ23", "secrets"), nil
		}
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "Ollama DZ23", "secrets"), nil
}

func loadProviderConfig() (multillm.Config, error) {
	var cfg multillm.Config
	path := providerConfigPath()
	if path == "" {
		return cfg, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read provider config: %w", err)
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("decode provider config: %w", err)
	}
	return cfg, nil
}

func providerStatuses(cfg multillm.Config) []ProviderStatus {
	out := make([]ProviderStatus, 0, len(cfg.Providers))
	for _, p := range cfg.Providers {
		needsKey := p.APIKeyEnv != ""
		out = append(out, ProviderStatus{
			Name:       p.Name,
			Type:       p.Type,
			BaseURL:    p.BaseURL,
			APIKeyEnv:  p.APIKeyEnv,
			Models:     len(p.Models),
			Enabled:    p.Enabled == nil || *p.Enabled,
			Configured: !needsKey || secrets.Configured(p.APIKeyEnv),
			NeedsKey:   needsKey,
		})
	}
	return out
}

func (s *Server) listProviders(w http.ResponseWriter, r *http.Request) error {
	cfg, err := loadProviderConfig()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]any{
		"config_path": providerConfigPath(),
		"providers":   providerStatuses(cfg),
	})
}

// providerKeyEnv resolves the credential variable for a provider declared in
// the config; arbitrary variable names are never accepted from the client.
func providerKeyEnv(name string) (string, error) {
	cfg, err := loadProviderConfig()
	if err != nil {
		return "", err
	}
	for _, p := range cfg.Providers {
		if p.Name == name {
			if p.APIKeyEnv == "" {
				return "", fmt.Errorf("provider %q does not use an API key", name)
			}
			return p.APIKeyEnv, nil
		}
	}
	return "", fmt.Errorf("unknown provider %q", name)
}

func (s *Server) setProviderKey(w http.ResponseWriter, r *http.Request) error {
	envName, err := providerKeyEnv(r.PathValue("name"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
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
	if _, err := secrets.Save(dir, envName, body.Key); err != nil {
		if errors.Is(err, secrets.ErrInvalid) {
			w.WriteHeader(http.StatusBadRequest)
		}
		return err
	}
	s.log().Info("provider credential saved", "provider", r.PathValue("name"))
	s.restartForProviders()
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Server) removeProviderKey(w http.ResponseWriter, r *http.Request) error {
	envName, err := providerKeyEnv(r.PathValue("name"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return err
	}
	dir, err := providerSecretsDir()
	if err != nil {
		return err
	}
	if err := secrets.Remove(dir, envName); err != nil {
		return err
	}
	s.log().Info("provider credential removed", "provider", r.PathValue("name"))
	s.restartForProviders()
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// restartForProviders reloads the Ollama server so the router re-reads
// credentials; it inherits the updated <ENV>_FILE from this process.
func (s *Server) restartForProviders() {
	if s.Restart != nil {
		go s.Restart()
	}
}
