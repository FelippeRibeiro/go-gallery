import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Calendar, ImageOff, Loader2 } from 'lucide-react'
import { listPublicAlbums } from '@/api/albums'
import PublicOrAppLayout from '@/components/PublicOrAppLayout'
import type { Album } from '@/types'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

export default function PublicGallery() {
  const [albums, setAlbums] = useState<Album[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    listPublicAlbums()
      .then(setAlbums)
      .finally(() => setLoading(false))
  }, [])

  return (
    <PublicOrAppLayout>
      <div className="space-y-8">
        <div>
          <h2 className="text-3xl font-bold">Galeria pública</h2>
          <p className="text-muted-foreground mt-1">Álbuns disponíveis para todos</p>
        </div>

        {loading && (
          <div className="flex justify-center py-20">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        )}

        {!loading && albums.length === 0 && (
          <div className="flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground">
            <ImageOff className="h-10 w-10 opacity-40" />
            <p>Nenhum álbum público disponível.</p>
          </div>
        )}

        {!loading && albums.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {albums.map((album) => (
              <Link key={album.ID} to={`/albums/${album.ID}`}>
                <Card className="overflow-hidden hover:ring-1 hover:ring-primary/40 transition-all cursor-pointer group">
                  {album.CapaUrl?.Valid ? (
                    <div className="h-44 overflow-hidden">
                      <img
                        src={album.CapaUrl.String}
                        alt={album.Titulo}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    </div>
                  ) : (
                    <div className="h-44 bg-muted flex items-center justify-center">
                      <ImageOff className="h-8 w-8 opacity-30" />
                    </div>
                  )}
                  <CardContent className="p-4 space-y-2">
                    <div className="flex items-start justify-between gap-2">
                      <h3 className="font-semibold leading-tight line-clamp-2">{album.Titulo}</h3>
                      <Badge variant="outline" className="shrink-0 text-xs">Público</Badge>
                    </div>
                    {album.Descricao?.Valid && (
                      <p className="text-xs text-muted-foreground line-clamp-2">{album.Descricao.String}</p>
                    )}
                    <div className="flex items-center gap-1 text-xs text-muted-foreground">
                      <Calendar className="h-3 w-3" />
                      {new Date(album.DataEvento).toLocaleDateString('pt-BR', {
                        day: '2-digit', month: 'short', year: 'numeric',
                      })}
                    </div>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </div>
    </PublicOrAppLayout>
  )
}
