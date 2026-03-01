<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'
import type { Conversation } from '../types/chat'

const chat = useChatStore()

const activeTitle = computed(() => {
  return (
    chat.conversations.find(
      (item: Conversation) => item.id === chat.activeConversationId,
    )?.name || '未选择会话'
  )
})
</script>

<template>
  <header class="chat-header">
    <div class="chat-title">
      <div class="chat-name">{{ activeTitle }}</div>
      <div class="chat-subtitle">在线</div>
    </div>
    <div class="chat-actions">
      <el-button size="small">语音</el-button>
      <el-button size="small">视频</el-button>
      <el-button size="small">更多</el-button>
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
  background: #f7f7f7;
}

.chat-name {
  font-size: 16px;
  color: #111827;
}

.chat-subtitle {
  font-size: 12px;
  color: #9ca3af;
}

.chat-actions {
  display: flex;
  gap: 8px;
}
</style>
