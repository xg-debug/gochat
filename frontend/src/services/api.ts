import type { UserProfile } from '../types/user'
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
    if (response.status === 401) {
      // 触发未授权事件，由上层处理（如跳转登录）
      window.dispatchEvent(new CustomEvent('unauthorized'))
    }
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

export async function logoutRequest() {
  return await request<{ message: string }>('/api/logout', {
    method: 'POST',
  })
}

export async function fetchConversations(): Promise<Conversation[]> {
  return await request<
    {
      id: string
      name: string
      avatar: string
      lastMessage: string
      unread: number
      online?: boolean
    }[]
  >('/api/conversations')
}

export async function fetchContacts(): Promise<Contact[]> {
  return await request<{ id: string; name: string; avatar: string; online?: boolean }[]>(
    '/api/contacts',
  )
}

export async function fetchMessages(conversationId: string): Promise<Message[]> {
  return await request<
    {
      id: string
      fromId: string
      content: string
      contentType: 'text' | 'file' | 'image' | 'video' | 'audio'
      time: number
      status: 'sent' | 'delivered' | 'read' | 'revoked'
    }[]
  >(`/api/messages?conversationId=${conversationId}`)
}

export type SearchUserResult = {
  id: number
  username: string
  nickname: string
  avatar: string
  isFriend: boolean
  pending: boolean
  pendingFromMe: boolean
}

export async function searchUser(keyword: string) {
  return await request<SearchUserResult[]>(`/api/user/search?keyword=${encodeURIComponent(keyword)}`)
}

export async function sendFriendRequest(toUserId: number) {
  return await request<{ message: string }>('/api/friend/request', {
    method: 'POST',
    body: JSON.stringify({ toUserId }),
  })
}

export type FriendRequestItem = {
  id: number
  fromUserId: number
  username: string
  nickname: string
  avatar: string
  time: number
}

export async function listFriendRequests() {
  return await request<FriendRequestItem[]>('/api/friend/requests')
}

export type GroupItem = {
  id: number
  name: string
  avatar?: string
  notice?: string
  role?: number
}

export async function searchGroup(keyword: string) {
  return await request<GroupItem[]>(`/api/group/search?keyword=${encodeURIComponent(keyword)}`)
}

export async function createGroup(name: string) {
  return await request<GroupItem>('/api/group/create', {
    method: 'POST',
    body: JSON.stringify({ name }),
  })
}

export async function joinGroup(groupId: number) {
  return await request<{ message: string }>('/api/group/join', {
    method: 'POST',
    body: JSON.stringify({ groupId }),
  })
}

export async function listGroups() {
  return await request<GroupItem[]>('/api/groups')
}

export async function handleFriendRequest(requestId: number, action: 'accept' | 'reject') {
  return await request<{ message: string }>('/api/friend/handle', {
    method: 'POST',
    body: JSON.stringify({ requestId, action }),
  })
}

export async function deleteFriend(friendId: number) {
  return await request<{ message: string }>('/api/friend/delete', {
    method: 'POST',
    body: JSON.stringify({ friendId }),
  })
}

export async function blockFriend(friendId: number) {
  return await request<{ message: string }>('/api/friend/block', {
    method: 'POST',
    body: JSON.stringify({ friendId }),
  })
}

export async function unblockFriend(friendId: number) {
  return await request<{ message: string }>('/api/friend/unblock', {
    method: 'POST',
    body: JSON.stringify({ friendId }),
  })
}

export async function updateProfile(data: { nickname?: string; avatar?: string; signature?: string; gender?: number }) {
  return await request<UserProfile>('/api/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function uploadAvatar(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  
  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  
  // Note: Content-Type is not set manually to let browser set boundary
  const response = await fetch('/api/upload/avatar', {
    method: 'POST',
    headers,
    body: formData,
  })
  
  if (!response.ok) {
    throw new Error('上传失败')
  }
  
  const data = await response.json()
  return data.url as string
}

export async function uploadChatImage(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch('/api/upload/chat/image', {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('上传失败')
  }

  const data = await response.json()
  return data.url as string
}

export async function uploadChatFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch('/api/upload/chat/file', {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('上传失败')
  }
  const data = await response.json()
  return data.url as string
}

export async function uploadChatAudio(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch('/api/upload/chat/audio', {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('上传失败')
  }
  const data = await response.json()
  return data.url as string
}

export async function uploadGroupAvatar(file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch('/api/upload/group/avatar', {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('上传失败')
  }
  const data = await response.json()
  return data.url as string
}

export async function getGroupProfile(groupId: number) {
  return await request<GroupItem>(`/api/group/profile?groupId=${groupId}`)
}

export async function updateGroupProfile(data: { groupId: number; name?: string; avatar?: string; notice?: string }) {
  return await request<{ message: string }>('/api/group/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export type GroupMember = {
  userId: number
  nickname: string
  username: string
  avatar: string
  role: number
}

export async function listGroupMembers(groupId: number) {
  return await request<GroupMember[]>(`/api/group/members?groupId=${groupId}`)
}

export async function kickGroupMember(groupId: number, userId: number) {
  return await request<{ message: string }>('/api/group/kick', {
    method: 'POST',
    body: JSON.stringify({ groupId, userId }),
  })
}

export async function setGroupAdmin(groupId: number, userId: number, action: 'set' | 'unset') {
  return await request<{ message: string }>('/api/group/admin', {
    method: 'POST',
    body: JSON.stringify({ groupId, userId, action }),
  })
}
