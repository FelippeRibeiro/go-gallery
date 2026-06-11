import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Clock, Users, Camera, Trash2, Loader2, UserPlus, Mail, RefreshCw } from 'lucide-react'
import { listInvites, revokeInvite, resendInvite, invitePhotographer, removePhotographer, removeClient } from '@/api/albums'
import type { Convite, Colaborador, Cliente } from '@/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

interface Props {
  albumId: number
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Tab = 'pending' | 'clients' | 'photographers'

function conviteStatus(c: Convite): 'pending' | 'accepted' | 'expired' {
  if (c.UsedAt?.Valid) return 'accepted'
  if (new Date(c.ExpiresAt) < new Date()) return 'expired'
  return 'pending'
}

export default function InvitesManager({ albumId, open, onOpenChange }: Props) {
  const [tab, setTab] = useState<Tab>('clients')
  const [convites, setConvites] = useState<Convite[]>([])
  const [fotografos, setFotografos] = useState<Colaborador[]>([])
  const [clientes, setClientes] = useState<Cliente[]>([])
  const [loading, setLoading] = useState(false)
  const [fotografoEmail, setFotografoEmail] = useState('')
  const [invitingPhotographer, setInvitingPhotographer] = useState(false)

  const reload = () => {
    setLoading(true)
    listInvites(albumId)
      .then((res) => {
        setConvites(res.convites ?? [])
        setFotografos(res.fotografos ?? [])
        setClientes(res.clientes ?? [])
      })
      .catch(() => toast.error('Erro ao carregar participantes'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    if (open) reload()
  }, [open, albumId])

  const handleRevoke = async (c: Convite) => {
    try {
      await revokeInvite(albumId, c.ID)
      setConvites((prev) => prev.filter((x) => x.ID !== c.ID))
      toast.success(`Convite de ${c.Email} revogado`)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao revogar convite'
      toast.error(msg)
    }
  }

  const handleResend = async (c: Convite) => {
    try {
      await resendInvite(albumId, c.ID)
      toast.success(`Email reenviado para ${c.Email}`)
    } catch {
      toast.error('Erro ao reenviar convite')
    }
  }

  const handleRemoveClient = async (c: Cliente) => {
    try {
      await removeClient(albumId, c.ID)
      setClientes((prev) => prev.filter((x) => x.ID !== c.ID))
      toast.success(`${c.Nome} removido do álbum`)
    } catch {
      toast.error('Erro ao remover cliente')
    }
  }

  const handleInvitePhotographer = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!fotografoEmail) return
    setInvitingPhotographer(true)
    try {
      const fa = await invitePhotographer(albumId, fotografoEmail) as Colaborador
      setFotografos((prev) => [...prev, fa])
      setFotografoEmail('')
      toast.success(`Fotógrafo ${fa.Email ?? fotografoEmail} adicionado ao álbum!`)
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Erro ao convidar fotógrafo'
      toast.error(msg)
    } finally {
      setInvitingPhotographer(false)
    }
  }

  const handleRemovePhotographer = async (fa: Colaborador) => {
    try {
      await removePhotographer(albumId, fa.ID)
      setFotografos((prev) => prev.filter((x) => x.ID !== fa.ID))
      toast.success(`${fa.Nome} removido do álbum`)
    } catch {
      toast.error('Erro ao remover fotógrafo')
    }
  }

  const pending = convites.filter((c) => conviteStatus(c) === 'pending')

  const tabs: { id: Tab; label: string; icon: React.ElementType; count?: number }[] = [
    { id: 'clients', label: 'Clientes', icon: Users, count: clientes.length },
    { id: 'pending', label: 'Pendentes', icon: Clock, count: pending.length },
    { id: 'photographers', label: 'Fotógrafos', icon: Camera, count: fotografos.length },
  ]

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>Participantes</DialogTitle>
          <DialogDescription>
            Gerencie clientes e fotógrafos colaboradores do álbum
          </DialogDescription>
        </DialogHeader>

