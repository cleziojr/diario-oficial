package api

// Matérias categorizadas pela LLM e a API pública de busca para jornalistas.
//   POST /api/v1/ingest                      (multipart: file=PDF)
//   GET  /api/v1/search?q=&tags=a,b&page=&limit=
//   GET  /api/v1/tags
//   GET  /api/v1/materias/{id}
//   GET  /api/v1/documents/{id}/materias

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cleziojr/diario-oficial/backend/pkg/aiclient"
	"github.com/cleziojr/diario-oficial/backend/pkg/pdfextract"
	"github.com/cleziojr/diario-oficial/backend/pkg/taxonomy"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	chunkSize    = 3000
	chunkOverlap = 200
)

// ---------------------------------------------------------------------------
// JSON
// ---------------------------------------------------------------------------

type documentRef struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	CreatedAt string `json:"created_at"`
}

type materiaJSON struct {
	ID             string          `json:"id"`
	DocumentID     string          `json:"document_id"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	Category       string          `json:"category"`
	Tags           []string        `json:"tags"`
	Entities       json.RawMessage `json:"entities"`
	MonetaryValues json.RawMessage `json:"monetary_values"`
	Dates          json.RawMessage `json:"dates"`
	Page           int             `json:"page"`
	Document       documentRef     `json:"document"`
	CreatedAt      string          `json:"created_at"`
}

type searchResponse struct {
	Items []materiaJSON `json:"items"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Total int           `json:"total"`
}

type tagCountJSON struct {
	Tag   string `json:"tag"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type ingestResponse struct {
	DocumentID    string `json:"document_id"`
	MateriasCount int    `json:"materias_count"`
	Pages         int    `json:"pages"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

type materiaHandlers struct {
	pool *pgxpool.Pool
	ai   *aiclient.Client
}

// scanMateria lê uma linha do SELECT padrão (materias JOIN documents).
func scanMateria(row pgx.Row) (materiaJSON, error) {
	var m materiaJSON
	var ent, mon, dat []byte
	if err := row.Scan(
		&m.ID, &m.DocumentID, &m.Title, &m.Summary, &m.Category, &m.Tags,
		&ent, &mon, &dat, &m.Page, &m.CreatedAt,
		&m.Document.ID, &m.Document.Filename, &m.Document.CreatedAt,
	); err != nil {
		return materiaJSON{}, err
	}
	m.Entities = jsonOr(ent, "{}")
	m.MonetaryValues = jsonOr(mon, "[]")
	m.Dates = jsonOr(dat, "[]")
	if m.Tags == nil {
		m.Tags = []string{}
	}
	return m, nil
}

func jsonOr(b []byte, fallback string) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage(fallback)
	}
	return json.RawMessage(b)
}

const selectMateria = `
SELECT m.id::text, m.document_id::text, m.title, m.summary, m.category, m.tags,
       m.entities, m.monetary_values, m.dates, coalesce(m.page, 0),
       to_char(m.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"'),
       d.id::text, d.filename,
       to_char(d.created_at, 'YYYY-MM-DD"T"HH24:MI:SS.US"Z"')
FROM materias m
JOIN documents d ON d.id = m.document_id`

// GET /api/v1/search
func (h *materiaHandlers) search(w http.ResponseWriter, r *http.Request) {
	page, limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var qArg any
	if q != "" {
		qArg = q
	}
	tags := parseTagsParam(r.URL.Query().Get("tags"))

	where := `
WHERE ($1::text IS NULL OR m.search_vector @@ websearch_to_tsquery('portuguese', $1))
  AND (cardinality($2::text[]) = 0 OR m.tags && $2::text[])`

	var total int
	if err := h.pool.QueryRow(r.Context(),
		`SELECT count(*) FROM materias m`+where, qArg, tags,
	).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao contar resultados")
		return
	}

	rows, err := h.pool.Query(r.Context(),
		selectMateria+where+` ORDER BY m.created_at DESC LIMIT $3 OFFSET $4`,
		qArg, tags, int32(limit), offset,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar matérias")
		return
	}
	defer rows.Close()

	items := make([]materiaJSON, 0, limit)
	for rows.Next() {
		m, err := scanMateria(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		items = append(items, m)
	}
	if rows.Err() != nil {
		writeError(w, http.StatusInternalServerError, "erro ao ler resultados")
		return
	}

	writeJSON(w, http.StatusOK, searchResponse{Items: items, Page: page, Limit: limit, Total: total})
}

