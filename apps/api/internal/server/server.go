// Package server wires the modular monolith together: one router, one process,
// modules mounted side by side (docs/05-api-and-backend-spec.md).
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/meracare/api/internal/auth"
	"github.com/meracare/api/internal/config"
	"github.com/meracare/api/internal/database"
	"github.com/meracare/api/internal/users"
	"github.com/meracare/api/pkg/httpx"
)

// Dependencies are the collaborators the router needs.
type Dependencies struct {
	Config   *config.Config
	Logger   *slog.Logger
	Pool     *database.Pool
	Verifier auth.Verifier
}

// New builds the fully wired HTTP handler.
func New(deps Dependencies) http.Handler {
	userRepo := users.NewRepository(deps.Pool)
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService)

	router := chi.NewRouter()
	router.NotFound(httpx.NotFoundHandler())
	router.MethodNotAllowed(httpx.MethodNotAllowedHandler())

	router.Use(middleware.RequestID)
	router.Use(httpx.RequestLogger(deps.Logger))
	router.Use(httpx.Recoverer)
	router.Use(httpx.CORS(deps.Config.CORSAllowedOrigins))
	router.Use(middleware.Timeout(deps.Config.RequestTimeout))

	// Unauthenticated operational endpoints.
	router.Get("/healthz", healthHandler)
	router.Get("/readyz", readyHandler(deps.Pool))

	router.Route("/v1", func(v1 chi.Router) {
		v1.Use(auth.RequireAuth(deps.Verifier, userService))
		v1.Mount("/me", userHandler.Routes())
	})

	return router
}

// healthHandler reports that the process is running. It performs no dependency
// checks so a liveness probe never restarts the API over a database blip.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

// readyHandler reports whether the API can serve traffic, i.e. the database is
// reachable.
func readyHandler(pool *database.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.Ping(ctx, pool); err != nil {
			httpx.WriteError(w, r, httpx.ErrUnavailable("The service is not ready.").WithCause(err))
			return
		}
		httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ready"})
	}
}
