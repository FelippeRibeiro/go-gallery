import api from './client'
import type { Album, AlbumComFotos, FotoPublica, InvitesResponse } from '../types'

export const listAlbums = () =>
  api.get<Album[]>('/albums').then((r) => r.data)

export const createAlbum = (data: {
  titulo: string
  descricao?: string
  data_evento: string
  lote: boolean
  publico?: boolean
  valor_album?: string
  valor_unitario_fotografia?: string
}) => api.post<Album>('/albums', data).then((r) => r.data)

export const getAlbum = (id: number) =>
  api.get<AlbumComFotos>(`/albums/${id}`).then((r) => r.data)

export const updateVisibility = (albumId: number, publico: boolean) =>
  api.patch<Album>(`/albums/${albumId}/visibility`, { publico }).then((r) => r.data)

export const inviteClient = (albumId: number, email: string) =>
  api.post(`/albums/${albumId}/invite`, { email }).then((r) => r.data)

export const uploadPhoto = (albumId: number, file: File) => {
  const form = new FormData()
  form.append('file', file)
  return api.post(`/albums/${albumId}/photos`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }).then((r) => r.data)
}

export const deletePhoto = (albumId: number, photoId: number) =>
  api.delete(`/albums/${albumId}/photos/${photoId}`)

export const uploadCover = (albumId: number, file: File) => {
  const form = new FormData()
  form.append('file', file)
  return api.post<Album>(`/albums/${albumId}/cover`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }).then((r) => r.data)
}

export const listPublicAlbums = () =>
  api.get<Album[]>('/public/albums').then((r) => r.data)

export const getPublicAlbum = (id: number) =>
  api.get<{ album: Album; fotos: FotoPublica[] }>(`/public/albums/${id}`).then((r) => r.data)

export const listInvites = (albumId: number) =>
  api.get<InvitesResponse>(`/albums/${albumId}/invites`).then((r) => r.data)

export const revokeInvite = (albumId: number, inviteId: number) =>
  api.delete(`/albums/${albumId}/invites/${inviteId}`)

export const invitePhotographer = (albumId: number, email: string) =>
  api.post(`/albums/${albumId}/invite-photographer`, { email }).then((r) => r.data)

export const removePhotographer = (albumId: number, faId: number) =>
  api.delete(`/albums/${albumId}/photographers/${faId}`)

export const removeClient = (albumId: number, caId: number) =>
  api.delete(`/albums/${albumId}/clients/${caId}`)
