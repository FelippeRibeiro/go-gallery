import { useState } from 'react'
import { ImageOff } from 'lucide-react'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface Props extends React.ImgHTMLAttributes<HTMLImageElement> {
  /** Classes do contêiner, que dita o layout; por padrão preenche o pai. */
  wrapperClassName?: string
}

/**
 * Substituto de <img> que mostra um skeleton enquanto o arquivo carrega e um
 * ícone de fallback se o carregamento falhar. O skeleton é um overlay — as
 * classes passadas em className continuam valendo só para o <img>.
 */
export default function SkeletonImage({ wrapperClassName, className, alt, onLoad, onError, ...props }: Props) {
  const [loaded, setLoaded] = useState(false)
  const [failed, setFailed] = useState(false)

  return (
    <div className={cn('relative h-full w-full overflow-hidden', wrapperClassName)}>
      {!failed && (
        <img
          {...props}
          alt={alt}
          draggable={false}
          // imagens vindas do cache podem não disparar onLoad após o mount
          ref={(img) => { if (img?.complete && img.naturalWidth > 0) setLoaded(true) }}
          onLoad={(e) => { setLoaded(true); onLoad?.(e) }}
          onError={(e) => { setFailed(true); onError?.(e) }}
          className={cn('h-full w-full', className)}
        />
      )}
      {!loaded && !failed && <Skeleton className="absolute inset-0 rounded-none" />}
      {failed && (
        <div className="absolute inset-0 flex items-center justify-center bg-muted">
          <ImageOff className="h-6 w-6 text-muted-foreground/40" />
        </div>
      )}
    </div>
  )
}
