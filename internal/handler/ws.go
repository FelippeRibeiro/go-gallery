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

	// Auth via cookie httpOnly — o navegador envia o cookie no handshake do WS
	// (mesma origem), então não é preciso passar o token na query string.
	c, err := r.Cookie(middleware.TokenCookieName)
	if err != nil || c.Value == "" {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	if _, err := middleware.ParseToken(c.Value); err != nil {
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	ws.GetHub().ServeWS(w, r, albumID)
}
