import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { Contact, Conversation, Message } from '../types/chat'
import { fetchContacts, fetchConversations, fetchMessages } from '../services/api'
import { decodePayload, encodePayload, WsClient, type WsMessage } from '../services/ws'

const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`

// 聊天数据状态与消息分发
export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const contacts = ref<Contact[]>([])
  const messageMap = ref<Record<string, Message[]>>({})
  const activeConversationId = ref<string>('')
  const wsClient = ref<WsClient | null>(null)
  const currentUserId = ref(0)

  const activeMessages = computed(() => {
    return messageMap.value[activeConversationId.value] || []
  })

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
  }

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

  function connect(userId: number, token: string) {
    if (wsClient.value || !token) return
    currentUserId.value = userId
    wsClient.value = new WsClient(`${wsUrl}?token=${encodeURIComponent(token)}`)
    wsClient.value.onMessage(handleIncomingMessage)
    wsClient.value.connect()
  }

  function disconnect() {
    wsClient.value?.disconnect()
    wsClient.value = null
  }

  function reset() {
    conversations.value = []
    contacts.value = []
    messageMap.value = {}
    activeConversationId.value = ''
    currentUserId.value = 0
    disconnect()
  }

  function sendMessage(
    toId: number,
    content: string,
    contentType: Message['contentType'],
  ) {
    const conversationId = activeConversationId.value
    if (!conversationId) return
    const newMessage: Message = {
      id: `local_${Date.now()}`,
      fromId: `u_${currentUserId.value || 1}`,
      content,
      contentType,
      time: Date.now(),
      status: 'sent',
    }
    if (!messageMap.value[conversationId]) {
      messageMap.value[conversationId] = []
    }
    messageMap.value[conversationId].push(newMessage)
    const conversation = conversations.value.find(
      (item: Conversation) => item.id === conversationId,
    )
    if (conversation) {
      conversation.lastMessage = contentType === 'image' ? '[图片]' : content
    }
    wsClient.value?.send({
      type: 'single',
      from_id: currentUserId.value || 1,
      to_id: toId,
      payload: encodePayload({ content, contentType }),
    })
  }

  function handleIncomingMessage(message: WsMessage) {
    const decoded = decodePayload(message.payload)
    if (!decoded) return
    const conversationId = message.from_id ? `u_${message.from_id}` : 'system'
    const incoming: Message = {
      id: `ws_${Date.now()}`,
      fromId: `u_${message.from_id}`,
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
      const name = contact?.name || (message.from_id ? `用户${message.from_id}` : '系统通知')
      const avatar = contact?.avatar || ''
      conversation = {
        id: conversationId,
        name,
        avatar,
        lastMessage: incoming.content,
        unread: 0,
      }
      conversations.value.unshift(conversation)
    }
    conversation.lastMessage =
      incoming.contentType === 'image' ? '[图片]' : incoming.content
    if (conversationId !== activeConversationId.value) {
      conversation.unread += 1
    }
  }

  return {
    conversations,
    contacts,
    activeConversationId,
    activeMessages,
    bootstrap,
    selectConversation,
    startConversation,
    connect,
    disconnect,
    sendMessage,
    reset,
  }
})
