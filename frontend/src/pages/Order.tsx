import { useCallback, useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Download, Package, Loader2, ExternalLink, QrCode, CreditCard, Copy, Check } from 'lucide-react'
import { getOrder, getDownloadLinks, createCheckout, createPix } from '@/api/orders'
import type { Pedido, PedidoFotoInfo, DownloadLink } from '@/types'
import type { Album } from '@/types'
import Layout from '@/components/Layout'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { toast } from 'sonner'

type PixData = { qr_code: string; qr_code_base64: string; ticket_url: string }

export default function Order() {
  const { id } = useParams<{ id: string }>()
  const orderId = Number(id)

  const [pedido, setPedido] = useState<Pedido | null>(null)
  const [album, setAlbum] = useState<Album | null>(null)
  const [fotos, setFotos] = useState<PedidoFotoInfo[]>([])
  const [links, setLinks] = useState<DownloadLink[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadingLinks, setLoadingLinks] = useState(false)

  // Pagamento
  const [payLoading, setPayLoading] = useState<'pix' | 'card' | null>(null)
  const [pix, setPix] = useState<PixData | null>(null)
  const [copied, setCopied] = useState(false)

  const refreshOrder = useCallback(async () => {
    const res = await getOrder(orderId)
    setPedido(res.pedido)
    setAlbum(res.album)
    setFotos(res.fotos)
    return res.pedido
  }, [orderId])

  useEffect(() => {
    refreshOrder()
      .catch(() => toast.error('Pedido não encontrado'))
      .finally(() => setLoading(false))
  }, [refreshOrder])

  // Polling: enquanto o PIX está exibido e o pedido pendente, consulta o status.
  const pago = pedido?.Status === 'pago'
  const pollingRef = useRef<number | null>(null)
  useEffect(() => {
    if (!pix || pago) {
      if (pollingRef.current) window.clearInterval(pollingRef.current)
      return
    }
    pollingRef.current = window.setInterval(async () => {
      try {
        const p = await refreshOrder()
        if (p.Status === 'pago') {
          toast.success('Pagamento confirmado!')
          setPix(null)
        }
      } catch { /* ignora erros transitórios de polling */ }
    }, 4000)
    return () => {
      if (pollingRef.current) window.clearInterval(pollingRef.current)
    }
  }, [pix, pago, refreshOrder])

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

  const handlePix = async () => {
    setPayLoading('pix')
    try {
      const data = await createPix(orderId)
      setPix({ qr_code: data.qr_code, qr_code_base64: data.qr_code_base64, ticket_url: data.ticket_url })
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao gerar PIX'
      toast.error(msg)
    } finally {
      setPayLoading(null)
    }
  }

  const handleCard = async () => {
    setPayLoading('card')
    try {
      const { init_point } = await createCheckout(orderId)
      window.location.href = init_point
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao iniciar checkout'
      toast.error(msg)
      setPayLoading(null)
    }
  }

  const copyPix = async () => {
    if (!pix) return
    await navigator.clipboard.writeText(pix.qr_code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
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
              <p className="font-semibold">
                {pago ? 'Pagamento confirmado' : pedido.Status === 'falhou' ? 'Pagamento não concluído' : 'Aguardando pagamento'}
              </p>
              <p className="text-sm text-muted-foreground">{fotos.length} foto{fotos.length !== 1 ? 's' : ''}</p>
            </div>
            <Badge className="ml-auto" variant={pago ? 'default' : pedido.Status === 'falhou' ? 'destructive' : 'secondary'}>
              {pedido.Status}
            </Badge>
          </div>

          <div className="flex items-center justify-between text-sm border-t border-border pt-4">
            <span className="text-muted-foreground">Total</span>
            <span className="font-semibold text-lg">R$ {pedido.ValorTotal}</span>
          </div>

          {pago ? (
            <>
              <Button className="w-full" onClick={handleDownloads} disabled={loadingLinks}>
                {loadingLinks ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Download className="h-4 w-4 mr-2" />}
                {links ? 'Atualizar links' : 'Gerar links de download'}
              </Button>
              <p className="text-xs text-muted-foreground text-center">Links expiram em 24 horas após a geração</p>
            </>
          ) : pix ? (
            /* QR Code do PIX embutido */
            <div className="flex flex-col items-center gap-4 border-t border-border pt-4">
              <p className="text-sm font-medium">Escaneie o QR Code para pagar com PIX</p>
              {pix.qr_code_base64 && (
                <img
                  src={`data:image/png;base64,${pix.qr_code_base64}`}
                  alt="QR Code PIX"
                  className="w-56 h-56 rounded-lg bg-white p-2"
                />
              )}
              <div className="w-full space-y-2">
                <p className="text-xs text-muted-foreground text-center">Ou copie o código PIX (copia e cola)</p>
                <div className="flex gap-2">
                  <input
                    readOnly
                    value={pix.qr_code}
                    className="flex-1 min-w-0 rounded-md border border-border bg-muted px-3 py-2 text-xs text-muted-foreground"
                  />
                  <Button size="icon" variant="outline" onClick={copyPix} title="Copiar código">
                    {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                  </Button>
                </div>
              </div>
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Loader2 className="h-3.5 w-3.5 animate-spin" />
                Aguardando confirmação do pagamento...
              </div>
            </div>
          ) : (
            /* Escolha do método de pagamento */
            <div className="space-y-3 border-t border-border pt-4">
              {pedido.Status === 'falhou' && (
                <p className="text-sm text-destructive text-center">
                  O pagamento anterior não foi concluído. Tente novamente abaixo.
                </p>
              )}
              <p className="text-sm font-medium text-center">Escolha como pagar</p>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <Button variant="outline" className="h-auto py-4 flex-col gap-2" onClick={handlePix} disabled={payLoading !== null}>
                  {payLoading === 'pix' ? <Loader2 className="h-5 w-5 animate-spin" /> : <QrCode className="h-5 w-5" />}
                  <span>PIX</span>
                </Button>
                <Button variant="outline" className="h-auto py-4 flex-col gap-2" onClick={handleCard} disabled={payLoading !== null}>
                  {payLoading === 'card' ? <Loader2 className="h-5 w-5 animate-spin" /> : <CreditCard className="h-5 w-5" />}
                  <span>Cartão / outros</span>
                </Button>
              </div>
              <p className="text-xs text-muted-foreground text-center">
                PIX é exibido aqui mesmo. "Cartão / outros" abre o checkout do Mercado Pago.
              </p>
            </div>
          )}
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
