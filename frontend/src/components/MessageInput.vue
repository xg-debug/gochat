<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useChatStore } from '../stores/chat'
import { uploadChatImage, uploadChatFile, uploadChatAudio } from '../services/api'

const chat = useChatStore()
const content = ref('')
const imageInput = ref<HTMLInputElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const recording = ref(false)
const showEmojiPanel = ref(false)
const recordStartAt = ref(0)
const recordSeconds = ref(0)
let recordTimer: number | null = null
let mediaRecorder: MediaRecorder | null = null
let recordChunks: Blob[] = []
const emojiList = ['😀', '😁', '😂', '🤣', '😊', '😍', '😘', '🤔', '😎', '😭', '👍', '👋', '🎉', '❤️', '🔥', '👏']

function send() {
  if (!content.value.trim()) return
  const conversationId = chat.activeConversationId
  if (!conversationId) return
  const toId = Number(conversationId.replace(/^[ug]_/, ''))
  if (!toId) return
  chat.sendMessage(toId, content.value, 'text')
  content.value = ''
}

function onImageClick() {
  showEmojiPanel.value = false
  imageInput.value?.click()
}

function onFileClick() {
  showEmojiPanel.value = false
  fileInput.value?.click()
}

function beginVoiceRecord() {
  showEmojiPanel.value = false
  if (recording.value) return
  startRecording()
}

function endVoiceRecord() {
  if (!recording.value) return
  mediaRecorder?.stop()
  recording.value = false
  if (recordTimer) {
    window.clearInterval(recordTimer)
    recordTimer = null
  }
}

function onVoiceClick() {
  if (recording.value) {
    endVoiceRecord()
  } else {
    beginVoiceRecord()
  }
}

async function onImageChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files || !input.files[0]) return
  const file = input.files[0]
  const conversationId = chat.activeConversationId
  if (!conversationId) return
  const toId = Number(conversationId.replace(/^[ug]_/, ''))
  if (!toId) return

  try {
    const url = await uploadChatImage(file)
    chat.sendMessage(toId, url, 'image')
  } catch (error) {
    const msg = error instanceof Error ? error.message : '上传失败'
    ElMessage.error(msg)
  } finally {
    input.value = ''
  }
}

async function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files || !input.files[0]) return
  const file = input.files[0]
  const conversationId = chat.activeConversationId
  if (!conversationId) return
  const toId = Number(conversationId.replace(/^[ug]_/, ''))
  if (!toId) return

  try {
    const url = await uploadChatFile(file)
    const contentType = file.type.startsWith('image/')
      ? 'image'
      : file.type.startsWith('audio/')
        ? 'audio'
        : file.type.startsWith('video/')
          ? 'video'
          : 'file'
    if (contentType === 'audio') {
      chat.sendMessage(toId, JSON.stringify({ url, duration: 0 }), 'audio')
    } else {
      chat.sendMessage(toId, url, contentType)
    }
  } catch (error) {
    const msg = error instanceof Error ? error.message : '上传失败'
    ElMessage.error(msg)
  } finally {
    input.value = ''
  }
}

async function startRecording() {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    recordStartAt.value = Date.now()
    recordSeconds.value = 0
    recordChunks = []
    const preferredType = MediaRecorder.isTypeSupported('audio/mp4')
      ? 'audio/mp4'
      : MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : ''
    mediaRecorder = preferredType
      ? new MediaRecorder(stream, { mimeType: preferredType })
      : new MediaRecorder(stream)
    mediaRecorder.ondataavailable = (e) => {
      if (e.data.size > 0) recordChunks.push(e.data)
    }
    mediaRecorder.onstop = async () => {
      stream.getTracks().forEach((t) => t.stop())
      const mime = mediaRecorder?.mimeType || 'audio/webm'
      const blob = new Blob(recordChunks, { type: mime })
      if (blob.size < 1024) {
        ElMessage.warning('录音过短')
        return
      }
      const ext = mime.includes('mp4') ? 'm4a' : 'webm'
      const file = new File([blob], `audio_${Date.now()}.${ext}`, { type: mime })
      const duration = Math.max(1, Math.round((Date.now() - recordStartAt.value) / 1000))
      const conversationId = chat.activeConversationId
      if (!conversationId) return
      const toId = Number(conversationId.replace(/^[ug]_/, ''))
      if (!toId) return
      try {
        const url = await uploadChatAudio(file)
        chat.sendMessage(toId, JSON.stringify({ url, duration }), 'audio')
      } catch (error) {
        const msg = error instanceof Error ? error.message : '上传失败'
        ElMessage.error(msg)
      }
    }
    mediaRecorder.start(200)
    recording.value = true
    recordTimer = window.setInterval(() => {
      recordSeconds.value += 1
    }, 1000)
  } catch {
    ElMessage.error('无法获取麦克风')
  }
}

