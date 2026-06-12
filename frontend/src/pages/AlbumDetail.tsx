import { useEffect, useRef, useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import {
  ArrowLeft, Calendar, DollarSign, Globe, ImagePlus, Images,
  Loader2, Lock, Mail, ShoppingCart, Users, X,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  DndContext,
  closestCenter,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  rectSortingStrategy,
  useSortable,
  arrayMove,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import PublicOrAppLayout from '@/components/PublicOrAppLayout'
import PhotoGrid from '@/components/PhotoGrid'
import PhotoUpload from '@/components/PhotoUpload'
import InviteModal from '@/components/InviteModal'
import InvitesManager from '@/components/InvitesManager'
import Lightbox from '@/components/Lightbox'
import { useAuth } from '@/context/AuthContext'
import { getAlbum, getPublicAlbum, updateVisibility, deletePhoto, uploadCover, listPhotos, reorderPhotos } from '@/api/albums'
import { createOrder } from '@/api/orders'
import type { Album, Foto, FotoPublica } from '@/types'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'

function AlbumDetailSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Skeleton className="h-9 w-9 rounded-md" />
        <div className="space-y-2 flex-1">
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-4 w-1/4" />
        </div>
      </div>
      <Skeleton className="h-32 w-full rounded-xl" />
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
        {Array.from({ length: 10 }).map((_, i) => (
          <Skeleton key={i} className="aspect-square rounded-lg" />
        ))}
      </div>
    </div>
  )
}

// ─── Client view ──────────────────────────────────────────────────────────────

