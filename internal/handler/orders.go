package handler

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/email"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	"github.com/FelippeRibeiro/go-gallery/internal/payment"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/google/uuid"
)

// appBaseURL devolve a URL pública do frontend (APP_URL), usada nas back_urls.
func appBaseURL() string {
	if u := os.Getenv("APP_URL"); u != "" {
		return u
	}
	return "http://localhost:5173"
}

// apiBaseURL devolve a URL pública do backend (API_URL), usada na notification_url
// do webhook do Mercado Pago. Em produção deve ser um domínio acessível externamente.
func apiBaseURL() string {
	if u := os.Getenv("API_URL"); u != "" {
		return u
	}
	return "http://localhost:8080"
}

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

	// Dono e colaboradores não compram o próprio álbum.
	if album.IDFotografo == userID {
		writeError(w, http.StatusForbidden, "você não pode comprar o seu próprio álbum")
		return
	}
	if isCollab, _ := queries.VerificarFotografoAlbum(r.Context(), userID, albumID); isCollab {
		writeError(w, http.StatusForbidden, "colaboradores não podem comprar fotos do álbum em que trabalham")
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
		// Venda completa: comprou uma vez, é dono do álbum inteiro — inclusive
		// das fotos adicionadas depois. Não existe recompra.
		comprou, _ := queries.ClienteComprouAlbum(r.Context(), userID, albumID)
		if comprou {
			writeError(w, http.StatusConflict, "você já comprou este álbum — todas as fotos estão disponíveis nos seus pedidos")
			return
		}
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

	// Validação anti-duplicidade (somente venda por foto): remove fotos que o
	// cliente já comprou (pedido com status 'pago') neste álbum. Rede de
	// segurança — o frontend já bloqueia essas fotos, mas o pedido também é
	// validado aqui.
	if !album.Lote {
		compradas, _ := queries.ListarFotoIDsCompradas(r.Context(), userID, albumID)
		if len(compradas) > 0 {
			jaComprada := make(map[int64]bool, len(compradas))
			for _, id := range compradas {
				jaComprada[id] = true
			}
			filtradas := fotosParaPedido[:0]
			for _, f := range fotosParaPedido {
				if !jaComprada[f.ID] {
					filtradas = append(filtradas, f)
				}
			}
			fotosParaPedido = filtradas
			if len(fotosParaPedido) == 0 {
				writeError(w, http.StatusConflict, "você já comprou essas fotos")
				return
			}
		}
	}

	// Total calculado a partir da MESMA fonte usada nos itens (valor_unitario de
	// cada foto), garantindo que a soma dos itens bata com o total cobrado.
	// Para lote, o total é o preço fechado do álbum e os itens valem 0.00.
	var valorTotal string
	if album.Lote {
		valorTotal = album.ValorAlbum
	} else {
		var soma float64
		for _, f := range fotosParaPedido {
			v, _ := strconv.ParseFloat(f.ValorUnitario, 64)
			soma += v
		}
		valorTotal = fmt.Sprintf("%.2f", soma)
	}

	// Pedido + itens em transação: o total cobrado nunca fica sem os itens
	// correspondentes (e vice-versa).
	tx, err := db.GetDB().BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar pedido")
		return
	}
	defer tx.Rollback() //nolint:errcheck

	qtx := queries.WithTx(tx)
	pedido, err := qtx.CriarPedido(r.Context(), userID, albumID, valorTotal, uuid.NewString())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar pedido")
		return
	}

	for _, f := range fotosParaPedido {
		v := f.ValorUnitario
		if album.Lote {
			v = "0.00"
		}
		if err := qtx.CriarPedidoFoto(r.Context(), pedido.ID, f.ID, v); err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao criar pedido")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao criar pedido")
		return
	}

	// NÃO enviamos email aqui: o pedido foi apenas criado (status 'pendente'),
	// ainda sem pagamento. O email "Pedido confirmado" é enviado somente quando
	// o pagamento é aprovado, em aplicarStatusPagamento (transição para 'pago').
	//
	// O pagamento é iniciado em seguida pelo frontend, na tela do pedido, onde o
	// cliente escolhe o método: Checkout Pro (POST /api/orders/{id}/checkout) ou
	// PIX embutido (POST /api/orders/{id}/pix).
	writeJSON(w, http.StatusCreated, map[string]any{
		"pedido_id":   pedido.ID,
		"valor_total": pedido.ValorTotal,
		"fotos":       len(fotosParaPedido),
	})
}

