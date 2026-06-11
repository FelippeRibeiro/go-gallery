import { useState } from 'react'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { ImageOff, Trash2, X, Check } from 'lucide-react'
import Lightbox from '@/components/Lightbox'
import type { Foto } from '@/types'

interface Props {
  fotos: Foto[]
  loading?: boolean
  onDelete?: (foto: Foto) => Promise<void>
}

function PhotoItem({
  foto,
  onDelete,
  onOpen,
}: {
  foto: Foto
  onDelete?: (foto: Foto) => Promise<void>
  onOpen: () => void
}) {
  const [confirming, setConfirming] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    if (!onDelete) return
    setDeleting(true)
    try {
      await onDelete(foto)
    } finally {
      setDeleting(false)
      setConfirming(false)
    }
  }

  return (
    <div
      className="group relative aspect-square rounded-lg overflow-hidden bg-muted cursor-zoom-in"
      onClick={() => !confirming && onOpen()}
    >
      <img
        src={foto.UrlBaixa}
        alt={foto.Descricao?.Valid ? foto.Descricao.String : `Foto ${foto.ID}`}
        className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
        loading="lazy"
      />

      {/* Hover overlay */}
      <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200 flex items-end justify-between p-2">
        <span className="text-[10px] text-white/80">
          {new Date(foto.CriadoEm).toLocaleDateString('pt-BR')}
        </span>
        {onDelete && !confirming && (
          <Button
            size="icon"
            variant="destructive"
            className="h-6 w-6"
            onClick={(e) => { e.stopPropagation(); setConfirming(true) }}
          >
            <Trash2 className="h-3 w-3" />
          </Button>
        )}
      </div>

      {/* Delete confirmation overlay */}
      {confirming && (
        <div
          className="absolute inset-0 bg-black/80 flex flex-col items-center justify-center gap-2 p-2"
          onClick={(e) => e.stopPropagation()}
        >
          <p className="text-white text-xs text-center font-medium">Excluir foto?</p>
          <div className="flex gap-1.5">
            <Button
              size="icon"
              variant="destructive"
              className="h-7 w-7"
              disabled={deleting}
              onClick={handleDelete}
            >
              <Check className="h-3.5 w-3.5" />
            </Button>
            <Button
              size="icon"
              variant="outline"
              className="h-7 w-7"
              disabled={deleting}
              onClick={() => setConfirming(false)}
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

export default function PhotoGrid({ fotos, loading, onDelete }: Props) {
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)

  if (loading) {
    return (
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
        {Array.from({ length: 12 }).map((_, i) => (
          <Skeleton key={i} className="aspect-square rounded-lg" />
        ))}
      </div>
    )
  }

  if (fotos.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-48 rounded-lg border border-dashed border-border text-muted-foreground gap-2">
        <ImageOff className="h-8 w-8 opacity-40" />
        <p className="text-sm">Nenhuma foto neste álbum ainda.</p>
      </div>
    )
  }

  const lightboxPhotos = fotos.map(f => ({ id: f.ID, url: f.UrlBaixa }))

  return (
    <>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
        {fotos.map((foto, i) => (
          <PhotoItem
            key={foto.ID}
            foto={foto}
            onDelete={onDelete}
            onOpen={() => setLightboxIndex(i)}
          />
        ))}
      </div>

      {lightboxIndex !== null && (
        <Lightbox
          photos={lightboxPhotos}
          initialIndex={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
        />
      )}
    </>
  )
}
