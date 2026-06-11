import { useEffect, useRef } from 'react'
import type { FotoPublica } from '@/types'

export interface WsEvent {
  type: 'new_photo'
  data: FotoPublica
}

export function useAlbumSocket(albumId: number, onNewPhoto: (foto: FotoPublica) => void) {
  const onNewPhotoRef = useRef(onNewPhoto)
  onNewPhotoRef.current = onNewPhoto

  useEffect(() => {
    // Só conecta se houver sessão; o cookie httpOnly é enviado automaticamente
    // no handshake do WebSocket (mesma origem), sem token na URL.
    if (!sessionStorage.getItem('user')) return

    const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${location.host}/ws/albums/${albumId}`

    let ws: WebSocket | null = new WebSocket(url)
    let dead = false

    ws.onopen = () => {
      // connected
    }

    ws.onmessage = (e) => {
      try {
        const event: WsEvent = JSON.parse(e.data)
        if (event.type === 'new_photo') {
          onNewPhotoRef.current(event.data)
        }
      } catch {
        // ignore malformed messages
      }
    }

    ws.onerror = () => {
      if (!dead) console.warn('[WS] album socket error')
    }

    ws.onclose = () => {
      ws = null
    }

    return () => {
      dead = true
      ws?.close()
    }
  }, [albumId])
}
