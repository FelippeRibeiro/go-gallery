// Package payment integra com o Mercado Pago. Suporta dois fluxos:
//
//   - Checkout Pro (redirect): CreatePreference cria uma "preference" e devolve
//     o init_point; o cliente é redirecionado ao site do MP, que mostra os
//     métodos e o QR do PIX.
//   - PIX transparente (embutido): CreatePixPayment cria um pagamento PIX e
//     devolve o QR Code (copia-e-cola + imagem base64) para renderizar na
//     própria UI.
//
// Em ambos, o MP notifica o webhook, que consulta o pagamento (GetPayment) e
// atualiza o status do pedido.
//
// Configuração via env: MP_ACCESS_TOKEN (token privado da conta) e API_URL
// (base pública do backend, usada para montar a notification_url do webhook).
package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	apiBase       = "https://api.mercadopago.com"
	preferencePth = "/checkout/preferences"
	paymentPth    = "/v1/payments/"
	paymentsPth   = "/v1/payments"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Configured informa se o Mercado Pago está habilitado (token presente).
func Configured() bool {
	return os.Getenv("MP_ACCESS_TOKEN") != ""
}

func accessToken() string { return os.Getenv("MP_ACCESS_TOKEN") }

// IsTest indica se devemos usar o ambiente de teste (sandbox). Controlado pela
// env MP_SANDBOX=true. Mantém também o fallback pelo prefixo "TEST-" (formato
// antigo) — nas credenciais novas do MP, tanto teste quanto produção usam o
// prefixo "APP_USR-", então o prefixo sozinho não é confiável.
func IsTest() bool {
	if v := strings.ToLower(os.Getenv("MP_SANDBOX")); v == "true" || v == "1" {
		return true
	}
	return strings.HasPrefix(accessToken(), "TEST-")
}

// ─── Preference (checkout) ────────────────────────────────────────────────────

type Item struct {
	Title      string  `json:"title"`
	Quantity   int     `json:"quantity"`
	UnitPrice  float64 `json:"unit_price"`
	CurrencyID string  `json:"currency_id"`
}

type BackURLs struct {
	Success string `json:"success"`
	Failure string `json:"failure"`
	Pending string `json:"pending"`
}

type PreferenceRequest struct {
	Items             []Item   `json:"items"`
	ExternalReference string   `json:"external_reference"`
	BackURLs          BackURLs `json:"back_urls"`
	AutoReturn        string   `json:"auto_return"`
	NotificationURL   string   `json:"notification_url,omitempty"`
}

type PreferenceResponse struct {
	ID          string `json:"id"`
	InitPoint   string `json:"init_point"`
	SandboxInit string `json:"sandbox_init_point"`
}

// CheckoutURL devolve o init_point apropriado: o de sandbox quando as credenciais
// são de teste, senão o de produção.
func (p *PreferenceResponse) CheckoutURL() string {
	if IsTest() && p.SandboxInit != "" {
		return p.SandboxInit
	}
	return p.InitPoint
}

// CreatePreference cria uma preference de checkout e devolve o init_point.
func CreatePreference(ctx context.Context, req PreferenceRequest) (*PreferenceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+preferencePth, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken())

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		return nil, fmt.Errorf("mercadopago preference falhou (%d): %s", res.StatusCode, buf.String())
	}

	var pref PreferenceResponse
	if err := json.NewDecoder(res.Body).Decode(&pref); err != nil {
		return nil, err
	}
	return &pref, nil
}

// ─── Payment (consulta para o webhook) ────────────────────────────────────────

type Payment struct {
	ID                int64   `json:"id"`
	Status            string  `json:"status"` // approved, pending, in_process, rejected, cancelled, refunded...
	ExternalReference string  `json:"external_reference"`
	TransactionAmount float64 `json:"transaction_amount"`
}

