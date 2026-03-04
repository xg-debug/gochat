<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const contacts = computed(() => chat.contacts)
const emit = defineEmits<{
  (event: 'open-group'): void
  (event: 'delete-friend', id: string): void
  (event: 'block-friend', id: string): void
  (event: 'unblock-friend', id: string): void
}>()
</script>

<template>
  <div class="contact-list">
    <div class="contact-section">
      <div class="section-title">通讯录</div>
      
      <!-- 群聊入口 -->
      <div class="contact-item group-chat-entry" @click="emit('open-group')">
        <div class="contact-avatar group-avatar">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
            <path d="M16 11c1.66 0 2.99-1.34 2.99-3S17.66 5 16 5c-1.66 0-3 1.34-3 3s1.34 3 3 3zm-8 0c1.66 0 2.99-1.34 2.99-3S9.66 5 8 5C6.34 5 5 6.34 5 8s1.34 3 3 3zm0 2c-2.33 0-7 1.17-7 3.5V19h14v-2.5c0-2.33-4.67-3.5-7-3.5zm8 0c-.29 0-.62.02-.97.05 1.16.84 1.97 1.97 1.97 3.45V19h6v-2.5c0-2.33-4.67-3.5-7-3.5z"/>
          </svg>
        </div>
        <div class="contact-name">群聊</div>
      </div>

      <!-- 联系人列表 -->
      <div v-for="item in contacts" :key="item.id" class="contact-item">
        <div class="contact-avatar">
          <img v-if="item.avatar" :src="item.avatar" />
          <span v-else>{{ item.name.slice(0, 1) }}</span>
          <span v-if="item.online" class="status-dot"></span>
        </div>
        <div class="contact-name" @click="chat.startConversation(item.id)">{{ item.name }}</div>
        <div class="contact-actions">
          <button class="contact-btn" @click="emit('delete-friend', item.id)">删除</button>
          <button class="contact-btn" @click="emit('block-friend', item.id)">拉黑</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contact-list {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.contact-section {
  display: flex;
  flex-direction: column;
}

.section-title {
  font-size: 12px;
  color: #999;
  padding: 10px 12px;
  background: #f7f7f7;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f0f0f0;
}

.contact-item:hover {
  background-color: #e9f5ee;
}

.contact-avatar {
  width: 36px;
  height: 36px;
  border-radius: 6px;
  background: #d9dde3; /* Fallback color matching generic avatar */
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  overflow: hidden;
  position: relative;
}

.contact-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.status-dot {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #22c55e;
  border-radius: 50%;
  right: -2px;
  bottom: -2px;
  border: 2px solid #f7f7f7;
}

.group-avatar {
  background: #07c160;
}

.contact-name {
  font-size: 14px;
  color: #333;
  flex: 1;
}

.contact-actions {
  display: flex;
  gap: 6px;
}

.contact-btn {
  border: none;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 12px;
  padding: 4px 6px;
  border-radius: 4px;
  cursor: pointer;
}
</style>
