package newapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	newapi "github.com/xy200303/go-newapi-sdk/newapi"
)

func TestNewAppliesDefaultsAndOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 3 * time.Second}

	client, err := newapi.New(
		"https://example.com/",
		newapi.WithHTTPClient(httpClient),
		newapi.WithAdminAuth("root-token", 99),
		newapi.WithUserAgent("custom-agent"),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if client.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q, want %q", client.BaseURL, "https://example.com")
	}
	if client.RootToken != "root-token" {
		t.Fatalf("RootToken = %q, want %q", client.RootToken, "root-token")
	}
	if client.RootUserID != 99 {
		t.Fatalf("RootUserID = %d, want %d", client.RootUserID, 99)
	}
	if client.HTTPClient != httpClient {
		t.Fatal("HTTPClient was not set by option")
	}
	if client.UserAgent != "custom-agent" {
		t.Fatalf("UserAgent = %q, want %q", client.UserAgent, "custom-agent")
	}
	if client.Management == nil || client.AIModel == nil {
		t.Fatal("generated service trees were not initialized")
	}
}

func TestGeneratedOperationReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"invalid token"}`))
	}))
	defer server.Close()

	client, err := newapi.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	var out map[string]any
	err = client.Management.System.StatusGet.Do(context.Background(), nil, &out)
	if err == nil {
		t.Fatal("Do() error = nil, want APIError")
	}

	apiErr, ok := err.(*newapi.APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
	if apiErr.Message != "invalid token" {
		t.Fatalf("Message = %q, want %q", apiErr.Message, "invalid token")
	}
}
