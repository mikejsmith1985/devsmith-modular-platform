package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestClient_Post verifies that Client.Post succeeds on 2xx responses and
// returns an error for non-2xx responses.
func TestClient_Post_SuccessAndFailure(t *testing.T) {
	t.Run("success-2xx", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			// Optionally assert headers or body here
			w.WriteHeader(201)
		}))
		defer srv.Close()

		c := NewClient(srv.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		payload := map[string]interface{}{"event": "test"}
		if err := c.Post(ctx, payload); err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})

	t.Run("failure-non2xx", func(t *testing.T) {
		srvFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer srvFail.Close()

		c2 := NewClient(srvFail.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		payload := map[string]interface{}{"event": "test"}
		if err := c2.Post(ctx, payload); err == nil {
			t.Fatalf("expected error for 500 response")
		}
	})
}
