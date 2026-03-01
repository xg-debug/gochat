export type Conversation = {
  id: string
  name: string
  avatar: string
  lastMessage: string
  unread: number
}

export type Contact = {
  id: string
  name: string
  avatar: string
}

export type Message = {
  id: string
  fromId: string
  content: string
  contentType: 'text' | 'file' | 'image' | 'video'
  time: number
  status: 'sent' | 'delivered' | 'read'
}
