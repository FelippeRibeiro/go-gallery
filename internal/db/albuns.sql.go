package db

import (
	"context"
	"database/sql"
	"time"
)

// scanAlbun lê todas as colunas de albuns (incluindo publico e capa_url).
func scanAlbun(row interface{ Scan(...any) error }, i *Albun) error {
	return row.Scan(
		&i.ID, &i.Titulo, &i.Descricao, &i.DataEvento, &i.CriadoEm,
		&i.Ativo, &i.IDFotografo, &i.Lote, &i.ValorAlbum, &i.ValorUnitarioFotografia,
		&i.Publico, &i.CapaUrl,
	)
}

const albumCols = `id, titulo, descricao, data_evento, criado_em, ativo, id_fotografo, lote, valor_album, valor_unitario_fotografia, publico, capa_url`

// ─── Álbuns ──────────────────────────────────────────────────────────────────

const criarAlbum = `
INSERT INTO albuns (titulo, descricao, data_evento, id_fotografo, lote, valor_album, valor_unitario_fotografia, publico)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING ` + albumCols

type CriarAlbumParams struct {
	Titulo                  string
	Descricao               sql.NullString
	DataEvento              time.Time
	IDFotografo             int64
	Lote                    bool
	ValorAlbum              string
	ValorUnitarioFotografia string
	Publico                 bool
}

func (q *Queries) CriarAlbum(ctx context.Context, arg CriarAlbumParams) (Albun, error) {
	row := q.db.QueryRowContext(ctx, criarAlbum,
		arg.Titulo, arg.Descricao, arg.DataEvento, arg.IDFotografo,
		arg.Lote, arg.ValorAlbum, arg.ValorUnitarioFotografia, arg.Publico,
	)
	var i Albun
	return i, scanAlbun(row, &i)
}

const listarAlbunsPorFotografo = `SELECT ` + albumCols + ` FROM albuns WHERE id_fotografo = $1 AND ativo = TRUE ORDER BY criado_em DESC`

