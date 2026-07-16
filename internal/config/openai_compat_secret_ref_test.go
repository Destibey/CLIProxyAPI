package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseConfigBytesResolvesOpenAICompatibilitySecretRefWithoutSerializingSecret(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(secretPath, []byte(`{"OPENAI_API_KEY":"synthetic-secret"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := ParseConfigBytes([]byte(`openai-compatibility:
  - name: relay
    base-url: https://relay.test/v1
    api-key-entries:
      - api-key-ref:
          path: ` + secretPath + `
          field: OPENAI_API_KEY
`))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.OpenAICompatibility[0].APIKeyEntries[0].APIKey; got != "synthetic-secret" {
		t.Fatalf("resolved key = %q", got)
	}
	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil || strings.Contains(string(yamlBytes), "synthetic-secret") || !strings.Contains(string(yamlBytes), "api-key-ref:") {
		t.Fatalf("serialized YAML leaked or lost ref: err=%v yaml=%s", err, yamlBytes)
	}
	jsonBytes, err := json.Marshal(cfg.OpenAICompatibility[0].APIKeyEntries[0])
	if err != nil || strings.Contains(string(jsonBytes), "synthetic-secret") || !strings.Contains(string(jsonBytes), "api-key-ref") {
		t.Fatalf("serialized JSON leaked or lost ref: err=%v json=%s", err, jsonBytes)
	}
}

func TestParseConfigBytesRejectsWeakOpenAICompatibilitySecretFile(t *testing.T) {
	secretPath := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(secretPath, []byte(`{"OPENAI_API_KEY":"synthetic-secret"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(secretPath, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ParseConfigBytes([]byte(`openai-compatibility:
  - name: relay
    base-url: https://relay.test/v1
    api-key-entries:
      - api-key-ref:
          path: ` + secretPath + `
          field: OPENAI_API_KEY
`))
	if err == nil || strings.Contains(err.Error(), "synthetic-secret") {
		t.Fatalf("weak secret file err=%v", err)
	}
}
