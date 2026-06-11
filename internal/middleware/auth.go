package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserTipoKey contextKey = "userTipo"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Tipo   string `json:"tipo"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return []byte("dev-secret-change-in-production")
	}
	return []byte(s)
}

func GenerateToken(userID int64, tipo string) (string, error) {
	claims := Claims{
		UserID: userID,
		Tipo:   tipo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"não autorizado"}`, http.StatusUnauthorized)
			return
		}
		claims, err := ParseToken(strings.TrimPrefix(authHeader, "Bearer "))
		if err != nil {
			http.Error(w, `{"error":"token inválido"}`, http.StatusUnauthorized)
			return
		}

		// Verify the user still exists and is active in the database.
		// This ensures deleted/deactivated accounts can't use old tokens.
		if _, err := db.GetQueries().ObterUsuarioPorID(r.Context(), claims.UserID); err != nil {
			http.Error(w, `{"error":"usuário não encontrado ou desativado"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserTipoKey, claims.Tipo)
		next(w, r.WithContext(ctx))
	}
}
