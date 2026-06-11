import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Calendar, Camera, Globe, ImageOff, Images, Loader2, Users } from 'lucide-react';
import { listPublicAlbums, listPublicPhotographers } from '@/api/albums';
import PublicOrAppLayout from '@/components/PublicOrAppLayout';
import type { Album, FotografoPublico } from '@/types';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

type Tab = 'albums' | 'fotografos';

function AlbumCard({ album }: { album: Album }) {
  return (
    <Link to={`/gallery/${album.ID}`}>
      <Card className="group overflow-hidden hover:ring-1 hover:ring-primary/40 transition-all cursor-pointer">
        <div className="h-44 overflow-hidden bg-muted">
          {album.CapaUrl?.Valid ? (
            <img src={album.CapaUrl.String} alt={album.Titulo} className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300" />
          ) : (
            <div className="w-full h-full flex items-center justify-center">
              <ImageOff className="h-8 w-8 opacity-20" />
            </div>
          )}
        </div>
        <CardContent className="p-4 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <h3 className="font-semibold leading-tight line-clamp-2">{album.Titulo}</h3>
            <Badge variant="outline" className="shrink-0 text-xs">
              Público
            </Badge>
          </div>
          {album.Descricao?.Valid && <p className="text-xs text-muted-foreground line-clamp-2">{album.Descricao.String}</p>}
          <div className="flex items-center gap-1 text-xs text-muted-foreground">
            <Calendar className="h-3 w-3" />
            {new Date(album.DataEvento).toLocaleDateString('pt-BR', {
              day: '2-digit',
              month: 'short',
              year: 'numeric',
            })}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

function PhotographerCard({ fotografo }: { fotografo: FotografoPublico }) {
  const avatarUrl = fotografo.FotoPerfil?.Valid ? fotografo.FotoPerfil.String : null;
  const bio = fotografo.Bio?.Valid ? fotografo.Bio.String : null;

  return (
    <Link to={`/p/${fotografo.ID}`}>
      <Card className="group hover:ring-1 hover:ring-primary/40 transition-all cursor-pointer">
        <CardContent className="p-5 flex gap-4 items-start">
          <div className="shrink-0 w-14 h-14 rounded-full overflow-hidden bg-muted flex items-center justify-center">
            {avatarUrl ? <img src={avatarUrl} alt={fotografo.Nome} className="w-full h-full object-cover" /> : <Camera className="h-6 w-6 text-muted-foreground/40" />}
          </div>
          <div className="flex-1 min-w-0">
            <h3 className="font-semibold group-hover:text-primary transition-colors truncate">{fotografo.Nome}</h3>
            {bio && <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">{bio}</p>}
            <div className="flex items-center gap-1 text-xs text-muted-foreground mt-2">
              <Images className="h-3 w-3" />
              {fotografo.TotalAlbuns} álbum{fotografo.TotalAlbuns !== 1 ? 's' : ''} público{fotografo.TotalAlbuns !== 1 ? 's' : ''}
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export default function Discover() {
  const [tab, setTab] = useState<Tab>('albums');
  const [albums, setAlbums] = useState<Album[]>([]);
  const [fotografos, setFotografos] = useState<FotografoPublico[]>([]);
  const [loadingAlbums, setLoadingAlbums] = useState(true);
  const [loadingFotografos, setLoadingFotografos] = useState(false);
  const [fotosLoaded, setFotosLoaded] = useState(false);

  useEffect(() => {
    listPublicAlbums()
      .then(setAlbums)
      .finally(() => setLoadingAlbums(false));
  }, []);

  useEffect(() => {
    if (tab === 'fotografos' && !fotosLoaded) {
      setLoadingFotografos(true);
      listPublicPhotographers()
        .then(setFotografos)
        .finally(() => {
          setLoadingFotografos(false);
          setFotosLoaded(true);
        });
    }
  }, [tab, fotosLoaded]);

  return (
    <PublicOrAppLayout>
      <div className="space-y-8">
        <div>
          <h2 className="text-3xl font-bold">Descobrir</h2>
          <p className="text-muted-foreground mt-1">Explore álbuns e fotógrafos</p>
        </div>

        {/* Tabs */}
        <div className="flex gap-1 border-b border-border">
          <button
            onClick={() => setTab('albums')}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === 'albums' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Globe className="h-4 w-4" />
            Álbuns
            {albums.length > 0 && (
              <Badge variant="secondary" className="text-xs">
                {albums.length}
              </Badge>
            )}
          </button>
          <button
            onClick={() => setTab('fotografos')}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 transition-colors ${
              tab === 'fotografos' ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
            }`}
          >
            <Users className="h-4 w-4" />
            Fotógrafos
          </button>
        </div>

        {/* Albums tab */}
        {tab === 'albums' && (
          <>
            {loadingAlbums && (
              <div className="flex justify-center py-20">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            )}
            {!loadingAlbums && albums.length === 0 && (
              <div className="flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground">
                <ImageOff className="h-10 w-10 opacity-40" />
                <p>Nenhum álbum público disponível.</p>
              </div>
            )}
            {!loadingAlbums && albums.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                {albums.map((album) => (
                  <AlbumCard key={album.ID} album={album} />
                ))}
              </div>
            )}
          </>
        )}

        {/* Photographers tab */}
        {tab === 'fotografos' && (
          <>
            {loadingFotografos && (
              <div className="flex justify-center py-20">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            )}
            {!loadingFotografos && fotografos.length === 0 && (
              <div className="flex flex-col items-center justify-center h-64 gap-3 text-muted-foreground">
                <Users className="h-10 w-10 opacity-40" />
                <p>Nenhum fotógrafo com álbuns públicos.</p>
              </div>
            )}
            {!loadingFotografos && fotografos.length > 0 && (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {fotografos.map((f) => (
                  <PhotographerCard key={f.ID} fotografo={f} />
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </PublicOrAppLayout>
  );
}
