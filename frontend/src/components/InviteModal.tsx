import { useState } from 'react'
import { Mail } from 'lucide-react'
import { toast } from 'sonner'
import { inviteClient } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

interface Props {
  albumId: number
  albumTitulo: string
  children?: React.ReactNode
}

export default function InviteModal({ albumId, albumTitulo, children }: Props) {
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await inviteClient(albumId, email)
      toast.success(`Convite enviado para ${email}!`, {
        description: 'O cliente tem 7 dias para aceitar.',
      })
      setEmail('')
      setOpen(false)
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'Não foi possível enviar o convite.'
      toast.error(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {children ?? (
          <Button variant="outline">
            <Mail className="h-4 w-4" />
            Convidar cliente
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Convidar cliente</DialogTitle>
          <DialogDescription>
            Envie um convite para o álbum <strong>{albumTitulo}</strong>. O cliente receberá um link para criar uma conta e acessar as fotos.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={submit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="invite-email">E-mail do cliente</Label>
            <Input
              id="invite-email"
              type="email"
              required
              placeholder="cliente@email.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? 'Enviando...' : 'Enviar convite'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
