export type WsMessage = {
  type: 'single' | 'group' | 'heartbeat' | 'ack' | 'read' | 'revoke' | 'call' | 'presence'
  from_id: number
  to_id: number
  payload: string
}

// 回调函数：当WebSocket收到消息时，就会触发所有的监听者
type MessageHandler = (message: WsMessage) => void

// WebSocket 客户端封装：负责管理 WebSocket连接、消息发送、消息接收、心跳检测、重连机制，以及消息编码解码
export class WsClient {
  // WebSocket 地址
  private url: string
  // 浏览器 WebSocket 对象
  private socket: WebSocket | null = null
  // 消息监听者列表
  private handlers: MessageHandler[] = []
  // 心跳定时器
  private heartbeatTimer: number | null = null
  // 重连定时器
  private reconnectTimer: number | null = null
  // 记录重连次数
  private reconnectAttempts = 0

  constructor(url: string) {
    this.url = url
  }

  // 用于连接 WebSocket
  connect() {
    // 防止重复连接：如果 WebSocket 已经打开，直接返回
    if (this.socket && this.socket.readyState === WebSocket.OPEN) return
    this.socket = new WebSocket(this.url)

    this.socket.onopen = () => {
      // 连接成功后重置重连次数
      this.reconnectAttempts = 0
      // 启动心跳
      this.startHeartbeat()
    }

    // 收到消息
    this.socket.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as WsMessage
        // 所有监听者都会收到消息
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

  // 注册消息监听者
  onMessage(handler: MessageHandler) {
    this.handlers.push(handler)
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    // 每25秒发送一次心跳，确保连接不会被断开
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

  // 计划重连
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
  extra?: Record<string, unknown> // 扩展字段
  tempId?: string // 临时消息ID
}

// 编码：将JSON->Base64
export function encodePayload(payload: ChatPayload) {
  const raw = JSON.stringify(payload)
  return btoa(
    Array.from(new TextEncoder().encode(raw))
      .map((byte) => String.fromCharCode(byte))
      .join(''),
  )
}

// 解码：将Base64->JSON
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
