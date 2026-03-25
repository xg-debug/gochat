import type { UserProfile } from '../types/user'
import type { Contact, Conversation, Message } from '../types/chat'

const baseUrl = ''

function normalizeResourceUrl(url?: string) {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://') || url.startsWith('blob:') || url.startsWith('data:')) {
    return url
  }
  if (url.startsWith('/')) {
    return url
  }
  return `/${url}`
}

// 通用请求封装
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

async function uploadResource(path: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)

  const token = localStorage.getItem('token') || ''
  const headers = new Headers()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const response = await fetch(path, {
    method: 'POST',
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw new Error('上传失败')
  }

  const data = (await response.json()) as { url: string }
  return normalizeResourceUrl(data.url)
}

export async function loginRequest(username: string, password: string) {
  const data = await request<{ token: string; user: UserProfile }>('/api/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  data.user.avatar = normalizeResourceUrl(data.user.avatar)
  return data
}

export async function registerRequest(
  username: string,
  password: string,
  nickname: string,
) {
  const data = await request<{ token: string; user: UserProfile }>('/api/register', {
    method: 'POST',
    body: JSON.stringify({ username, password, nickname }),
  })
  data.user.avatar = normalizeResourceUrl(data.user.avatar)
  return data
}

export async function profileRequest() {
  const profile = await request<UserProfile>('/api/profile')
  profile.avatar = normalizeResourceUrl(profile.avatar)
  return profile
}

export async function logoutRequest() {
  return await request<{ message: string }>('/api/logout', {
    method: 'POST',
  })
}

export async function fetchConversations(): Promise<Conversation[]> {
  const list = await request<
    {
      id: string
      name: string
      avatar: string
      lastMessage: string
      unread: number
      online?: boolean
    }[]
  >('/api/conversations')
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
}

export async function searchConversations(keyword: string): Promise<Conversation[]> {
  const list = await request<
    {
      id: string
      name: string
      avatar: string
      lastMessage: string
      unread: number
      online?: boolean
    }[]
  >(`/api/conversations/search?keyword=${encodeURIComponent(keyword)}`)
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
}

export async function fetchContacts(): Promise<Contact[]> {
  const list = await request<{ id: string; name: string; avatar: string; online?: boolean }[]>(
    '/api/contacts',
  )
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
}

export async function fetchMessages(conversationId: string): Promise<Message[]> {
  const list = await request<
    {
      id: string
      fromId: string
      fromAvatar?: string
      content: string
      contentType: 'text' | 'file' | 'image' | 'video' | 'audio'
      time: number
      status: 'sent' | 'delivered' | 'read' | 'revoked'
    }[]
  >(`/api/messages?conversationId=${conversationId}`)
  return list.map((item) => {
    const normalizedAvatar = normalizeResourceUrl(item.fromAvatar)
    if (item.contentType === 'image' || item.contentType === 'file' || item.contentType === 'video') {
      return { ...item, fromAvatar: normalizedAvatar, content: normalizeResourceUrl(item.content) }
    }
    if (item.contentType === 'audio') {
      try {
        const parsed = JSON.parse(item.content) as { url?: string; duration?: number }
        if (parsed.url) {
          parsed.url = normalizeResourceUrl(parsed.url)
          return { ...item, fromAvatar: normalizedAvatar, content: JSON.stringify(parsed) }
        }
      } catch {
        return { ...item, fromAvatar: normalizedAvatar, content: normalizeResourceUrl(item.content) }
      }
    }
    return { ...item, fromAvatar: normalizedAvatar }
  })
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
  const list = await request<SearchUserResult[]>(`/api/user/search?keyword=${encodeURIComponent(keyword)}`)
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
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
  const list = await request<FriendRequestItem[] | null>('/api/friend/requests')
  const safeList = Array.isArray(list) ? list : []
  return safeList.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
}

export type GroupItem = {
  id: number
  name: string
  avatar?: string
  notice?: string
  role?: number
}

export async function searchGroup(keyword: string) {
  const list = await request<GroupItem[]>(`/api/group/search?keyword=${encodeURIComponent(keyword)}`)
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
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
  const list = await request<GroupItem[]>('/api/groups')
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
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

export async function updateProfile(data: {
  nickname?: string
  avatar?: string
  signature?: string
  gender?: number
  phone?: string
  location?: string
  birthday?: string
}) {
  const profile = await request<UserProfile>('/api/profile', {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  profile.avatar = normalizeResourceUrl(profile.avatar)
  return profile
}

export async function uploadAvatar(file: File) {
  return uploadResource('/api/upload/avatar', file)
}

export async function uploadChatImage(file: File) {
  return uploadResource('/api/upload/chat/image', file)
}

export async function uploadChatFile(file: File) {
  return uploadResource('/api/upload/chat/file', file)
}

export async function uploadChatAudio(file: File) {
  return uploadResource('/api/upload/chat/audio', file)
}

export async function uploadGroupAvatar(file: File) {
  return uploadResource('/api/upload/group/avatar', file)
}

export async function getGroupProfile(groupId: number) {
  const group = await request<GroupItem>(`/api/group/profile?groupId=${groupId}`)
  group.avatar = normalizeResourceUrl(group.avatar)
  return group
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

export async function listInviteableFriends(groupId: number) {
  const list = await request<GroupMember[]>(`/api/group/inviteable?groupId=${groupId}`)
  return list.map((item) => ({
    ...item,
    avatar: normalizeResourceUrl(item.avatar),
  }))
}

export async function inviteGroupMember(groupId: number, userId: number) {
  return await request<{ message: string }>('/api/group/invite', {
    method: 'POST',
    body: JSON.stringify({ groupId, userId }),
  })
}