// GetPayment consulta um pagamento pelo ID.
func GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+paymentPth+paymentID, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken())

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		return nil, fmt.Errorf("mercadopago payment falhou (%d): %s", res.StatusCode, buf.String())
	}

	var p Payment
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// FindPaymentByExternalReference busca o pagamento mais relevante associado a
// um external_reference (id do pedido): prioriza um pagamento aprovado; na
// ausência, devolve o mais recente. Retorna (nil, nil) quando não há nenhum.
// Usado para reconciliar pedidos do Checkout Pro, em que o payment id só
// chegaria pelo webhook.
func FindPaymentByExternalReference(ctx context.Context, externalRef string) (*Payment, error) {
	u := apiBase + paymentsPth + "/search?sort=date_created&criteria=desc&external_reference=" + url.QueryEscape(externalRef)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken())

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		return nil, fmt.Errorf("mercadopago search falhou (%d): %s", res.StatusCode, buf.String())
	}

	var out struct {
		Results []Payment `json:"results"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Results) == 0 {
		return nil, nil
	}
	for i := range out.Results {
		if out.Results[i].Status == "approved" {
			return &out.Results[i], nil
		}
	}
	return &out.Results[0], nil
}

// ─── Assinatura do webhook ────────────────────────────────────────────────────

// VerifyWebhookSignature valida o header x-signature das notificações do
// Mercado Pago (HMAC-SHA256 do manifesto "id:...;request-id:...;ts:...;" com a
// chave secreta do webhook). Validação opt-in: sem MP_WEBHOOK_SECRET
// configurado, retorna sempre true.
func VerifyWebhookSignature(xSignature, xRequestID, dataID string) bool {
	secret := os.Getenv("MP_WEBHOOK_SECRET")
	if secret == "" {
		return true
	}

	var ts, v1 string
	for _, part := range strings.Split(xSignature, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "ts":
			ts = strings.TrimSpace(v)
		case "v1":
			v1 = strings.TrimSpace(v)
		}
	}
	if ts == "" || v1 == "" {
		return false
	}

	// Manifesto conforme a documentação do MP; segmentos com valor ausente são
	// omitidos. O data.id alfanumérico entra em minúsculas.
	manifest := ""
	if dataID != "" {
		manifest += "id:" + strings.ToLower(dataID) + ";"
	}
	if xRequestID != "" {
		manifest += "request-id:" + xRequestID + ";"
	}
	manifest += "ts:" + ts + ";"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(v1))
}

// ─── Pagamento PIX (checkout transparente) ───────────────────────────────────

type PixPayer struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name,omitempty"`
}

type PixRequest struct {
	TransactionAmount float64  `json:"transaction_amount"`
	Description       string   `json:"description"`
	PaymentMethodID   string   `json:"payment_method_id"` // sempre "pix"
	Payer             PixPayer `json:"payer"`
	ExternalReference string   `json:"external_reference"`
	NotificationURL   string   `json:"notification_url,omitempty"`
}

// PixResponse traz os dados do QR Code retornados pelo Mercado Pago.
type PixResponse struct {
	ID                int64  `json:"id"`
	Status            string `json:"status"`
	QRCode            string // copia-e-cola (EMV)
	QRCodeBase64      string // imagem PNG do QR em base64
	TicketURL         string // página do MP com o QR
	ExternalReference string
}

// raw da resposta do MP para extrair point_of_interaction.transaction_data.
type pixRawResponse struct {
	ID                 int64  `json:"id"`
	Status             string `json:"status"`
	ExternalReference  string `json:"external_reference"`
	PointOfInteraction struct {
		TransactionData struct {
			QRCode       string `json:"qr_code"`
			QRCodeBase64 string `json:"qr_code_base64"`
			TicketURL    string `json:"ticket_url"`
		} `json:"transaction_data"`
	} `json:"point_of_interaction"`
}

// CreatePixPayment cria um pagamento PIX e devolve os dados do QR Code.
func CreatePixPayment(ctx context.Context, req PixRequest) (*PixResponse, error) {
	req.PaymentMethodID = "pix"
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+paymentsPth, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken())
	// Chave de idempotência evita pagamentos duplicados em retries.
	httpReq.Header.Set("X-Idempotency-Key", uuid.NewString())

	res, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(res.Body)
		return nil, fmt.Errorf("mercadopago pix falhou (%d): %s", res.StatusCode, buf.String())
	}

	var raw pixRawResponse
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return &PixResponse{
		ID:                raw.ID,
		Status:            raw.Status,
		QRCode:            raw.PointOfInteraction.TransactionData.QRCode,
		QRCodeBase64:      raw.PointOfInteraction.TransactionData.QRCodeBase64,
		TicketURL:         raw.PointOfInteraction.TransactionData.TicketURL,
		ExternalReference: raw.ExternalReference,
	}, nil
}

// MapStatus converte o status do Mercado Pago no status interno do pedido.
func MapStatus(mpStatus string) string {
	switch mpStatus {
	case "approved":
		return "pago"
	case "pending", "in_process", "authorized":
		return "pendente"
	default: // rejected, cancelled, refunded, charged_back
		return "falhou"
	}
}
