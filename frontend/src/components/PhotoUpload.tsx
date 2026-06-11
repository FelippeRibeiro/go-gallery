import { useRef, useState } from 'react'
import { CloudUpload, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { uploadPhoto } from '@/api/albums'
import { cn } from '@/lib/utils'
import type { Foto } from '@/types'

interface Props {
  albumId: number
  onUploaded: (foto: Foto) => void
}

export default function PhotoUpload({ albumId, onUploaded }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [queue, setQueue] = useState(0)
  const [done, setDone] = useState(0)

  const handle = async (file: File) => {
    if (!file.type.startsWith('image/')) {
      toast.error(`${file.name} não é uma imagem.`)
      return
    }
    setQueue((q) => q + 1)
    try {
      const foto = await uploadPhoto(albumId, file)
      onUploaded(foto)
      setDone((d) => d + 1)
      toast.success(`${file.name} enviada com sucesso!`)
    } catch (e: unknown) {
      const msg =
        (e as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        `Falha ao enviar ${file.name}`
      toast.error(msg)
    } finally {
      setQueue((q) => q - 1)
    }
  }

  const onFiles = (files: FileList | null) => {
    if (!files) return
    setDone(0)
    setUploading(true)
    Promise.all(Array.from(files).map(handle)).finally(() => setUploading(false))
  }

  return (
    <div
      className={cn(
        'relative flex flex-col items-center justify-center gap-3 rounded-xl border-2 border-dashed p-8 text-center cursor-pointer transition-colors select-none',
        dragging
          ? 'border-primary bg-primary/5 text-primary'
          : 'border-border hover:border-primary/50 text-muted-foreground hover:text-foreground'
      )}
      onClick={() => !uploading && inputRef.current?.click()}
      onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={(e) => {
        e.preventDefault()
        setDragging(false)
        onFiles(e.dataTransfer.files)
      }}
    >
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        multiple
        className="hidden"
        onChange={(e) => onFiles(e.target.files)}
      />

      {uploading ? (
        <>
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
          <p className="text-sm font-medium">
            Enviando {done}/{queue + done} foto{queue + done !== 1 ? 's' : ''}…
          </p>
        </>
      ) : (
        <>
          <CloudUpload className="h-8 w-8" />
          <div>
            <p className="text-sm font-medium">Clique ou arraste as fotos aqui</p>
            <p className="text-xs text-muted-foreground mt-0.5">JPEG, PNG, GIF, WebP · máx 10 MB por arquivo</p>
          </div>
        </>
      )}
    </div>
  )
}
