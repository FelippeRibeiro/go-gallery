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

// TokenCookieName é o nome do cookie httpOnly que guarda o JWT.
const TokenCookieName = "token"

const tokenTTL = 24 * time.Hour

// cookieSecure indica se o cookie deve ter o atributo Secure (HTTPS).
// Controlado por COOKIE_SECURE=true em produção.
func cookieSecure() bool {
	v := strings.ToLower(os.Getenv("COOKIE_SECURE"))
	return v == "true" || v == "1"
}

// SetTokenCookie grava o JWT num cookie httpOnly. O token nunca é exposto ao JS.
func SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     TokenCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(tokenTTL.Seconds()),
	})
}

// ClearTokenCookie remove o cookie de autenticação (logout).
func ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     TokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// tokenFromRequest extrai o JWT do cookie httpOnly ou, como fallback, do header
// Authorization: Bearer (útil para clientes de API/testes).
func tokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(TokenCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
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
		tokenStr := tokenFromRequest(r)
		if tokenStr == "" {
			http.Error(w, `{"error":"não autorizado"}`, http.StatusUnauthorized)
			return
		}
		claims, err := ParseToken(tokenStr)
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
