package newapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	newapi "github.com/xy200303/go-newapi-sdk/newapi"
)

func TestOperationDoSupportsPathParamsAndQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1beta/models/demo:generateContent" {
			t.Fatalf("path = %q, want %q", r.URL.Path, "/v1beta/models/demo:generateContent")
		}
		if r.URL.Query().Get("stream") != "true" {
			t.Fatalf("stream = %q, want %q", r.URL.Query().Get("stream"), "true")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client, err := newapi.New(server.URL, newapi.WithBearerToken("token"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var out map[string]bool
	err = client.AIModel.Chat.Gemini.GeminiRelayV1BetaPost.Do(t.Context(), &newapi.CallConfig{
		PathParams: map[string]any{"model": "demo"},
		Query: map[string]any{
			"stream": true,
		},
		JSONBody: map[string]any{"input": "hello"},
	}, &out)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if !out["ok"] {
		t.Fatal("response not decoded")
	}
}
