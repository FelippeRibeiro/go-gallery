package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/email"
	imgproc "github.com/FelippeRibeiro/go-gallery/internal/image"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/google/uuid"
)

// GET /api/albums
func ListAlbums(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	tipo, _ := r.Context().Value(middleware.UserTipoKey).(string)

	queries := db.GetQueries()
	var albums []db.Albun

	switch tipo {
	case "cliente":
		albums, _ = queries.ListarAlbunsPorCliente(r.Context(), userID)
	default:
		own, _ := queries.ListarAlbunsPorFotografo(r.Context(), userID)
		collab, _ := queries.ListarAlbunsColaborador(r.Context(), userID)
		seen := make(map[int64]bool)
		for _, a := range own {
			seen[a.ID] = true
			albums = append(albums, a)
		}
		for _, a := range collab {
			if !seen[a.ID] {
				albums = append(albums, a)
			}
		}
	}

	if albums == nil {
		albums = []db.Albun{}
	}

	search := strings.ToLower(r.URL.Query().Get("search"))
	if search != "" {
		filtered := albums[:0]
		for _, a := range albums {
			if strings.Contains(strings.ToLower(a.Titulo), search) {
				filtered = append(filtered, a)
			}
		}
		albums = filtered
	}

	writeJSON(w, http.StatusOK, comURLAlbuns(albums))
}

// POST /api/albums
func CreateAlbum(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	var req struct {
		Titulo                  string `json:"titulo"`
		Descricao               string `json:"descricao"`
		DataEvento              string `json:"data_evento"`
		Lote                    bool   `json:"lote"`
		Publico                 bool   `json:"publico"`
		ValorAlbum              string `json:"valor_album"`
		ValorUnitarioFotografia string `json:"valor_unitario_fotografia"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if req.Titulo == "" || req.DataEvento == "" {
		writeError(w, http.StatusBadRequest, "titulo e data_evento são obrigatórios")
		return
	}

	dataEvento, err := time.Parse("2006-01-02", req.DataEvento)
	if err != nil {
		writeError(w, http.StatusBadRequest, "data_evento inválida, use o formato AAAA-MM-DD")
		return
	}

	valorAlbum := req.ValorAlbum
	if valorAlbum == "" {
		valorAlbum = "0.00"
	}
	valorUnitario := req.ValorUnitarioFotografia
	if valorUnitario == "" {
		valorUnitario = "0.00"
	}

	album, err := db.GetQueries().CriarAlbum(r.Context(), db.CriarAlbumParams{
		Titulo:                  req.Titulo,
		Descricao:               sql.NullString{String: req.Descricao, Valid: req.Descricao != ""},
		DataEvento:              dataEvento,
		IDFotografo:             userID,
		Lote:                    req.Lote,
		Publico:                 req.Publico,
		ValorAlbum:              valorAlbum,
		ValorUnitarioFotografia: valorUnitario,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar álbum")
		return
	}
	writeJSON(w, http.StatusCreated, comURLAlbum(album))
}

// GET /api/albums/{id}
func GetAlbum(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "álbum não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar álbum")
		return
	}

	// Determina o papel do usuário em relação ao álbum
	papel := "visitante"
	if album.IDFotografo == userID {
		papel = "dono"
	} else if isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID); isCollab {
		papel = "colaborador"
	} else if hasAccess, _ := queries.VerificarAcessoClienteAlbum(r.Context(), userID, albumID); hasAccess {
		papel = "cliente"
	}

	// Sem vínculo com o álbum: só pode visualizar se for público
	if papel == "visitante" && !album.Publico {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	fotos, err := queries.ListarFotografiasPorAlbum(r.Context(), albumID)
	if err != nil || fotos == nil {
		fotos = []db.Fotografia{}
	}

	// Dono e colaboradores recebem as URLs originais (url_alta)
	if papel == "dono" || papel == "colaborador" {
		writeJSON(w, http.StatusOK, map[string]any{"album": comURLAlbum(album), "fotos": comURLFotos(fotos), "papel": papel})
		return
	}

	// Clientes e visitantes recebem apenas previews (sem url_alta)
	fotosPublicas := make([]fotoPublicaDTO, len(fotos))
	for i, f := range fotos {
		fotosPublicas[i] = fotoPublicaDTO{ID: f.ID, UrlBaixa: s3client.PublicURL(f.UrlBaixa), CriadoEm: f.CriadoEm}
	}
	writeJSON(w, http.StatusOK, map[string]any{"album": comURLAlbum(album), "fotos": fotosPublicas, "papel": papel})
}

// PATCH /api/albums/{id}/visibility  — somente dono
func UpdateVisibility(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	var req struct {
		Publico bool `json:"publico"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	updated, err := db.GetQueries().AtualizarPublicoAlbum(r.Context(), albumID, req.Publico)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar álbum")
		return
	}
	writeJSON(w, http.StatusOK, comURLAlbum(updated))
}

