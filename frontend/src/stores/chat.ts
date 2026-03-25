import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Contact, Conversation, Message } from '../types/chat'
import { fetchContacts, fetchConversations, fetchMessages, searchConversations as searchConversationsRequest } from '../services/api'
import { decodePayload, encodePayload, WsClient, type WsMessage } from '../services/ws'
import { messagePreview, parseConversationId } from '../utils/chat'

const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`

// 聊天数据管理 + WebSocket连接 + 消息分发 + 会话管理
export const useChatStore = defineStore('chat', () => {
  // 会话列表
  const conversations = ref<Conversation[]>([])
  // 联系人列表
  const contacts = ref<Contact[]>([])
  // 消息映射表（会话ID -> 消息列表）
  const messageMap = ref<Record<string, Message[]>>({})
  // 当前激活会话ID（当前聊天）
  const activeConversationId = ref<string>('')
  // WebSocket客户端实例
  const wsClient = ref<WsClient | null>(null)
  // 当前用户ID
  const currentUserId = ref(0)

  const activeMessages = computed(() => {
    return messageMap.value[activeConversationId.value] || []
  })

  const activeConversation = computed(() => {
    return conversations.value.find((item) => item.id === activeConversationId.value)
  })

  // 初始化聊天数据：加载会话->加载联系人->设置默认激活会话
  async function bootstrap() {
    conversations.value = await fetchConversations()
    contacts.value = await fetchContacts()
    if (!activeConversationId.value && conversations.value.length > 0) {
      const firstConversation = conversations.value[0]
      if (firstConversation) {
        activeConversationId.value = firstConversation.id
      }
    }
    if (activeConversationId.value) {
      messageMap.value[activeConversationId.value] = await fetchMessages(
        activeConversationId.value,
      )
    }
  }

  // 切换聊天窗口（切换会话）：更新激活会话ID->加载消息->更新未读计数
  async function selectConversation(conversationId: string) {
    activeConversationId.value = conversationId
    if (!messageMap.value[conversationId]) {
      messageMap.value[conversationId] = await fetchMessages(conversationId)
    }
    const conversation = conversations.value.find(
      (item: Conversation) => item.id === conversationId,
    )
    if (conversation) {
      conversation.unread = 0
    }
    if (conversationId.startsWith('u_')) {
      const peerId = Number(conversationId.replace('u_', ''))
      if (peerId > 0) {
        sendReadReceipt(peerId)
      }
      const list = messageMap.value[conversationId] || []
      list.forEach((item) => {
        if (item.status !== 'revoked') {
          item.status = 'delivered'
        }
      })
    }
  }

  // 从联系人发起聊天-启动新会话（单聊）：点击联系人->生成会话ID->判断会话是否存在
  function startConversation(contactId: string) {
    // 规范化会话ID：单聊用户ID统一加 u_ 前缀
    const targetConvId = contactId.startsWith('u_') ? contactId : `u_${contactId}`
    
    const existing = conversations.value.find(c => c.id === targetConvId)
    if (existing) {
      selectConversation(targetConvId)
    } else {
      // 查找联系人信息（使用原始ID）
      const rawId = contactId.startsWith('u_') ? contactId.slice(2) : contactId
      const contact = contacts.value.find(c => c.id === rawId)
      
      if (contact) {
        const newConv: Conversation = {
          id: targetConvId,
          name: contact.name,
          avatar: contact.avatar,
          lastMessage: '',
          unread: 0,
        }
        conversations.value.unshift(newConv)
        selectConversation(targetConvId)
      }
    }
  }

  // 建立WebSocket连接：初始化WS客户端->设置消息处理->连接
  function connect(userId: number, token: string) {
    if (wsClient.value || !token) return
    currentUserId.value = userId
    wsClient.value = new WsClient(`${wsUrl}?token=${encodeURIComponent(token)}`)
    wsClient.value.onMessage(handleIncomingMessage)
    wsClient.value.connect()
  }

  // 断开WebSocket连接：关闭WS客户端->清除实例
  function disconnect() {
    wsClient.value?.disconnect()
    wsClient.value = null
  }

  // 退出登录时调用：重置聊天数据->断开WS连接
  function reset() {
    conversations.value = []
    contacts.value = []
    messageMap.value = {}
    activeConversationId.value = ''
    currentUserId.value = 0
    disconnect()
  }

  // 发送消息：创建本地消息->添加到映射messageMap->更新会话的lastMessage->发送WS消息
  function sendMessage(
    toId: number,
    content: string,
    contentType: Message['contentType'],
  ) {
    const conversationId = activeConversationId.value
    if (!conversationId) return
    const parsed = parseConversationId(conversationId)
    const isGroup = parsed.kind === 'group'
    // tempId 是客户端生成的临时消息ID，用来在服务器返回真实 messageId 之前，标识本地消息。
    const tempId = `local_${Date.now()}`
    const newMessage: Message = {
      id: tempId,
      fromId: `u_${currentUserId.value || 1}`,
      content,
      contentType,
      time: Date.now(),
      status: 'sent',
    }
    if (!messageMap.value[conversationId]) {
      messageMap.value[conversationId] = []
    }
    // 添加到会话的消息列表（UI界面），此时聊天界面立即显示发送的消息
    messageMap.value[conversationId].push(newMessage)
    const conversation = conversations.value.find(
      (item: Conversation) => item.id === conversationId,
    )
    if (conversation) {
      conversation.lastMessage = messagePreview(content, contentType)
    }
    wsClient.value?.send({
      type: isGroup ? 'group' : 'single',
      from_id: currentUserId.value,
      to_id: toId,
      payload: encodePayload({ content, contentType, tempId }),
    })
  }


  // 处理WebSocket收到的所有消息 ｜ 接收消息：根据类型分类处理->更新本地消息状态->触发事件
  function handleIncomingMessage(message: WsMessage) {
    if (message.type === 'ack') { // 确认消息送达
      try {
        const payload = JSON.parse(atob(message.payload)) as {
          tempId?: string
          messageId?: number
        }
        if (!payload?.tempId || !payload?.messageId) return
        Object.keys(messageMap.value).forEach((convId) => {
          const list = messageMap.value[convId]
          const msg = list?.find((item) => item.id === payload.tempId)
          if (msg) {
            msg.id = `m_${payload.messageId}`
            msg.status = 'delivered'
          }
        })
      } catch {
        return
      }
      return
    }
    if (message.type === 'call') { // 语音/视频信令
      let payload: Record<string, unknown> = {}
      const decoded = decodePayload(message.payload)
      if (decoded?.extra && typeof decoded.extra === 'object') {
        payload = decoded.extra as Record<string, unknown>
      } else {
        try {
          payload = JSON.parse(message.payload) as Record<string, unknown>
        } catch {
          return
        }
      }
      try {
        window.dispatchEvent(
          new CustomEvent('call-signal', {
            detail: {
              fromId: message.from_id,
              toId: message.to_id,
              payload,
            },
          }),
        )
      } catch {
        return
      }
      return
    }
    if (message.type === 'presence') { // 好友在线状态更新
      let online = false
      const decoded = decodePayload(message.payload)
      if (decoded?.extra && typeof decoded.extra.online !== 'undefined') {
        online = Boolean(decoded.extra.online)
      } else {
        try {
          const raw = JSON.parse(message.payload) as { online?: boolean }
          online = Boolean(raw.online)
        } catch {
          online = true
        }
      }
      const userID = `u_${message.from_id}`
      contacts.value.forEach((item) => {
        if (item.id === userID) {
          item.online = online
        }
      })
      conversations.value.forEach((item) => {
        if (item.id === userID) {
          item.online = online
        }
      })
      return
    }
    if (message.type === 'read') {
      const conversationId = `u_${message.from_id}`
      const list = messageMap.value[conversationId] || []
      list.forEach((item) => {
        if (item.fromId === `u_${currentUserId.value}` && item.status !== 'revoked') {
          item.status = 'read'
        }
      })
      return
    }
    if (message.type === 'revoke') { // 暂未实现
      return
    }
    const decoded = decodePayload(message.payload)
    if (!decoded) return
    const senderName = decoded.extra?.fromName as string | undefined
    const conversationId =
      message.type === 'group' ? `g_${message.to_id}` : message.from_id ? `u_${message.from_id}` : 'system'
    const incoming: Message = {
      id: `ws_${Date.now()}`,
      fromId: `u_${message.from_id}`,
      fromAvatar:
        typeof decoded.extra?.fromAvatar === 'string'
          ? decoded.extra.fromAvatar
          : '',
      content: decoded.content,
      contentType: decoded.contentType,
      time: Date.now(),
      status: 'delivered',
    }
    if (!messageMap.value[conversationId]) {
      messageMap.value[conversationId] = []
    }
    messageMap.value[conversationId].push(incoming)
    let conversation = conversations.value.find(
      (item: Conversation) => item.id === conversationId,
    )
    if (!conversation) {
      const contact = contacts.value.find((item) => item.id === conversationId)
      const name =
        message.type === 'group'
          ? `群聊${message.to_id}`
          : contact?.name || senderName || (message.from_id ? `用户${message.from_id}` : '系统通知')
      const avatar = contact?.avatar || ''
      conversation = {
        id: conversationId,
        name,
        avatar,
        lastMessage: incoming.content,
        unread: 0,
        online: message.type === 'group' ? false : true,
      }
      conversations.value.unshift(conversation)
    }
    conversation.lastMessage = messagePreview(incoming.content, incoming.contentType)
    if (message.type === 'single') {
      conversation.online = true
      contacts.value.forEach((item) => {
        if (item.id === conversationId) {
          item.online = true
        }
      })
    }
    if (conversationId !== activeConversationId.value) {
      conversation.unread += 1
      window.dispatchEvent(
        new CustomEvent('incoming-message', {
          detail: {
            conversationId,
            name: message.type === 'group' ? conversation.name : senderName || conversation.name,
            content: incoming.content,
            contentType: incoming.contentType,
          },
        }),
      )
    } else if (message.type === 'single') {
      const peerId = Number(conversationId.replace('u_', ''))
      if (peerId > 0) {
        sendReadReceipt(peerId)
      }
    }
  }

  // 发送语音/视频信令：创建WS消息->发送
  function sendCallSignal(toId: number, payload: Record<string, unknown>) {
    wsClient.value?.send({
      type: 'call',
      from_id: currentUserId.value || 1,
      to_id: toId,
      payload: encodePayload({
        content: '',
        contentType: 'text',
        extra: payload,
      }),
    })
  }

  function sendReadReceipt(toId: number) {
    if (!toId) return
    wsClient.value?.send({
      type: 'read',
      from_id: currentUserId.value || 1,
      to_id: toId,
      payload: encodePayload({
        content: '',
        contentType: 'text',
      }),
    })
  }

  async function searchConversations(keyword: string) {
    const trimmed = keyword.trim()
    if (!trimmed) {
      conversations.value = await fetchConversations()
      return
    }
    try {
      const list = await searchConversationsRequest(trimmed)
      if (list.length > 0) {
        conversations.value = list
        return
      }
    } catch {
      // fallback to local filter below
    }
    const fullList = await fetchConversations()
    const lower = trimmed.toLowerCase()
    conversations.value = fullList.filter((item) => {
      return (
        item.name.toLowerCase().includes(lower) ||
        item.lastMessage.toLowerCase().includes(lower)
      )
    })
  }

  return {
    conversations,
    contacts,
    activeConversationId,
    activeMessages,
    activeConversation,
    bootstrap,
    selectConversation,
    startConversation,
    connect,
    disconnect,
    sendMessage,
    sendCallSignal,
    sendReadReceipt,
    searchConversations,
    reset,
  }
})
