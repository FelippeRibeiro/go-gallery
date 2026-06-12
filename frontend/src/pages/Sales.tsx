import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ChevronRight, ImageIcon, ShoppingBag, Clock, Wallet } from 'lucide-react'
import { listSales } from '@/api/sales'
import type { Venda, VendasResumo } from '@/types'
import Layout from '@/components/Layout'
import SkeletonImage from '@/components/SkeletonImage'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

function formatDate(iso: string) {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: 'numeric' })
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  if (status === 'pago') return 'default'
  if (status === 'falhou') return 'destructive'
  return 'secondary'
}

export default function Sales() {
  const [vendas, setVendas] = useState<Venda[] | null>(null)
  const [resumo, setResumo] = useState<VendasResumo | null>(null)

  useEffect(() => {
    listSales()
      .then((data) => {
        setVendas(data.pedidos)
        setResumo(data.resumo)
      })
      .catch(() => {
        toast.error('Erro ao carregar vendas')
        setVendas([])
      })
  }, [])

  return (
    <Layout>
      <div className="max-w-3xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Wallet className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-2xl font-bold">Vendas</h1>
            <p className="text-sm text-muted-foreground">Pedidos feitos nos seus álbuns e saldo recebido</p>
          </div>
        </div>

        {/* Resumo */}
        {resumo ? (
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div className="rounded-xl border border-border p-4 space-y-1">
              <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                <Wallet className="h-3.5 w-3.5" />
                Saldo recebido
              </div>
              <p className="text-2xl font-bold text-emerald-500">R$ {resumo.TotalRecebido}</p>
              <p className="text-xs text-muted-foreground">
                {resumo.PedidosPagos} pedido{resumo.PedidosPagos !== 1 ? 's' : ''} pago{resumo.PedidosPagos !== 1 ? 's' : ''}
              </p>
            </div>
            <div className="rounded-xl border border-border p-4 space-y-1">
              <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                <Clock className="h-3.5 w-3.5" />
                Aguardando pagamento
              </div>
              <p className="text-2xl font-bold">R$ {resumo.TotalPendente}</p>
              <p className="text-xs text-muted-foreground">
                {resumo.PedidosPendentes} pedido{resumo.PedidosPendentes !== 1 ? 's' : ''} pendente{resumo.PedidosPendentes !== 1 ? 's' : ''}
              </p>
            </div>
            <div className="rounded-xl border border-border p-4 space-y-1">
              <div className="flex items-center gap-2 text-xs text-muted-foreground uppercase tracking-wider">
                <ShoppingBag className="h-3.5 w-3.5" />
                Total de pedidos
              </div>
              <p className="text-2xl font-bold">{vendas?.length ?? 0}</p>
              <p className="text-xs text-muted-foreground">em todos os seus álbuns</p>
            </div>
          </div>
        ) : (
          vendas === null && (
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              {[0, 1, 2].map((i) => <Skeleton key={i} className="h-24 rounded-xl" />)}
            </div>
          )
        )}

        {/* Lista */}
        {vendas === null ? (
          <div className="space-y-3">
            {[0, 1, 2].map((i) => (
              <Skeleton key={i} className="h-24 w-full rounded-xl" />
            ))}
          </div>
        ) : vendas.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border py-16 text-center">
            <ShoppingBag className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="font-medium">Nenhuma venda ainda</p>
            <p className="text-sm text-muted-foreground">Quando alguém comprar fotos dos seus álbuns, aparecerá aqui.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {vendas.map((v) => (
              <Link
                key={v.ID}
                to={`/sales/${v.ID}`}
                className="flex items-center gap-4 rounded-xl border border-border p-3 transition-colors hover:bg-accent"
              >
                <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg bg-muted">
                  {v.CapaUrl?.Valid ? (
                    <SkeletonImage src={v.CapaUrl.String} alt={v.AlbumTitulo} className="object-cover" />
                  ) : (
                    <div className="flex h-full w-full items-center justify-center text-muted-foreground">
                      <ImageIcon className="h-5 w-5" />
                    </div>
                  )}
                </div>

                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <p className="truncate font-semibold">{v.AlbumTitulo}</p>
                    <Badge variant={statusVariant(v.Status)} className="shrink-0">
                      {v.Status}
                    </Badge>
                  </div>
                  <p className="truncate text-sm text-muted-foreground">
                    {v.ClienteNome} · {v.ClienteEmail}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Pedido #{v.ID} · {v.TotalFotos} foto{v.TotalFotos !== 1 ? 's' : ''} · {formatDate(v.CriadoEm)}
                  </p>
                </div>

                <div className="flex shrink-0 items-center gap-2">
                  <span className="font-semibold">R$ {v.ValorTotal}</span>
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