function ClientAlbumView({ album, fotos: allFotos, compradas, comprado, authenticated }: { album: Album; fotos: FotoPublica[]; compradas: number[]; comprado: boolean; authenticated: boolean }) {
  const navigate = useNavigate()
  const [selected, setSelected] = useState<Set<number>>(new Set())
  const [selectMode, setSelectMode] = useState(false)
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)
  const [ordering, setOrdering] = useState(false)

  // Fotos já compradas continuam visíveis na lista, mas não podem ser
  // selecionadas novamente para uma nova compra.
  const compradasSet = new Set(compradas)
  const fotos = allFotos
  const lightboxPhotos = fotos.map(f => ({ id: f.id, url: f.url_baixa }))

  const toggle = (id: number) =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  const handlePhotoClick = (foto: FotoPublica, index: number) => {
    if (!selectMode) {
      setLightboxIndex(index)
      return
    }
    if (compradasSet.has(foto.id)) {
      toast.info('Esta foto já foi comprada')
      return
    }
    toggle(foto.id)
  }

  const unitPrice = Number(album.ValorUnitarioFotografia) || 0
  const total = album.Lote
    ? Number(album.ValorAlbum)
    : selected.size * unitPrice

  const handleFinalize = async () => {
    if (!authenticated) {
      navigate(`/login?redirect=/albums/${album.ID}`)
      return
    }
    setOrdering(true)
    try {
      const fotoIds = album.Lote ? [] : Array.from(selected)
      const result = await createOrder(album.ID, fotoIds)
      toast.success('Pedido criado! Escolha a forma de pagamento.')
      // A escolha do método (PIX ou cartão) acontece na tela do pedido.
      navigate(`/orders/${result.pedido_id}`)
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao criar pedido'
      toast.error(msg)
    } finally {
      setOrdering(false)
    }
  }

  return (
    <div className="space-y-6 max-w-6xl mx-auto pb-24">
      {/* Cover */}
      {album.CapaUrl?.Valid && (
        <div className="w-full h-48 sm:h-64 rounded-xl overflow-hidden">
          <img src={album.CapaUrl.String} alt="Capa" className="w-full h-full object-cover" />
        </div>
      )}

      {/* Header */}
      <div className="flex items-start gap-3">
        <Button variant="ghost" size="icon" asChild className="shrink-0 mt-0.5">
          <Link to={authenticated ? '/dashboard' : '/discover'}><ArrowLeft className="h-4 w-4" /></Link>
        </Button>
        <div className="min-w-0">
          <h1 className="text-2xl font-bold truncate">{album.Titulo}</h1>
          {album.Descricao?.Valid && (
            <p className="text-sm text-muted-foreground mt-1">{album.Descricao.String}</p>
          )}
          <div className="flex flex-wrap gap-4 text-xs text-muted-foreground mt-2">
            <span className="flex items-center gap-1">
              <Calendar className="h-3.5 w-3.5" />
              {new Date(album.DataEvento).toLocaleDateString('pt-BR', {
                day: '2-digit', month: 'long', year: 'numeric',
              })}
            </span>
            <span className="flex items-center gap-1">
              <Images className="h-3.5 w-3.5" />
              {fotos.length} foto{fotos.length !== 1 ? 's' : ''}
            </span>
            <span className="flex items-center gap-1">
              <DollarSign className="h-3.5 w-3.5" />
              {album.Lote
                ? `R$ ${album.ValorAlbum} (álbum completo)`
                : `R$ ${unitPrice.toFixed(2)} / foto`}
            </span>
          </div>
        </div>
      </div>

      <Separator />

      {/* Visitante deslogado: apenas visualização */}
      {!authenticated && fotos.length > 0 && (
        <p className="text-sm text-muted-foreground">
          <Link to={`/login?redirect=/albums/${album.ID}`} className="text-primary hover:underline font-medium">
            Entre na sua conta
          </Link>{' '}
          para selecionar e comprar fotos.
        </p>
      )}

      {/* Venda completa (lote) já comprada: novas fotos entram automaticamente
          no pedido pago — nada de marcação "comprada" nem recompra. */}
      {authenticated && album.Lote && comprado && (
        <div className="flex items-center justify-between gap-4 rounded-lg border border-emerald-700/50 bg-emerald-950/30 p-4">
          <div>
            <p className="text-sm font-medium">Você já comprou este álbum</p>
            <p className="text-xs text-muted-foreground">
              Todas as fotos — inclusive as adicionadas depois da compra — ficam disponíveis para download no seu pedido.
            </p>
          </div>
          <Button size="sm" variant="outline" asChild className="shrink-0">
            <Link to="/orders">Meus pedidos</Link>
          </Button>
        </div>
      )}

      {/* Compra de álbum completo (lote) — somente autenticado */}
      {authenticated && fotos.length > 0 && album.Lote && !comprado && (
        <div className="flex items-center justify-between rounded-lg border border-border p-4">
          <div>
            <p className="text-sm font-medium">Álbum completo</p>
            <p className="text-xs text-muted-foreground">
              {fotos.length} foto{fotos.length !== 1 ? 's' : ''} por R$ {Number(album.ValorAlbum).toFixed(2)}
            </p>
          </div>
          <Button size="sm" onClick={handleFinalize} disabled={ordering}>
            {ordering && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
            <ShoppingCart className="h-4 w-4 mr-1" />
            Comprar álbum
          </Button>
        </div>
      )}

      {/* Select mode toggle (compra por foto) — somente autenticado */}
      {authenticated && fotos.length > 0 && !album.Lote && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            {selectMode ? 'Clique nas fotos para selecionar' : 'Ative a seleção para escolher fotos'}
          </p>
          <div className="flex items-center gap-2">
            <Label htmlFor="select-mode" className="text-sm cursor-pointer">
              Modo seleção
            </Label>
            <Switch id="select-mode" checked={selectMode} onCheckedChange={setSelectMode} />
          </div>
        </div>
      )}

      {/* Photos */}
      {fotos.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 rounded-lg border border-dashed border-border text-muted-foreground gap-2">
          <Images className="h-8 w-8 opacity-40" />
          <p className="text-sm">Nenhuma foto neste álbum ainda.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
          {fotos.map((foto, i) => {
            const comprada = compradasSet.has(foto.id)
            return (
            <div
              key={foto.id}
              onClick={() => handlePhotoClick(foto, i)}
              className={`relative aspect-square rounded-lg overflow-hidden bg-muted group ${
                selectMode ? (comprada ? 'cursor-not-allowed' : 'cursor-pointer') : 'cursor-zoom-in'
              }`}
            >
              <img
                src={foto.url_baixa}
                alt={`Foto ${foto.id}`}
                className={`w-full h-full object-cover transition-transform duration-300 ${comprada ? 'opacity-60' : 'group-hover:scale-105'}`}
                loading="lazy"
              />
              {/* Marca de foto já comprada */}
              {comprada && (
                <div className="absolute top-2 left-2 rounded-full bg-emerald-600/90 px-2 py-0.5 text-[10px] font-semibold text-white">
                  Comprada
                </div>
              )}
              {/* Selection overlay (não disponível para fotos já compradas) */}
              {selectMode && !comprada && (
                <div className={`absolute inset-0 transition-colors ${selected.has(foto.id) ? 'bg-primary/40' : 'bg-transparent group-hover:bg-white/10'}`}>
                  {selected.has(foto.id) && (
                    <div className="absolute top-2 right-2 h-5 w-5 rounded-full bg-primary flex items-center justify-center">
                      <svg className="h-3 w-3 text-white" viewBox="0 0 12 12" fill="none">
                        <path d="M2 6l3 3 5-5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                      </svg>
                    </div>
                  )}
                </div>
              )}
            </div>
            )
          })}
        </div>
      )}

      {/* Lightbox */}
      {lightboxIndex !== null && (
        <Lightbox
          photos={lightboxPhotos}
          initialIndex={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
        />
      )}

      {/* Sticky selection bar */}
      {selectMode && selected.size > 0 && (
        <div className="fixed bottom-0 left-0 right-0 bg-card/95 backdrop-blur border-t border-border p-4 flex items-center justify-between gap-4 z-50">
          <div className="flex items-center gap-3">
            <ShoppingCart className="h-5 w-5 text-primary shrink-0" />
            <div>
              <p className="text-sm font-semibold">
                {selected.size} foto{selected.size !== 1 ? 's' : ''} selecionada{selected.size !== 1 ? 's' : ''}
              </p>
              <p className="text-xs text-muted-foreground">R$ {total.toFixed(2)}</p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={() => { setSelected(new Set()); setSelectMode(false) }}>
              <X className="h-4 w-4" />
              Limpar
            </Button>
            <Button size="sm" onClick={handleFinalize} disabled={ordering}>
              {ordering && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
              Finalizar seleção
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Sortable photo item (for drag-and-drop reorder) ─────────────────────────

function SortablePhotoItem({ foto }: { foto: Foto }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: foto.ID })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }
  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="relative aspect-square rounded-lg overflow-hidden bg-muted cursor-grab active:cursor-grabbing"
    >
      <img src={foto.UrlBaixa} alt="" className="w-full h-full object-cover pointer-events-none" loading="lazy" />
    </div>
  )
}

