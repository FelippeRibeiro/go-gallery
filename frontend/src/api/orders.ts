import api from './client'
import type { Pedido, PedidoFotoInfo, DownloadLink, Album } from '../types'

export const createOrder = (albumId: number, fotoIds: number[]) =>
  api.post<{ pedido_id: number; valor_total: string; fotos: number }>(
    `/albums/${albumId}/orders`,
    { foto_ids: fotoIds }
  ).then((r) => r.data)

export const getOrder = (orderId: number) =>
  api.get<{ pedido: Pedido; album: Album; fotos: PedidoFotoInfo[] }>(
    `/orders/${orderId}`
  ).then((r) => r.data)

export const getDownloadLinks = (orderId: number) =>
  api.get<{ downloads: DownloadLink[] }>(`/orders/${orderId}/downloads`).then((r) => r.data)
