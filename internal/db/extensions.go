package db

// Este arquivo NÃO é gerado pelo sqlc — é código manual que complementa o código gerado.
// Contém:
//  1. Type aliases para os Row types de JOINs (mantém compatibilidade com os handlers).
//  2. Wrappers exportados com a assinatura posicional original, delegando aos helpers
//     internos (minúsculos) gerados pelo sqlc que recebem params struct.
//
// Execute "sqlc generate" para regenerar os demais arquivos deste pacote.

import (
	"context"
	"database/sql"
)

// ─── Type aliases para resultados de JOIN ─────────────────────────────────────
// Criados após "sqlc generate", estes tipos existirão nos arquivos gerados.

type ClienteAlbumInfo = ListarClientesAlbumRow
type FotografoAlbumInfo = ListarFotografosAlbumRow
type PedidoFotoInfo = ListarFotosPedidoRow
type PedidoResumo = ListarPedidosResumoPorClienteRow
type FotografoPublico = ListarFotografosComAlbumPublicoRow

// ─── Wrappers com args posicionais ────────────────────────────────────────────

// VerificarFotografoAlbum retorna true se o fotógrafo é colaborador do álbum.
func (q *Queries) VerificarFotografoAlbum(ctx context.Context, idFotografo, idAlbum int64) (bool, error) {
	return q.verificarFotografoAlbum(ctx, verificarFotografoAlbumParams{
		IDFotografo: idFotografo,
		IDAlbum:     idAlbum,
	})
}

// VerificarAcessoClienteAlbum retorna true se o cliente tem acesso ao álbum.
func (q *Queries) VerificarAcessoClienteAlbum(ctx context.Context, idCliente, idAlbum int64) (bool, error) {
	return q.verificarAcessoClienteAlbum(ctx, verificarAcessoClienteAlbumParams{
		IDCliente: idCliente,
		IDAlbum:   idAlbum,
	})
}

// VerificarConvitePendente retorna true se existe convite válido e não utilizado.
func (q *Queries) VerificarConvitePendente(ctx context.Context, email string, idAlbum int64) (bool, error) {
	return q.verificarConvitePendente(ctx, verificarConvitePendenteParams{
		Email:   email,
		IDAlbum: idAlbum,
	})
}

// AssociarClienteAlbum insere o vínculo cliente↔álbum (ON CONFLICT DO NOTHING).
func (q *Queries) AssociarClienteAlbum(ctx context.Context, idCliente, idAlbum int64) (ClientesAlbun, error) {
	return q.associarClienteAlbum(ctx, associarClienteAlbumParams{
		IDCliente: idCliente,
		IDAlbum:   idAlbum,
	})
}

// AssociarFotografoAlbum insere o vínculo fotógrafo↔álbum (ON CONFLICT DO NOTHING).
func (q *Queries) AssociarFotografoAlbum(ctx context.Context, idFotografo, idAlbum int64) (FotografosAlbun, error) {
	return q.associarFotografoAlbum(ctx, associarFotografoAlbumParams{
		IDFotografo: idFotografo,
		IDAlbum:     idAlbum,
	})
}

// RemoverFotografoAlbum remove o vínculo fotógrafo↔álbum pelo ID da relação.
func (q *Queries) RemoverFotografoAlbum(ctx context.Context, id, idAlbum int64) error {
	return q.removerFotografoAlbum(ctx, removerFotografoAlbumParams{
		ID:      id,
		IDAlbum: idAlbum,
	})
}

// RemoverClienteAlbum remove o vínculo cliente↔álbum pelo ID da relação.
func (q *Queries) RemoverClienteAlbum(ctx context.Context, id, idAlbum int64) error {
	return q.removerClienteAlbum(ctx, removerClienteAlbumParams{
		ID:      id,
		IDAlbum: idAlbum,
	})
}

// RevogarConvite exclui um convite pelo ID e id_album.
func (q *Queries) RevogarConvite(ctx context.Context, id, idAlbum int64) error {
	return q.revogarConvite(ctx, revogarConviteParams{
		ID:      id,
		IDAlbum: idAlbum,
	})
}

// AtualizarPublicoAlbum alterna a visibilidade pública do álbum.
func (q *Queries) AtualizarPublicoAlbum(ctx context.Context, id int64, publico bool) (Albun, error) {
	return q.atualizarPublicoAlbum(ctx, atualizarPublicoAlbumParams{
		ID:      id,
		Publico: publico,
	})
}

// AtualizarCapaAlbum define a capa do álbum; capa_url é TEXT NULL no banco.
func (q *Queries) AtualizarCapaAlbum(ctx context.Context, id int64, capaUrl string) (Albun, error) {
	return q.atualizarCapaAlbum(ctx, atualizarCapaAlbumParams{
		ID:      id,
		CapaUrl: sql.NullString{String: capaUrl, Valid: capaUrl != ""},
	})
}

// AtualizarOrdemFotografia define a ordem de exibição de uma foto no álbum.
func (q *Queries) AtualizarOrdemFotografia(ctx context.Context, id, idAlbum, ordem int64) error {
	return q.atualizarOrdemFotografia(ctx, atualizarOrdemFotografiaParams{
		ID:      id,
		IDAlbum: idAlbum,
		Ordem:   ordem,
	})
}

