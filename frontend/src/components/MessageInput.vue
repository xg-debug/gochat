<script setup lang="ts">
import { ref } from 'vue'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const content = ref('')

function send() {
  if (!content.value.trim()) return
  const conversationId = chat.activeConversationId
  if (!conversationId || !conversationId.startsWith('u_')) return
  const toId = Number(conversationId.replace('u_', ''))
  if (!toId) return
  chat.sendMessage(toId, content.value, 'text')
  content.value = ''
}
</script>

<template>
  <div class="message-input">
    <div class="message-toolbar">
      <el-button size="small">表情</el-button>
      <el-button size="small">图片</el-button>
      <el-button size="small">文件</el-button>
      <el-button size="small">截图</el-button>
    </div>
    <el-input
      v-model="content"
      type="textarea"
      :rows="4"
      placeholder="输入消息内容"
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
  background: #f7f7f7;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message-toolbar {
  display: flex;
  gap: 8px;
}

.message-actions {
  display: flex;
  justify-content: flex-end;
}
</style>
