import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Download, Package, Loader2, ExternalLink } from 'lucide-react'
import { getOrder, getDownloadLinks } from '@/api/orders'
import type { Pedido, PedidoFotoInfo, DownloadLink } from '@/types'
import type { Album } from '@/types'
import Layout from '@/components/Layout'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

export default function Order() {
  const { id } = useParams<{ id: string }>()
  const orderId = Number(id)

  const [pedido, setPedido] = useState<Pedido | null>(null)
  const [album, setAlbum] = useState<Album | null>(null)
  const [fotos, setFotos] = useState<PedidoFotoInfo[]>([])
  const [links, setLinks] = useState<DownloadLink[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingLinks, setLoadingLinks] = useState(false)

  useEffect(() => {
    getOrder(orderId)
      .then((res) => {
        setPedido(res.pedido)
        setAlbum(res.album)
        setFotos(res.fotos)
      })
      .catch(() => toast.error('Pedido não encontrado'))
      .finally(() => setLoading(false))
  }, [orderId])

  const handleDownloads = async () => {
    setLoadingLinks(true)
    try {
      const res = await getDownloadLinks(orderId)
      setLinks(res.downloads)
    } catch {
      toast.error('Erro ao gerar links de download')
    } finally {
      setLoadingLinks(false)
    }
  }

  if (loading) {
    return (
      <Layout>
        <div className="space-y-4 max-w-2xl mx-auto">
          <Skeleton className="h-8 w-1/3" />
          <Skeleton className="h-32 w-full rounded-xl" />
        </div>
      </Layout>
    )
  }

  if (!pedido || !album) {
    return (
      <Layout>
        <div className="flex flex-col items-center justify-center h-64 gap-4">
          <p className="text-muted-foreground">Pedido não encontrado.</p>
          <Button asChild variant="outline"><Link to="/orders">Voltar</Link></Button>
        </div>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/orders"><ArrowLeft className="h-4 w-4" /></Link>
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Pedido #{pedido.ID}</h1>
            <p className="text-sm text-muted-foreground">{album.Titulo}</p>
          </div>
        </div>

        {/* Summary card */}
        <div className="rounded-xl border border-border p-6 space-y-4">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Package className="h-6 w-6" />
            </div>
            <div>
              <p className="font-semibold">Pedido confirmado</p>
              <p className="text-sm text-muted-foreground">{fotos.length} foto{fotos.length !== 1 ? 's' : ''}</p>
            </div>
            <Badge className="ml-auto" variant={pedido.Status === 'pendente' ? 'secondary' : 'default'}>
              {pedido.Status}
            </Badge>
          </div>

          <div className="flex items-center justify-between text-sm border-t border-border pt-4">
            <span className="text-muted-foreground">Total</span>
            <span className="font-semibold text-lg">R$ {pedido.ValorTotal}</span>
          </div>

          <Button className="w-full" onClick={handleDownloads} disabled={loadingLinks}>
            {loadingLinks ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Download className="h-4 w-4 mr-2" />}
            {links ? 'Atualizar links' : 'Gerar links de download'}
          </Button>

          <p className="text-xs text-muted-foreground text-center">Links expiram em 24 horas após a geração</p>
        </div>

        {/* Download grid */}
        {links && (
          <div className="space-y-3">
            <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
              Fotos para download ({links.length})
            </h2>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
              {links.map((l) => (
                <div key={l.foto_id} className="group relative rounded-lg overflow-hidden border border-border">
                  <img
                    src={l.url_baixa}
                    alt={`Foto ${l.foto_id}`}
                    className="w-full aspect-square object-cover"
                  />
                  <a
                    href={l.url_download}
                    target="_blank"
                    rel="noreferrer"
                    className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center"
                  >
                    <div className="flex flex-col items-center gap-1 text-white">
                      <ExternalLink className="h-5 w-5" />
                      <span className="text-xs font-medium">Baixar original</span>
                    </div>
                  </a>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Layout>
  )
}
