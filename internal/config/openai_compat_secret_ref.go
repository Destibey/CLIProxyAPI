package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func (cfg *Config) resolveOpenAICompatibilityAPIKeys() error {
	for providerIndex := range cfg.OpenAICompatibility {
		for keyIndex := range cfg.OpenAICompatibility[providerIndex].APIKeyEntries {
			entry := &cfg.OpenAICompatibility[providerIndex].APIKeyEntries[keyIndex]
			if entry.APIKeyRef == nil {
				continue
			}
			if strings.TrimSpace(entry.APIKey) != "" {
				return fmt.Errorf("openai-compatibility[%d].api-key and api-key-ref are mutually exclusive", providerIndex)
			}
			value, err := resolveAPIKeyReference(*entry.APIKeyRef)
			if err != nil {
				return fmt.Errorf("openai-compatibility[%d].api-key-ref: %w", providerIndex, err)
			}
			entry.APIKey = value
		}
	}
	return nil
}

func resolveAPIKeyReference(ref APIKeyReference) (string, error) {
	env := strings.TrimSpace(ref.Env)
	path := strings.TrimSpace(ref.Path)
	field := strings.TrimSpace(ref.Field)
	if env != "" && (path != "" || field != "") {
		return "", errors.New("env cannot be combined with path or field")
	}
	if env != "" {
		value, ok := os.LookupEnv(env)
		if !ok || strings.TrimSpace(value) == "" {
			return "", errors.New("environment variable is unset")
		}
		return strings.TrimSpace(value), nil
	}
	if path == "" || field == "" {
		return "", errors.New("path and field are required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", errors.New("secret file unavailable")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return "", errors.New("secret file permissions must exclude group and other access")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("secret file unreadable")
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return "", errors.New("secret file is not JSON")
	}
	raw, ok := document[field]
	if !ok {
		return "", errors.New("secret field is missing")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || strings.TrimSpace(value) == "" {
		return "", errors.New("secret field is not a non-empty string")
	}
	return strings.TrimSpace(value), nil
}
