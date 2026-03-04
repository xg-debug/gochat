export type WsMessage = {
  type: 'single' | 'group' | 'heartbeat' | 'ack' | 'read' | 'revoke' | 'call'
  from_id: number
  to_id: number
  payload: string
}

type MessageHandler = (message: WsMessage) => void

// WebSocket 客户端封装，支持重连与心跳
export class WsClient {
  private url: string
  private socket: WebSocket | null = null
  private handlers: MessageHandler[] = []
  private heartbeatTimer: number | null = null
  private reconnectTimer: number | null = null
  private reconnectAttempts = 0

  constructor(url: string) {
    this.url = url
  }

  connect() {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) return
    this.socket = new WebSocket(this.url)

    this.socket.onopen = () => {
      this.reconnectAttempts = 0
      this.startHeartbeat()
    }

    this.socket.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as WsMessage
        this.handlers.forEach((handler) => handler(parsed))
      } catch {
        return
      }
    }

    this.socket.onclose = () => {
      this.stopHeartbeat()
      this.scheduleReconnect()
    }

    this.socket.onerror = () => {
      this.socket?.close()
    }
  }

  disconnect() {
    this.stopHeartbeat()
    this.socket?.close()
    this.socket = null
  }

  send(message: WsMessage) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return
    }
    this.socket.send(JSON.stringify(message))
  }

  onMessage(handler: MessageHandler) {
    this.handlers.push(handler)
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = window.setInterval(() => {
      this.send({
        type: 'heartbeat',
        from_id: 0,
        to_id: 0,
        payload: encodePayload({ content: 'ping', contentType: 'text' }),
      })
    }, 25000)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      window.clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return
    const delay = Math.min(10000, 1000 * (this.reconnectAttempts + 1))
    this.reconnectAttempts += 1
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
  }
}

export type ChatPayload = {
  content: string
  contentType: 'text' | 'image' | 'file' | 'video' | 'audio'
  extra?: Record<string, unknown>
  tempId?: string
}

export function encodePayload(payload: ChatPayload) {
  const raw = JSON.stringify(payload)
  return btoa(
    Array.from(new TextEncoder().encode(raw))
      .map((byte) => String.fromCharCode(byte))
      .join(''),
  )
}

export function decodePayload(payload: string): ChatPayload | null {
  try {
    const decoded = atob(payload)
    const bytes = Uint8Array.from(decoded, (c) => c.charCodeAt(0))
    const text = new TextDecoder().decode(bytes)
    return JSON.parse(text) as ChatPayload
  } catch {
    return null
  }
}