// carregarPedidoDoCliente busca o pedido e valida que pertence ao usuário e está
// aguardando pagamento. Centraliza a validação dos endpoints de checkout.
func carregarPedidoDoCliente(w http.ResponseWriter, r *http.Request) (*db.Pedido, bool) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return nil, false
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	pedido, err := db.GetQueries().ObterPedidoPorID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pedido não encontrado")
		} else {
			writeError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		}
		return nil, false
	}
	if pedido.IDCliente != userID {
		writeError(w, http.StatusForbidden, "acesso negado")
		return nil, false
	}
	// Sincroniza com o MP antes de aceitar um novo pagamento: se o pedido já
	// foi pago e o webhook ainda não chegou, evita uma cobrança duplicada.
	reconciliarPagamento(r.Context(), &pedido)
	if pedido.Status == "pago" {
		writeError(w, http.StatusConflict, "pedido já foi pago")
		return nil, false
	}
	return &pedido, true
}

// POST /api/orders/{id}/checkout — cria a preference do Checkout Pro (redirect)
// e devolve o init_point.
func CreateCheckout(w http.ResponseWriter, r *http.Request) {
	if !payment.Configured() {
		writeError(w, http.StatusServiceUnavailable, "pagamento não configurado")
		return
	}
	pedido, ok := carregarPedidoDoCliente(w, r)
	if !ok {
		return
	}

	queries := db.GetQueries()
	album, _ := queries.ObterAlbumPorID(r.Context(), pedido.IDAlbum)
	precoTotal, _ := strconv.ParseFloat(pedido.ValorTotal, 64)
	appURL := appBaseURL()

	pref, err := payment.CreatePreference(r.Context(), payment.PreferenceRequest{
		Items: []payment.Item{{
			Title:      album.Titulo,
			Quantity:   1,
			UnitPrice:  precoTotal,
			CurrencyID: "BRL",
		}},
		ExternalReference: referenciaPedido(pedido),
		BackURLs: payment.BackURLs{
			Success: fmt.Sprintf("%s/orders/%d", appURL, pedido.ID),
			Failure: fmt.Sprintf("%s/orders/%d", appURL, pedido.ID),
			Pending: fmt.Sprintf("%s/orders/%d", appURL, pedido.ID),
		},
		AutoReturn:      "approved",
		NotificationURL: apiBaseURL() + "/api/webhooks/mercadopago",
	})
	if err != nil {
		fmt.Printf("[WARN] falha ao criar preference do Mercado Pago: %v\n", err)
		writeError(w, http.StatusBadGateway, "erro ao iniciar checkout")
		return
	}

	_ = queries.AtualizarPreferenciaPedido(r.Context(), pedido.ID, pref.ID)
	writeJSON(w, http.StatusOK, map[string]any{"init_point": pref.CheckoutURL()})
}

