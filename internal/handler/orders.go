package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/email"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
)

// POST /api/albums/{id}/orders — cliente cria pedido
func CreateOrder(w http.ResponseWriter, r *http.Request) {
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
		} else {
			writeError(w, http.StatusInternalServerError, "erro ao buscar álbum")
		}
		return
	}

	// Verify client has access
	hasAccess, _ := queries.VerificarAcessoClienteAlbum(r.Context(), userID, albumID)
	if !hasAccess && !album.Publico {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	var req struct {
		FotoIDs []int64 `json:"foto_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo inválido")
		return
	}

	// For lote albums, get all photos; for per-photo, use selection
	var fotosParaPedido []db.Fotografia
	if album.Lote {
		fotosParaPedido, _ = queries.ListarFotografiasPorAlbum(r.Context(), albumID)
	} else {
		if len(req.FotoIDs) == 0 {
			writeError(w, http.StatusBadRequest, "selecione ao menos uma foto")
			return
		}
		for _, fid := range req.FotoIDs {
			f, err := queries.ObterFotografiaPorID(r.Context(), fid)
			if err != nil || f.IDAlbum != albumID {
				continue
			}
			fotosParaPedido = append(fotosParaPedido, f)
		}
	}

	if len(fotosParaPedido) == 0 {
		writeError(w, http.StatusBadRequest, "nenhuma foto válida selecionada")
		return
	}

	// Calculate total
	valorTotal := album.ValorAlbum
	if !album.Lote {
		total := float64(len(fotosParaPedido)) * func() float64 {
			v, _ := strconv.ParseFloat(album.ValorUnitarioFotografia, 64)
			return v
		}()
		valorTotal = fmt.Sprintf("%.2f", total)
	}

	pedido, err := queries.CriarPedido(r.Context(), userID, albumID, valorTotal)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar pedido")
		return
	}

	for _, f := range fotosParaPedido {
		v := f.ValorUnitario
		if album.Lote {
			v = "0.00"
		}
		_ = queries.CriarPedidoFoto(r.Context(), pedido.ID, f.ID, v)
	}

	// Send confirmation email async
	usuario, _ := queries.ObterUsuarioPorID(r.Context(), userID)
	appURL := func() string {
		u := os.Getenv("APP_URL")
		if u == "" {
			return "http://localhost:5173"
		}
		return u
	}()
	downloadURL := fmt.Sprintf("%s/orders/%d", appURL, pedido.ID)
	go func() {
		if err := email.EnviarConfirmacaoPedido(usuario.Email, album.Titulo, downloadURL); err != nil {
			fmt.Printf("[WARN] falha ao enviar email de pedido: %v\n", err)
		}
	}()

	writeJSON(w, http.StatusCreated, map[string]any{
		"pedido_id":   pedido.ID,
		"valor_total": pedido.ValorTotal,
		"fotos":       len(fotosParaPedido),
	})
}

// GET /api/orders/{id} — cliente vê pedido
func GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	pedido, err := queries.ObterPedidoPorID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pedido não encontrado")
		} else {
			writeError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		}
		return
	}

	if pedido.IDCliente != userID {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	fotos, _ := queries.ListarFotosPedido(r.Context(), orderID)
	if fotos == nil {
		fotos = []db.PedidoFotoInfo{}
	}
	for i := range fotos {
		fotos[i].UrlBaixa = s3client.PublicURL(fotos[i].UrlBaixa)
	}

	album, _ := queries.ObterAlbumPorID(r.Context(), pedido.IDAlbum)

	writeJSON(w, http.StatusOK, map[string]any{
		"pedido": pedido,
		"album":  comURLAlbum(album),
		"fotos":  fotos,
	})
}

// GET /api/orders/{id}/downloads — gera URLs pré-assinadas para download
func GetDownloadLinks(w http.ResponseWriter, r *http.Request) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	pedido, err := queries.ObterPedidoPorID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pedido não encontrado")
		} else {
			writeError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		}
		return
	}

	if pedido.IDCliente != userID {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	fotos, err := queries.ListarFotosPedido(r.Context(), orderID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar fotos")
		return
	}

	type DownloadLink struct {
		FotoID      int64  `json:"foto_id"`
		UrlBaixa    string `json:"url_baixa"`
		UrlDownload string `json:"url_download"`
	}

	links := make([]DownloadLink, 0, len(fotos))
	for _, f := range fotos {
		// url_alta guarda apenas a key (path) — usada direto para gerar o presign.
		presignedURL, err := s3client.PresignGetObject(r.Context(), s3client.KeyFromURL(f.UrlAlta), 24*time.Hour)
		if err != nil {
			presignedURL = ""
		}
		links = append(links, DownloadLink{
			FotoID:      f.IDFotografia,
			UrlBaixa:    s3client.PublicURL(f.UrlBaixa),
			UrlDownload: presignedURL,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"downloads": links})
}