// ListarFotografiasPorAlbumPaginado retorna fotos de um álbum com paginação.
func (q *Queries) ListarFotografiasPorAlbumPaginado(ctx context.Context, idAlbum, limit, offset int64) ([]Fotografia, error) {
	return q.listarFotografiasPorAlbumPaginado(ctx, listarFotografiasPorAlbumPaginadoParams{
		IDAlbum: idAlbum,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
}

// CriarPedido cria um pedido com status 'pendente'. A referência (UUID) é a
// external_reference usada no Mercado Pago — única por pedido, sobrevive a
// resets do banco (o ID numérico colide).
func (q *Queries) CriarPedido(ctx context.Context, idCliente, idAlbum int64, valorTotal, referencia string) (Pedido, error) {
	return q.criarPedido(ctx, criarPedidoParams{
		IDCliente:  idCliente,
		IDAlbum:    idAlbum,
		ValorTotal: valorTotal,
		Referencia: sql.NullString{String: referencia, Valid: referencia != ""},
	})
}

// CriarPedidoFoto associa uma foto a um pedido (ON CONFLICT DO NOTHING).
func (q *Queries) CriarPedidoFoto(ctx context.Context, idPedido, idFotografia int64, valorUnitario string) error {
	return q.criarPedidoFoto(ctx, criarPedidoFotoParams{
		IDPedido:      idPedido,
		IDFotografia:  idFotografia,
		ValorUnitario: valorUnitario,
	})
}

// ListarFotoIDsCompradas retorna os IDs das fotos que o cliente já comprou
// (pedido com status 'pago') em um álbum.
func (q *Queries) ListarFotoIDsCompradas(ctx context.Context, idCliente, idAlbum int64) ([]int64, error) {
	return q.listarFotoIDsCompradas(ctx, listarFotoIDsCompradasParams{
		IDCliente: idCliente,
		IDAlbum:   idAlbum,
	})
}

// ClienteComprouAlbum retorna true se o cliente tem algum pedido pago no álbum.
// Usado em álbuns de venda completa (lote): comprou uma vez, é dono do álbum
// inteiro — inclusive de fotos adicionadas depois.
func (q *Queries) ClienteComprouAlbum(ctx context.Context, idCliente, idAlbum int64) (bool, error) {
	return q.clienteComprouAlbum(ctx, clienteComprouAlbumParams{
		IDCliente: idCliente,
		IDAlbum:   idAlbum,
	})
}

// AtualizarPreferenciaPedido grava o ID da preference do Mercado Pago no pedido.
func (q *Queries) AtualizarPreferenciaPedido(ctx context.Context, id int64, preferenceID string) error {
	return q.atualizarPreferenciaPedido(ctx, atualizarPreferenciaPedidoParams{
		ID:             id,
		MpPreferenceID: sql.NullString{String: preferenceID, Valid: preferenceID != ""},
	})
}

// RegistrarPagamentoPedido atualiza o status do pedido e grava o ID do pagamento do MP.
func (q *Queries) RegistrarPagamentoPedido(ctx context.Context, id int64, status, paymentID string) error {
	return q.registrarPagamentoPedido(ctx, registrarPagamentoPedidoParams{
		ID:          id,
		Status:      status,
		MpPaymentID: sql.NullString{String: paymentID, Valid: paymentID != ""},
	})
}

// MarcarPedidoPagoSeNaoPago marca o pedido como 'pago' de forma atômica, apenas
// se ele ainda não estiver pago. Retorna true somente quando ESTA chamada
// efetuou a transição (linha afetada). Isso garante que os efeitos colaterais
// do pagamento (e-mail de confirmação, vínculo cliente↔álbum) ocorram uma única
// vez, mesmo quando o webhook do Mercado Pago e o polling do frontend chegam
// concorrentemente.
func (q *Queries) MarcarPedidoPagoSeNaoPago(ctx context.Context, id int64, paymentID string) (bool, error) {
	res, err := q.db.ExecContext(ctx,
		`UPDATE pedidos SET status = 'pago', mp_payment_id = $2 WHERE id = $1 AND status <> 'pago'`,
		id, sql.NullString{String: paymentID, Valid: paymentID != ""},
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// AtualizarBioFotografo salva o texto de bio; bio é TEXT NULL no banco.
func (q *Queries) AtualizarBioFotografo(ctx context.Context, id int64, bio string) error {
	return q.atualizarBioFotografo(ctx, atualizarBioFotografoParams{
		ID:  id,
		Bio: sql.NullString{String: bio, Valid: bio != ""},
	})
}

// AtualizarFotoPerfilFotografo salva a key/URL da foto de perfil; foto_perfil é TEXT NULL no banco.
func (q *Queries) AtualizarFotoPerfilFotografo(ctx context.Context, id int64, fotoPerfil string) error {
	return q.atualizarFotoPerfilFotografo(ctx, atualizarFotoPerfilFotografoParams{
		ID:         id,
		FotoPerfil: sql.NullString{String: fotoPerfil, Valid: fotoPerfil != ""},
	})
}
