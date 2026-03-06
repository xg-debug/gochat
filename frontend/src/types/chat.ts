export type Conversation = {
  id: string
  name: string
  avatar: string
  lastMessage: string
  unread: number
  online?: boolean
}

export type Contact = {
  id: string
  name: string
  avatar: string
  online?: boolean
}

export type Message = {
  id: string
  fromId: string
  fromAvatar?: string
  content: string
  contentType: 'text' | 'file' | 'image' | 'video' | 'audio'
  time: number
  status: 'sent' | 'delivered' | 'read' | 'revoked'
}
