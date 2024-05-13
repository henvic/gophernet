package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(log *slog.Logger, api *API, httpAddress string) *HTTPServer {
	router := http.NewServeMux()
	server := &HTTPServer{
		http: &http.Server{
			Addr:    httpAddress,
			Handler: router,

			ReadHeaderTimeout: 5 * time.Second, // mitigate risk of Slowloris Attack
		},
		log: log,
		api: api,
	}

	router.HandleFunc("GET /burrows", server.statusHandler)
	router.HandleFunc("POST /burrows/{name}/rent", server.rentHandler)

	return server
}

// HTTPServer with HTTP handlers.
type HTTPServer struct {
	http *http.Server
	log  *slog.Logger
	api  *API
}

// Run API server, initializing an HTTP API.
func (s *HTTPServer) Run(ctx context.Context) error {
	s.log.Info("HTTP server listening", slog.Any("address", s.http.Addr))
	if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown HTTP server.
func (s *HTTPServer) Shutdown(ctx context.Context) {
	s.log.Info("shutting down HTTP server gracefully")
	if s.http != nil {
		if err := s.http.Shutdown(ctx); err != nil {
			s.log.Error("graceful shutdown of HTTP server failed", slog.Any("error", err))
		}
	}
}

func (s *HTTPServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	var burrows []Burrow
	s.api.Status(&burrows)
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "	")
	enc.Encode(&burrows)
}

func (s *HTTPServer) rentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "	")

	var burrow Burrow
	if err := s.api.RentBurrow(r.PathValue("name"), &burrow); err != nil {
		// If error type isn't APIError, return a generic 500 error without exposing any internal error details.
		var apiErr APIError
		if !errors.As(err, &apiErr) {
			apiErr = APIError{
				HTTPCode: http.StatusInternalServerError,
				Message:  http.StatusText(http.StatusInternalServerError),
			}

			// Don't log context errors, as they are expected when the client cancels the request.
			if err != context.Canceled && err != context.DeadlineExceeded {
				s.log.LogAttrs(r.Context(), slog.LevelError, "unexpected error handling HTTP request", slog.Any("error", err))
			}
		}

		// Otherwise, return friendly error message.
		w.WriteHeader(apiErr.HTTPCode)
		enc.Encode(&apiErr)
		return
	}

	enc.Encode(&burrow)
}