function onEmojiClick() {
  showEmojiPanel.value = !showEmojiPanel.value
}

function appendEmoji(emoji: string) {
  content.value += emoji
}

async function onPaste(event: ClipboardEvent) {
  const items = event.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (!file) continue
      const conversationId = chat.activeConversationId
      if (!conversationId) return
      const toId = Number(conversationId.replace(/^[ug]_/, ''))
      if (!toId) return
      try {
        const url = await uploadChatImage(file)
        chat.sendMessage(toId, url, 'image')
      } catch (error) {
        const msg = error instanceof Error ? error.message : '上传失败'
        ElMessage.error(msg)
      }
    }
  }
}

onBeforeUnmount(() => {
  if (recordTimer) {
    window.clearInterval(recordTimer)
    recordTimer = null
  }
  if (recording.value) {
    mediaRecorder?.stop()
    recording.value = false
  }
})
</script>

<template>
  <div class="message-input">
    <div class="message-toolbar">
      <button class="icon-btn" title="表情" @click="onEmojiClick">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M12 2a10 10 0 1 1 0 20 10 10 0 0 1 0-20Zm-4 9a1.2 1.2 0 1 0 0-2.4A1.2 1.2 0 0 0 8 11Zm8 0a1.2 1.2 0 1 0 0-2.4A1.2 1.2 0 0 0 16 11Zm-8.2 3.3a1 1 0 1 0 1.4 1.4A4 4 0 0 0 12 16a4 4 0 0 0 2.8-1.3 1 1 0 1 0-1.4-1.4A2 2 0 0 1 12 14a2 2 0 0 1-1.2-.7Z"
          />
        </svg>
      </button>
      <button
        class="icon-btn voice-btn"
        :class="{ active: recording }"
        :title="recording ? '点击结束录音并发送' : '点击开始录音'"
        @click="onVoiceClick"
      >
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M12 3a3 3 0 0 1 3 3v6a3 3 0 1 1-6 0V6a3 3 0 0 1 3-3Zm-7 9a1 1 0 0 1 2 0 5 5 0 0 0 10 0 1 1 0 1 1 2 0 7 7 0 0 1-6 6.9V21a1 1 0 1 1-2 0v-2.1A7 7 0 0 1 5 12Z"
          />
        </svg>
      </button>
      <button class="icon-btn" title="图片" @click="onImageClick">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M4 5h16a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2Zm0 2v10h16V7H4Zm3 8 3-4 3 3 2-2 3 3H7Z"
          />
        </svg>
      </button>
      <button class="icon-btn" title="文件(图片/文档/视频)" @click="onFileClick">
        <svg viewBox="0 0 24 24" class="icon">
          <path
            fill="currentColor"
            d="M14 2H7a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V7l-5-5Zm1 6h4v12H7V4h6v4Z"
          />
        </svg>
      </button>
    </div>
    <div v-if="recording" class="recording-tip">
      正在录音 {{ recordSeconds }}s，松开按钮发送
    </div>
    <div v-if="showEmojiPanel" class="emoji-panel">
      <button v-for="emoji in emojiList" :key="emoji" class="emoji-item" @click="appendEmoji(emoji)">
        {{ emoji }}
      </button>
    </div>
    <input
      ref="imageInput"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      style="display: none"
      @change="onImageChange"
    />
    <input
      ref="fileInput"
      type="file"
      style="display: none"
      @change="onFileChange"
    />
    <el-input
      v-model="content"
      type="textarea"
      :rows="4"
      placeholder="输入消息内容"
      @paste="onPaste"
    />
    <div class="message-actions">
      <el-button type="success" size="small" @click="send">发送</el-button>
    </div>
  </div>
</template>

<style scoped>
.message-input {
  padding: 12px 16px;
  border-top: 1px solid #e5e7eb;
  background: #fff;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message-toolbar {
  display: flex;
  gap: 8px;
}

.emoji-panel {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 6px;
  padding: 8px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fff;
}

.emoji-item {
  border: none;
  background: #f8fafc;
  border-radius: 6px;
  padding: 4px 0;
  cursor: pointer;
  font-size: 18px;
}

.emoji-item:hover {
  background: #eef2ff;
}

.icon-btn {
  width: 32px;
  height: 32px;
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

.icon-btn.active {
  border-color: #ef4444;
  color: #ef4444;
}

.voice-btn.active {
  background: #fee2e2;
}

.recording-tip {
  font-size: 12px;
  color: #ef4444;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  padding: 6px 8px;
}

.icon-btn .icon {
  width: 18px;
  height: 18px;
}

.message-actions {
  display: flex;
  justify-content: flex-end;
}
</style>
