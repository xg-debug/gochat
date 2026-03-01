<script setup lang="ts">
import { computed } from 'vue'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const contacts = computed(() => chat.contacts)
</script>

<template>
  <div class="contact-list">
    <div class="contact-section">
      <div class="section-title">常用分类</div>
      <div class="section-item">
        <span>群聊</span>
        <span class="section-count">20</span>
      </div>
      <div class="section-item">
        <span>联系人</span>
        <span class="section-count">211</span>
      </div>
    </div>

    <div class="contact-section">
      <div class="section-title">最近联系人</div>
      <div v-for="item in contacts" :key="item.id" class="contact-item">
        <div class="contact-avatar">{{ item.name.slice(0, 1) }}</div>
        <div class="contact-name">{{ item.name }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.contact-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.contact-section {
  background: #fff;
  border-radius: 12px;
  padding: 10px;
}

.section-title {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 8px;
}

.section-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 4px;
  font-size: 13px;
  color: #111827;
}

.section-count {
  color: #9ca3af;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 2px;
}

.contact-avatar {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  background: #1f2937;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
}

.contact-name {
  flex: 1;
  font-size: 13px;
  color: #111827;
}
</style>
