import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Package, ChevronRight, ImageIcon } from 'lucide-react'
import { listMyOrders } from '@/api/orders'
import type { PedidoResumo } from '@/types'
import Layout from '@/components/Layout'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

function formatDate(iso: string) {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: 'numeric' })
}

export default function Orders() {
  const [pedidos, setPedidos] = useState<PedidoResumo[] | null>(null)

  useEffect(() => {
    listMyOrders()
      .then(setPedidos)
      .catch(() => {
        toast.error('Erro ao carregar pedidos')
        setPedidos([])
      })
  }, [])

  return (
    <Layout>
      <div className="max-w-3xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Package className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">Meus pedidos</h1>
            <p className="text-sm text-muted-foreground">Acompanhe suas compras e baixe as fotos</p>
          </div>
        </div>

        {pedidos === null ? (
          <div className="space-y-3">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-24 w-full rounded-xl" />
            ))}
          </div>
        ) : pedidos.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 text-center">
            <Package className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="font-medium">Nenhum pedido ainda</p>
            <p className="text-sm text-muted-foreground">Quando você comprar fotos, elas aparecerão aqui.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {pedidos.map((p) => (
              <Link
                key={p.ID}
                to={`/orders/${p.ID}`}
                className="flex items-center gap-4 rounded-xl border border-border p-3 transition-colors hover:bg-accent"
              >
                <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg bg-muted">
                  {p.CapaUrl?.Valid ? (
                    <img src={p.CapaUrl.String} alt={p.AlbumTitulo} className="h-full w-full object-cover" />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                      <ImageIcon className="h-5 w-5" />
                    </div>
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="truncate font-semibold">{p.AlbumTitulo}</p>
                    <Badge variant={p.Status === 'pendente' ? 'secondary' : 'default'} className="shrink-0">
                      {p.Status}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Pedido #{p.ID} · {p.TotalFotos} foto{p.TotalFotos !== 1 ? 's' : ''} · {formatDate(p.CriadoEm)}
                  </p>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <span className="font-semibold">R$ {p.ValorTotal}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </Layout>
  )
}
