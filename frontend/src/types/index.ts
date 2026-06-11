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
  Ordem: number
}

export interface AlbumComFotos {
  album: Album
  fotos: Foto[]
}

export interface AuthResponse {
  token: string
  user: User
}

export interface Pedido {
  ID: number
  IDCliente: number
  IDAlbum: number
  Status: string
  ValorTotal: string
  CriadoEm: string
}

export interface PedidoFotoInfo {
  ID: number
  IDPedido: number
  IDFotografia: number
  ValorUnitario: string
  UrlAlta: string
  UrlBaixa: string
}

export interface DownloadLink {
  foto_id: number
  url_baixa: string
  url_download: string
}

export interface FotosPage {
  fotos: Foto[] | FotoPublica[]
  total: number
  offset: number
}

export interface Fotografo {
  id: number
  nome: string
  foto_perfil: { String: string; Valid: boolean } | null
  bio: { String: string; Valid: boolean } | null
}

export interface FotografoPublicoResponse {
  fotografo: Fotografo
  albums: Album[]
}

export interface FotografoPublico {
  ID: number
  Nome: string
  FotoPerfil: { String: string; Valid: boolean }
  Bio: { String: string; Valid: boolean }
  TotalAlbuns: number
}

export interface Perfil {
  id: number
  nome: string
  email: string
  tipo: string
  foto_perfil: { String: string; Valid: boolean } | null
  bio: { String: string; Valid: boolean } | null
}
