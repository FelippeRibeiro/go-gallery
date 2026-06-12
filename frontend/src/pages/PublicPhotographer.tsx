import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Calendar, Camera, Images } from 'lucide-react'
import { getPublicPhotographer } from '@/api/albums'
import PublicOrAppLayout from '@/components/PublicOrAppLayout'
import SkeletonImage from '@/components/SkeletonImage'
import type { Album, FotografoPublicoResponse } from '@/types'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

function AlbumCard({ album }: { album: Album }) {
  return (
    <Link to={`/albums/${album.ID}`}>
      <Card className="group overflow-hidden hover:border-primary/50 transition-colors cursor-pointer">
        <div className="aspect-video bg-muted overflow-hidden relative">
          {album.CapaUrl?.Valid ? (
            <SkeletonImage src={album.CapaUrl.String} alt={album.Titulo} className="object-cover group-hover:scale-105 transition-transform duration-300" />
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <Images className="h-8 w-8 text-muted-foreground/20" />
            </div>
          )}
        </div>
        <CardContent className="pt-3 pb-4">
          <h3 className="font-semibold line-clamp-1 group-hover:text-primary transition-colors">{album.Titulo}</h3>
          <div className="flex items-center gap-1 text-xs text-muted-foreground mt-2">
            <Calendar className="h-3 w-3" />
            {new Date(album.DataEvento).toLocaleDateString('pt-BR', { day: '2-digit', month: 'short', year: 'numeric' })}
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}

export default function PublicPhotographer() {
  const { id } = useParams<{ id: string }>()
  const [data, setData] = useState<FotografoPublicoResponse | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getPublicPhotographer(Number(id))
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false))
  }, [id])

  const avatarUrl = data?.fotografo?.foto_perfil?.Valid ? data.fotografo.foto_perfil.String : null
  const bio = data?.fotografo?.bio?.Valid ? data.fotografo.bio.String : null

  return (
    <PublicOrAppLayout>
      <div className="space-y-8 max-w-5xl mx-auto">
        {loading && (
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <Skeleton className="h-20 w-20 rounded-full" />
              <div className="space-y-2">
                <Skeleton className="h-8 w-48" />
                <Skeleton className="h-4 w-32" />
              </div>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
              {Array.from({ length: 6 }).map((_, i) => <Skeleton key={i} className="aspect-video rounded-xl" />)}
            </div>
          </div>
        )}

        {!loading && !data && (
          <div className="flex flex-col items-center justify-center h-64 gap-4 text-muted-foreground">
            <p>Fotógrafo não encontrado.</p>
            <Button variant="outline" asChild>
              <Link to="/discover"><ArrowLeft className="h-4 w-4 mr-1" />Voltar</Link>
            </Button>
          </div>
        )}

        {!loading && data && (
          <>
            {/* Profile header */}
            <div className="flex items-start gap-5">
              <Button variant="ghost" size="icon" asChild className="mt-1 shrink-0">
                <Link to="/discover"><ArrowLeft className="h-4 w-4" /></Link>
              </Button>
              <div className="shrink-0 w-20 h-20 rounded-full overflow-hidden bg-muted flex items-center justify-center">
                {avatarUrl ? (
                  <SkeletonImage src={avatarUrl} alt={data.fotografo.nome} className="object-cover" />
                ) : (
                  <Camera className="h-8 w-8 text-muted-foreground/30" />
                )}
              </div>
              <div>
                <h2 className="text-3xl font-bold">{data.fotografo.nome}</h2>
                {bio && <p className="text-muted-foreground mt-1 max-w-lg text-sm">{bio}</p>}
                <p className="text-muted-foreground text-sm mt-2">
                  {data.albums.length} álbum{data.albums.length !== 1 ? 's' : ''} público{data.albums.length !== 1 ? 's' : ''}
                </p>
              </div>
            </div>

            {data.albums.length === 0 ? (
              <div className="flex items-center justify-center h-48 rounded-xl border border-dashed border-border text-muted-foreground">
                <div className="text-center">
                  <Images className="h-8 w-8 opacity-30 mx-auto mb-2" />
                  <p className="text-sm">Nenhum álbum público ainda.</p>
                </div>
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                {data.albums.map((a) => <AlbumCard key={a.ID} album={a} />)}
              </div>
            )}
          </>
        )}
      </div>
    </PublicOrAppLayout>
  )
}
