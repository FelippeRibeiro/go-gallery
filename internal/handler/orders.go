package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	"github.com/FelippeRibeiro/go-gallery/internal/email"
	"github.com/FelippeRibeiro/go-gallery/internal/middleware"
	"github.com/FelippeRibeiro/go-gallery/internal/payment"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
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

	// Validação anti-duplicidade: remove fotos que o cliente já comprou
	// (pedido com status 'pago') neste álbum. Rede de segurança — o frontend
	// já oculta essas fotos, mas o pedido também é validado aqui.
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

	// Email de pedido criado (aguardando pagamento) — async.
	usuario, _ := queries.ObterUsuarioPorID(r.Context(), userID)
	downloadURL := fmt.Sprintf("%s/orders/%d", appBaseURL(), pedido.ID)
	go func() {
		if err := email.EnviarConfirmacaoPedido(usuario.Email, album.Titulo, downloadURL); err != nil {
			fmt.Printf("[WARN] falha ao enviar email de pedido: %v\n", err)
		}
	}()

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
		ExternalReference: strconv.FormatInt(pedido.ID, 10),
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
		ExternalReference: strconv.FormatInt(pedido.ID, 10),
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

	// Download liberado apenas para pedidos pagos.
	if pedido.Status != "pago" {
		writeError(w, http.StatusForbidden, "pagamento ainda não confirmado")
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

// POST /api/webhooks/mercadopago — notificação de pagamento do Mercado Pago.
// Endpoint público (sem autenticação). O MP envia o tipo e o ID do recurso via
// query params e/ou corpo JSON; consultamos o pagamento, resolvemos o pedido
// pelo external_reference e atualizamos o status.
func MercadoPagoWebhook(w http.ResponseWriter, r *http.Request) {
	fmt.Println("MercadoPagoWebhook")
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

	pedidoID, err := strconv.ParseInt(pay.ExternalReference, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	queries := db.GetQueries()
	pedido, err := queries.ObterPedidoPorID(r.Context(), pedidoID)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	novoStatus := payment.MapStatus(pay.Status)
	if err := queries.RegistrarPagamentoPedido(r.Context(), pedidoID, novoStatus, strconv.FormatInt(pay.ID, 10)); err != nil {
		fmt.Printf("[WARN] webhook MP: falha ao atualizar pedido %d: %v\n", pedidoID, err)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Email de confirmação apenas na transição para 'pago' (evita duplicatas).
	if novoStatus == "pago" && pedido.Status != "pago" {
		usuario, _ := queries.ObterUsuarioPorID(r.Context(), pedido.IDCliente)
		album, _ := queries.ObterAlbumPorID(r.Context(), pedido.IDAlbum)
		downloadURL := fmt.Sprintf("%s/orders/%d", appBaseURL(), pedido.ID)
		go func() {
			if err := email.EnviarConfirmacaoPedido(usuario.Email, album.Titulo, downloadURL); err != nil {
				fmt.Printf("[WARN] falha ao enviar email de pagamento: %v\n", err)
			}
		}()
	}

	w.WriteHeader(http.StatusOK)
}
