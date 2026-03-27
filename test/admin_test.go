package newapi_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	newapi "github.com/xy200303/go-newapi-sdk/newapi"
)

func TestAdminSearchUserContextReturnsErrNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/search" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"","data":{"items":[],"total":0}}`))
	}))
	defer server.Close()

	client, err := newapi.New(server.URL)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.AdminSearchUserContext(t.Context(), "missing")
	if err == nil {
		t.Fatal("AdminSearchUserContext() error = nil, want ErrNotFound")
	}
	if !errors.Is(err, newapi.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}
