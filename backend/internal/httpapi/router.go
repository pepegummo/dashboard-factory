package httpapi

import (
	"database/sql"
	"net/http"

	"dashboard-builder/backend/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter wires up all HTTP routes for the dashboard builder API.
func NewRouter(database *sql.DB) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	factoryH := handlers.NewFactoryHandler(database)
	templateH := handlers.NewTemplateHandler(database)
	dashboardH := handlers.NewDashboardHandler(database)
	telemetryH := handlers.NewTelemetryHandler(database)

	r.Get("/api/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/factories", func(r chi.Router) {
		r.Get("/", factoryH.List)
		r.Get("/{id}/machines", factoryH.ListMachines)
	})

	r.Get("/api/metrics", templateH.Metrics)

	r.Route("/api/templates", func(r chi.Router) {
		r.Get("/", templateH.List)
		r.Post("/", templateH.Create)
		r.Get("/{id}", templateH.Get)
		r.Put("/{id}", templateH.Update)
		r.Delete("/{id}", templateH.Delete)
	})

	r.Route("/api/dashboards", func(r chi.Router) {
		r.Get("/", dashboardH.List)
		r.Post("/", dashboardH.Create)
		r.Get("/{id}", dashboardH.Get)
		r.Delete("/{id}", dashboardH.Delete)
	})

	r.Get("/api/machines/{id}/telemetry", telemetryH.Get)

	return r
}