// ─── Photographer view ────────────────────────────────────────────────────────

function PhotographerAlbumView({ album: initialAlbum, fotos: initialFotos }: { album: Album; fotos: Foto[] }) {
  const { user } = useAuth()
  const albumId = initialAlbum.ID
  const coverInputRef = useRef<HTMLInputElement>(null)
  const [showParticipants, setShowParticipants] = useState(false)

  const [album, setAlbum] = useState(initialAlbum)
  const [fotos, setFotos] = useState(initialFotos)
  const [visibilityLoading, setVisibilityLoading] = useState(false)
  const [coverUploading, setCoverUploading] = useState(false)
  const [totalFotos, setTotalFotos] = useState(initialFotos.length)
  const [loadingMore, setLoadingMore] = useState(false)
  const [reorderMode, setReorderMode] = useState(false)
  const [savingOrder, setSavingOrder] = useState(false)

  const isOwner = album.IDFotografo === user?.id

  const sensors = useSensors(useSensor(PointerSensor))

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return
    setFotos((prev) => {
      const oldIndex = prev.findIndex((f) => f.ID === active.id)
      const newIndex = prev.findIndex((f) => f.ID === over.id)
      return arrayMove(prev, oldIndex, newIndex)
    })
  }

  const handleSaveOrder = async () => {
    setSavingOrder(true)
    try {
      await reorderPhotos(albumId, fotos.map((f) => f.ID))
      setReorderMode(false)
      toast.success('Ordem salva!')
    } catch {
      toast.error('Erro ao salvar ordem')
    } finally {
      setSavingOrder(false)
    }
  }

  const handleVisibilityToggle = async (publico: boolean) => {
    setVisibilityLoading(true)
    try {
      const updated = await updateVisibility(albumId, publico)
      setAlbum(updated)
      toast.success(publico ? 'Álbum tornado público' : 'Álbum tornado privado')
    } catch {
      toast.error('Erro ao alterar visibilidade')
    } finally {
      setVisibilityLoading(false)
    }
  }

  const handleDeletePhoto = async (foto: Foto) => {
    try {
      await deletePhoto(albumId, foto.ID)
      setFotos((prev) => prev.filter((f) => f.ID !== foto.ID))
      toast.success('Foto removida')
    } catch (e) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao remover foto'
      toast.error(msg)
    }
  }

  const handleCoverChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setCoverUploading(true)
    try {
      const updated = await uploadCover(albumId, file)
      setAlbum(updated)
      toast.success('Capa atualizada!')
    } catch {
      toast.error('Erro ao enviar capa')
    } finally {
      setCoverUploading(false)
      if (coverInputRef.current) coverInputRef.current.value = ''
    }
  }

  const handleLoadMore = async () => {
    setLoadingMore(true)
    try {
      const res = await listPhotos(albumId, fotos.length)
      setFotos((prev) => [...prev, ...res.fotos as Foto[]])
      setTotalFotos(res.total)
    } catch {
      toast.error('Erro ao carregar mais fotos')
    } finally {
      setLoadingMore(false)
    }
  }

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Cover */}
      {album.CapaUrl?.Valid && (
        <div className="relative w-full h-48 sm:h-64 rounded-xl overflow-hidden group">
          <img src={album.CapaUrl.String} alt="Capa do álbum" className="w-full h-full object-cover" />
          {isOwner && (
            <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
              <Button variant="secondary" size="sm" disabled={coverUploading} onClick={() => coverInputRef.current?.click()}>
                <ImagePlus className="h-4 w-4" />
                Trocar capa
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-start gap-4">
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <Button variant="ghost" size="icon" asChild className="shrink-0 mt-0.5">
            <Link to="/dashboard"><ArrowLeft className="h-4 w-4" /></Link>
          </Button>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2 mb-1">
              <h1 className="text-2xl font-bold truncate">{album.Titulo}</h1>
              <Badge variant={album.Lote ? 'default' : 'secondary'}>
                {album.Lote ? 'Lote' : 'Por foto'}
              </Badge>
              <Badge variant="outline" className="gap-1">
                {album.Publico ? <><Globe className="h-3 w-3" /> Público</> : <><Lock className="h-3 w-3" /> Privado</>}
              </Badge>
            </div>
            {album.Descricao?.Valid && (
              <p className="text-sm text-muted-foreground mb-2">{album.Descricao.String}</p>
            )}
            <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <Calendar className="h-3.5 w-3.5" />
                {new Date(album.DataEvento).toLocaleDateString('pt-BR', { day: '2-digit', month: 'long', year: 'numeric' })}
              </span>
              <span className="flex items-center gap-1">
                <Images className="h-3.5 w-3.5" />
                {fotos.length} foto{fotos.length !== 1 ? 's' : ''}
              </span>
              <span className="flex items-center gap-1">
                <DollarSign className="h-3.5 w-3.5" />
                {album.Lote ? `R$ ${album.ValorAlbum} (lote)` : `R$ ${album.ValorUnitarioFotografia} / foto`}
              </span>
            </div>
          </div>
        </div>

        {/* Owner-only controls */}
        {isOwner && (
          <div className="flex flex-wrap gap-2 shrink-0">
            <div className="flex items-center gap-2 rounded-lg border px-3 py-2">
              <Label htmlFor="visibility" className="text-sm cursor-pointer select-none">
                {album.Publico ? 'Público' : 'Privado'}
              </Label>
              <Switch id="visibility" checked={album.Publico} disabled={visibilityLoading} onCheckedChange={handleVisibilityToggle} />
            </div>

            <Button variant="outline" size="sm" onClick={() => setShowParticipants(true)}>
              <Users className="h-4 w-4" />
              Participantes
            </Button>

            {!album.Publico && (
              <InviteModal albumId={albumId} albumTitulo={album.Titulo}>
                <Button variant="outline" size="sm">
                  <Mail className="h-4 w-4" />
                  Convidar cliente
                </Button>
              </InviteModal>
            )}
          </div>
        )}
      </div>

      {isOwner && (
        <InvitesManager albumId={albumId} open={showParticipants} onOpenChange={setShowParticipants} />
      )}

      <Separator />

      {/* Cover upload (owner, no cover yet) */}
      {isOwner && !album.CapaUrl?.Valid && (
        <>
          <div className="space-y-2">
            <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Imagem de capa</h2>
            <button
              type="button"
              disabled={coverUploading}
              onClick={() => coverInputRef.current?.click()}
              className="w-full h-28 rounded-xl border-2 border-dashed border-border hover:border-primary/50 transition-colors flex flex-col items-center justify-center gap-2 text-muted-foreground hover:text-foreground disabled:opacity-50 cursor-pointer"
            >
              <ImagePlus className="h-6 w-6" />
              <span className="text-sm">{coverUploading ? 'Enviando…' : 'Adicionar capa do álbum'}</span>
            </button>
          </div>
          <Separator />
        </>
      )}

      <input ref={coverInputRef} type="file" accept="image/jpeg,image/png,image/webp" className="hidden" onChange={handleCoverChange} />

      {/* Upload photos (owner + collaborators) */}
      <div className="space-y-2">
        <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Adicionar fotos</h2>
        <PhotoUpload albumId={albumId} onUploaded={(foto) => setFotos((prev) => [foto, ...prev])} />
      </div>

      <Separator />

      {/* Gallery */}
      <div className="space-y-3">
        {/* Reorder toggle */}
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">
            Galeria · {fotos.length} foto{fotos.length !== 1 ? 's' : ''}
            {totalFotos > fotos.length ? ` (${totalFotos} total)` : ''}
          </h2>
          {fotos.length > 1 && (
            <div className="flex gap-2">
              {reorderMode ? (
                <>
                  <Button size="sm" variant="ghost" onClick={() => setReorderMode(false)} disabled={savingOrder}>Cancelar</Button>
                  <Button size="sm" onClick={handleSaveOrder} disabled={savingOrder}>
                    {savingOrder ? <Loader2 className="h-3.5 w-3.5 animate-spin mr-1" /> : null}
                    Salvar ordem
                  </Button>
                </>
              ) : (
                <Button size="sm" variant="ghost" onClick={() => setReorderMode(true)}>Reordenar</Button>
              )}
            </div>
          )}
        </div>

        {reorderMode ? (
          <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            <SortableContext items={fotos.map((f) => f.ID)} strategy={rectSortingStrategy}>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
                {fotos.map((foto) => <SortablePhotoItem key={foto.ID} foto={foto} />)}
              </div>
            </SortableContext>
          </DndContext>
        ) : (
          <PhotoGrid fotos={fotos} onDelete={handleDeletePhoto} />
        )}

        {fotos.length < totalFotos && (
          <div className="flex justify-center pt-2">
            <Button variant="outline" size="sm" onClick={handleLoadMore} disabled={loadingMore}>
              {loadingMore ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : null}
              Carregar mais ({totalFotos - fotos.length} restantes)
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}

// ─── Main (route-level) ───────────────────────────────────────────────────────

type ViewMode = 'manage' | 'view'

export default function AlbumDetail() {
  const { id } = useParams<{ id: string }>()
  const albumId = Number(id)
  const { isAuthenticated } = useAuth()

  const [album, setAlbum] = useState<Album | null>(null)
  const [fotos, setFotos] = useState<Foto[] | FotoPublica[]>([])
  const [compradas, setCompradas] = useState<number[]>([])
  const [comprado, setComprado] = useState(false)
  const [mode, setMode] = useState<ViewMode>('view')
  const [loading, setLoading] = useState(true)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function load() {
      setLoading(true)
      setNotFound(false)

      // Usuário autenticado: tenta acesso por vínculo (dono/colaborador/cliente)
      // ou como visitante de álbum público — o backend retorna o "papel".
      if (isAuthenticated) {
        try {
          const res = await getAlbum(albumId)
          if (cancelled) return
          setAlbum(res.album)
          setFotos(res.fotos)
          setCompradas(res.compradas ?? [])
          setComprado(res.comprado ?? false)
          setMode(res.papel === 'dono' || res.papel === 'colaborador' ? 'manage' : 'view')
          setLoading(false)
          return
        } catch {
          // Sem acesso autenticado — tenta como álbum público
        }
      }

      try {
        const res = await getPublicAlbum(albumId)
        if (cancelled) return
        setAlbum(res.album)
        setFotos(res.fotos)
        setMode('view')
      } catch {
        if (!cancelled) setNotFound(true)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    load()
    return () => { cancelled = true }
  }, [albumId, isAuthenticated])

  if (loading) return <PublicOrAppLayout><AlbumDetailSkeleton /></PublicOrAppLayout>

  if (notFound || !album) return (
    <PublicOrAppLayout>
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-muted-foreground">Álbum não encontrado.</p>
        <Button asChild variant="outline">
          <Link to={isAuthenticated ? '/dashboard' : '/discover'}>Voltar</Link>
        </Button>
      </div>
    </PublicOrAppLayout>
  )

  return (
    <PublicOrAppLayout>
      {mode === 'manage' ? (
        <PhotographerAlbumView album={album} fotos={fotos as Foto[]} />
      ) : (
        <ClientAlbumView album={album} fotos={fotos as FotoPublica[]} compradas={compradas} comprado={comprado} authenticated={isAuthenticated} />
      )}
    </PublicOrAppLayout>
  )
}
