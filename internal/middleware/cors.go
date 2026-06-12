package middleware

import (
	"net/http"
	"os"
	"strings"
)

// appOrigin devolve a origem do frontend (APP_URL), única origem externa
// autorizada a fazer requisições cross-origin.
func appOrigin() string {
	if u := os.Getenv("APP_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:5173"
}

// CORS restringe cross-origin à origem do frontend (APP_URL). Em dev (proxy do
// Vite) e em produção (Go serve a SPA) tudo é same-origin e estes headers nem
// são consultados pelo browser — liberar "*" aqui só abriria a leitura da API
// para qualquer site.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && origin == appOrigin() {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
