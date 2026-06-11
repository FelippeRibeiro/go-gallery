import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import Layout from '@/components/Layout'
import { createAlbum } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'

export default function AlbumCreate() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState({
    titulo: '',
    descricao: '',
    data_evento: '',
    lote: false,
    publico: false,
    valor_album: '',
    valor_unitario_fotografia: '',
  })

  const set = <K extends keyof typeof form>(key: K, value: (typeof form)[K]) =>
    setForm((f) => ({ ...f, [key]: value }))

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const album = await createAlbum(form)
      toast.success('Álbum criado!', { description: form.titulo })
      navigate(`/albums/${album.ID}`)
    } catch {
      toast.error('Erro ao criar álbum', { description: 'Verifique os dados e tente novamente.' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Layout>
      <div className="max-w-xl mx-auto space-y-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" asChild>
            <Link to="/dashboard"><ArrowLeft className="h-4 w-4" /></Link>
          </Button>
          <div>
            <h1 className="text-2xl font-bold">Novo Álbum</h1>
            <p className="text-sm text-muted-foreground">Preencha os dados do álbum</p>
          </div>
        </div>

        <form onSubmit={submit}>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Informações</CardTitle>
              <CardDescription>Detalhes básicos do álbum</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="titulo">Título *</Label>
                <Input
                  id="titulo"
                  required
                  placeholder="Ex: Casamento João & Maria"
                  value={form.titulo}
                  onChange={(e) => set('titulo', e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="descricao">Descrição</Label>
                <Textarea
                  id="descricao"
                  rows={3}
                  placeholder="Informações adicionais sobre o evento…"
                  value={form.descricao}
                  onChange={(e) => set('descricao', e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="data">Data do evento *</Label>
                <Input
                  id="data"
                  type="date"
                  required
                  value={form.data_evento}
                  onChange={(e) => set('data_evento', e.target.value)}
                />
              </div>
            </CardContent>

            <Separator />

            <CardHeader className="pt-4">
              <CardTitle className="text-base">Precificação</CardTitle>
              <CardDescription>Como as fotos serão vendidas</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <Label htmlFor="lote" className="text-sm font-medium cursor-pointer">
                    Vender álbum completo (lote)
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    {form.lote ? 'Preço fixo para todas as fotos' : 'Preço individual por foto'}
                  </p>
                </div>
                <Switch
                  id="lote"
                  checked={form.lote}
                  onCheckedChange={(v) => set('lote', v)}
                />
              </div>

              <div className="flex items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <Label htmlFor="publico" className="text-sm font-medium cursor-pointer">
                    Álbum público
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    {form.publico ? 'Visível para qualquer pessoa' : 'Acesso apenas por convite'}
                  </p>
                </div>
                <Switch
                  id="publico"
                  checked={form.publico}
                  onCheckedChange={(v) => set('publico', v)}
                />
              </div>

              {form.lote ? (
                <div className="space-y-2">
                  <Label htmlFor="valor-album">Valor do álbum (R$)</Label>
                  <Input
                    id="valor-album"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="0,00"
                    value={form.valor_album}
                    onChange={(e) => set('valor_album', e.target.value)}
                  />
                </div>
              ) : (
                <div className="space-y-2">
                  <Label htmlFor="valor-unitario">Valor por foto (R$)</Label>
                  <Input
                    id="valor-unitario"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="0,00"
                    value={form.valor_unitario_fotografia}
                    onChange={(e) => set('valor_unitario_fotografia', e.target.value)}
                  />
                </div>
              )}
            </CardContent>

            <Separator />

            <CardContent className="pt-4 flex flex-col-reverse sm:flex-row gap-2 sm:justify-end">
              <Button type="button" variant="outline" onClick={() => navigate(-1)}>
                Cancelar
              </Button>
              <Button type="submit" disabled={loading}>
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                {loading ? 'Criando...' : 'Criar álbum'}
              </Button>
            </CardContent>
          </Card>
        </form>
      </div>
    </Layout>
  )
}
