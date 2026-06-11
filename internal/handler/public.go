package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
)

type fotoPublicaDTO struct {
	ID       int64     `json:"id"`
	UrlBaixa string    `json:"url_baixa"`
	CriadoEm time.Time `json:"criado_em"`
}

// GET /api/public/albums
func ListPublicAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := db.GetQueries().ListarAlbunsPublicos(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar álbuns públicos")
		return
	}
	if albums == nil {
		albums = []db.Albun{}
	}
	writeJSON(w, http.StatusOK, albums)
}

// GET /api/public/albums/{id}
func GetPublicAlbum(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	queries := db.GetQueries()

	// Só retorna se o álbum for público
	album, err := queries.ObterAlbumPublico(r.Context(), albumID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "álbum não encontrado ou não é público")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar álbum")
		return
	}

	fotos, err := queries.ListarFotografiasPorAlbum(r.Context(), albumID)
	if err != nil || fotos == nil {
		fotos = []db.Fotografia{}
	}

	// Retorna apenas previews públicos — sem url_alta (original privada)
	fotosPublicas := make([]fotoPublicaDTO, len(fotos))
	for i, f := range fotos {
		fotosPublicas[i] = fotoPublicaDTO{
			ID:       f.ID,
			UrlBaixa: f.UrlBaixa,
			CriadoEm: f.CriadoEm,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"album": album,
		"fotos": fotosPublicas,
	})
}