// POST /api/albums/{id}/invite  — convite de cliente, somente dono, somente privado
func InviteClient(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := albumDoFotografo(r, albumID, userID)
	if err != nil {
		respondAlbumErr(w, err)
		return
	}

	if album.Publico {
		writeError(w, http.StatusBadRequest, "álbuns públicos não precisam de convite")
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "email é obrigatório")
		return
	}

	fotografo, err := queries.ObterUsuarioPorID(r.Context(), userID)
	if err == nil && fotografo.Email == req.Email {
		writeError(w, http.StatusBadRequest, "você não pode se convidar para seu próprio álbum")
		return
	}

	existente, err := queries.ObterUsuarioPorEmail(r.Context(), req.Email)
	if err == nil {
		// Se já é fotógrafo, bloqueia convite de cliente
		if fmt.Sprintf("%v", existente.Tipo) == "fotografo" {
			writeError(w, http.StatusBadRequest, "este e-mail pertence a um fotógrafo — use a opção de convidar fotógrafo")
			return
		}
		temAcesso, _ := queries.VerificarAcessoClienteAlbum(r.Context(), existente.ID, albumID)
		if temAcesso {
			writeError(w, http.StatusConflict, "este usuário já tem acesso ao álbum")
			return
		}
	}

	pendente, _ := queries.VerificarConvitePendente(r.Context(), req.Email, albumID)
	if pendente {
		writeError(w, http.StatusConflict, "já existe um convite pendente para este e-mail")
		return
	}

	token := uuid.New().String()
	convite, err := queries.CriarConvite(r.Context(), db.CriarConviteParams{
		Token:     token,
		IDAlbum:   albumID,
		Email:     req.Email,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar convite")
		return
	}

	capaURL := ""
	if album.CapaUrl.Valid {
		capaURL = s3client.PublicURL(album.CapaUrl.String)
	}
	go func() {
		if err := email.EnviarConviteAlbum(req.Email, album.Titulo, token, capaURL); err != nil {
			fmt.Printf("[WARN] falha ao enviar email para %s: %v\n", req.Email, err)
		}
	}()

	writeJSON(w, http.StatusCreated, map[string]any{
		"convite_id": convite.ID,
		"email":      req.Email,
		"expires_at": convite.ExpiresAt,
	})
}

// POST /api/albums/{id}/invite-photographer  — somente dono, fotógrafo já cadastrado
func InvitePhotographer(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "email é obrigatório")
		return
	}

	queries := db.GetQueries()
	user, err := queries.ObterUsuarioPorEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusNotFound, "fotógrafo não encontrado — o usuário deve estar cadastrado na plataforma")
		return
	}
	if fmt.Sprintf("%v", user.Tipo) != "fotografo" {
		writeError(w, http.StatusBadRequest, "o usuário deve ter cadastro como fotógrafo")
		return
	}
	if user.ID == userID {
		writeError(w, http.StatusBadRequest, "você já é o dono deste álbum")
		return
	}

	isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), user.ID, albumID)
	if isCollab {
		writeError(w, http.StatusConflict, "este fotógrafo já colabora neste álbum")
		return
	}

	fa, err := queries.AssociarFotografoAlbum(r.Context(), user.ID, albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao associar fotógrafo")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":           fa.ID,
		"id_fotografo": fa.IDFotografo,
		"nome":         user.Nome,
		"email":        user.Email,
		"criado_em":    fa.CriadoEm,
	})
}

