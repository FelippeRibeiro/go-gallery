package handler

import (
	"net/http"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	"github.com/FelippeRibeiro/go-gallery/internal/ws"
)

// GET /ws/albums/{id}
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
	claims, err := middleware.ParseToken(c.Value)
	if err != nil {
		http.Error(w, "token inválido", http.StatusUnauthorized)
		return
	}

	queries := db.GetQueries()
	if _, err := queries.ObterUsuarioPorID(r.Context(), claims.UserID); err != nil {
		http.Error(w, "usuário não encontrado ou desativado", http.StatusUnauthorized)
		return
	}

	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		http.Error(w, "álbum não encontrado", http.StatusNotFound)
		return
	}

	// Mesmo critério de acesso do GetAlbum: dono, colaborador, cliente com
	// vínculo ou álbum público. Sem isso, qualquer usuário autenticado poderia
	// entrar na sala de um álbum privado alheio e receber os eventos de upload.
	autorizado := album.Publico || album.IDFotografo == claims.UserID
	if !autorizado {
		if ok, _ := queries.VerificarFotografoAlbum(r.Context(), claims.UserID, albumID); ok {
			autorizado = true
		}
	}
	if !autorizado {
		if ok, _ := queries.VerificarAcessoClienteAlbum(r.Context(), claims.UserID, albumID); ok {
			autorizado = true
		}
	}
	if !autorizado {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	ws.GetHub().ServeWS(w, r, albumID)
}
