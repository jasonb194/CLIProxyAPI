package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

func TestManagerExecute_DefaultFillFirstFallsThroughProvidersAfterQuota(t *testing.T) {
	withQuotaCooldownEnabled(t)

	model := t.Name() + "-model"
	manager := NewManager(nil, nil, nil)
	manager.SetRetryConfig(0, 0, 0)
	attempts := make([]string, 0, 3)
	newExecutor := func(provider string) *mockCustomErrorExecutor {
		return &mockCustomErrorExecutor{
			identifier: provider,
			executeFn: func(_ context.Context, auth *Auth, req cliproxyexecutor.Request, _ cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
				if req.Model != model {
					t.Errorf("request model = %q, want %q", req.Model, model)
				}
				attempts = append(attempts, provider+":"+auth.ID)
				if provider == "gemini" {
					return cliproxyexecutor.Response{}, &Error{HTTPStatus: http.StatusTooManyRequests, Message: "quota exhausted"}
				}
				return cliproxyexecutor.Response{Payload: []byte(`{"ok":true}`)}, nil
			},
		}
	}
	manager.RegisterExecutor(newExecutor("gemini"))
	manager.RegisterExecutor(newExecutor("claude"))

	geminiA := t.Name() + "-gemini-a"
	geminiB := t.Name() + "-gemini-b"
	claudeA := t.Name() + "-claude-a"
	auths := []*Auth{
		{ID: geminiA, Provider: "gemini", Status: StatusActive},
		{ID: geminiB, Provider: "gemini", Status: StatusActive},
		{ID: claudeA, Provider: "claude", Status: StatusActive},
	}
	registryRef := registry.GetGlobalRegistry()
	for _, auth := range auths {
		registryRef.RegisterClient(auth.ID, auth.Provider, []*registry.ModelInfo{{ID: model}})
		t.Cleanup(func() { registryRef.UnregisterClient(auth.ID) })
		if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
			t.Fatalf("Register(%s) error = %v", auth.ID, errRegister)
		}
	}

	response, errExecute := manager.Execute(context.Background(), []string{"gemini", "claude"}, cliproxyexecutor.Request{
		Model:   model,
		Payload: []byte(`{"model":"` + model + `","messages":[{"role":"user","content":"hello"}]}`),
	}, cliproxyexecutor.Options{})
	if errExecute != nil {
		t.Fatalf("Execute() error = %v", errExecute)
	}
	if string(response.Payload) != `{"ok":true}` {
		t.Fatalf("Execute() payload = %s, want successful Claude response", response.Payload)
	}
	wantAttempts := []string{"gemini:" + geminiA, "gemini:" + geminiB, "claude:" + claudeA}
	if len(attempts) != len(wantAttempts) {
		t.Fatalf("attempts = %v, want %v", attempts, wantAttempts)
	}
	for index, want := range wantAttempts {
		if attempts[index] != want {
			t.Fatalf("attempt %d = %q, want %q (all attempts %v)", index, attempts[index], want, attempts)
		}
	}
	for _, authID := range []string{geminiA, geminiB} {
		cooled, ok := manager.GetByID(authID)
		if !ok || cooled == nil {
			t.Fatalf("auth %q is missing after quota response", authID)
		}
		if state := cooled.ModelStates[model]; state == nil || !state.Unavailable {
			t.Fatalf("auth %q model state = %+v, want unavailable after quota response", authID, state)
		}
	}
}