// GET /api/albums/{id}/invites  — somente dono
func ListInvites(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	queries := db.GetQueries()
	convites, _ := queries.ListarConvitesPorAlbum(r.Context(), albumID)
	fotografos, _ := queries.ListarFotografosAlbum(r.Context(), albumID)
	clientes, _ := queries.ListarClientesAlbum(r.Context(), albumID)

	if convites == nil {
		convites = []db.Convite{}
	}
	if fotografos == nil {
		fotografos = []db.FotografoAlbumInfo{}
	}
	if clientes == nil {
		clientes = []db.ClienteAlbumInfo{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"convites":   convites,
		"fotografos": fotografos,
		"clientes":   clientes,
	})
}

// DELETE /api/albums/{id}/invites/{inviteId}  — somente dono
func RevokeInvite(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	inviteID, err := strconv.ParseInt(r.PathValue("inviteId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	if err := db.GetQueries().RevogarConvite(r.Context(), inviteID, albumID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao revogar convite")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/albums/{id}/photographers/{faId}  — somente dono
func RemovePhotographer(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	faID, err := strconv.ParseInt(r.PathValue("faId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	if err := db.GetQueries().RemoverFotografoAlbum(r.Context(), faID, albumID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover fotógrafo")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/albums/{id}/clients/{clientId}  — somente dono
func RemoveClient(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	clientID, err := strconv.ParseInt(r.PathValue("clientId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	if err := db.GetQueries().RemoverClienteAlbum(r.Context(), clientID, albumID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover cliente")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/albums/{id}/photos  — dono OU colaborador fotógrafo
func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "álbum não encontrado")
		return
	}
	// Must be owner or collaborator
	if album.IDFotografo != userID {
		isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID)
		if !isCollab {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
	}

	data, contentType, ext, err := readImageForm(w, r)
	if err != nil {
		return
	}

	img, _, err := imgproc.Decode(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo de imagem inválido")
		return
	}

	photoID := uuid.New().String()
	originalKey, previewKey := s3client.AlbumPhotoKeys(albumID, photoID, ext)
	ctx := r.Context()

	if err := s3client.UploadPrivate(ctx, originalKey, data, contentType); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao enviar imagem para o S3")
		return
	}

	preview, err := imgproc.GeneratePreview(img)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao gerar preview")
		return
	}
	if err := s3client.UploadPublic(ctx, previewKey, preview, "image/jpeg"); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao enviar preview para o S3")
		return
	}

	valorUnitario := album.ValorUnitarioFotografia
	if valorUnitario == "" {
		valorUnitario = "0.00"
	}

	foto, err := queries.CriarFotografia(ctx, db.CriarFotografiaParams{
		UrlAlta:       originalKey,
		UrlBaixa:      previewKey,
		IDFotografo:   userID,
		IDAlbum:       albumID,
		ValorUnitario: valorUnitario,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao salvar foto no banco")
		return
	}

	writeJSON(w, http.StatusCreated, comURLFoto(foto))
}

// DELETE /api/albums/{id}/photos/{photoId}  — dono OU colaborador
func DeletePhoto(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de álbum inválido")
		return
	}
	photoID, err := strconv.ParseInt(r.PathValue("photoId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de foto inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "álbum não encontrado")
		return
	}
	isOwner := album.IDFotografo == userID
	if !isOwner {
		isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID)
		if !isCollab {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
	}

	foto, err := queries.ObterFotografiaPorID(r.Context(), photoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "foto não encontrada")
		return
	}
	if foto.IDAlbum != albumID {
		writeError(w, http.StatusForbidden, "esta foto não pertence ao álbum informado")
		return
	}
	// Dono do álbum pode remover qualquer foto; colaborador só as que ele enviou.
	if !isOwner && foto.IDFotografo != userID {
		writeError(w, http.StatusForbidden, "acesso negado: esta foto pertence a outro fotógrafo")
		return
	}

	ctx := r.Context()

	// Se a foto já faz parte de algum pedido, não pode ser apagada de vez
	// (quebraria o histórico do pedido e os downloads de quem comprou).
	// Nesse caso fazemos soft delete (ativo = false) e mantemos os arquivos no S3.
	emPedido, err := queries.ContarPedidosDaFotografia(ctx, photoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao verificar pedidos da foto")
		return
	}

	if emPedido > 0 {
		if err := queries.DesativarFotografia(ctx, photoID); err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao remover foto do banco")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Sem pedidos: remove definitivamente do banco e os objetos do S3.
	if err := queries.DeletarFotografia(ctx, photoID); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover foto do banco")
		return
	}
	_ = s3client.DeleteObject(ctx, s3client.KeyFromURL(foto.UrlAlta))
	_ = s3client.DeleteObject(ctx, s3client.KeyFromURL(foto.UrlBaixa))

	w.WriteHeader(http.StatusNoContent)
}

