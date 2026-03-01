import type { UserProfile } from '../stores/auth'
import type { Contact, Conversation, Message } from '../types/chat'

const baseUrl = ''

// 通用请求封装，后端接口完善后可直接替换实现
async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token') || ''
  const headers = new Headers(options.headers)
  headers.set('Content-Type', 'application/json')
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(`${baseUrl}${path}`, { ...options, headers })
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const data = (await response.json()) as { error?: string }
      if (data?.error) {
        message = data.error
      }
    } catch {
      message = `HTTP ${response.status}`
    }
    throw new Error(message)
  }
  return (await response.json()) as T
}

export async function loginRequest(username: string, password: string) {
  return await request<{ token: string; user: UserProfile }>('/api/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export async function registerRequest(
  username: string,
  password: string,
  nickname: string,
) {
  return await request<{ token: string; user: UserProfile }>('/api/register', {
    method: 'POST',
    body: JSON.stringify({ username, password, nickname }),
  })
}

export async function profileRequest() {
  return await request<UserProfile>('/api/profile')
}

export async function fetchConversations(): Promise<Conversation[]> {
  return await request<
    {
      id: string
      name: string
      avatar: string
      lastMessage: string
      unread: number
    }[]
  >('/api/conversations')
}

export async function fetchContacts(): Promise<Contact[]> {
  return await request<{ id: string; name: string; avatar: string }[]>(
    '/api/contacts',
  )
}

export async function fetchMessages(conversationId: string): Promise<Message[]> {
  return await request<
    {
      id: string
      fromId: string
      content: string
      contentType: 'text' | 'file' | 'image' | 'video'
      time: number
      status: 'sent' | 'delivered' | 'read'
    }[]
  >(`/api/messages?conversationId=${conversationId}`)
}