// POST /api/orders/{id}/pix — cria um pagamento PIX e devolve os dados do QR Code.
func CreatePix(w http.ResponseWriter, r *http.Request) {
	if !payment.Configured() {
		writeError(w, http.StatusServiceUnavailable, "pagamento não configurado")
		return
	}
	pedido, ok := carregarPedidoDoCliente(w, r)
	if !ok {
		return
	}

	queries := db.GetQueries()
	album, _ := queries.ObterAlbumPorID(r.Context(), pedido.IDAlbum)
	usuario, _ := queries.ObterUsuarioPorID(r.Context(), pedido.IDCliente)
	precoTotal, _ := strconv.ParseFloat(pedido.ValorTotal, 64)

	pix, err := payment.CreatePixPayment(r.Context(), payment.PixRequest{
		TransactionAmount: precoTotal,
		Description:       fmt.Sprintf("Pedido #%d — %s", pedido.ID, album.Titulo),
		Payer:             payment.PixPayer{Email: usuario.Email, FirstName: usuario.Nome},
		ExternalReference: referenciaPedido(pedido),
		NotificationURL:   apiBaseURL() + "/api/webhooks/mercadopago",
	})
	if err != nil {
		fmt.Printf("[WARN] falha ao criar pagamento PIX: %v\n", err)
		// Surfaca a mensagem do Mercado Pago para facilitar o diagnóstico
		// (ex.: chave PIX ausente, pagador = vendedor, credenciais test/prod).
		writeError(w, http.StatusBadGateway, "erro ao gerar PIX: "+err.Error())
		return
	}

	// Guarda o id do pagamento e o status inicial (normalmente 'pendente').
	_ = queries.RegistrarPagamentoPedido(r.Context(), pedido.ID, payment.MapStatus(pix.Status), strconv.FormatInt(pix.ID, 10))

	writeJSON(w, http.StatusOK, map[string]any{
		"qr_code":        pix.QRCode,
		"qr_code_base64": pix.QRCodeBase64,
		"ticket_url":     pix.TicketURL,
		"status":         payment.MapStatus(pix.Status),
	})
}

// GET /api/orders — lista todos os pedidos do usuário autenticado
func ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	queries := db.GetQueries()
	pedidos, err := queries.ListarPedidosResumoPorCliente(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar pedidos")
		return
	}

	for i := range pedidos {
		if pedidos[i].CapaUrl.Valid {
			pedidos[i].CapaUrl.String = s3client.PublicURL(pedidos[i].CapaUrl.String)
		}
	}
	if pedidos == nil {
		pedidos = []db.PedidoResumo{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"pedidos": pedidos})
}

// fotosDoPedido devolve as fotos cobertas pelo pedido. Em álbuns de venda
// completa (lote) o cliente comprou o álbum inteiro: fotos adicionadas depois
// da compra também entram, unindo as fotos ativas do álbum aos itens do pedido
// (os itens preservam fotos desativadas que já foram compradas). Em venda por
// foto, são exatamente os itens do pedido.
func fotosDoPedido(ctx context.Context, pedido *db.Pedido) ([]db.PedidoFotoInfo, error) {
	queries := db.GetQueries()
	itens, err := queries.ListarFotosPedido(ctx, pedido.ID)
	if err != nil {
		return nil, err
	}

	album, err := queries.ObterAlbumPorID(ctx, pedido.IDAlbum)
	if err != nil || !album.Lote {
		return itens, nil
	}

	fotosAtuais, err := queries.ListarFotografiasPorAlbum(ctx, pedido.IDAlbum)
	if err != nil {
		return itens, nil
	}
	noPedido := make(map[int64]bool, len(itens))
	for _, it := range itens {
		noPedido[it.IDFotografia] = true
	}
	for _, f := range fotosAtuais {
		if noPedido[f.ID] {
			continue
		}
		itens = append(itens, db.PedidoFotoInfo{
			IDPedido:      pedido.ID,
			IDFotografia:  f.ID,
			ValorUnitario: "0.00",
			UrlAlta:       f.UrlAlta,
			UrlBaixa:      f.UrlBaixa,
		})
	}
	return itens, nil
}

