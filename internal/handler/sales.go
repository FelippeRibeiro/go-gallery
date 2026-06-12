package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
)

// O saldo considera apenas álbuns dos quais o fotógrafo é dono — não há
// divisão de receita com colaboradores.

// GET /api/sales — pedidos feitos nos álbuns do fotógrafo + resumo de saldo
func ListSales(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)
	tipo, _ := r.Context().Value(middleware.UserTipoKey).(string)
	if tipo != "fotografo" {
		writeError(w, http.StatusForbidden, "apenas fotógrafos têm vendas")
		return
	}

	queries := db.GetQueries()
	vendas, err := queries.ListarVendasPorFotografo(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar vendas")
		return
	}
	resumo, err := queries.ResumoVendasFotografo(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao calcular saldo")
		return
	}

	for i := range vendas {
		if vendas[i].CapaUrl.Valid {
			vendas[i].CapaUrl.String = s3client.PublicURL(vendas[i].CapaUrl.String)
		}
	}
	if vendas == nil {
		vendas = []db.VendaInfo{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"resumo":  resumo,
		"pedidos": vendas,
	})
}

// GET /api/sales/{id} — detalhe de uma venda: comprador e fotos do pedido.
// Apenas o dono do álbum do pedido pode ver.
func GetSale(w http.ResponseWriter, r *http.Request) {
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

	album, err := queries.ObterAlbumPorID(r.Context(), pedido.IDAlbum)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar álbum")
		return
	}
	if album.IDFotografo != userID {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	cliente, err := queries.ObterUsuarioPorID(r.Context(), pedido.IDCliente)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao buscar comprador")
		return
	}

	// Itens do pedido (o que foi comprado), sem a união de fotos posteriores
	// dos álbuns lote — aqui interessa a venda registrada, não o download.
	fotos, _ := queries.ListarFotosPedido(r.Context(), pedido.ID)
	if fotos == nil {
		fotos = []db.PedidoFotoInfo{}
	}
	for i := range fotos {
		fotos[i].UrlBaixa = s3client.PublicURL(fotos[i].UrlBaixa)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"pedido": pedido,
		"album":  comURLAlbum(album),
		// Usuario completo tem SenhaHash — expõe só o necessário.
		"cliente": map[string]any{
			"id":    cliente.ID,
			"nome":  cliente.Nome,
			"email": cliente.Email,
		},
		"fotos": fotos,
	})
}
