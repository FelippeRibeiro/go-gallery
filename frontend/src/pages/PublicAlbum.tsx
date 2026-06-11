import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Calendar, ImageOff, Loader2 } from 'lucide-react'
import { getPublicAlbum } from '@/api/albums'
import PublicOrAppLayout from '@/components/PublicOrAppLayout'
import type { Album, FotoPublica } from '@/types'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import Lightbox from '@/components/Lightbox'

export default function PublicAlbum() {
  const { id } = useParams<{ id: string }>()
  const albumId = Number(id)

  const [album, setAlbum] = useState<Album | null>(null)
  const [fotos, setFotos] = useState<FotoPublica[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)

  useEffect(() => {
    getPublicAlbum(albumId)
      .then((res) => {
        setAlbum(res.album)
        setFotos(res.fotos)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [albumId])

  return (
    <PublicOrAppLayout>
      <div className="space-y-8">
        {loading && (
          <div className="space-y-4">
            <Skeleton className="h-8 w-1/3" />
            <Skeleton className="h-64 w-full rounded-xl" />
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
              {Array.from({ length: 10 }).map((_, i) => (
                <Skeleton key={i} className="aspect-square rounded-lg" />
              ))}
            </div>
          </div>
        )}

        {!loading && error && (
          <div className="flex flex-col items-center justify-center h-64 gap-4 text-muted-foreground">
            <ImageOff className="h-10 w-10 opacity-40" />
            <p>Álbum não encontrado ou não é público.</p>
            <Button asChild variant="outline">
              <Link to="/gallery">Ver galeria</Link>
            </Button>
          </div>
        )}

        {!loading && !error && album && (
          <>
            {/* Cover */}
            {album.CapaUrl?.Valid && (
              <div className="w-full h-48 sm:h-72 rounded-xl overflow-hidden">
                <img
                  src={album.CapaUrl.String}
                  alt="Capa do álbum"
                  className="w-full h-full object-cover"
                />
              </div>
            )}

            {/* Header */}
            <div className="flex items-start gap-3">
              <Button variant="ghost" size="icon" asChild className="shrink-0 mt-0.5">
                <Link to="/gallery"><ArrowLeft className="h-4 w-4" /></Link>
              </Button>
              <div>
                <h2 className="text-2xl font-bold">{album.Titulo}</h2>
                {album.Descricao?.Valid && (
                  <p className="text-sm text-muted-foreground mt-1">{album.Descricao.String}</p>
                )}
                <div className="flex items-center gap-1 text-xs text-muted-foreground mt-2">
                  <Calendar className="h-3.5 w-3.5" />
                  {new Date(album.DataEvento).toLocaleDateString('pt-BR', {
                    day: '2-digit', month: 'long', year: 'numeric',
                  })}
                </div>
              </div>
            </div>

            {/* Photos */}
            {fotos.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-48 rounded-lg border border-dashed border-border text-muted-foreground gap-2">
                <ImageOff className="h-8 w-8 opacity-40" />
                <p className="text-sm">Nenhuma foto neste álbum ainda.</p>
              </div>
            ) : (
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
                {fotos.map((foto, i) => (
                  <div
                    key={foto.id}
                    onClick={() => setLightboxIndex(i)}
                    className="relative aspect-square rounded-lg overflow-hidden bg-muted cursor-zoom-in group"
                  >
                    <img
                      src={foto.url_baixa}
                      alt={`Foto ${foto.id}`}
                      className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
                      loading="lazy"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity flex items-end p-2">
                      <span className="text-[10px] text-white/80">
                        {new Date(foto.criado_em).toLocaleDateString('pt-BR')}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {lightboxIndex !== null && (
              <Lightbox
                photos={fotos.map(f => ({ id: f.id, url: f.url_baixa }))}
                initialIndex={lightboxIndex}
                onClose={() => setLightboxIndex(null)}
              />
            )}
          </>
        )}

        {loading && !error && (
          <div className="flex justify-center py-10">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        )}
      </div>
    </PublicOrAppLayout>
  )
}