// aplicarStatusPagamento grava o novo status do pedido e o id do pagamento,
// sem nunca rebaixar um pedido já pago (uma notificação atrasada de um
// pagamento cancelado não pode desfazer um pagamento confirmado). Na transição
// para 'pago' envia o email de confirmação. Atualiza pedido.Status in place.
func aplicarStatusPagamento(ctx context.Context, pedido *db.Pedido, novoStatus, paymentID string) {
	if pedido.Status == "pago" {
		return
	}
	queries := db.GetQueries()

	if novoStatus == "pago" {
		// Transição atômica no banco: apenas um chamador consegue mudar de
		// !pago -> pago. Sem isso, webhook do MP e polling do frontend chegando
		// juntos leem 'pendente' ao mesmo tempo e ambos enviam o email.
		mudou, err := queries.MarcarPedidoPagoSeNaoPago(ctx, pedido.ID, paymentID)
		if err != nil {
			fmt.Printf("[WARN] falha ao marcar pedido %d como pago: %v\n", pedido.ID, err)
			return
		}
		pedido.Status = "pago"
		if !mudou {
			// Outro processo já confirmou este pagamento — efeitos já disparados.
			return
		}

		// Comprador passa a ter vínculo com o álbum (idempotente) — o álbum
		// aparece no dashboard dele como os de convite.
		_, _ = queries.AssociarClienteAlbum(ctx, pedido.IDCliente, pedido.IDAlbum)

		usuario, _ := queries.ObterUsuarioPorID(ctx, pedido.IDCliente)
		album, _ := queries.ObterAlbumPorID(ctx, pedido.IDAlbum)
		downloadURL := fmt.Sprintf("%s/orders/%d", appBaseURL(), pedido.ID)
		go func() {
			if err := email.EnviarConfirmacaoPedido(usuario.Email, album.Titulo, downloadURL); err != nil {
				fmt.Printf("[WARN] falha ao enviar email de pagamento: %v\n", err)
			}
		}()
		return
	}

	// Status não-pago (pendente/falhou): grava sem rebaixar um pedido já pago.
	if err := queries.RegistrarPagamentoPedido(ctx, pedido.ID, novoStatus, paymentID); err != nil {
		fmt.Printf("[WARN] falha ao atualizar status do pedido %d: %v\n", pedido.ID, err)
		return
	}
	pedido.Status = novoStatus
}

// referenciaPedido devolve a external_reference do pedido: o UUID gravado em
// `referencia` ou, para pedidos antigos (pré-migração), o ID numérico.
func referenciaPedido(p *db.Pedido) string {
	if p.Referencia.Valid && p.Referencia.String != "" {
		return p.Referencia.String
	}
	return strconv.FormatInt(p.ID, 10)
}

// pagamentoConfere valida que o pagamento corresponde ao pedido: referência e
// valor exatos. Sem isso, um pagamento antigo da mesma conta MP com a mesma
// external_reference (ex.: IDs numéricos reaproveitados após reset do banco)
// marcaria o pedido errado como pago.
func pagamentoConfere(pedido *db.Pedido, pay *payment.Payment) bool {
	if pay.ExternalReference != referenciaPedido(pedido) {
		return false
	}
	total, err := strconv.ParseFloat(pedido.ValorTotal, 64)
	if err != nil {
		return false
	}
	diff := pay.TransactionAmount - total
	if diff < 0 {
		diff = -diff
	}
	if diff >= 0.01 {
		fmt.Printf("[WARN] pagamento %d ignorado: valor %.2f difere do pedido %d (%.2f)\n",
			pay.ID, pay.TransactionAmount, pedido.ID, total)
		return false
	}
	return true
}

