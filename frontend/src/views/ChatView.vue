<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import ChatHeader from '../components/ChatHeader.vue'
import ConversationList from '../components/ConversationList.vue'
import ContactList from '../components/ContactList.vue'
import MessageList from '../components/MessageList.vue'
import MessageInput from '../components/MessageInput.vue'

const auth = useAuthStore()
const chat = useChatStore()
const activeTab = ref('conversations')

onMounted(async () => {
  if (!auth.user && auth.token) {
    await auth.fetchProfile()
  }
  await chat.bootstrap()
  if (auth.user?.id && auth.token) {
    chat.connect(auth.user.id, auth.token)
  }
})

onBeforeUnmount(() => {
  chat.disconnect()
})
</script>

<template>
  <div class="wechat-shell">
    <aside class="wechat-nav">
      <div class="nav-avatar">
        <div class="avatar-circle">{{ auth.user?.nickname?.slice(0, 1) || '我' }}</div>
      </div>
      <div class="nav-list">
        <div class="nav-icon active" title="聊天">
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M6 4h9a5 5 0 0 1 5 5v4a5 5 0 0 1-5 5H9l-4 3v-3H6a5 5 0 0 1-5-5V9a5 5 0 0 1 5-5Z"
              fill="currentColor"
            />
          </svg>
        </div>
        <div class="nav-icon" title="通讯录">
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M8 3a3 3 0 1 1 0 6 3 3 0 0 1 0-6Zm8 1h3v17h-3v-1.5h-1v-2h1v-3h-1v-2h1v-3h-1v-2h1V4ZM4 14c0-2 2-4 4-4s4 2 4 4v3H4v-3Z"
              fill="currentColor"
            />
          </svg>
        </div>
      </div>
      <div class="nav-bottom">
        <div class="nav-icon" title="设置">
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M12 8.2a3.8 3.8 0 1 1 0 7.6 3.8 3.8 0 0 1 0-7.6Zm8.6 3.3-1.7-.6a6.7 6.7 0 0 0-.6-1.4l.8-1.6-1.7-1.7-1.6.8c-.5-.3-1-.5-1.5-.6l-.6-1.7h-2.4l-.6 1.7c-.5.1-1 .3-1.5.6l-1.6-.8-1.7 1.7.8 1.6c-.3.5-.5 1-.6 1.5l-1.7.6v2.4l1.7.6c.1.5.3 1 .6 1.5l-.8 1.6 1.7 1.7 1.6-.8c.5.3 1 .5 1.5.6l.6 1.7h2.4l.6-1.7c.5-.1 1-.3 1.5-.6l1.6.8 1.7-1.7-.8-1.6c.3-.5.5-1 .6-1.5l1.7-.6v-2.4Z"
              fill="currentColor"
            />
          </svg>
        </div>
      </div>
    </aside>

    <section class="wechat-panel">
      <div class="panel-search">
        <el-input placeholder="搜索" size="small" />
      </div>
      <el-tabs v-model="activeTab" class="panel-tabs">
        <el-tab-pane label="会话" name="conversations">
          <ConversationList />
        </el-tab-pane>
        <el-tab-pane label="通讯录" name="contacts">
          <div class="contact-toolbar">
            <el-button size="small" plain>通讯录管理</el-button>
          </div>
          <ContactList />
        </el-tab-pane>
      </el-tabs>
    </section>

    <main class="wechat-chat">
      <ChatHeader />
      <MessageList />
      <MessageInput />
    </main>
  </div>
</template>

<style scoped>
.wechat-shell {
  display: grid;
  grid-template-columns: 72px 280px 1fr;
  height: 100vh;
  background: #e6e7ea;
}

.wechat-nav {
  background: #d8dbe2;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 10px;
  gap: 16px;
}

.nav-avatar {
  margin-bottom: 12px;
}

.avatar-circle {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: #1f2937;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.nav-icon {
  width: 46px;
  height: 46px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.6);
  color: #475569;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.nav-icon.active {
  background: #10b981;
  color: #fff;
}

.icon {
  width: 22px;
  height: 22px;
}

.nav-bottom {
  margin-top: auto;
}

.wechat-panel {
  background: #f7f7f7;
  border-right: 1px solid #e4e7ec;
  display: flex;
  flex-direction: column;
  padding: 12px;
  gap: 8px;
}

.panel-search {
  padding: 4px 0;
}

.panel-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.panel-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: auto;
}

.contact-toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 8px;
}

.wechat-chat {
  display: flex;
  flex-direction: column;
  background: #f2f3f5;
}
</style>
