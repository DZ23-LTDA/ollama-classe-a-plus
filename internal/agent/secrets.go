package agent

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type SecretMetadata struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SecretStore struct {
	mu     sync.RWMutex
	root   string
	values map[string]SecretValue
}

type SecretValue struct {
	Ciphertext string    `json:"ciphertext"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func NewSecretStore(root string) (*SecretStore, error) {
	store := &SecretStore{root: strings.TrimSpace(root), values: map[string]SecretValue{}}
	if store.root != "" {
		if err := os.MkdirAll(store.root, 0o700); err != nil {
			return nil, err
		}
		if err := readJSON(filepathJoin(store.root, "secrets.json"), &store.values); err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return store, nil
}

func (s *SecretStore) Set(name, value string) error {
	name = normalizeSecretName(name)
	if name == "" || strings.TrimSpace(value) == "" {
		return errors.New("secret name and value are required")
	}
	ciphertext, err := encryptCredential(value)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.values[name] = SecretValue{Ciphertext: ciphertext, UpdatedAt: time.Now().UTC()}
	err = s.persistLocked()
	s.mu.Unlock()
	return err
}

func (s *SecretStore) Get(name string) (string, error) {
	name = normalizeSecretName(name)
	s.mu.RLock()
	value, ok := s.values[name]
	s.mu.RUnlock()
	if !ok {
		return "", os.ErrNotExist
	}
	return decryptCredential(value.Ciphertext)
}

func (s *SecretStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = normalizeSecretName(name)
	if _, ok := s.values[name]; !ok {
		return os.ErrNotExist
	}
	delete(s.values, name)
	return s.persistLocked()
}

func (s *SecretStore) List() []SecretMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]SecretMetadata, 0, len(s.values))
	for name, value := range s.values {
		items = append(items, SecretMetadata{Name: name, UpdatedAt: value.UpdatedAt})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func (s *SecretStore) persistLocked() error {
	if s.root == "" {
		return nil
	}
	return writeJSONAtomic(filepathJoin(s.root, "secrets.json"), s.values)
}

func normalizeSecretName(name string) string {
	name = strings.TrimSpace(name)
	for _, character := range name {
		if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '_' && character != '-' && character != '.' {
			return ""
		}
	}
	return name
}

type DLPFinding struct {
	Kind     string `json:"kind"`
	Redacted string `json:"redacted"`
}

var dlpPatterns = []struct {
	kind    string
	pattern *regexp.Regexp
}{
	{"private_key", regexp.MustCompile(`(?s)-----BEGIN [A-Z ]+PRIVATE KEY-----.*?-----END [A-Z ]+PRIVATE KEY-----`)},
	{"github_token", regexp.MustCompile(`(?i)\b(?:ghp|github_pat)_[A-Za-z0-9_]{20,}\b`)},
	{"openrouter_token", regexp.MustCompile(`\bsk-or-v1-[A-Za-z0-9_-]{20,}\b`)},
	{"openai_token", regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{20,}\b`)},
	{"xai_token", regexp.MustCompile(`\bxai-[A-Za-z0-9_-]{20,}\b`)},
	{"aws_access_key", regexp.MustCompile(`\bAKIA[A-Z0-9]{16}\b`)},
	{"slack_token", regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{16,}\b`)},
	{"bearer_token", regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]{16,}`)},
	{"credential_assignment", regexp.MustCompile(`(?i)(?:password|passwd|secret|api[_-]?key|access[_-]?token|refresh[_-]?token)\s*[:=]\s*["']?[A-Za-z0-9._~+/=-]{8,}["']?`)},
}

func ScanDLP(text string) []DLPFinding {
	findings := []DLPFinding{}
	for _, item := range dlpPatterns {
		if match := item.pattern.FindString(text); match != "" {
			findings = append(findings, DLPFinding{Kind: item.kind, Redacted: item.pattern.ReplaceAllString(match, "[REDACTED]")})
		}
	}
	return findings
}

func RedactDLP(text string) string {
	for _, item := range dlpPatterns {
		text = item.pattern.ReplaceAllString(text, "[REDACTED]")
	}
	return text
}

// RedactValue recursively removes credential-shaped strings from values that
// are about to be persisted, emitted as events, or serialized externally.
// JSON-compatible values are normalized to JSON-compatible maps/slices so
// nested tool responses cannot bypass the string redaction path.
func RedactValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case string:
		return RedactDLP(typed)
	case []byte:
		return RedactDLP(string(typed))
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			if sensitiveDLPKey(key) {
				result[key] = "[REDACTED]"
				continue
			}
			result[key] = RedactValue(item)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, item := range typed {
			result[index] = RedactValue(item)
		}
		return result
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return value
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return value
		}
		var normalized any
		if err := json.Unmarshal(encoded, &normalized); err != nil {
			return value
		}
		return RedactValue(normalized)
	}
}

func sensitiveDLPKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(key, "token") || strings.Contains(key, "secret") || strings.Contains(key, "password") || strings.Contains(key, "passwd") || strings.Contains(key, "api_key") || strings.Contains(key, "access_key") || strings.Contains(key, "private_key")
}