// reconciliarPagamento consulta o Mercado Pago quando o pedido ainda está
// pendente e sincroniza o status local. Permite confirmar o pagamento pelo
// polling do frontend mesmo quando o webhook não alcança o servidor (ex.: dev
// em localhost, sem túnel público).
func reconciliarPagamento(ctx context.Context, pedido *db.Pedido) {
	if pedido.Status != "pendente" || !payment.Configured() {
		return
	}

	var pay *payment.Payment
	var err error
	if pedido.MpPaymentID.Valid && pedido.MpPaymentID.String != "" {
		pay, err = payment.GetPayment(ctx, pedido.MpPaymentID.String)
	} else if pedido.MpPreferenceID.Valid && pedido.Referencia.Valid {
		// Checkout Pro: o payment id só chegaria pelo webhook — busca pela
		// referência UUID. Pedidos legados sem referência não usam a busca:
		// o ID numérico colide com pagamentos antigos da conta.
		pay, err = payment.FindPaymentByExternalReference(ctx, pedido.Referencia.String)
	}
	if err != nil || pay == nil {
		return
	}
	if !pagamentoConfere(pedido, pay) {
		return
	}

	aplicarStatusPagamento(ctx, pedido, payment.MapStatus(pay.Status), strconv.FormatInt(pay.ID, 10))
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

	// Pedido pendente: confere o status direto no Mercado Pago (o frontend
	// faz polling desta rota enquanto espera a confirmação).
	reconciliarPagamento(r.Context(), &pedido)

	fotos, _ := fotosDoPedido(r.Context(), &pedido)
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

// carregarPedidoPago busca o pedido, valida que pertence ao usuário e exige
// status 'pago' — sincronizando com o Mercado Pago antes de negar (o webhook
// pode não ter chegado). Centraliza a validação dos endpoints de download.
func carregarPedidoPago(w http.ResponseWriter, r *http.Request) (*db.Pedido, bool) {
	orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return nil, false
	}
	userID := r.Context().Value(middleware.UserIDKey).(int64)

	pedido, err := db.GetQueries().ObterPedidoPorID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pedido não encontrado")
		} else {
			writeError(w, http.StatusInternalServerError, "erro ao buscar pedido")
		}
		return nil, false
	}
	if pedido.IDCliente != userID {
		writeError(w, http.StatusForbidden, "acesso negado")
		return nil, false
	}

	reconciliarPagamento(r.Context(), &pedido)
	if pedido.Status != "pago" {
		writeError(w, http.StatusForbidden, "pagamento ainda não confirmado")
		return nil, false
	}
	return &pedido, true
}

