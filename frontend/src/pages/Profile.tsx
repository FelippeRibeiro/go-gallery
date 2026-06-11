import { useEffect, useRef, useState } from 'react'
import { Camera, ExternalLink, Loader2, Pencil, Upload } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'
import Layout from '@/components/Layout'
import { getProfile, updateProfile, uploadProfilePhoto } from '@/api/profile'
import { useAuth } from '@/context/AuthContext'
import type { Perfil } from '@/types'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Skeleton } from '@/components/ui/skeleton'

export default function Profile() {
  const { user } = useAuth()
  const [perfil, setPerfil] = useState<Perfil | null>(null)
  const [loading, setLoading] = useState(true)
  const [bio, setBio] = useState('')
  const [editingBio, setEditingBio] = useState(false)
  const [savingBio, setSavingBio] = useState(false)
  const [uploadingPhoto, setUploadingPhoto] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    getProfile()
      .then((p) => {
        setPerfil(p)
        setBio(p.bio?.Valid ? p.bio.String : '')
      })
      .finally(() => setLoading(false))
  }, [])

  const handleSaveBio = async () => {
    setSavingBio(true)
    try {
      await updateProfile(bio)
      setPerfil((prev) => prev ? { ...prev, bio: { String: bio, Valid: bio.length > 0 } } : prev)
      setEditingBio(false)
      toast.success('Perfil atualizado')
    } catch {
      toast.error('Erro ao salvar bio')
    } finally {
      setSavingBio(false)
    }
  }

  const handlePhotoChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploadingPhoto(true)
    try {
      const res = await uploadProfilePhoto(file)
      setPerfil((prev) => prev ? { ...prev, foto_perfil: { String: res.foto_perfil, Valid: true } } : prev)
      toast.success('Foto atualizada')
    } catch {
      toast.error('Erro ao enviar foto')
    } finally {
      setUploadingPhoto(false)
      if (fileRef.current) fileRef.current.value = ''
    }
  }

  const avatarUrl = perfil?.foto_perfil?.Valid ? perfil.foto_perfil.String : null

  return (
    <Layout>
      <div className="max-w-2xl mx-auto space-y-8">
        <div>
          <h1 className="text-2xl font-bold">Meu perfil</h1>
          <p className="text-muted-foreground text-sm mt-1">
            Personalize como você aparece para os clientes
          </p>
        </div>

        {loading && (
          <div className="space-y-4">
            <Skeleton className="h-24 w-24 rounded-full" />
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-20 w-full" />
          </div>
        )}

        {!loading && perfil && (
          <div className="space-y-8">
            {/* Avatar */}
            <div className="space-y-3">
              <p className="text-sm font-medium">Foto de perfil</p>
              <div className="flex items-center gap-5">
                <div className="relative w-24 h-24 rounded-full overflow-hidden bg-muted flex items-center justify-center shrink-0">
                  {uploadingPhoto ? (
                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                  ) : avatarUrl ? (
                    <img src={avatarUrl} alt={perfil.nome} className="w-full h-full object-cover" />
                  ) : (
                    <Camera className="h-10 w-10 text-muted-foreground/30" />
                  )}
                </div>
                <div className="space-y-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => fileRef.current?.click()}
                    disabled={uploadingPhoto}
                    className="gap-2"
                  >
                    <Upload className="h-4 w-4" />
                    {avatarUrl ? 'Trocar foto' : 'Adicionar foto'}
                  </Button>
                  <p className="text-xs text-muted-foreground">JPG, PNG ou WebP. Será cortada em quadrado.</p>
                </div>
                <input
                  ref={fileRef}
                  type="file"
                  accept="image/*"
                  className="hidden"
                  onChange={handlePhotoChange}
                />
              </div>
            </div>

            {/* Nome (somente leitura) */}
            <div className="space-y-1">
              <p className="text-sm font-medium">Nome</p>
              <p className="text-foreground">{perfil.nome}</p>
            </div>

            {/* Email (somente leitura) */}
            <div className="space-y-1">
              <p className="text-sm font-medium">Email</p>
              <p className="text-muted-foreground text-sm">{perfil.email}</p>
            </div>

            {/* Bio */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <p className="text-sm font-medium">Bio</p>
                {!editingBio && (
                  <Button variant="ghost" size="sm" onClick={() => setEditingBio(true)} className="gap-1 h-7 text-xs">
                    <Pencil className="h-3 w-3" />
                    Editar
                  </Button>
                )}
              </div>
              {editingBio ? (
                <div className="space-y-2">
                  <Textarea
                    value={bio}
                    onChange={(e) => setBio(e.target.value)}
                    placeholder="Conte um pouco sobre você e seu trabalho..."
                    rows={4}
                    maxLength={500}
                    className="resize-none"
                  />
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-muted-foreground">{bio.length}/500</span>
                    <div className="flex gap-2">
                      <Button variant="ghost" size="sm" onClick={() => { setEditingBio(false); setBio(perfil.bio?.Valid ? perfil.bio.String : '') }}>
                        Cancelar
                      </Button>
                      <Button size="sm" onClick={handleSaveBio} disabled={savingBio}>
                        {savingBio && <Loader2 className="h-3 w-3 animate-spin mr-1" />}
                        Salvar
                      </Button>
                    </div>
                  </div>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  {perfil.bio?.Valid && perfil.bio.String
                    ? perfil.bio.String
                    : <span className="italic">Sem bio ainda.</span>
                  }
                </p>
              )}
            </div>

            {/* Link público */}
            {user?.tipo === 'fotografo' && (
              <div className="rounded-lg border border-border p-4 space-y-1">
                <p className="text-sm font-medium">Seu perfil público</p>
                <p className="text-xs text-muted-foreground">
                  Clientes podem ver seus álbuns públicos neste link:
                </p>
                <Link
                  to={`/p/${user.id}`}
                  className="text-sm text-primary hover:underline flex items-center gap-1 w-fit"
                >
                  /p/{user.id}
                  <ExternalLink className="h-3 w-3" />
                </Link>
              </div>
            )}
          </div>
        )}
      </div>
    </Layout>
  )
}
