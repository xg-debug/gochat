<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import type { Message } from '../types/chat'

const chat = useChatStore()
const auth = useAuthStore()
const listRef = ref<HTMLDivElement | null>(null)
const playingId = ref('')
let playingAudio: HTMLAudioElement | null = null

const messages = computed(() => chat.activeMessages)

function normalizeAvatarUrl(url?: string) {
  if (!url) return ''
  if (
    url.startsWith('http://') ||
    url.startsWith('https://') ||
    url.startsWith('blob:') ||
    url.startsWith('data:')
  ) {
    return url
  }
  return url.startsWith('/') ? url : `/${url}`
}

function isSelfMessage(item: Message) {
  return item.fromId === `u_${auth.user?.id || 0}`
}

function resolveMessageAvatar(item: Message) {
  if (isSelfMessage(item)) {
    return normalizeAvatarUrl(auth.user?.avatar)
  }
  const fromAvatar = normalizeAvatarUrl(item.fromAvatar)
  if (fromAvatar) {
    return fromAvatar
  }
  const sender = chat.contacts.find((contact) => {
    return (
      contact.id === item.fromId ||
      `u_${contact.id}` === item.fromId ||
      contact.id === item.fromId.replace(/^u_/, '')
    )
  })
  if (sender?.avatar) {
    return normalizeAvatarUrl(sender.avatar)
  }
  if (chat.activeConversationId.startsWith('u_')) {
    return normalizeAvatarUrl(chat.activeConversation?.avatar)
  }
  return ''
}

function resolveMessageFallback(item: Message) {
  if (isSelfMessage(item)) {
    return auth.user?.nickname?.slice(0, 1) || '我'
  }
  return '对'
}

function getAudioMeta(content: string) {
  try {
    const parsed = JSON.parse(content) as { url?: string; duration?: number }
    if (parsed.url) {
      return {
        url: parsed.url,
        duration: parsed.duration || 0,
      }
    }
  } catch {
    // keep legacy format
  }
  return { url: content, duration: 0 }
}

function formatDuration(seconds: number) {
  if (!seconds || seconds <= 0) return ''
  return `${seconds}s`
}

function playVoice(id: string, content: string) {
  const meta = getAudioMeta(content)
  if (!meta.url) return
  if (playingAudio) {
    playingAudio.pause()
    playingAudio = null
  }
  if (playingId.value === id) {
    playingId.value = ''
    return
  }
  const audio = new Audio(meta.url)
  playingAudio = audio
  playingId.value = id
  audio.onended = () => {
    if (playingId.value === id) {
      playingId.value = ''
    }
    if (playingAudio === audio) {
      playingAudio = null
    }
  }
  void audio.play().catch(() => {
    playingId.value = ''
    if (playingAudio === audio) {
      playingAudio = null
    }
  })
}

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
        :class="{ self: isSelfMessage(item) }"
      >
        <div class="message-avatar">
          <img v-if="resolveMessageAvatar(item)" :src="resolveMessageAvatar(item)" class="avatar-img" />
          <span v-else>{{ resolveMessageFallback(item) }}</span>
        </div>
        <div class="message-bubble">
          <div v-if="item.status === 'revoked'" class="message-content">[已撤回]</div>
          <div v-else-if="item.contentType === 'image'" class="message-image">
            <img :src="item.content" alt="image" />
          </div>
          <div
            v-else-if="item.contentType === 'audio'"
            class="voice-bubble"
            :class="{ playing: playingId === item.id }"
            @click="playVoice(item.id, item.content)"
          >
            <span class="voice-icon">▶</span>
            <span class="voice-text">{{ formatDuration(getAudioMeta(item.content).duration || 1) }}</span>
          </div>
          <div v-else-if="item.contentType === 'video'" class="message-video">
            <video controls :src="item.content"></video>
          </div>
          <div v-else-if="item.contentType === 'file'" class="message-file">
            <a :href="item.content" target="_blank">下载文件</a>
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
  background: #f5f5f5;
}

.message-wrap {
  margin-bottom: 16px;
}

.message-time {
  text-align: center;
  font-size: 12px;
  color: #9aa0a6;
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
  border-radius: 6px;
  background: #cbd5e1;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  overflow: hidden;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.message-bubble {
  max-width: 60%;
  padding: 10px 12px;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(15, 23, 42, 0.08);
}

.message-item.self .message-bubble {
  background: #95ec69;
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

.voice-bubble {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 70px;
  padding: 6px 10px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.9);
  cursor: pointer;
}

.message-item.self .voice-bubble {
  background: rgba(255, 255, 255, 0.75);
}

.voice-bubble.playing .voice-icon {
  animation: pulse 1s infinite;
}

.voice-icon {
  font-size: 12px;
  color: #374151;
}

.voice-text {
  font-size: 12px;
  color: #374151;
}

@keyframes pulse {
  0% { opacity: 0.35; }
  50% { opacity: 1; }
  100% { opacity: 0.35; }
}

.message-video video {
  width: 260px;
  max-width: 100%;
  border-radius: 8px;
}

.message-file a {
  color: #2563eb;
  text-decoration: none;
  font-size: 13px;
}

</style>
