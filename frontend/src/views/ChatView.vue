<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import { searchUser, sendFriendRequest, listFriendRequests, handleFriendRequest, updateProfile, uploadAvatar, logoutRequest, type SearchUserResult, type FriendRequestItem } from '../services/api'
import ChatHeader from '../components/ChatHeader.vue'
import ConversationList from '../components/ConversationList.vue'
import ContactList from '../components/ContactList.vue'
import MessageList from '../components/MessageList.vue'
import MessageInput from '../components/MessageInput.vue'

const auth = useAuthStore()
const chat = useChatStore()
const router = useRouter()
const activeTab = ref('conversations')

// 搜索加好友相关
const showAddFriend = ref(false)
const searchKeyword = ref('')
const searchResults = ref<SearchUserResult[]>([])
const searching = ref(false)

// 好友申请列表相关
const showFriendRequests = ref(false)
const friendRequests = ref<FriendRequestItem[]>([])

// 个人信息编辑相关
const showProfileEdit = ref(false)
const profileForm = ref({
  nickname: '',
  avatar: '',
  signature: '',
  gender: 0,
})

const switchTab = (tab: string) => {
  activeTab.value = tab
}

const onSearchUser = async () => {
  if (!searchKeyword.value) return
  searching.value = true
  try {
    searchResults.value = await searchUser(searchKeyword.value)
  } catch (error) {
    ElMessage.error('搜索失败')
  } finally {
    searching.value = false
  }
}

const onAddFriend = async (user: SearchUserResult) => {
  try {
    await sendFriendRequest(user.id)
    ElMessage.success('发送请求成功')
    user.isFriend = true // 临时标记，实际需要对方同意
  } catch (error) {
    const msg = error instanceof Error ? error.message : '发送失败'
    ElMessage.error(msg)
  }
}

const openFriendRequests = async () => {
  showFriendRequests.value = true
  try {
    friendRequests.value = await listFriendRequests()
  } catch (error) {
    ElMessage.error('获取列表失败')
  }
}

