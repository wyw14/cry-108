package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/wyw14/fabchem/internal/api"
)

func TestHealthAndCollections(t *testing.T) {
	runtime, err := api.NewRuntime(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.NewServer(runtime, nil).Handler())
	defer server.Close()
	for _, path := range []string{"/healthz", "/api/blends", "/api/manifolds", "/api/dispenses", "/api/incidents"} {
		response, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			t.Fatalf("GET %s returned %d", path, response.StatusCode)
		}
		var body map[string]any
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			response.Body.Close()
			t.Fatalf("decode %s: %v", path, err)
		}
		response.Body.Close()
	}
}