func (q *Queries) ListarAlbunsPorFotografo(ctx context.Context, idFotografo int64) ([]Albun, error) {
	rows, err := q.db.QueryContext(ctx, listarAlbunsPorFotografo, idFotografo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Albun
	for rows.Next() {
		var i Albun
		if err := scanAlbun(rows, &i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const obterAlbumPorID = `SELECT ` + albumCols + ` FROM albuns WHERE id = $1 AND ativo = TRUE`

func (q *Queries) ObterAlbumPorID(ctx context.Context, id int64) (Albun, error) {
	row := q.db.QueryRowContext(ctx, obterAlbumPorID, id)
	var i Albun
	return i, scanAlbun(row, &i)
}

const listarAlbunsPorCliente = `
SELECT a.id, a.titulo, a.descricao, a.data_evento, a.criado_em, a.ativo, a.id_fotografo, a.lote, a.valor_album, a.valor_unitario_fotografia, a.publico, a.capa_url
FROM albuns a
INNER JOIN clientes_albuns ca ON ca.id_album = a.id
WHERE ca.id_cliente = $1 AND a.ativo = TRUE
ORDER BY a.criado_em DESC
`

func (q *Queries) ListarAlbunsPorCliente(ctx context.Context, idCliente int64) ([]Albun, error) {
	rows, err := q.db.QueryContext(ctx, listarAlbunsPorCliente, idCliente)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Albun
	for rows.Next() {
		var i Albun
		if err := scanAlbun(rows, &i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listarAlbunsPublicos = `SELECT ` + albumCols + ` FROM albuns WHERE publico = TRUE AND ativo = TRUE ORDER BY criado_em DESC`

func (q *Queries) ListarAlbunsPublicos(ctx context.Context) ([]Albun, error) {
	rows, err := q.db.QueryContext(ctx, listarAlbunsPublicos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Albun
	for rows.Next() {
		var i Albun
		if err := scanAlbun(rows, &i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const obterAlbumPublico = `SELECT ` + albumCols + ` FROM albuns WHERE id = $1 AND publico = TRUE AND ativo = TRUE`

func (q *Queries) ObterAlbumPublico(ctx context.Context, id int64) (Albun, error) {
	row := q.db.QueryRowContext(ctx, obterAlbumPublico, id)
	var i Albun
	return i, scanAlbun(row, &i)
}

const atualizarPublicoAlbum = `UPDATE albuns SET publico = $2 WHERE id = $1 RETURNING ` + albumCols

func (q *Queries) AtualizarPublicoAlbum(ctx context.Context, id int64, publico bool) (Albun, error) {
	row := q.db.QueryRowContext(ctx, atualizarPublicoAlbum, id, publico)
	var i Albun
	return i, scanAlbun(row, &i)
}

const atualizarCapaAlbum = `UPDATE albuns SET capa_url = $2 WHERE id = $1 RETURNING ` + albumCols

func (q *Queries) AtualizarCapaAlbum(ctx context.Context, id int64, capaURL string) (Albun, error) {
	row := q.db.QueryRowContext(ctx, atualizarCapaAlbum, id, capaURL)
	var i Albun
	return i, scanAlbun(row, &i)
}

// ─── Fotografias ─────────────────────────────────────────────────────────────

const criarFotografia = `
INSERT INTO fotografias (url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario, ativo, criado_em
`

type CriarFotografiaParams struct {
	UrlAlta       string
	UrlBaixa      string
	Descricao     sql.NullString
	IDFotografo   int64
	IDAlbum       int64
	ValorUnitario string
}

func (q *Queries) CriarFotografia(ctx context.Context, arg CriarFotografiaParams) (Fotografia, error) {
	row := q.db.QueryRowContext(ctx, criarFotografia,
		arg.UrlAlta, arg.UrlBaixa, arg.Descricao, arg.IDFotografo, arg.IDAlbum, arg.ValorUnitario,
	)
	var i Fotografia
	err := row.Scan(&i.ID, &i.UrlAlta, &i.UrlBaixa, &i.Descricao,
		&i.IDFotografo, &i.IDAlbum, &i.ValorUnitario, &i.Ativo, &i.CriadoEm)
	return i, err
}

const obterFotografiaPorID = `
SELECT id, url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario, ativo, criado_em
FROM fotografias WHERE id = $1
`

func (q *Queries) ObterFotografiaPorID(ctx context.Context, id int64) (Fotografia, error) {
	row := q.db.QueryRowContext(ctx, obterFotografiaPorID, id)
	var i Fotografia
	err := row.Scan(&i.ID, &i.UrlAlta, &i.UrlBaixa, &i.Descricao,
		&i.IDFotografo, &i.IDAlbum, &i.ValorUnitario, &i.Ativo, &i.CriadoEm)
	return i, err
}

const listarFotografiasPorAlbum = `
SELECT id, url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario, ativo, criado_em
FROM fotografias WHERE id_album = $1 AND ativo = TRUE ORDER BY criado_em DESC
`

func (q *Queries) ListarFotografiasPorAlbum(ctx context.Context, idAlbum int64) ([]Fotografia, error) {
	rows, err := q.db.QueryContext(ctx, listarFotografiasPorAlbum, idAlbum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Fotografia
	for rows.Next() {
		var i Fotografia
		if err := rows.Scan(&i.ID, &i.UrlAlta, &i.UrlBaixa, &i.Descricao,
			&i.IDFotografo, &i.IDAlbum, &i.ValorUnitario, &i.Ativo, &i.CriadoEm); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const deletarFotografia = `DELETE FROM fotografias WHERE id = $1`

func (q *Queries) DeletarFotografia(ctx context.Context, id int64) error {
	_, err := q.db.ExecContext(ctx, deletarFotografia, id)
	return err
}

// ─── Clientes ────────────────────────────────────────────────────────────────

const associarClienteAlbum = `
INSERT INTO clientes_albuns (id_cliente, id_album)
VALUES ($1, $2)
ON CONFLICT (id_cliente, id_album) DO NOTHING
RETURNING id, id_cliente, id_album, criado_em
`

func (q *Queries) AssociarClienteAlbum(ctx context.Context, idCliente, idAlbum int64) (ClientesAlbun, error) {
	row := q.db.QueryRowContext(ctx, associarClienteAlbum, idCliente, idAlbum)
	var i ClientesAlbun
	err := row.Scan(&i.ID, &i.IDCliente, &i.IDAlbum, &i.CriadoEm)
	return i, err
}

const verificarAcessoCliente = `SELECT COUNT(*) FROM clientes_albuns WHERE id_cliente = $1 AND id_album = $2`

func (q *Queries) VerificarAcessoClienteAlbum(ctx context.Context, idCliente, idAlbum int64) (bool, error) {
	var n int
	err := q.db.QueryRowContext(ctx, verificarAcessoCliente, idCliente, idAlbum).Scan(&n)
	return n > 0, err
}

const verificarConvitePendente = `
SELECT COUNT(*) FROM convites WHERE email = $1 AND id_album = $2 AND used_at IS NULL AND expires_at > NOW()
`

func (q *Queries) VerificarConvitePendente(ctx context.Context, email string, idAlbum int64) (bool, error) {
	var n int
	err := q.db.QueryRowContext(ctx, verificarConvitePendente, email, idAlbum).Scan(&n)
	return n > 0, err
}