// POST /api/albums/{id}/cover  — somente dono
func UploadCover(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	if _, err := albumDoFotografo(r, albumID, userID); err != nil {
		respondAlbumErr(w, err)
		return
	}

	data, _, _, err := readImageForm(w, r)
	if err != nil {
		return
	}

	img, _, err := imgproc.Decode(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo de imagem inválido")
		return
	}

	resized := imgproc.Resize(img, 1200)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 85}); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao processar capa")
		return
	}

	coverKey := s3client.AlbumCoverKey(albumID)
	if err := s3client.UploadPublic(r.Context(), coverKey, buf.Bytes(), "image/jpeg"); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao enviar capa para o S3")
		return
	}

	// Grava apenas a key; a URL é montada na leitura.
	album, err := db.GetQueries().AtualizarCapaAlbum(r.Context(), albumID, coverKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao atualizar capa no banco")
		return
	}
	writeJSON(w, http.StatusOK, comURLAlbum(album))
}

// POST /api/albums/{id}/invites/{inviteId}/resend — somente dono
func ResendInvite(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	inviteID, err := strconv.ParseInt(r.PathValue("inviteId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := albumDoFotografo(r, albumID, userID)
	if err != nil {
		respondAlbumErr(w, err)
		return
	}

	convites, err := queries.ListarConvitesPorAlbum(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar convites")
		return
	}

	var convite *db.Convite
	for i, c := range convites {
		if c.ID == inviteID {
			convite = &convites[i]
			break
		}
	}
	if convite == nil || convite.UsedAt.Valid {
		writeError(w, http.StatusNotFound, "convite não encontrado ou já utilizado")
		return
	}

	capaURL := ""
	if album.CapaUrl.Valid {
		capaURL = s3client.PublicURL(album.CapaUrl.String)
	}
	go func() {
		if err := email.EnviarConviteAlbum(convite.Email, album.Titulo, convite.Token, capaURL); err != nil {
			fmt.Printf("[WARN] falha ao reenviar email para %s: %v\n", convite.Email, err)
		}
	}()

	writeJSON(w, http.StatusOK, map[string]any{"status": "enviado"})
}

// PUT /api/albums/{id}/photos/reorder — dono OU colaborador fotógrafo
func ReorderPhotos(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "álbum não encontrado")
		return
	}
	if album.IDFotografo != userID {
		isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID)
		if !isCollab {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
	}

	var req struct {
		FotoIDs []int64 `json:"foto_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	for i, id := range req.FotoIDs {
		_ = queries.AtualizarOrdemFotografia(r.Context(), id, albumID, int64(i+1))
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/albums/{id}/photos — paginado
func ListPhotos(w http.ResponseWriter, r *http.Request) {
	albumID, err := parseAlbumID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	tipo, _ := r.Context().Value(middleware.UserTipoKey).(string)

	queries := db.GetQueries()
	album, err := queries.ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		writeError(w, http.StatusNotFound, "álbum não encontrado")
		return
	}

	if tipo == "cliente" {
		hasAccess, _ := queries.VerificarAcessoClienteAlbum(r.Context(), userID, albumID)
		if !hasAccess {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
	} else if album.IDFotografo != userID {
		isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID)
		if !isCollab {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
	}

	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	const pageSize = 50

	fotos, err := queries.ListarFotografiasPorAlbumPaginado(r.Context(), albumID, pageSize, offset)
	if err != nil || fotos == nil {
		fotos = []db.Fotografia{}
	}
	total, _ := queries.ContarFotografiasPorAlbum(r.Context(), albumID)

	if tipo == "cliente" {
		fotosCliente := make([]fotoPublicaDTO, len(fotos))
		for i, f := range fotos {
			fotosCliente[i] = fotoPublicaDTO{ID: f.ID, UrlBaixa: s3client.PublicURL(f.UrlBaixa), CriadoEm: f.CriadoEm}
		}
		writeJSON(w, http.StatusOK, map[string]any{"fotos": fotosCliente, "total": total, "offset": offset})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"fotos": comURLFotos(fotos), "total": total, "offset": offset})
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// comURLAlbum monta a URL pública da capa a partir da key gravada no banco.
func comURLAlbum(a db.Albun) db.Albun {
	if a.CapaUrl.Valid && a.CapaUrl.String != "" {
		a.CapaUrl.String = s3client.PublicURL(a.CapaUrl.String)
	}
	return a
}

func comURLAlbuns(albuns []db.Albun) []db.Albun {
	out := make([]db.Albun, len(albuns))
	for i, a := range albuns {
		out[i] = comURLAlbum(a)
	}
	return out
}

// comURLFoto monta as URLs públicas da foto a partir das keys gravadas no banco.
func comURLFoto(f db.Fotografia) db.Fotografia {
	f.UrlAlta = s3client.PublicURL(f.UrlAlta)
	f.UrlBaixa = s3client.PublicURL(f.UrlBaixa)
	return f
}

func comURLFotos(fotos []db.Fotografia) []db.Fotografia {
	out := make([]db.Fotografia, len(fotos))
	for i, f := range fotos {
		out[i] = comURLFoto(f)
	}
	return out
}

var errAlbumNotFound = errors.New("not_found")
var errForbidden = errors.New("forbidden")

func albumDoFotografo(r *http.Request, albumID, userID int64) (db.Albun, error) {
	album, err := db.GetQueries().ObterAlbumPorID(r.Context(), albumID)
	if err != nil {
		return db.Albun{}, errAlbumNotFound
	}
	if album.IDFotografo != userID {
		return db.Albun{}, errForbidden
	}
	return album, nil
}

func respondAlbumErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errAlbumNotFound):
		writeError(w, http.StatusNotFound, "álbum não encontrado")
	case errors.Is(err, errForbidden):
		writeError(w, http.StatusForbidden, "acesso negado")
	default:
		writeError(w, http.StatusInternalServerError, "erro interno")
	}
}

func readImageForm(w http.ResponseWriter, r *http.Request) (data []byte, contentType, ext string, err error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize+1024)
	if parseErr := r.ParseMultipartForm(maxFileSize); parseErr != nil {
		var mbe *http.MaxBytesError
		if errors.As(parseErr, &mbe) {
			writeError(w, http.StatusRequestEntityTooLarge, "arquivo muito grande (máximo 10 MB)")
		} else {
			writeError(w, http.StatusBadRequest, "erro ao processar formulário")
		}
		return nil, "", "", parseErr
	}

	file, _, fileErr := r.FormFile("file")
	if fileErr != nil {
		writeError(w, http.StatusBadRequest, "campo 'file' é obrigatório")
		return nil, "", "", fileErr
	}
	defer file.Close()

	data, readErr := io.ReadAll(file)
	if readErr != nil {
		writeError(w, http.StatusBadRequest, "erro ao ler arquivo")
		return nil, "", "", readErr
	}

	contentType = http.DetectContentType(data)
	var ok bool
	ext, ok = allowedTypes[contentType]
	if !ok {
		writeError(w, http.StatusBadRequest, "tipo de arquivo não permitido")
		return nil, "", "", errors.New("unsupported type")
	}
	return data, contentType, ext, nil
}

func parseAlbumID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}
