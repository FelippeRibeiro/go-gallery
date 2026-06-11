package db

import (
	"database/sql"
	"time"
)

type Albun struct {
	ID                      int64
	Titulo                  string
	Descricao               sql.NullString
	DataEvento              time.Time
	CriadoEm                time.Time
	Ativo                   bool
	IDFotografo             int64
	Lote                    bool
	ValorAlbum              string
	ValorUnitarioFotografia string
	Publico                 bool
	CapaUrl                 sql.NullString
}

type ClientesAlbun struct {
	ID        int64
	IDCliente int64
	IDAlbum   int64
	CriadoEm  time.Time
}

type Fotografia struct {
	ID            int64
	UrlAlta       string
	UrlBaixa      string
	Descricao     sql.NullString
	IDFotografo   int64
	IDAlbum       int64
	ValorUnitario string
	Ativo         bool
	CriadoEm      time.Time
}

type FotografosAlbun struct {
	ID          int64
	IDFotografo int64
	IDAlbum     int64
	CriadoEm    time.Time
}

type Usuario struct {
	ID        int64
	Nome      string
	Email     string
	SenhaHash sql.NullString
	Ativo     bool
	CriadoEm  time.Time
	Tipo      interface{}
}

type Convite struct {
	ID        int64
	Token     string
	IDAlbum   int64
	Email     string
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CriadoEm  time.Time
}
