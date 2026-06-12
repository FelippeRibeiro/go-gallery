import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Calendar, Images, Mail, ShoppingBag, User } from 'lucide-react'
import { getSale } from '@/api/sales'
import type { VendaDetalhe } from '@/types'
import Layout from '@/components/Layout'
import SkeletonImage from '@/components/SkeletonImage'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

function formatDate(iso: string) {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ''
  return d.toLocaleDateString('pt-BR', { day: '2-digit', month: 'long', year: 'numeric' })
}

function statusVariant(status: string): 'default' | 'secondary' | 'destructive' {
  if (status === 'pago') return 'default'
  if (status === 'falhou') return 'destructive'
  return 'secondary'
}

export default function Sale() {
  const { id } = useParams<{ id: string }>()
  const [venda, setVenda] = useState<VendaDetalhe | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getSale(Number(id))
      .then(setVenda)
      .catch(() => toast.error('Venda não encontrada'))
      .finally(() => setLoading(false))
  }, [id])

  if (loading) {
    return (
      <Layout>
        <div className="space-y-4 max-w-2xl mx-auto">
          <Skeleton className="h-8 w-1/3" />
          <Skeleton className="h-40 w-full rounded-xl" />
        </div>
      </Layout>
    )
  }

  if (!venda) {
    return (
      <Layout>
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <p className="text-muted-foreground">Venda não encontrada.</p>
          <Button asChild variant="outline"><Link to="/sales">Voltar</Link></Button>
        </div>
      </Layout>
    )
  }

  const { pedido, album, cliente, fotos } = venda

  return (
    <Layout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/sales"><ArrowLeft className="h-4 w-4" /></Link>
          </Button>
          <div className="min-w-0">
            <h1 className="text-2xl font-bold">Venda #{pedido.ID}</h1>
            <Link to={`/albums/${album.ID}`} className="text-sm text-muted-foreground hover:text-primary hover:underline">
              {album.Titulo}
            </Link>
          </div>
          <Badge className="ml-auto shrink-0" variant={statusVariant(pedido.Status)}>
            {pedido.Status}
          </Badge>
        </div>

        {/* Comprador + resumo */}
        <div className="rounded-xl border border-border p-6 space-y-4">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <User className="h-6 w-6" />
            </div>
            <div className="min-w-0">
              <p className="font-semibold truncate">{cliente.nome}</p>
              <p className="text-sm text-muted-foreground flex items-center gap-1 truncate">
                <Mail className="h-3.5 w-3.5 shrink-0" />
                {cliente.email}
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 border-t border-border pt-4 text-sm">
            <div className="flex items-center gap-2 text-muted-foreground">
              <Calendar className="h-4 w-4 shrink-0" />
              {formatDate(pedido.CriadoEm)}
            </div>
            <div className="flex items-center gap-2 text-muted-foreground">
              <Images className="h-4 w-4 shrink-0" />
              {fotos.length} foto{fotos.length !== 1 ? 's' : ''}
            </div>
            <div className="flex items-center gap-2 sm:justify-end">
              <ShoppingBag className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="font-semibold text-lg">R$ {pedido.ValorTotal}</span>
            </div>
          </div>
        </div>

        {/* Fotos vendidas */}
        {fotos.length > 0 && (
          <div className="space-y-3">
            <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Fotos do pedido ({fotos.length})
            </h2>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
              {fotos.map((f) => (
                <div key={f.IDFotografia} className="relative rounded-lg overflow-hidden border border-border">
                  <SkeletonImage
                    src={f.UrlBaixa}
                    alt={`Foto ${f.IDFotografia}`}
                    wrapperClassName="w-full aspect-square"
                    className="object-cover"
                    loading="lazy"
                  />
                  <span className="absolute bottom-1.5 right-1.5 rounded-md bg-black/70 px-1.5 py-0.5 text-[11px] font-medium text-white backdrop-blur-sm">
                    R$ {f.ValorUnitario}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Layout>
  )
}
