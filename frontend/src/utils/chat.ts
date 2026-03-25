import type { Message } from '../types/chat'

export function parseConversationId(conversationId: string): { kind: 'single' | 'group' | 'unknown'; id: number } {
  if (conversationId.startsWith('u_')) {
    return { kind: 'single', id: Number(conversationId.slice(2)) || 0 }
  }
  if (conversationId.startsWith('g_')) {
    return { kind: 'group', id: Number(conversationId.slice(2)) || 0 }
  }
  return { kind: 'unknown', id: 0 }
}

export function messagePreview(content: string, contentType: Message['contentType']): string {
  switch (contentType) {
    case 'image':
      return '[图片]'
    case 'audio':
      return '[语音]'
    case 'file':
      return '[文件]'
    case 'video':
      return '[视频]'
    default:
      return content
  }
}