// GET /api/orders/{id}/downloads — gera URLs pré-assinadas para download
func GetDownloadLinks(w http.ResponseWriter, r *http.Request) {
	pedido, ok := carregarPedidoPago(w, r)
	if !ok {
		return
	}

	fotos, err := fotosDoPedido(r.Context(), pedido)
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

// nomeArquivoSeguro reduz um texto a um nome de arquivo ASCII seguro para o
// header Content-Disposition (minúsculas, alfanumérico e hífens).
func nomeArquivoSeguro(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

// GET /api/orders/{id}/download — baixa todas as fotos do pedido num único ZIP,
// transmitido pelo backend (o link do bucket nunca chega ao navegador).
func DownloadOrderZip(w http.ResponseWriter, r *http.Request) {
	pedido, ok := carregarPedidoPago(w, r)
	if !ok {
		return
	}

	fotos, err := fotosDoPedido(r.Context(), pedido)
	if err != nil || len(fotos) == 0 {
		writeError(w, http.StatusNotFound, "nenhuma foto disponível para download")
		return
	}

	album, _ := db.GetQueries().ObterAlbumPorID(r.Context(), pedido.IDAlbum)
	nomeZip := fmt.Sprintf("pedido-%d.zip", pedido.ID)
	if slug := nomeArquivoSeguro(album.Titulo); slug != "" {
		nomeZip = fmt.Sprintf("%s-pedido-%d.zip", slug, pedido.ID)
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+nomeZip+`"`)

	// Zip em streaming: cada original é copiado do S3 direto para a resposta.
	// JPEG já é comprimido, então as entradas usam Store (sem deflate).
	zw := zip.NewWriter(w)
	defer zw.Close()
	for _, f := range fotos {
		key := s3client.KeyFromURL(f.UrlAlta)
		body, _, _, err := s3client.GetObject(r.Context(), key)
		if err != nil {
			fmt.Printf("[WARN] zip do pedido %d: falha ao ler %s: %v\n", pedido.ID, key, err)
			continue
		}
		entry, err := zw.CreateHeader(&zip.FileHeader{
			Name:     fmt.Sprintf("foto-%d%s", f.IDFotografia, path.Ext(key)),
			Method:   zip.Store,
			Modified: time.Now(),
		})
		if err == nil {
			_, err = io.Copy(entry, body)
		}
		body.Close()
		if err != nil {
			// Resposta já iniciada — não dá para sinalizar erro HTTP; encerra.
			fmt.Printf("[WARN] zip do pedido %d: stream interrompido: %v\n", pedido.ID, err)
			return
		}
	}
}

// GET /api/orders/{id}/download/{fotoId} — baixa uma foto original do pedido,
// transmitida pelo backend com Content-Disposition: attachment.
func DownloadOrderPhoto(w http.ResponseWriter, r *http.Request) {
	pedido, ok := carregarPedidoPago(w, r)
	if !ok {
		return
	}
	fotoID, err := strconv.ParseInt(r.PathValue("fotoId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id de foto inválido")
		return
	}

	// A foto precisa estar coberta pelo pedido (itens ou, em álbum lote, fotos
	// atuais do álbum — mesma regra de fotosDoPedido).
	fotos, err := fotosDoPedido(r.Context(), pedido)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar fotos")
		return
	}
	var alvo *db.PedidoFotoInfo
	for i := range fotos {
		if fotos[i].IDFotografia == fotoID {
			alvo = &fotos[i]
			break
		}
	}
	if alvo == nil {
		writeError(w, http.StatusNotFound, "foto não pertence a este pedido")
		return
	}

	key := s3client.KeyFromURL(alvo.UrlAlta)
	body, contentType, contentLength, err := s3client.GetObject(r.Context(), key)
	if err != nil {
		writeError(w, http.StatusBadGateway, "erro ao buscar arquivo")
		return
	}
	defer body.Close()

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="foto-%d%s"`, fotoID, path.Ext(key)))
	if contentLength > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}
	_, _ = io.Copy(w, body)
}

// POST /api/webhooks/mercadopago — notificação de pagamento do Mercado Pago.
// Endpoint público (sem autenticação). O MP envia o tipo e o ID do recurso via
// query params e/ou corpo JSON; consultamos o pagamento, resolvemos o pedido
// pelo external_reference e atualizamos o status.
func MercadoPagoWebhook(w http.ResponseWriter, r *http.Request) {
	// Validação de assinatura (opt-in via MP_WEBHOOK_SECRET). Mesmo sem ela o
	// fluxo é seguro: o status vem sempre de GetPayment na API do MP, nunca do
	// corpo da notificação.
	if !payment.VerifyWebhookSignature(
		r.Header.Get("x-signature"),
		r.Header.Get("x-request-id"),
		r.URL.Query().Get("data.id"),
	) {
		fmt.Println("[WARN] webhook MP: assinatura inválida — notificação ignorada")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Responde 200 sempre que possível — o MP reenvia em caso de erro.
	tipo := r.URL.Query().Get("type")
	if tipo == "" {
		tipo = r.URL.Query().Get("topic")
	}
	paymentID := r.URL.Query().Get("data.id")
	if paymentID == "" {
		paymentID = r.URL.Query().Get("id")
	}

	// Notificações mais recentes trazem os dados no corpo JSON.
	if paymentID == "" || tipo == "" {
		var body struct {
			Type   string `json:"type"`
			Action string `json:"action"`
			Data   struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if tipo == "" {
				tipo = body.Type
			}
			if paymentID == "" {
				paymentID = body.Data.ID
			}
			if tipo == "" && strings.HasPrefix(body.Action, "payment") {
				tipo = "payment"
			}
		}
	}

	// Só tratamos notificações de pagamento.
	if tipo != "payment" || paymentID == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	pay, err := payment.GetPayment(r.Context(), paymentID)
	if err != nil {
		fmt.Printf("[WARN] webhook MP: falha ao consultar pagamento %s: %v\n", paymentID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Resolve o pedido pela referência UUID; pedidos legados (pré-migração)
	// usavam o ID numérico como external_reference.
	queries := db.GetQueries()
	pedido, err := queries.ObterPedidoPorReferencia(r.Context(), sql.NullString{String: pay.ExternalReference, Valid: true})
	if err != nil {
		pedidoID, perr := strconv.ParseInt(pay.ExternalReference, 10, 64)
		if perr != nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		pedido, err = queries.ObterPedidoPorID(r.Context(), pedidoID)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if !pagamentoConfere(&pedido, pay) {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Aplica o status sem rebaixar pedidos pagos e envia o email de
	// confirmação na transição para 'pago'.
	aplicarStatusPagamento(r.Context(), &pedido, payment.MapStatus(pay.Status), strconv.FormatInt(pay.ID, 10))

	w.WriteHeader(http.StatusOK)
}
