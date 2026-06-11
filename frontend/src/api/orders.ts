import api from './client'
import type { Pedido, PedidoResumo, PedidoFotoInfo, DownloadLink, Album } from '../types'

export const listMyOrders = () =>
  api.get<{ pedidos: PedidoResumo[] }>('/orders').then((r) => r.data.pedidos)

export const createOrder = (albumId: number, fotoIds: number[]) =>
  api.post<{ pedido_id: number; valor_total: string; fotos: number }>(
    `/albums/${albumId}/orders`,
    { foto_ids: fotoIds }
  ).then((r) => r.data)

// Checkout Pro (redirect): devolve o init_point do Mercado Pago.
export const createCheckout = (orderId: number) =>
  api.post<{ init_point: string }>(`/orders/${orderId}/checkout`).then((r) => r.data)

// PIX transparente: devolve os dados do QR Code para render na própria UI.
export const createPix = (orderId: number) =>
  api.post<{ qr_code: string; qr_code_base64: string; ticket_url: string; status: string }>(
    `/orders/${orderId}/pix`
  ).then((r) => r.data)

export const getOrder = (orderId: number) =>
  api.get<{ pedido: Pedido; album: Album; fotos: PedidoFotoInfo[] }>(
    `/orders/${orderId}`
  ).then((r) => r.data)

export const getDownloadLinks = (orderId: number) =>
  api.get<{ downloads: DownloadLink[] }>(`/orders/${orderId}/downloads`).then((r) => r.data)
