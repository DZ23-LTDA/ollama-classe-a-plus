package secrets

import (
	"errors"
	"os"
	"testing"
)

func TestValidateRejectsBadInput(t *testing.T) {
	for _, tc := range []struct{ env, key string }{
		{"lower_case", "k"},
		{"A", "k"},
		{"OPENAI_API_KEY", ""},
		{"OPENAI_API_KEY", "two\nlines"},
		{"OPENAI_API_KEY", string(make([]byte, MaxKeyBytes+1))},
	} {
		if err := validate(tc.env, tc.key); !errors.Is(err, ErrInvalid) {
			t.Errorf("validate(%q, len %d) = %v, want ErrInvalid", tc.env, len(tc.key), err)
		}
	}
	if err := validate("OPENAI_API_KEY", "  sk-test  "); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}
}

func TestProtectDoesNotStoreKeyInClear(t *testing.T) {
	payload, err := protect("sk-unit-test-value")
	if err != nil {
		t.Fatal(err)
	}
	if fileExtension == ".dpapi" && string(payload) == "sk-unit-test-value" {
		t.Fatal("windows payload must be DPAPI-protected")
	}
}

func TestConfiguredReadsFileVariable(t *testing.T) {
	const env = "DZ23_SECRETS_TEST_KEY"
	t.Setenv(env, "")
	path := t.TempDir() + "/k" + fileExtension
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(env+"_FILE", path)
	if !Configured(env) {
		t.Fatal("expected configured via _FILE")
	}
	t.Setenv(env+"_FILE", path+".missing")
	if Configured(env) {
		t.Fatal("missing file must not count as configured")
	}
}
