package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	imgproc "github.com/FelippeRibeiro/go-gallery/internal/image"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/google/uuid"
)

const maxFileSize = 10 << 20 // 10 MB

var allowedTypes = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

type uploadResponse struct {
	ID          string `json:"id"`
	OriginalKey string `json:"original_key"`
	PreviewKey  string `json:"preview_key"`
	OriginalURL string `json:"original_url"`
	PreviewURL  string `json:"preview_url"`
}

func Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1024)

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "arquivo muito grande (máximo 10 MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "erro ao processar formulário", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	// albumID := r.FormValue("album_id")
	// if albumID == "" {
	// 	http.Error(w, "campo 'album_id' é obrigatório", http.StatusBadRequest)
	// 	return
	// }
	if err != nil {
		http.Error(w, "campo 'file' é obrigatório", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "arquivo muito grande (máximo 10 MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "erro ao ler arquivo", http.StatusBadRequest)
		return
	}

	contentType := http.DetectContentType(data)
	ext, ok := allowedTypes[contentType]
	if !ok {
		http.Error(w, "tipo de arquivo não permitido; envie uma imagem (jpeg, png, gif ou webp)", http.StatusBadRequest)
		return
	}

	img, _, err := imgproc.Decode(data)
	if err != nil {
		http.Error(w, "arquivo de imagem inválido", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	originalKey := fmt.Sprintf("uploads/%s/original.%s", id, ext)
	previewKey := fmt.Sprintf("uploads/%s/preview.jpg", id)

	ctx := r.Context()

	if err := s3client.UploadPrivate(ctx, originalKey, data, contentType); err != nil {
		http.Error(w, "erro ao enviar imagem original para o S3", http.StatusInternalServerError)
		return
	}

	preview, err := imgproc.GeneratePreview(img)
	if err != nil {
		http.Error(w, "erro ao gerar preview da imagem", http.StatusInternalServerError)
		return
	}

	if err := s3client.UploadPublic(ctx, previewKey, preview, "image/jpeg"); err != nil {
		http.Error(w, "erro ao enviar preview para o S3", http.StatusInternalServerError)
		return
	}

	resp := uploadResponse{
		ID:          id,
		OriginalKey: originalKey,
		PreviewKey:  previewKey,
		OriginalURL: s3client.ObjectURL(originalKey),
		PreviewURL:  s3client.ObjectURL(previewKey),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
