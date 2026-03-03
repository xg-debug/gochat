<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const auth = useAuthStore()
const listRef = ref<HTMLDivElement | null>(null)

const messages = computed(() => chat.activeMessages)

watch(
  () => messages.value.length,
  async () => {
    await nextTick()
    listRef.value?.scrollTo({ top: listRef.value.scrollHeight })
  },
)
</script>

<template>
  <div ref="listRef" class="message-list">
    <div v-for="item in messages" :key="item.id" class="message-wrap">
      <div class="message-time">
        {{ new Date(item.time).toLocaleTimeString() }}
      </div>
      <div
        class="message-item"
        :class="{ self: item.fromId === `u_${auth.user?.id || 1}` }"
      >
        <div class="message-avatar">
          {{ item.fromId === `u_${auth.user?.id || 1}` ? '我' : '对' }}
        </div>
        <div class="message-bubble">
          <div v-if="item.contentType === 'image'" class="message-image">
            <img :src="item.content" alt="image" />
          </div>
          <div v-else class="message-content">{{ item.content }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-list {
  flex: 1;
  padding: 16px 20px;
  overflow-y: auto;
  background: #f2f3f5;
}

.message-wrap {
  margin-bottom: 16px;
}

.message-time {
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 8px;
}

.message-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.message-item.self {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: #111827;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
}

.message-bubble {
  max-width: 60%;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.06);
}

.message-item.self .message-bubble {
  background: #9fe870;
}

.message-content {
  font-size: 14px;
  color: #111827;
  white-space: pre-wrap;
}

.message-image img {
  max-width: 220px;
  border-radius: 6px;
  display: block;
}
</style>
