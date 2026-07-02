// Package auth cuida de hashing de senha, emissão/validação de JWT (access token)
// e geração de refresh tokens opacos (com rotação).
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	// AccessTTL é a validade curta do access token (JWT).
	AccessTTL = 15 * time.Minute
	// RefreshTTL é a validade do refresh token.
	RefreshTTL = 7 * 24 * time.Hour
)

// ErrInvalidToken indica JWT ausente/expirado/adulterado.
var ErrInvalidToken = errors.New("token inválido")

// Claims são as claims do access token.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// HashPassword gera o hash bcrypt da senha.
func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compara senha em claro com o hash bcrypt.
func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// GenerateAccessToken assina um JWT HS256 para o usuário.
func GenerateAccessToken(secret, userID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(secret))
}

// ParseAccessToken valida a assinatura e a expiração, devolvendo as claims.
func ParseAccessToken(secret, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// GenerateRefreshToken cria um token opaco aleatório e devolve (valor, hash).
// Só o hash é persistido; o valor vai para o cliente.
func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, HashToken(raw), nil
}

// HashToken devolve o SHA-256 (hex) de um refresh token.
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
