package handler

import (
	"net/http"

	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	"github.com/FelippeRibeiro/go-gallery/internal/ws"
)

// GET /ws/albums/{id}?token=JWT
func ServeAlbumWS(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}

	// WebSocket auth via query param (browsers can't send custom headers)
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "token ausente", http.StatusUnauthorized)
		return
	}
	if _, err := middleware.ParseToken(tokenStr); err != nil {
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	ws.GetHub().ServeWS(w, r, albumID)
}