        {/* Tab bar */}
        <div className="flex gap-1 rounded-lg bg-muted p-1 shrink-0">
          {tabs.map(({ id, label, icon: Icon, count }) => (
            <button
              key={id}
              onClick={() => setTab(id)}
              className={cn(
                'flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors',
                tab === id
                  ? 'bg-background text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Icon className="h-3.5 w-3.5 shrink-0" />
              <span className="hidden sm:inline">{label}</span>
              {count !== undefined && count > 0 && (
                <span className={cn(
                  'text-[10px] rounded-full px-1.5 py-0.5 min-w-[18px] text-center',
                  tab === id ? 'bg-primary text-primary-foreground' : 'bg-muted-foreground/20'
                )}>
                  {count}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Scrollable content */}
        <div className="overflow-y-auto flex-1 min-h-0">
          {loading ? (
            <div className="flex justify-center py-8">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <>
              {/* Clientes */}
              {tab === 'clients' && (
                <div className="space-y-2">
                  {clientes.length === 0 ? (
                    <p className="text-sm text-muted-foreground text-center py-6">
                      Nenhum cliente com acesso ao álbum.
                    </p>
                  ) : (
                    clientes.map((c) => (
                      <div key={c.ID} className="flex items-center justify-between rounded-lg border border-border px-4 py-3 gap-3">
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground text-xs font-bold">
                            {c.Nome?.charAt(0).toUpperCase()}
                          </div>
                          <div className="min-w-0">
                            <p className="text-sm font-medium truncate">{c.Nome}</p>
                            <p className="text-xs text-muted-foreground truncate">{c.Email}</p>
                          </div>
                        </div>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="h-7 w-7 text-destructive hover:text-destructive shrink-0"
                          onClick={() => handleRemoveClient(c)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    ))
                  )}
                </div>
              )}

              {/* Convites pendentes */}
              {tab === 'pending' && (
                <div className="space-y-2">
                  {pending.length === 0 ? (
                    <p className="text-sm text-muted-foreground text-center py-6">Nenhum convite pendente.</p>
                  ) : (
                    pending.map((c) => (
                      <div key={c.ID} className="flex items-center justify-between rounded-lg border border-border px-4 py-3 gap-3">
                        <div className="min-w-0">
                          <p className="text-sm font-medium truncate">{c.Email}</p>
                          <p className="text-xs text-muted-foreground">
                            Expira em {new Date(c.ExpiresAt).toLocaleDateString('pt-BR')}
                          </p>
                        </div>
                        <div className="flex items-center gap-2 shrink-0">
                          <Badge variant="outline" className="text-xs">Pendente</Badge>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="h-7 w-7 text-muted-foreground hover:text-foreground"
                            onClick={() => handleResend(c)}
                            title="Reenviar email"
                          >
                            <RefreshCw className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="h-7 w-7 text-destructive hover:text-destructive"
                            onClick={() => handleRevoke(c)}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              )}

              {/* Fotógrafos colaboradores */}
              {tab === 'photographers' && (
                <div className="space-y-4">
                  <form onSubmit={handleInvitePhotographer} className="flex gap-2">
                    <div className="relative flex-1">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
                      <Input
                        type="email"
                        required
                        placeholder="email@fotografo.com"
                        className="pl-9"
                        value={fotografoEmail}
                        onChange={(e) => setFotografoEmail(e.target.value)}
                      />
                    </div>
                    <Button type="submit" disabled={invitingPhotographer} size="sm">
                      {invitingPhotographer
                        ? <Loader2 className="h-4 w-4 animate-spin" />
                        : <UserPlus className="h-4 w-4" />
                      }
                      <span className="hidden sm:inline ml-1">Adicionar</span>
                    </Button>
                  </form>

                  <div className="space-y-2">
                    {fotografos.length === 0 ? (
                      <p className="text-sm text-muted-foreground text-center py-4">
                        Nenhum fotógrafo colaborador. Use o formulário acima para adicionar.
                      </p>
                    ) : (
                      fotografos.map((fa) => (
                        <div key={fa.ID} className="flex items-center justify-between rounded-lg border border-border px-4 py-3 gap-3">
                          <div className="flex items-center gap-3 min-w-0">
                            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary text-xs font-bold">
                              {fa.Nome?.charAt(0).toUpperCase()}
                            </div>
                            <div className="min-w-0">
                              <p className="text-sm font-medium truncate">{fa.Nome}</p>
                              <p className="text-xs text-muted-foreground truncate">{fa.Email}</p>
                            </div>
                          </div>
                          <Button
                            size="icon"
                            variant="ghost"
                            className="h-7 w-7 text-destructive hover:text-destructive shrink-0"
                            onClick={() => handleRemovePhotographer(fa)}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