// GET /api/v1/tags
func (h *materiaHandlers) tags(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(),
		`SELECT tag, count(*) FROM materias, unnest(tags) AS tag GROUP BY tag ORDER BY count(*) DESC, tag`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar tags")
		return
	}
	defer rows.Close()

	out := make([]tagCountJSON, 0, 16)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		out = append(out, tagCountJSON{Tag: tag, Label: taxonomy.Label(tag), Count: count})
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/v1/materias/{id}
func (h *materiaHandlers) get(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if _, err := parseUUIDParam(id); err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	m, err := scanMateria(h.pool.QueryRow(r.Context(), selectMateria+` WHERE m.id = $1::uuid`, id))
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "matéria não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar matéria")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// GET /api/v1/documents/{id}/materias
func (h *materiaHandlers) listByDocument(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if _, err := parseUUIDParam(id); err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	rows, err := h.pool.Query(r.Context(),
		selectMateria+` WHERE m.document_id = $1::uuid ORDER BY coalesce(m.page,0), m.created_at`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar matérias")
		return
	}
	defer rows.Close()

	items := make([]materiaJSON, 0)
	for rows.Next() {
		m, err := scanMateria(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		items = append(items, m)
	}
	writeJSON(w, http.StatusOK, items)
}

// POST /api/v1/ingest  (multipart/form-data: file=PDF, filename opcional)
func (h *materiaHandlers) ingest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "envie um multipart/form-data com o campo 'file'")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "campo 'file' (PDF) é obrigatório")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, 64<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "erro ao ler o arquivo")
		return
	}

	filename := strings.TrimSpace(r.FormValue("filename"))
	if filename == "" {
		filename = header.Filename
	}
	if filename == "" {
		filename = "documento.pdf"
	}

	extracted, err := pdfextract.Extract(r.Context(), data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "falha ao extrair PDF: "+err.Error())
		return
	}

	var documentID string
	if err := h.pool.QueryRow(r.Context(),
		`INSERT INTO documents (filename) VALUES ($1) RETURNING id::text`, filename,
	).Scan(&documentID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar documento")
		return
	}

	count, err := h.categorizeAndStore(r.Context(), documentID, extracted.Text)
	if err != nil {
		writeError(w, http.StatusBadGateway, "erro ao categorizar: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ingestResponse{
		DocumentID:    documentID,
		MateriasCount: count,
		Pages:         extracted.Pages,
	})
}

// categorizeAndStore percorre páginas (separadas por form-feed) e chunks,
// chama o ai-service e persiste as matérias. Devolve quantas foram criadas.
func (h *materiaHandlers) categorizeAndStore(ctx context.Context, documentID, text string) (int, error) {
	maxChunks := envInt("INGEST_MAX_CHUNKS", 0) // 0 = sem limite
	chunksDone := 0
	stored := 0

	pages := strings.Split(text, "\f")
	for pi, pageText := range pages {
		pageNum := pi + 1
		for _, chunk := range splitIntoChunks(pageText) {
			if strings.TrimSpace(chunk) == "" {
				continue
			}
			if maxChunks > 0 && chunksDone >= maxChunks {
				return stored, nil
			}
			chunksDone++

			res, err := h.ai.Categorize(ctx, chunk, pageNum)
			if err != nil {
				return stored, err
			}
			for _, m := range res.Materias {
				// body fica vazio: a busca full-text usa title+summary, que são
				// específicos por matéria (usar o chunk inteiro casaria todas as
				// matérias do mesmo trecho, derrubando a precisão).
				if err := h.insertMateria(ctx, documentID, pageNum, "", m); err != nil {
					return stored, err
				}
				stored++
			}
		}
	}
	return stored, nil
}

func (h *materiaHandlers) insertMateria(ctx context.Context, documentID string, page int, body string, m aiclient.Materia) error {
	category := taxonomy.Normalize(m.Category)
	tags := taxonomy.NormalizeList(append(m.Tags, category))

	entities, _ := json.Marshal(m.Entities)
	monetary, _ := json.Marshal(orEmptyArr(m.MonetaryValues))
	dates, _ := json.Marshal(orEmptyArr(m.Dates))

	_, err := h.pool.Exec(ctx, `
INSERT INTO materias
  (document_id, title, summary, body, category, tags, entities, monetary_values, dates, page)
VALUES ($1::uuid, $2, $3, $4, $5, $6::text[], $7::jsonb, $8::jsonb, $9::jsonb, $10)`,
		documentID, m.Title, m.Summary, body, category, tags,
		entities, monetary, dates, page,
	)
	return err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseTagsParam(raw string) []string {
	out := []string{}
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, taxonomy.Normalize(t))
		}
	}
	return out
}

func splitIntoChunks(text string) []string {
	runes := []rune(text)
	total := len(runes)
	if total == 0 {
		return nil
	}
	if total <= chunkSize {
		return []string{text}
	}
	step := chunkSize - chunkOverlap
	var chunks []string
	for start := 0; start < total; start += step {
		end := start + chunkSize
		if end > total {
			end = total
		}
		chunks = append(chunks, string(runes[start:end]))
		if end == total {
			break
		}
	}
	return chunks
}

func orEmptyArr(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// ---------------------------------------------------------------------------
// Mount
// ---------------------------------------------------------------------------

func mountMaterias(r chi.Router, pool *pgxpool.Pool, ai *aiclient.Client) {
	h := &materiaHandlers{pool: pool, ai: ai}
	r.Post("/api/v1/ingest", h.ingest)
	r.Get("/api/v1/search", h.search)
	r.Get("/api/v1/tags", h.tags)
	r.Get("/api/v1/materias/{id}", h.get)
	r.Get("/api/v1/documents/{id}/materias", h.listByDocument)
}
