export interface User {
  id: number
  nome: string
  email: string
  tipo: 'fotografo' | 'cliente' | 'admin'
}

export interface Album {
  ID: number
  Titulo: string
  Descricao: { String: string; Valid: boolean }
  DataEvento: string
  CriadoEm: string
  Ativo: boolean
  IDFotografo: number
  Lote: boolean
  Publico: boolean
  CapaUrl: { String: string; Valid: boolean }
  ValorAlbum: string
  ValorUnitarioFotografia: string
}

export interface FotoPublica {
  id: number
  url_baixa: string
  criado_em: string
}

export interface Convite {
  ID: number
  Token: string
  IDAlbum: number
  Email: string
  ExpiresAt: string
  UsedAt: { Time: string; Valid: boolean }
  CriadoEm: string
}

export interface Colaborador {
  ID: number
  IDFotografo: number
  IDAlbum: number
  Nome: string
  Email: string
  CriadoEm: string
}

export interface Cliente {
  ID: number
  IDCliente: number
  IDAlbum: number
  Nome: string
  Email: string
  CriadoEm: string
}

export interface InvitesResponse {
  convites: Convite[]
  fotografos: Colaborador[]
  clientes: Cliente[]
}

export interface Foto {
  ID: number
  UrlAlta: string
  UrlBaixa: string
  Descricao: { String: string; Valid: boolean }
  IDFotografo: number
  IDAlbum: number
  ValorUnitario: string
  Ativo: boolean
  CriadoEm: string
}

export interface AlbumComFotos {
  album: Album
  fotos: Foto[]
}

export interface AuthResponse {
  token: string
  user: User
}
