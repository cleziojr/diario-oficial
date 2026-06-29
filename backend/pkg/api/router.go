package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/cleziojr/diario-oficial/backend/pkg/aiclient"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Get("/ready", func(w http.ResponseWriter, req *http.Request) {
		if err := pool.Ping(req.Context()); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	})

	q := sqlc.New(pool)
	aiClient := aiclient.New(os.Getenv("AI_SERVICE_URL"))

	// API pública de busca (matérias categorizadas) + ingestão de PDF.
	mountMaterias(r, pool, aiClient)

	r.Route("/api/v1/documents", func(r chi.Router) {
		mountDocuments(r, q)
		mountDocumentInsights(r, q)     // GET  /documents/{id}/insights  (agregador legacy)
		mountDocumentInsightsCRUD(r, q) // POST /documents/{documentId}/insights
		                                // GET  /documents/{documentId}/insights[?model=]
		r.Route("/{documentId}/analyses", func(r chi.Router) {
			mountDocumentAnalyses(r, q, aiClient)
		})
	})

	r.Route("/api/v1/analyses", func(r chi.Router) {
		mountAnalyses(r, q)
		mountAnalysisInsights(r, q) // GET /analyses/{analysisId}/insights
	})

	r.Route("/api/v1/insights", func(r chi.Router) {
		mountInsights(r, q) // GET/PATCH/DELETE /insights/{id}
	})

	return r
}