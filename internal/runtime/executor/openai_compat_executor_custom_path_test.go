package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v8/sdk/translator"
)

func TestOpenAICompatExecutorUsesPathPrefixedBaseURL(t *testing.T) {
	var gotPath, gotAuthorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	t.Cleanup(server.Close)

	executor := NewOpenAICompatExecutor("my-custom-gateway", &config.Config{})
	auth := &cliproxyauth.Auth{
		Provider: "openai-compatibility",
		Attributes: map[string]string{
			"base_url": server.URL + "/custom/v1",
			"api_key":  "test-api-key",
		},
	}
	_, errExecute := executor.Execute(context.Background(), auth, cliproxyexecutor.Request{
		Model:   "my-model",
		Payload: []byte(`{"model":"my-model","messages":[{"role":"user","content":"hello"}]}`),
	}, cliproxyexecutor.Options{SourceFormat: sdktranslator.FromString("openai")})
	if errExecute != nil {
		t.Fatalf("Execute() error = %v", errExecute)
	}
	if gotPath != "/custom/v1/chat/completions" {
		t.Fatalf("request path = %q, want /custom/v1/chat/completions", gotPath)
	}
	if gotAuthorization != "Bearer test-api-key" {
		t.Fatalf("Authorization = %q, want Bearer test-api-key", gotAuthorization)
	}
}
