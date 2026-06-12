import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Calendar, FolderOpen, Globe, Images, Lock, PlusCircle, Search } from 'lucide-react'
import { toast } from 'sonner'
import Layout from '@/components/Layout'
import { listAlbums } from '@/api/albums'
import { useAuth } from '@/context/AuthContext'
import type { Album } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import SkeletonImage from '@/components/SkeletonImage'

function AlbumCard({ album }: { album: Album }) {
  return (
    <Link to={`/albums/${album.ID}`}>
      <Card className="group h-full overflow-hidden transition-colors hover:border-primary/50 cursor-pointer">
        {/* Cover */}
        <div className="relative aspect-video bg-muted overflow-hidden">
          {album.CapaUrl?.Valid ? (
            <SkeletonImage
              src={album.CapaUrl.String}
              alt={album.Titulo}
              className="object-cover transition-transform duration-300 group-hover:scale-105"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <Images className="h-10 w-10 text-muted-foreground/20" />
            </div>
          )}
          {/* Badges overlay */}
          <div className="absolute top-2 left-2 flex gap-1.5">
            <Badge variant={album.Lote ? 'default' : 'secondary'} className="text-[11px]">
              {album.Lote ? 'Lote' : 'Por foto'}
            </Badge>
          </div>
          <div className="absolute top-2 right-2">
            <Badge variant="outline" className="text-[11px] bg-black/40 border-white/10 text-white/80 backdrop-blur-sm gap-1">
              {album.Publico
                ? <><Globe className="h-3 w-3" />Público</>
                : <><Lock className="h-3 w-3" />Privado</>}
            </Badge>
          </div>
        </div>

        {/* Info */}
        <CardContent className="pt-3 pb-4">
          <h3 className="font-semibold line-clamp-1 group-hover:text-primary transition-colors">
            {album.Titulo}
          </h3>
          {album.Descricao?.Valid && (
            <p className="text-sm text-muted-foreground line-clamp-2 mt-0.5">{album.Descricao.String}</p>
          )}
          <div className="flex items-center justify-between mt-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <Calendar className="h-3.5 w-3.5" />
              {new Date(album.DataEvento).toLocaleDateString('pt-BR', {
                day: '2-digit',
                month: 'short',
                year: 'numeric',
              })}
            </span>
            <span>
              {album.Lote ? `R$ ${album.ValorAlbum}` : `R$ ${album.ValorUnitarioFotografia}/foto`}
            </span>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}

function DashboardSkeleton() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <Card key={i} className="overflow-hidden">
          <Skeleton className="aspect-video w-full rounded-none" />
          <CardContent className="pt-3 pb-4">
            <Skeleton className="h-4 w-3/4" />
            <Skeleton className="h-3 w-full mt-2" />
            <Skeleton className="h-3 w-1/2 mt-3" />
          </CardContent>
        </Card>
      ))}
    </div>
  )
}

export default function Dashboard() {
  const { user } = useAuth()
  const isClient = user?.tipo === 'cliente'
  const [albums, setAlbums] = useState<Album[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    listAlbums()
      .then(setAlbums)
      .catch(() => toast.error('Erro ao carregar álbuns'))
      .finally(() => setLoading(false))
  }, [])

  const filtered = albums.filter((a) =>
    a.Titulo.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <Layout>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">{isClient ? 'Meus álbuns' : 'Meus Álbuns'}</h1>
            <p className="text-sm text-muted-foreground mt-1">
              {loading
                ? '…'
                : search
                  ? `${filtered.length} de ${albums.length} álbum${albums.length !== 1 ? 's' : ''}`
                  : `${albums.length} álbum${albums.length !== 1 ? 's' : ''}`}
            </p>
          </div>
          {!isClient && (
            <Button asChild>
              <Link to="/albums/new">
                <PlusCircle className="h-4 w-4" />
                Novo álbum
              </Link>
            </Button>
          )}
        </div>

        <div className="relative max-w-sm">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <Input
            placeholder="Buscar álbuns…"
            className="pl-9"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>

        {loading && <DashboardSkeleton />}

        {!loading && albums.length === 0 && (
          <Card className="border-dashed">
            <CardContent className="flex flex-col items-center justify-center py-16 gap-4 text-center">
              <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <FolderOpen className="h-7 w-7" />
              </div>
              <div>
                <p className="font-medium">Nenhum álbum ainda</p>
                <p className="text-sm text-muted-foreground mt-1">
                  {isClient
                    ? 'Você ainda não foi convidado para nenhum álbum.'
                    : 'Crie seu primeiro álbum para começar a organizar suas fotos.'}
                </p>
              </div>
              {!isClient && (
                <Button asChild>
                  <Link to="/albums/new">
                    <PlusCircle className="h-4 w-4" />
                    Criar primeiro álbum
                  </Link>
                </Button>
              )}
            </CardContent>
          </Card>
        )}

        {!loading && albums.length > 0 && filtered.length === 0 && (
          <Card className="border-dashed">
            <CardContent className="flex flex-col items-center justify-center py-12 gap-3 text-center">
              <FolderOpen className="h-8 w-8 text-muted-foreground/40" />
              <p className="text-sm text-muted-foreground">Nenhum álbum encontrado para &quot;{search}&quot;</p>
            </CardContent>
          </Card>
        )}

        {!loading && filtered.length > 0 && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {filtered.map((album) => (
              <AlbumCard key={album.ID} album={album} />
            ))}
          </div>
        )}
      </div>
    </Layout>
  )
}
