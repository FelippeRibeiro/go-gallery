import { useCallback, useEffect, useRef, useState } from 'react'
import { ChevronLeft, ChevronRight, X, ZoomIn, ZoomOut } from 'lucide-react'

export interface LightboxPhoto {
  id: number
  url: string
}

interface Props {
  photos: LightboxPhoto[]
  initialIndex: number
  onClose: () => void
}

export default function Lightbox({ photos, initialIndex, onClose }: Props) {
  const [index, setIndex] = useState(initialIndex)
  const [scale, setScale] = useState(1)
  const [offset, setOffset] = useState({ x: 0, y: 0 })
  const [dragging, setDragging] = useState(false)

  const containerRef = useRef<HTMLDivElement>(null)
  const dragRef = useRef<{ sx: number; sy: number; ox: number; oy: number } | null>(null)
  // Touch state for pinch-zoom and swipe
  const touchRef = useRef<{
    touches: { id: number; x: number; y: number }[]
    pinchDist: number | null
    swipeStartX: number
  }>({ touches: [], pinchDist: null, swipeStartX: 0 })

  const photo = photos[index]
  const canPrev = index > 0
  const canNext = index < photos.length - 1

  const resetTransform = () => { setScale(1); setOffset({ x: 0, y: 0 }) }
  const goTo = (i: number) => { setIndex(i); resetTransform() }

  // Lock body scroll
  useEffect(() => {
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => { document.body.style.overflow = prev }
  }, [])

  // Keyboard
  useEffect(() => {
    const handle = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
      if (e.key === 'ArrowLeft' && canPrev) goTo(index - 1)
      if (e.key === 'ArrowRight' && canNext) goTo(index + 1)
      if (e.key === '+' || e.key === '=') setScale(s => Math.min(6, s + 0.5))
      if (e.key === '-') setScale(s => { const n = Math.max(1, s - 0.5); if (n === 1) resetTransform(); return n })
    }
    window.addEventListener('keydown', handle)
    return () => window.removeEventListener('keydown', handle)
  }, [index, canPrev, canNext])

  // Wheel zoom toward cursor
  const onWheel = useCallback((e: WheelEvent) => {
    e.preventDefault()
    const el = containerRef.current
    if (!el) return

    const rect = el.getBoundingClientRect()
    // Cursor position relative to element center
    const cx = e.clientX - rect.left - rect.width / 2
    const cy = e.clientY - rect.top - rect.height / 2

    setScale(prevScale => {
      const factor = e.deltaY < 0 ? 1.12 : 0.89
      const newScale = Math.max(1, Math.min(6, prevScale * factor))

      if (newScale === 1) {
        setOffset({ x: 0, y: 0 })
        return 1
      }

      // Keep the image point under the cursor fixed
      // Point in image-space: p = (cursor - offset) / prevScale
      // After zoom: offset_new = cursor - newScale * p
      setOffset(prev => ({
        x: cx - (newScale / prevScale) * (cx - prev.x),
        y: cy - (newScale / prevScale) * (cy - prev.y),
      }))
      return newScale
    })
  }, [])

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
  }, [onWheel])

  // ── Mouse drag ───────────────────────────────────────────────────────────────
  const onMouseDown = (e: React.MouseEvent) => {
    if (scale <= 1) return
    e.preventDefault()
    setDragging(true)
    dragRef.current = { sx: e.clientX, sy: e.clientY, ox: offset.x, oy: offset.y }
  }

  const onMouseMove = (e: React.MouseEvent) => {
    if (!dragging || !dragRef.current) return
    setOffset({
      x: dragRef.current.ox + (e.clientX - dragRef.current.sx),
      y: dragRef.current.oy + (e.clientY - dragRef.current.sy),
    })
  }

  const stopDrag = () => { setDragging(false); dragRef.current = null }

  const onDoubleClick = (e: React.MouseEvent) => {
    const el = containerRef.current
    if (!el) return
    const rect = el.getBoundingClientRect()
    const cx = e.clientX - rect.left - rect.width / 2
    const cy = e.clientY - rect.top - rect.height / 2

    if (scale > 1) {
      resetTransform()
    } else {
      const newScale = 2.5
      setScale(newScale)
      // Zoom toward double-click point
      setOffset({ x: cx * (1 - newScale), y: cy * (1 - newScale) })
    }
  }

  // ── Touch handlers ───────────────────────────────────────────────────────────
  const onTouchStart = (e: React.TouchEvent) => {
    const t = touchRef.current
    t.touches = Array.from(e.touches).map(x => ({ id: x.identifier, x: x.clientX, y: x.clientY }))
    if (t.touches.length === 1) {
      t.swipeStartX = t.touches[0].x
      if (scale > 1) {
        dragRef.current = { sx: t.touches[0].x, sy: t.touches[0].y, ox: offset.x, oy: offset.y }
      }
      t.pinchDist = null
    }
    if (t.touches.length === 2) {
      const dx = t.touches[0].x - t.touches[1].x
      const dy = t.touches[0].y - t.touches[1].y
      t.pinchDist = Math.hypot(dx, dy)
      dragRef.current = null
    }
  }

  const onTouchMove = (e: React.TouchEvent) => {
    e.preventDefault()
    const t = touchRef.current
    const touches = Array.from(e.touches)

    if (touches.length === 1 && scale > 1 && dragRef.current) {
      const touch = touches[0]
      setOffset({
        x: dragRef.current.ox + (touch.clientX - dragRef.current.sx),
        y: dragRef.current.oy + (touch.clientY - dragRef.current.sy),
      })
    }

    if (touches.length === 2 && t.pinchDist !== null) {
      const dx = touches[0].clientX - touches[1].clientX
      const dy = touches[0].clientY - touches[1].clientY
      const newDist = Math.hypot(dx, dy)
      const factor = newDist / t.pinchDist
      t.pinchDist = newDist

      // Pinch center
      const el = containerRef.current
      if (el) {
        const rect = el.getBoundingClientRect()
        const cx = (touches[0].clientX + touches[1].clientX) / 2 - rect.left - rect.width / 2
        const cy = (touches[0].clientY + touches[1].clientY) / 2 - rect.top - rect.height / 2

        setScale(prev => {
          const newScale = Math.max(1, Math.min(6, prev * factor))
          if (newScale === 1) { setOffset({ x: 0, y: 0 }); return 1 }
          setOffset(o => ({
            x: cx - (newScale / prev) * (cx - o.x),
            y: cy - (newScale / prev) * (cy - o.y),
          }))
          return newScale
        })
      }
    }
  }

  const onTouchEnd = (e: React.TouchEvent) => {
    const t = touchRef.current
    const remaining = Array.from(e.touches).length
    if (remaining === 0 && t.touches.length === 1 && scale <= 1) {
      // Swipe navigation
      const dx = e.changedTouches[0].clientX - t.swipeStartX
      if (Math.abs(dx) > 60) {
        if (dx > 0 && canPrev) goTo(index - 1)
        if (dx < 0 && canNext) goTo(index + 1)
      }
    }
    t.touches = []
    t.pinchDist = null
    dragRef.current = null
  }

  return (
    <div className="fixed inset-0 z-50 bg-black flex items-center justify-center select-none">
      {/* Close button */}
      <button
        onClick={onClose}
        className="absolute top-4 right-4 z-20 rounded-full bg-white/10 hover:bg-white/20 p-2 text-white/70 hover:text-white transition-colors"
      >
        <X className="h-5 w-5" />
      </button>

      {/* Image area */}
      <div
        ref={containerRef}
        className="absolute inset-0 flex items-center justify-center overflow-hidden"
        onMouseDown={onMouseDown}
        onMouseMove={onMouseMove}
        onMouseUp={stopDrag}
        onMouseLeave={stopDrag}
        onDoubleClick={onDoubleClick}
        onTouchStart={onTouchStart}
        onTouchMove={onTouchMove}
        onTouchEnd={onTouchEnd}
        style={{ cursor: scale > 1 ? (dragging ? 'grabbing' : 'grab') : 'zoom-in' }}
        onClick={(e) => { if (e.target === e.currentTarget && scale === 1) onClose() }}
      >
        <img
          key={photo.id}
          src={photo.url}
          alt={`Foto ${photo.id}`}
          className="max-w-full max-h-full object-contain pointer-events-none"
          style={{
            transform: `translate(${offset.x}px, ${offset.y}px) scale(${scale})`,
            transformOrigin: 'center center',
            transition: dragging ? 'none' : 'transform 0.08s ease-out',
          }}
          draggable={false}
        />
      </div>

      {/* Prev / Next */}
      {canPrev && (
        <button
          onClick={() => goTo(index - 1)}
          className="absolute left-3 top-1/2 -translate-y-1/2 z-20 rounded-full bg-white/10 hover:bg-white/20 p-3 text-white/70 hover:text-white transition-colors"
        >
          <ChevronLeft className="h-6 w-6" />
        </button>
      )}
      {canNext && (
        <button
          onClick={() => goTo(index + 1)}
          className="absolute right-3 top-1/2 -translate-y-1/2 z-20 rounded-full bg-white/10 hover:bg-white/20 p-3 text-white/70 hover:text-white transition-colors"
        >
          <ChevronRight className="h-6 w-6" />
        </button>
      )}

      {/* Bottom bar */}
      <div className="absolute bottom-5 left-1/2 -translate-x-1/2 z-20 flex items-center gap-3 rounded-full bg-black/60 backdrop-blur-sm px-4 py-2 text-sm">
        <button
          onClick={() => setScale(s => { const n = Math.max(1, +(s - 0.5).toFixed(1)); if (n === 1) resetTransform(); return n })}
          disabled={scale <= 1}
          className="text-white/60 hover:text-white disabled:opacity-30 transition-colors"
        >
          <ZoomOut className="h-4 w-4" />
        </button>

        <button
          onClick={resetTransform}
          className="text-white/60 hover:text-white transition-colors tabular-nums min-w-[46px] text-center text-xs"
        >
          {Math.round(scale * 100)}%
        </button>

        <button
          onClick={() => setScale(s => {
            const newScale = Math.min(6, +(s + 0.5).toFixed(1))
            if (s === 1) setOffset({ x: 0, y: 0 })
            return newScale
          })}
          disabled={scale >= 6}
          className="text-white/60 hover:text-white disabled:opacity-30 transition-colors"
        >
          <ZoomIn className="h-4 w-4" />
        </button>

        {photos.length > 1 && (
          <>
            <span className="text-white/20 mx-1">|</span>
            <span className="text-white/50 text-xs tabular-nums">
              {index + 1} / {photos.length}
            </span>
          </>
        )}
      </div>

      {/* Hint (only shows briefly) */}
      {scale === 1 && (
        <p className="absolute top-4 left-1/2 -translate-x-1/2 z-20 text-white/30 text-xs pointer-events-none">
          Scroll para zoom · Duplo clique para ampliar · ← → navegar
        </p>
      )}
    </div>
  )
}
