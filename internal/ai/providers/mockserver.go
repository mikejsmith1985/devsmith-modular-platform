package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// StartMockOllamaServer starts a minimal deterministic HTTP server that
// implements a tiny subset of the Ollama HTTP API used by the review service.
// It is intended for local test runs where a full Ollama instance is unavailable
// or slow. Returns a shutdown function to stop the server.
func StartMockOllamaServer(listenAddr, model string) (func(context.Context) error, error) {
	mux := http.NewServeMux()

	// /api/tags - return available models
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}{
			Models: []struct {
				Name string `json:"name"`
			}{{Name: model}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// /api/generate - accept prompt and return deterministic response quickly
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Build a small deterministic response based on prompt length
		resp := struct {
			Response           string `json:"response"`
			Model              string `json:"model"`
			StopReason         string `json:"stop_reason,omitempty"`
			PromptEvalCount    int    `json:"prompt_eval_count"`
			EvalCount          int    `json:"eval_count"`
			EvalDuration       int64  `json:"eval_duration"`
			PromptEvalDuration int64  `json:"prompt_eval_duration"`
			Done               bool   `json:"done"`
		}{
			Response: fmt.Sprintf("{\"mocked\":true,\"preview\":\"%s\"}", truncateForMock(req.Prompt)),
			Model:    model,
			Done:     true,
		}

		// Simulate a very small processing delay so tests still exercise async waits
		time.Sleep(200 * time.Millisecond)

		_ = json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	// Try to listen on the requested address to fail fast if port is unavailable
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = server.Serve(ln) // Serve will return when server is Shutdown
	}()

	shutdown := func(ctx context.Context) error {
		return server.Shutdown(ctx)
	}

	return shutdown, nil
}

// truncateForMock keeps returned mocked payloads small and predictable
func truncateForMock(s string) string {
	if len(s) <= 120 {
		return s
	}
	return s[:120] + "..."
}
