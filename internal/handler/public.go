package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
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

// GET /api/public/photographers
func ListPublicPhotographers(w http.ResponseWriter, r *http.Request) {
	fotografos, err := db.GetQueries().ListarFotografosComAlbumPublico(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar fotógrafos")
		return
	}
	if fotografos == nil {
		fotografos = []db.FotografoPublico{}
	}
	writeJSON(w, http.StatusOK, fotografos)
}

// GET /api/public/photographers/{id}
func GetPublicPhotographer(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	fotografoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	queries := db.GetQueries()
	fotografo, err := queries.ObterFotografoPorID(r.Context(), fotografoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "fotógrafo não encontrado")
		return
	}

	albums, _ := queries.ListarAlbunsPublicosPorFotografo(r.Context(), fotografoID)
	if albums == nil {
		albums = []db.Albun{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"fotografo": map[string]any{
			"id":          fotografo.ID,
			"nome":        fotografo.Nome,
			"foto_perfil": fotografo.FotoPerfil,
			"bio":         fotografo.Bio,
		},
		"albums": albums,
	})
}
