import api from './client'
import type { Venda, VendasResumo, VendaDetalhe } from '../types'

export const listSales = () =>
  api.get<{ resumo: VendasResumo; pedidos: Venda[] }>('/sales').then((r) => r.data)

export const getSale = (orderId: number) =>
  api.get<VendaDetalhe>(`/sales/${orderId}`).then((r) => r.data)