const onHandleRequest = async (id: number, action: 'accept' | 'reject') => {
  try {
    await handleFriendRequest(id, action)
    ElMessage.success(action === 'accept' ? '已同意' : '已拒绝')
    friendRequests.value = friendRequests.value.filter(item => item.id !== id)
    // 刷新通讯录
    if (action === 'accept') {
      await chat.bootstrap()
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const openProfileEdit = () => {
  if (auth.user) {
    profileForm.value = {
      nickname: auth.user.nickname || '',
      avatar: auth.user.avatar || '',
      signature: '', // 目前 UserProfile 类型还没加上 signature 字段，暂留空
      gender: 0,
    }
  }
  showProfileEdit.value = true
}

const fileInput = ref<HTMLInputElement | null>(null)

const onAvatarClick = () => {
  fileInput.value?.click()
}

const onFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (input.files && input.files[0]) {
    const file = input.files[0]
    try {
      const url = await uploadAvatar(file)
      profileForm.value.avatar = url
      ElMessage.success('头像上传成功')
    } catch (error) {
      ElMessage.error('头像上传失败')
    }
  }
}

const onSaveProfile = async () => {
  try {
    const updated = await updateProfile(profileForm.value)
    auth.user = updated
    localStorage.setItem('user', JSON.stringify(updated))
    ElMessage.success('保存成功')
    showProfileEdit.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const onLogout = async () => {
  try {
    await logoutRequest()
  } catch {
    // ignore logout errors
  } finally {
    chat.reset()
    auth.logout()
    router.push('/login')
  }
}

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
      <div class="nav-avatar" @click="openProfileEdit" title="修改个人信息" style="cursor: pointer;">
        <img v-if="auth.user?.avatar" :src="auth.user.avatar" class="avatar-img" />
        <div v-else class="avatar-circle">{{ auth.user?.nickname?.slice(0, 1) || '我' }}</div>
      </div>
      <div class="nav-list">
        <div 
          class="nav-icon" 
          :class="{ active: activeTab === 'conversations' }" 
          @click="switchTab('conversations')"
          title="聊天"
        >
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M6 4h9a5 5 0 0 1 5 5v4a5 5 0 0 1-5 5H9l-4 3v-3H6a5 5 0 0 1-5-5V9a5 5 0 0 1 5-5Z"
              fill="currentColor"
            />
          </svg>
        </div>
        <div 
          class="nav-icon" 
          :class="{ active: activeTab === 'contacts' }" 
          @click="switchTab('contacts')"
          title="通讯录"
        >
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M8 3a3 3 0 1 1 0 6 3 3 0 0 1 0-6Zm8 1h3v17h-3v-1.5h-1v-2h1v-3h-1v-2h1v-3h-1v-2h1V4ZM4 14c0-2 2-4 4-4s4 2 4 4v3H4v-3Z"
              fill="currentColor"
            />
          </svg>
        </div>
      </div>
      <div class="nav-bottom">
        <div class="nav-icon" title="加好友" @click="showAddFriend = true">
          <svg viewBox="0 0 24 24" class="icon">
             <path d="M15 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm-9-2V7H4v3H1v2h3v3h2v-3h3v-2H6zm9 4c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" fill="currentColor"/>
          </svg>
        </div>
        <div class="nav-icon" title="设置">
          <svg viewBox="0 0 24 24" class="icon">
            <path
              d="M12 8.2a3.8 3.8 0 1 1 0 7.6 3.8 3.8 0 0 1 0-7.6Zm8.6 3.3-1.7-.6a6.7 6.7 0 0 0-.6-1.4l.8-1.6-1.7-1.7-1.6.8c-.5-.3-1-.5-1.5-.6l-.6-1.7h-2.4l-.6 1.7c-.5.1-1 .3-1.5.6l-1.6-.8-1.7 1.7.8 1.6c-.3.5-.5 1-.6 1.5l-1.7.6v2.4l1.7.6c.1.5.3 1 .6 1.5l-.8 1.6 1.7 1.7 1.6-.8c.5.3 1 .5 1.5.6l.6 1.7h2.4l.6-1.7c.5-.1 1-.3 1.5-.6l1.6.8 1.7-1.7-.8-1.6c.3-.5.5-1 .6-1.5l1.7-.6v-2.4Z"
              fill="currentColor"
            />
          </svg>
        </div>
        <div class="nav-icon logout" title="退出登录" @click="onLogout">
          <svg viewBox="0 0 24 24" class="icon">
            <path
              fill="currentColor"
              d="M10 3H5a2 2 0 00-2 2v14a2 2 0 002 2h5v-2H5V5h5V3zm4.59 4.59L13.17 9l2.59 3-2.59 3 1.42 1.41L19 12l-4.41-4.41zM9 13h8v-2H9v2z"
            />
          </svg>
        </div>
      </div>
    </aside>

    <section class="wechat-panel">
      <div class="panel-search">
        <el-input placeholder="搜索" size="small" prefix-icon="Search" />
      </div>
      
      <!-- Custom Content Switching instead of el-tabs with headers -->
      <div class="panel-content">
        <ConversationList v-if="activeTab === 'conversations'" />
        <div v-if="activeTab === 'contacts'">
           <div class="contact-toolbar">
            <el-button size="small" plain @click="openFriendRequests">新朋友</el-button>
          </div>
          <ContactList />
        </div>
      </div>
    </section>

    <main class="wechat-chat">
      <template v-if="chat.activeConversationId">
        <ChatHeader />
        <MessageList />
        <MessageInput />
      </template>
      <div v-else class="empty-state">
        <div class="empty-icon">
          <svg viewBox="0 0 24 24" width="64" height="64" fill="#e5e5e5">
            <path d="M6 4h9a5 5 0 0 1 5 5v4a5 5 0 0 1-5 5H9l-4 3v-3H6a5 5 0 0 1-5-5V9a5 5 0 0 1 5-5Z" />
          </svg>
        </div>
      </div>
    </main>

    <!-- 弹窗：搜索加好友 -->
    <el-dialog v-model="showAddFriend" title="添加好友" width="400px">
      <div class="search-box">
        <el-input v-model="searchKeyword" placeholder="输入用户名搜索" @keyup.enter="onSearchUser">
          <template #append>
            <el-button @click="onSearchUser" :loading="searching">搜索</el-button>
          </template>
        </el-input>
      </div>
      <div class="search-results">
        <div v-for="user in searchResults" :key="user.id" class="result-item">
          <div class="result-avatar">
            <img v-if="user.avatar" :src="user.avatar" />
            <span v-else>{{ user.nickname.slice(0, 1) }}</span>
          </div>
          <div class="result-info">
            <div class="name">{{ user.nickname }}</div>
            <div class="username">ID: {{ user.username }}</div>
          </div>
          <el-button 
            size="small" 
            type="primary" 
            :disabled="user.isFriend"
            @click="onAddFriend(user)"
          >
            {{ user.isFriend ? '已添加' : '添加' }}
          </el-button>
        </div>
        <div v-if="searchResults.length === 0 && !searching && searchKeyword" class="no-result">
          未找到用户
        </div>
      </div>
    </el-dialog>

    <!-- 弹窗：新朋友请求 -->
    <el-dialog v-model="showFriendRequests" title="新朋友" width="400px">
      <div class="request-list">
        <div v-for="req in friendRequests" :key="req.id" class="request-item">
          <div class="result-avatar">
            <img v-if="req.avatar" :src="req.avatar" />
            <span v-else>{{ req.nickname.slice(0, 1) }}</span>
          </div>
          <div class="result-info">
            <div class="name">{{ req.nickname }}</div>
            <div class="msg">请求添加你为好友</div>
          </div>
          <div class="actions">
            <el-button size="small" type="success" @click="onHandleRequest(req.id, 'accept')">同意</el-button>
            <el-button size="small" type="danger" @click="onHandleRequest(req.id, 'reject')">拒绝</el-button>
          </div>
        </div>
        <div v-if="friendRequests.length === 0" class="no-result">暂无好友请求</div>
      </div>
    </el-dialog>

    <!-- 弹窗：修改个人信息 -->
    <el-dialog v-model="showProfileEdit" title="个人信息" width="400px">
      <el-form label-width="60px">
        <el-form-item label="头像">
          <div class="avatar-uploader" @click="onAvatarClick">
            <img v-if="profileForm.avatar" :src="profileForm.avatar" class="avatar-preview" />
            <div v-else class="avatar-placeholder">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="currentColor">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z"/>
              </svg>
            </div>
            <div class="upload-mask">
              <span>更换</span>
            </div>
          </div>
          <input 
            ref="fileInput"
            type="file" 
            accept="image/jpeg,image/png,image/gif" 
            style="display: none" 
            @change="onFileChange"
          />
        </el-form-item>
        <el-form-item label="昵称">
          <el-input v-model="profileForm.nickname" />
        </el-form-item>
        <el-form-item label="签名">
          <el-input v-model="profileForm.signature" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProfileEdit = false">取消</el-button>
        <el-button type="primary" @click="onSaveProfile">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.wechat-shell {
  display: grid;
  grid-template-columns: 60px 250px 1fr;
  height: 100vh;
  background: #f5f5f5;
  overflow: hidden;
}

.wechat-nav {
  background: #2e2e2e;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 0;
  justify-content: space-between;
}

.nav-avatar {
  margin-bottom: 20px;
}

.avatar-circle {
  width: 36px;
  height: 36px;
  border-radius: 4px;
  background: #fff;
  color: #2e2e2e;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 14px;
}

.avatar-img {
  width: 36px;
  height: 36px;
  border-radius: 4px;
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
  flex: 1;
  width: 100%;
  align-items: center;
}

.nav-bottom {
  margin-bottom: 10px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  align-items: center;
}

.nav-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #9ca3af;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.nav-icon:hover {
  color: #fff;
}

.nav-icon.active {
  color: #07c160;
}

.icon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
}

.wechat-panel {
  background: #f7f7f7;
  border-right: 1px solid #dcdfe6;
  display: flex;
  flex-direction: column;
}

.panel-search {
  padding: 12px;
  background: #f7f7f7;
  -webkit-app-region: drag;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
}

.contact-toolbar {
  padding: 10px 12px;
  border-bottom: 1px solid #e5e5e5;
}

.wechat-chat {
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

/* 搜索结果样式 */
.search-results {
  margin-top: 20px;
  max-height: 300px;
  overflow-y: auto;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.result-avatar {
  width: 40px;
  height: 40px;
  border-radius: 4px;
  background: #eee;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  overflow: hidden;
}

.result-avatar img {
  width: 100%;
  height: 100%;
}

.result-info {
  flex: 1;
}

.result-info .name {
  font-weight: bold;
  font-size: 14px;
}

.result-info .username {
  font-size: 12px;
  color: #999;
}

.no-result {
  text-align: center;
  color: #999;
  padding: 20px;
}

/* 好友请求列表样式 */
.request-list {
  max-height: 400px;
  overflow-y: auto;
}

.request-item {
  display: flex;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.request-item .msg {
  font-size: 12px;
  color: #999;
}

.actions {
  display: flex;
  gap: 8px;
}

.avatar-uploader {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  cursor: pointer;
  border: 1px dashed #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: border-color 0.2s;
}

.avatar-uploader:hover {
  border-color: #409eff;
}

.avatar-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  color: #8c939d;
}

.upload-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s;
  font-size: 12px;
}

.avatar-uploader:hover .upload-mask {
  opacity: 1;
}
</style>
