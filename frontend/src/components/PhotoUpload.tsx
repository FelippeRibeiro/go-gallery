import { useRef, useState } from 'react'
import { CloudUpload, CheckCircle2, XCircle } from 'lucide-react'
import { uploadPhoto } from '@/api/albums'
import { cn } from '@/lib/utils'
import type { Foto } from '@/types'

interface Props {
  albumId: number
  onUploaded: (foto: Foto) => void
}

interface FileEntry {
  id: string
  name: string
  progress: number
  status: 'uploading' | 'done' | 'error'
  error?: string
}

export default function PhotoUpload({ albumId, onUploaded }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [files, setFiles] = useState<FileEntry[]>([])

  const updateFile = (id: string, patch: Partial<FileEntry>) =>
    setFiles((prev) => prev.map((f) => (f.id === id ? { ...f, ...patch } : f)))

  const handle = async (file: File) => {
    if (!file.type.startsWith('image/')) return
    const id = Math.random().toString(36).slice(2)
    setFiles((prev) => [...prev, { id, name: file.name, progress: 0, status: 'uploading' }])
    try {
      const foto = await uploadPhoto(albumId, file, (pct) => updateFile(id, { progress: pct }))
      updateFile(id, { progress: 100, status: 'done' })
      onUploaded(foto)
    } catch (e: unknown) {
      const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Falha'
      updateFile(id, { status: 'error', error: msg })
    }
  }

  const onFiles = (list: FileList | null) => {
    if (!list) return
    setFiles([])
    Array.from(list).filter((f) => f.type.startsWith('image/')).forEach(handle)
  }

  const active = files.some((f) => f.status === 'uploading')

  return (
    <div className="space-y-3">
      <div
        className={cn(
          'relative flex flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed p-8 text-center cursor-pointer transition-colors select-none',
          dragging
            ? 'border-primary bg-primary/5 text-primary'
            : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
        )}
        onClick={() => !active && inputRef.current?.click()}
        onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
        onDragLeave={() => setDragging(false)}
        onDrop={(e) => { e.preventDefault(); setDragging(false); onFiles(e.dataTransfer.files) }}
      >
        <input ref={inputRef} type="file" accept="image/*" multiple className="hidden" onChange={(e) => onFiles(e.target.files)} />
        <CloudUpload className="h-8 w-8" />
        <div>
          <p className="text-sm font-medium">Clique ou arraste as fotos aqui</p>
          <p className="text-xs text-muted-foreground mt-0.5">JPEG, PNG, WebP · máx 10 MB por arquivo</p>
        </div>
      </div>

      {files.length > 0 && (
        <div className="space-y-1.5 max-h-48 overflow-y-auto pr-1">
          {files.map((f) => (
            <div key={f.id} className="flex items-center gap-2 text-sm">
              {f.status === 'done'
                ? <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />
                : f.status === 'error'
                ? <XCircle className="h-4 w-4 text-destructive shrink-0" />
                : <div className="h-4 w-4 rounded-full border-2 border-primary border-t-transparent animate-spin shrink-0" />
              }
              <span className="truncate flex-1 text-muted-foreground">{f.name}</span>
              {f.status === 'uploading' && (
                <div className="w-20 shrink-0">
                  <div className="h-1 rounded-full bg-muted overflow-hidden">
                    <div className="h-full bg-primary transition-all duration-150" style={{ width: `${f.progress}%` }} />
                  </div>
                </div>
              )}
              {f.status === 'error' && <span className="text-xs text-destructive shrink-0">{f.error}</span>}
              {f.status === 'done' && <span className="text-xs text-green-500 shrink-0">Enviado</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
