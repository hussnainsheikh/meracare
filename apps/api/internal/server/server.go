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
	"github.com/meracare/api/internal/authz"
	"github.com/meracare/api/internal/config"
	"github.com/meracare/api/internal/database"
	"github.com/meracare/api/internal/invitations"
	"github.com/meracare/api/internal/members"
	"github.com/meracare/api/internal/relationships"
	"github.com/meracare/api/internal/seniors"
	"github.com/meracare/api/internal/tasks"
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

	relationshipRepo := relationships.NewRepository(deps.Pool)
	// Every senior-scoped route resolves access through this guard.
	guard := authz.NewGuard(relationshipRepo)

	seniorRepo := seniors.NewRepository(deps.Pool)
	seniorHandler := seniors.NewHandler(seniors.NewService(seniorRepo, relationshipRepo), guard)

	memberHandler := members.NewHandler(members.NewService(relationshipRepo), guard)

	requireAuth := auth.RequireAuth(deps.Verifier, userService)
	invitationService := invitations.NewService(
		invitations.NewRepository(deps.Pool),
		relationshipRepo,
		userLookup{repo: userRepo},
		seniorRepo,
	)
	invitationHandler := invitations.NewHandler(invitationService, guard, requireAuth)

	taskHandler := tasks.NewHandler(
		tasks.NewService(tasks.NewRepository(deps.Pool), seniorRepo, relationshipRepo),
		guard,
	)

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

	// Invitations are mounted outside the authenticated group: previewing an
	// invitation must work before the recipient has an account. The routes that
	// need a signed-in user apply the middleware themselves.
	router.Route("/v1/invitations", func(tokens chi.Router) {
		tokens.Mount("/", invitationHandler.TokenRoutes())
	})

	router.Route("/v1", func(v1 chi.Router) {
		v1.Use(requireAuth)
		v1.Mount("/me", userHandler.Routes())
		v1.Mount("/tasks", taskHandler.TaskRoutes())
		v1.Mount("/seniors", seniorHandler.Routes(seniors.SubRoutes{
			Members:     memberHandler.Routes(),
			Invitations: invitationHandler.SeniorRoutes(),
			Tasks:       taskHandler.SeniorRoutes(),
		}))
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
