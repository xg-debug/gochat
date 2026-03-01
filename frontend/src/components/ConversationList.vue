<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const conversations = computed(() => chat.conversations)

function handleSelect(conversationId: string) {
  chat.selectConversation(conversationId)
}

function formatTime() {
  const date = new Date()
  return `${date.getHours().toString().padStart(2, '0')}:${date
    .getMinutes()
    .toString()
    .padStart(2, '0')}`
}
</script>

<template>
  <div class="conversation-list">
    <el-scrollbar height="100%">
      <div
        v-for="item in conversations"
        :key="item.id"
        class="conversation-item"
        :class="{ active: item.id === chat.activeConversationId }"
        @click="handleSelect(item.id)"
      >
        <div class="conversation-avatar">{{ item.name.slice(0, 1) }}</div>
        <div class="conversation-body">
          <div class="conversation-row">
            <div class="conversation-name">{{ item.name }}</div>
            <div class="conversation-time">{{ formatTime() }}</div>
          </div>
          <div class="conversation-row">
            <div class="conversation-last">{{ item.lastMessage }}</div>
            <div v-if="item.unread" class="conversation-unread">{{ item.unread }}</div>
          </div>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<style scoped>
.conversation-list {
  flex: 1;
}

.conversation-item {
  display: flex;
  gap: 12px;
  padding: 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.2s;
}

.conversation-item.active,
.conversation-item:hover {
  background: #e9eef4;
}

.conversation-avatar {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: #111827;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.conversation-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.conversation-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.conversation-name {
  font-size: 14px;
  color: #111827;
}

.conversation-time {
  font-size: 12px;
  color: #94a3b8;
}

.conversation-last {
  font-size: 12px;
  color: #6b7280;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conversation-unread {
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  background: #ef4444;
  color: #fff;
  font-size: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}
</style>
