package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	imgutil "github.com/FelippeRibeiro/go-gallery/internal/image"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
)

// GET /api/profile
func GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	u, err := db.GetQueries().ObterUsuarioPorID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "usuário não encontrado")
		return
	}

	fotoPerfil := u.FotoPerfil
	if fotoPerfil.Valid {
		fotoPerfil.String = s3client.PublicURL(fotoPerfil.String)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":          u.ID,
		"nome":        u.Nome,
		"email":       u.Email,
		"tipo":        u.Tipo,
		"foto_perfil": fotoPerfil,
		"bio":         u.Bio,
	})
}

// PATCH /api/profile
func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var body struct {
		Bio string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	if err := db.GetQueries().AtualizarBioFotografo(r.Context(), userID, body.Bio); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar perfil")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/profile/photo
func UploadProfilePhoto(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "multipart inválido")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo ausente")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao ler arquivo")
		return
	}

	img, _, err := imgutil.Decode(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "formato de imagem inválido")
		return
	}

	avatarData, err := imgutil.GenerateAvatar(img)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar imagem")
		return
	}

	key := fmt.Sprintf("profiles/%d/avatar.jpg", userID)
	if err := s3client.UploadPublic(r.Context(), key, avatarData, "image/jpeg"); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao salvar imagem")
		return
	}

	// Grava apenas a key; a URL é montada na leitura.
	if err := db.GetQueries().AtualizarFotoPerfilFotografo(r.Context(), userID, key); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao salvar url da foto")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"foto_perfil": s3client.PublicURL(key)})
}
