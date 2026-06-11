import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Camera, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { acceptInvite } from '@/api/auth'
import { useAuth } from '@/context/AuthContext'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

export default function AcceptInvite() {
  const [searchParams] = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const navigate = useNavigate()
  const { loginWith } = useAuth()

  const [nome, setNome] = useState('')
  const [senha, setSenha] = useState('')
  const [loading, setLoading] = useState(false)

  if (!token) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="w-full max-w-sm text-center">
          <CardContent className="pt-6 space-y-2">
            <p className="font-medium">Link inválido</p>
            <p className="text-sm text-muted-foreground">
              Este link de convite é inválido ou já expirou.
            </p>
          </CardContent>
        </Card>
      </div>
    )
  }

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      const res = await acceptInvite(token, nome, senha)
      loginWith(res.user)
      toast.success('Bem-vindo ao Go Gallery!')
      navigate(`/albums/${res.album_id}`)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Convite inválido ou expirado.'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="w-full max-w-sm space-y-6">
        <div className="flex flex-col items-center gap-2">
          <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
            <Camera className="h-6 w-6" />
          </div>
          <h1 className="text-2xl font-bold">Go Gallery</h1>
          <p className="text-sm text-muted-foreground text-center">
            Você foi convidado! Crie sua conta para acessar as fotos.
          </p>
        </div>

        <Card>
          <form onSubmit={submit}>
            <CardHeader className="pb-4">
              <CardTitle className="text-base">Criar conta de acesso</CardTitle>
              <CardDescription>
                Escolha um nome e uma senha para sua conta.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="nome">Seu nome</Label>
                <Input
                  id="nome"
                  required
                  autoComplete="name"
                  placeholder="Como quer ser chamado"
                  value={nome}
                  onChange={(e) => setNome(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="senha">Senha</Label>
                <Input
                  id="senha"
                  type="password"
                  required
                  minLength={6}
                  autoComplete="new-password"
                  placeholder="Mínimo 6 caracteres"
                  value={senha}
                  onChange={(e) => setSenha(e.target.value)}
                />
              </div>
            </CardContent>
            <CardFooter>
              <Button type="submit" className="w-full" disabled={loading}>
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                {loading ? 'Acessando...' : 'Criar conta e ver álbum'}
              </Button>
            </CardFooter>
          </form>
        </Card>
      </div>
    </div>
  )
}
