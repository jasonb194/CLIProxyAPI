package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAPIKeyModelIsCompatConfigDecoding(t *testing.T) {
	const yamlConfig = `gemini-api-key:
  - models:
      - name: gemini-upstream
        alias: gemini-alias
        is-compat: true
      - name: gemini-native
        alias: gemini-native
interactions-api-key:
  - models:
      - name: interactions-upstream
        alias: interactions-alias
        is-compat: true
xai-api-key:
  - models:
      - name: xai-upstream
        alias: xai-alias
        is-compat: true
claude-api-key:
  - models:
      - name: claude-upstream
        alias: claude-alias
        is-compat: true
codex-api-key:
  - models:
      - name: codex-upstream
        alias: codex-alias
        is-compat: true
openai-compatibility:
  - name: deepseek
    models:
      - name: deepseek-upstream
        alias: deepseek-alias
        is-compat: true
      - name: openai-native
        alias: openai-native
`

	var cfg Config
	if errDecode := yaml.Unmarshal([]byte(yamlConfig), &cfg); errDecode != nil {
		t.Fatalf("decode error: %v", errDecode)
	}

	if len(cfg.GeminiKey) != 1 || !cfg.GeminiKey[0].Models[0].IsCompat {
		t.Fatalf("gemini-api-key IsCompat = %+v, want true", cfg.GeminiKey)
	}
	if cfg.GeminiKey[0].Models[1].IsCompat {
		t.Fatal("gemini-api-key omitted IsCompat = true, want default false")
	}
	if len(cfg.InteractionsKey) != 1 || !cfg.InteractionsKey[0].Models[0].IsCompat {
		t.Fatalf("interactions-api-key IsCompat = %+v, want true", cfg.InteractionsKey)
	}
	if len(cfg.XAIKey) != 1 || !cfg.XAIKey[0].Models[0].IsCompat {
		t.Fatalf("xai-api-key IsCompat = %+v, want true", cfg.XAIKey)
	}
	if len(cfg.ClaudeKey) != 1 || !cfg.ClaudeKey[0].Models[0].IsCompat {
		t.Fatalf("claude-api-key IsCompat = %+v, want true", cfg.ClaudeKey)
	}
	if len(cfg.CodexKey) != 1 || !cfg.CodexKey[0].Models[0].IsCompat {
		t.Fatalf("codex-api-key IsCompat = %+v, want true", cfg.CodexKey)
	}
	if len(cfg.OpenAICompatibility) != 1 || !cfg.OpenAICompatibility[0].Models[0].IsCompat {
		t.Fatalf("openai-compatibility IsCompat = %+v, want true", cfg.OpenAICompatibility)
	}
	if cfg.OpenAICompatibility[0].Models[1].IsCompat {
		t.Fatal("openai-compatibility omitted IsCompat = true, want default false")
	}
}

func TestOpenAICompatibilityConfigDecodesCustomPathAndAPIKey(t *testing.T) {
	cfg, errParse := ParseConfigBytes([]byte(`openai-compatibility:
  - name: "my-custom-gateway"
    base-url: "https://gateway.example.com/custom/v1"
    api-key-entries:
      - api-key: "test-api-key"
    models:
      - name: "upstream-model"
        alias: "my-model"
`))
	if errParse != nil {
		t.Fatalf("ParseConfigBytes() error = %v", errParse)
	}
	if len(cfg.OpenAICompatibility) != 1 {
		t.Fatalf("OpenAICompatibility entries = %d, want 1", len(cfg.OpenAICompatibility))
	}
	provider := cfg.OpenAICompatibility[0]
	if provider.Name != "my-custom-gateway" {
		t.Fatalf("provider name = %q, want my-custom-gateway", provider.Name)
	}
	if provider.BaseURL != "https://gateway.example.com/custom/v1" {
		t.Fatalf("provider base-url = %q, want https://gateway.example.com/custom/v1", provider.BaseURL)
	}
	if len(provider.APIKeyEntries) != 1 || provider.APIKeyEntries[0].APIKey != "test-api-key" {
		t.Fatalf("provider API key entries = %+v, want one test API key", provider.APIKeyEntries)
	}
}
