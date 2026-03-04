<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'
import type { Conversation } from '../types/chat'

const chat = useChatStore()
const emit = defineEmits<{
  (event: 'more'): void
  (event: 'audio-call'): void
  (event: 'video-call'): void
  (event: 'group-manage'): void
}>()

const activeTitle = computed(() => {
  return (
    chat.conversations.find(
      (item: Conversation) => item.id === chat.activeConversationId,
    )?.name || '未选择会话'
  )
})

const activeSubtitle = computed(() => {
  const conv = chat.activeConversation
  if (!conv) return ''
  if (conv.id.startsWith('g_')) {
    return '群聊'
  }
  return conv.online ? '在线' : '离线'
})

const isGroupChat = computed(() => {
  return chat.activeConversation?.id?.startsWith('g_')
})
</script>

<template>
  <header class="chat-header">
    <div class="chat-title">
      <div class="chat-name">{{ activeTitle }}</div>
      <div class="chat-subtitle">{{ activeSubtitle }}</div>
    </div>
    <div class="chat-actions">
      <button v-if="isGroupChat" class="icon-btn" title="群成员" @click="emit('group-manage')">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3Zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3Zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5Zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5Z"
          />
        </svg>
      </button>
      <button class="icon-btn" title="语音" @click="emit('audio-call')">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M12 3a3 3 0 0 1 3 3v6a3 3 0 1 1-6 0V6a3 3 0 0 1 3-3Zm-7 9a1 1 0 0 1 2 0 5 5 0 0 0 10 0 1 1 0 1 1 2 0 7 7 0 0 1-6 6.9V21a1 1 0 1 1-2 0v-2.1A7 7 0 0 1 5 12Z"
          />
        </svg>
      </button>
      <button class="icon-btn" title="视频" @click="emit('video-call')">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M4 6h10a2 2 0 0 1 2 2v1.4l4-2.4v10l-4-2.4V16a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2Z"
          />
        </svg>
      </button>
      <button class="icon-btn" title="更多" @click="emit('more')">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M5 12a2 2 0 1 0 0.001-3.999A2 2 0 0 0 5 12Zm7 0a2 2 0 1 0 0.001-3.999A2 2 0 0 0 12 12Zm7 0a2 2 0 1 0 0.001-3.999A2 2 0 0 0 19 12Z"
          />
        </svg>
      </button>
    </div>
  </header>
</template>

<style scoped>
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #e5e7eb;
  background: #fafafa;
}

.chat-name {
  font-size: 16px;
  color: #111827;
  font-weight: 600;
}

.chat-subtitle {
  font-size: 12px;
  color: #9ca3af;
}

.chat-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 1px solid #e5e7eb;
  background: #fff;
  color: #64748b;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.icon-btn:hover {
  border-color: #cbd5f5;
  color: #2563eb;
}

.icon-btn .icon {
  width: 16px;
  height: 16px;
}
</style>
