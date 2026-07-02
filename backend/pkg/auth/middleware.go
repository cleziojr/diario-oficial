package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ctxKey string

const (
	userIDKey ctxKey = "uid"
	emailKey  ctxKey = "email"
)

// Require é um middleware que exige um access token (JWT) válido no header
// Authorization: Bearer <token>. Sem token válido, responde 401.
func Require(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				unauthorized(w)
				return
			}
			tokenStr := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			claims, err := ParseAccessToken(secret, tokenStr)
			if err != nil {
				unauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, emailKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID devolve o id do usuário autenticado (ou "" se não houver).
func UserID(ctx context.Context) string {
	v, _ := ctx.Value(userIDKey).(string)
	return v
}

// Email devolve o e-mail do usuário autenticado (ou "").
func Email(ctx context.Context) string {
	v, _ := ctx.Value(emailKey).(string)
	return v
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "autenticação necessária"})
}
