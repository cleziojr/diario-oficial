package api

// Autenticação: registro, login, refresh (com rotação) e /me.
// Leitura é pública; escrita/ingestão exige access token válido.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cleziojr/diario-oficial/backend/pkg/auth"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authHandlers struct {
	pool   *pgxpool.Pool
	secret string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userJSON struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type tokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         userJSON `json:"user"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validCredentials(c credentials) (string, bool) {
	email := normalizeEmail(c.Email)
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return "e-mail inválido", false
	}
	if len(c.Password) < 6 {
		return "a senha deve ter ao menos 6 caracteres", false
	}
	return "", true
}

// issueTokens gera access + refresh e persiste o hash do refresh (rotação).
func (h *authHandlers) issueTokens(ctx context.Context, userID, email string) (tokenResponse, error) {
	access, err := auth.GenerateAccessToken(h.secret, userID, email)
	if err != nil {
		return tokenResponse{}, err
	}
	raw, hash, err := auth.GenerateRefreshToken()
	if err != nil {
		return tokenResponse{}, err
	}
	if _, err := h.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1::uuid, $2, $3)`,
		userID, hash, time.Now().Add(auth.RefreshTTL),
	); err != nil {
		return tokenResponse{}, err
	}
	return tokenResponse{
		AccessToken:  access,
		RefreshToken: raw,
		User:         userJSON{ID: userID, Email: email},
	}, nil
}

// POST /api/v1/auth/register
func (h *authHandlers) register(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	if msg, ok := validCredentials(body); !ok {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	email := normalizeEmail(body.Email)
	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar senha")
		return
	}
	var userID string
	err = h.pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id::text`,
		email, hash,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao criar usuário")
		return
	}
	tokens, err := h.issueTokens(r.Context(), userID, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar tokens")
		return
	}
	writeJSON(w, http.StatusCreated, tokens)
}

// POST /api/v1/auth/login
func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	email := normalizeEmail(body.Email)

	var userID, hash string
	err := h.pool.QueryRow(r.Context(),
		`SELECT id::text, password_hash FROM users WHERE email = $1`, email,
	).Scan(&userID, &hash)
	if err != nil || !auth.CheckPassword(hash, body.Password) {
		// Mesma resposta para usuário inexistente e senha errada (evita enumeração).
		writeError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
		return
	}
	tokens, err := h.issueTokens(r.Context(), userID, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar tokens")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/v1/auth/refresh  (rotação: revoga o antigo e emite um novo par)
func (h *authHandlers) refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.RefreshToken) == "" {
		writeError(w, http.StatusBadRequest, "refresh_token é obrigatório")
		return
	}
	tokenHash := auth.HashToken(strings.TrimSpace(body.RefreshToken))

	var tokenID, userID, email string
	err := h.pool.QueryRow(r.Context(), `
SELECT rt.id::text, u.id::text, u.email
FROM refresh_tokens rt
JOIN users u ON u.id = rt.user_id
WHERE rt.token_hash = $1 AND rt.revoked = FALSE AND rt.expires_at > NOW()`,
		tokenHash,
	).Scan(&tokenID, &userID, &email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "refresh token inválido ou expirado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao validar token")
		return
	}

	// Rotação: revoga o refresh usado antes de emitir um novo.
	if _, err := h.pool.Exec(r.Context(),
		`UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1::uuid`, tokenID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao rotacionar token")
		return
	}

	tokens, err := h.issueTokens(r.Context(), userID, email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar tokens")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// GET /api/v1/auth/me  (rota protegida)
func (h *authHandlers) me(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, userJSON{
		ID:    auth.UserID(r.Context()),
		Email: auth.Email(r.Context()),
	})
}

func mountAuth(r chi.Router, pool *pgxpool.Pool, secret string) {
	h := &authHandlers{pool: pool, secret: secret}
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", h.register)
		r.Post("/login", h.login)
		r.Post("/refresh", h.refresh)
		r.With(auth.Require(secret)).Get("/me", h.me)
	})
}
