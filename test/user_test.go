package newapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	newapi "xy200303/go-newapi-sdk/newapi"
)

func TestUserCreateTokenContextFallsBackToTokenList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/token/":
			w.Header().Set("Content-Type", "application/json")
			if r.Method == http.MethodPost {
				_, _ = w.Write([]byte(`{"success":true,"message":"","data":{}}`))
				return
			}
			_, _ = w.Write([]byte(`{"success":true,"message":"","data":{"items":[{"id":7,"name":"demo","key":"abc123","unlimited_quota":true}],"total":1,"page":1,"page_size":100}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := newapi.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	token, err := client.UserCreateTokenContext(t.Context(), "user-token", 1, newapi.CreateTokenRequest{
		Name:           "demo",
		RemainQuota:    0,
		UnlimitedQuota: true,
	})
	if err != nil {
		t.Fatalf("UserCreateTokenContext() error = %v", err)
	}

	if token == nil {
		t.Fatal("UserCreateTokenContext() token = nil")
	}
	if token.ID != 7 {
		t.Fatalf("token.ID = %d, want %d", token.ID, 7)
	}
	if token.Key != "sk-abc123" {
		t.Fatalf("token.Key = %q, want %q", token.Key, "sk-abc123")
	}
}

func TestUserListModelsContextSupportsGroupedPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/models" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"","data":{"default":["gpt-4o","gpt-4.1"],"vision":["gpt-4o"]}}`))
	}))
	defer server.Close()

	client, err := newapi.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	models, err := client.UserListModelsContext(t.Context(), "user-token", 1)
	if err != nil {
		t.Fatalf("UserListModelsContext() error = %v", err)
	}

	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want %d", len(models), 2)
	}
	if models[0].ID != "gpt-4o" {
		t.Fatalf("models[0].ID = %q, want %q", models[0].ID, "gpt-4o")
	}
}
