<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import { searchUser, sendFriendRequest, listFriendRequests, handleFriendRequest, updateProfile, uploadAvatar, logoutRequest, searchGroup, joinGroup, createGroup, listGroups, deleteFriend, blockFriend, uploadGroupAvatar, getGroupProfile, updateGroupProfile, listGroupMembers, kickGroupMember, setGroupAdmin, type SearchUserResult, type FriendRequestItem, type GroupItem, type GroupMember } from '../services/api'
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
const pendingFriendCount = ref(0)
let friendPollTimer: number | null = null

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
    ElMessage.success('已发送好友请求')
    user.pending = true
    user.pendingFromMe = true
  } catch (error) {
    const msg = error instanceof Error ? error.message : '发送失败'
    ElMessage.error(msg)
  }
}

const openFriendRequests = async () => {
  showFriendRequests.value = true
  try {
    friendRequests.value = await listFriendRequests()
    pendingFriendCount.value = friendRequests.value.length
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
    pendingFriendCount.value = friendRequests.value.length
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

const onDeleteFriend = async (id: string) => {
  const rawId = Number(id.replace('u_', ''))
  if (!rawId) return
  try {
    await deleteFriend(rawId)
    ElMessage.success('已删除好友')
    await chat.bootstrap()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const onBlockFriend = async (id: string) => {
  const rawId = Number(id.replace('u_', ''))
  if (!rawId) return
  try {
    await blockFriend(rawId)
    ElMessage.success('已拉黑')
    await chat.bootstrap()
  } catch (error) {
    ElMessage.error('拉黑失败')
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

const refreshFriendRequests = async () => {
  try {
    const list = await listFriendRequests()
    pendingFriendCount.value = list.length
  } catch {
    // ignore
  }
}

// 群聊相关
const showGroupDialog = ref(false)
const groupKeyword = ref('')
const groupResults = ref<GroupItem[]>([])
const groupCreating = ref(false)
const groupName = ref('')
const myGroups = ref<GroupItem[]>([])
const showGroupInfo = ref(false)
const groupProfile = ref<GroupItem | null>(null)
const groupMembers = ref<GroupMember[]>([])
const groupForm = ref({ name: '', avatar: '', notice: '' })
const groupAvatarInput = ref<HTMLInputElement | null>(null)

// 通话相关
const callVisible = ref(false)
const callStatus = ref<'idle' | 'calling' | 'ringing' | 'in-call'>('idle')
const callType = ref<'audio' | 'video'>('video')
const callPeerId = ref<number | null>(null)
const localVideoRef = ref<HTMLVideoElement | null>(null)
const remoteVideoRef = ref<HTMLVideoElement | null>(null)
const localStream = ref<MediaStream | null>(null)
const remoteStream = ref<MediaStream | null>(null)
let peerConnection: RTCPeerConnection | null = null
let pendingOffer: RTCSessionDescriptionInit | null = null
const isMuted = ref(false)
const isCameraOff = ref(false)

const openGroupDialog = async () => {
  showGroupDialog.value = true
  try {
    myGroups.value = await listGroups()
  } catch {
    myGroups.value = []
  }
}

const onSearchGroup = async () => {
  if (!groupKeyword.value.trim()) return
  try {
    groupResults.value = await searchGroup(groupKeyword.value.trim())
  } catch {
    ElMessage.error('搜索群聊失败')
  }
}

const onJoinGroup = async (group: GroupItem) => {
  try {
    await joinGroup(group.id)
    ElMessage.success('已加入群聊')
    await chat.bootstrap()
  } catch (error) {
    const msg = error instanceof Error ? error.message : '加入失败'
    ElMessage.error(msg)
  }
}

const onCreateGroup = async () => {
  if (!groupName.value.trim()) return
  groupCreating.value = true
  try {
    const created = await createGroup(groupName.value.trim())
    ElMessage.success('群聊已创建')
    groupName.value = ''
    await chat.bootstrap()
    showGroupDialog.value = false
    chat.selectConversation(`g_${created.id}`)
  } catch (error) {
    const msg = error instanceof Error ? error.message : '创建失败'
    ElMessage.error(msg)
  } finally {
    groupCreating.value = false
  }
}

const openGroupInfo = async () => {
  const conversationId = chat.activeConversationId
  if (!conversationId || !conversationId.startsWith('g_')) return
  const groupId = Number(conversationId.replace('g_', ''))
  if (!groupId) return
  try {
    groupProfile.value = await getGroupProfile(groupId)
    groupMembers.value = await listGroupMembers(groupId)
    groupForm.value = {
      name: groupProfile.value.name || '',
      avatar: groupProfile.value.avatar || '',
      notice: groupProfile.value.notice || '',
    }
    showGroupInfo.value = true
  } catch {
    ElMessage.error('获取群信息失败')
  }
}

const onGroupAvatarClick = () => {
  groupAvatarInput.value?.click()
}

const onGroupAvatarChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  if (!input.files || !input.files[0]) return
  try {
    const url = await uploadGroupAvatar(input.files[0])
    groupForm.value.avatar = url
  } catch {
    ElMessage.error('上传失败')
  } finally {
    input.value = ''
  }
}

const onSaveGroupProfile = async () => {
  if (!groupProfile.value) return
  try {
    await updateGroupProfile({
      groupId: groupProfile.value.id,
      name: groupForm.value.name,
      avatar: groupForm.value.avatar,
      notice: groupForm.value.notice,
    })
    ElMessage.success('已更新')
    await chat.bootstrap()
    showGroupInfo.value = false
  } catch {
    ElMessage.error('更新失败')
  }
}

const onKickMember = async (userId: number) => {
  if (!groupProfile.value) return
  try {
    await kickGroupMember(groupProfile.value.id, userId)
    groupMembers.value = groupMembers.value.filter((m) => m.userId !== userId)
  } catch {
    ElMessage.error('踢人失败')
  }
}

const onSetAdmin = async (userId: number, action: 'set' | 'unset') => {
  if (!groupProfile.value) return
  try {
    await setGroupAdmin(groupProfile.value.id, userId, action)
    groupMembers.value = await listGroupMembers(groupProfile.value.id)
  } catch {
    ElMessage.error('操作失败')
  }
}

const ensurePeerConnection = () => {
  if (peerConnection) return peerConnection
  peerConnection = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
  })
  peerConnection.onicecandidate = (event) => {
    if (event.candidate && callPeerId.value) {
      chat.sendCallSignal(callPeerId.value, {
        action: 'candidate',
        candidate: event.candidate,
      })
    }
  }
  peerConnection.ontrack = (event) => {
    remoteStream.value = event.streams[0]
    if (remoteVideoRef.value) {
      remoteVideoRef.value.srcObject = remoteStream.value
    }
  }
  return peerConnection
}

const startLocalStream = async (type: 'audio' | 'video') => {
  const stream = await navigator.mediaDevices.getUserMedia({
    audio: true,
    video: type === 'video',
  })
  localStream.value = stream
  if (localVideoRef.value) {
    localVideoRef.value.srcObject = stream
  }
  const pc = ensurePeerConnection()
  stream.getTracks().forEach((track) => pc.addTrack(track, stream))
}

const cleanupCall = () => {
  callStatus.value = 'idle'
  callPeerId.value = null
  pendingOffer = null
  isMuted.value = false
  isCameraOff.value = false
  if (peerConnection) {
    peerConnection.close()
    peerConnection = null
  }
  if (localStream.value) {
    localStream.value.getTracks().forEach((t) => t.stop())
    localStream.value = null
  }
  if (remoteStream.value) {
    remoteStream.value.getTracks().forEach((t) => t.stop())
    remoteStream.value = null
  }
  callVisible.value = false
}

const toggleMute = () => {
  if (!localStream.value) return
  const enabled = !isMuted.value
  localStream.value.getAudioTracks().forEach((t) => {
    t.enabled = enabled
  })
  isMuted.value = !isMuted.value
}

const toggleCamera = () => {
  if (!localStream.value) return
  if (callType.value !== 'video') return
  const enabled = !isCameraOff.value
  localStream.value.getVideoTracks().forEach((t) => {
    t.enabled = enabled
  })
  isCameraOff.value = !isCameraOff.value
}

const startCall = async (type: 'audio' | 'video') => {
  const conversationId = chat.activeConversationId
  if (!conversationId || !conversationId.startsWith('u_')) {
    ElMessage.error('请选择单人会话')
    return
  }
  const peerId = Number(conversationId.replace('u_', ''))
  if (!peerId) return
  callType.value = type
  callPeerId.value = peerId
  callVisible.value = true
  callStatus.value = 'calling'
  await startLocalStream(type)
  const pc = ensurePeerConnection()
  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)
  chat.sendCallSignal(peerId, {
    action: 'offer',
    sdp: offer,
    callType: type,
  })
}

const acceptCall = async () => {
  if (!pendingOffer || !callPeerId.value) return
  await startLocalStream(callType.value)
  const pc = ensurePeerConnection()
  await pc.setRemoteDescription(pendingOffer)
  const answer = await pc.createAnswer()
  await pc.setLocalDescription(answer)
  chat.sendCallSignal(callPeerId.value, {
    action: 'answer',
    sdp: answer,
  })
  callStatus.value = 'in-call'
  pendingOffer = null
}

const rejectCall = () => {
  if (callPeerId.value) {
    chat.sendCallSignal(callPeerId.value, { action: 'reject' })
  }
  cleanupCall()
}

const hangupCall = () => {
  if (callPeerId.value) {
    chat.sendCallSignal(callPeerId.value, { action: 'hangup' })
  }
  cleanupCall()
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
  await refreshFriendRequests()
  friendPollTimer = window.setInterval(refreshFriendRequests, 15000)
})

onBeforeUnmount(() => {
  chat.disconnect()
  if (friendPollTimer) {
    window.clearInterval(friendPollTimer)
    friendPollTimer = null
  }
})

const onIncomingMessage = (event: Event) => {
  const detail = (event as CustomEvent).detail as {
    name: string
    content: string
    contentType: string
  }
  const text =
    detail.contentType === 'image'
      ? '[图片]'
      : detail.contentType === 'audio'
        ? '[语音]'
        : detail.contentType === 'file'
          ? '[文件]'
          : detail.contentType === 'video'
            ? '[视频]'
            : detail.content
  ElNotification({
    title: detail.name || '新消息',
    message: text,
    duration: 2500,
  })
}

onMounted(() => {
  window.addEventListener('incoming-message', onIncomingMessage)
})

onBeforeUnmount(() => {
  window.removeEventListener('incoming-message', onIncomingMessage)
})

const onCallSignal = async (event: Event) => {
  const detail = (event as CustomEvent).detail as {
    fromId: number
    payload: Record<string, unknown>
  }
  const action = detail.payload?.action as string
  if (!detail.fromId) return
  if (action === 'offer') {
    callPeerId.value = detail.fromId
    callType.value = (detail.payload?.callType as 'audio' | 'video') || 'video'
    pendingOffer = detail.payload?.sdp as RTCSessionDescriptionInit
    callVisible.value = true
    callStatus.value = 'ringing'
    return
  }
  if (action === 'answer') {
    if (!peerConnection) return
    const sdp = detail.payload?.sdp as RTCSessionDescriptionInit
    if (sdp) {
      await peerConnection.setRemoteDescription(sdp)
      callStatus.value = 'in-call'
    }
    return
  }
  if (action === 'candidate') {
    if (!peerConnection) return
    const candidate = detail.payload?.candidate as RTCIceCandidateInit
    if (candidate) {
      await peerConnection.addIceCandidate(candidate)
    }
    return
  }
  if (action === 'reject' || action === 'hangup') {
    ElMessage.info(action === 'reject' ? '对方已拒绝' : '通话已结束')
    cleanupCall()
  }
}

onMounted(() => {
  window.addEventListener('call-signal', onCallSignal)
})

onBeforeUnmount(() => {
  window.removeEventListener('call-signal', onCallSignal)
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
          <span v-if="pendingFriendCount > 0" class="nav-badge"></span>
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
          <ContactList
            @open-group="openGroupDialog"
            @delete-friend="onDeleteFriend"
            @block-friend="onBlockFriend"
          />
        </div>
      </div>
    </section>

    <main class="wechat-chat">
      <template v-if="chat.activeConversationId">
        <ChatHeader
          @more="openGroupInfo"
          @group-manage="openGroupInfo"
          @audio-call="startCall('audio')"
          @video-call="startCall('video')"
        />
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
        <el-input v-model="searchKeyword" placeholder="输入账号搜索" @keyup.enter="onSearchUser">
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
            <div class="username">账号: {{ user.username }}</div>
          </div>
          <el-button 
            size="small" 
            type="primary" 
            :disabled="user.isFriend || user.pending"
            @click="onAddFriend(user)"
          >
            {{ user.isFriend ? '已添加' : user.pending ? (user.pendingFromMe ? '等待验证' : '对方已申请') : '添加' }}
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

    <!-- 弹窗：群聊 -->
    <el-dialog v-model="showGroupDialog" title="群聊" width="420px">
      <div class="group-section">
        <div class="group-title">搜索群聊</div>
        <el-input v-model="groupKeyword" placeholder="输入群名称搜索" @keyup.enter="onSearchGroup">
          <template #append>
            <el-button @click="onSearchGroup">搜索</el-button>
          </template>
        </el-input>
        <div class="group-results">
          <div v-for="g in groupResults" :key="g.id" class="group-item">
            <div class="group-name">{{ g.name }}</div>
            <el-button size="small" type="primary" @click="onJoinGroup(g)">加入</el-button>
          </div>
          <div v-if="groupResults.length === 0 && groupKeyword" class="no-result">
            未找到群聊
          </div>
        </div>
      </div>
      <div class="group-section">
        <div class="group-title">创建群聊</div>
        <div class="group-create">
          <el-input v-model="groupName" placeholder="请输入群名称" />
          <el-button type="success" :loading="groupCreating" @click="onCreateGroup">创建</el-button>
        </div>
      </div>
      <div class="group-section">
        <div class="group-title">我的群聊</div>
        <div class="group-results">
          <div v-for="g in myGroups" :key="g.id" class="group-item">
            <div class="group-name">{{ g.name }}</div>
            <el-button size="small" @click="chat.selectConversation(`g_${g.id}`); showGroupDialog = false">进入</el-button>
          </div>
          <div v-if="myGroups.length === 0" class="no-result">
            暂无群聊
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 弹窗：群聊信息 -->
    <el-dialog v-model="showGroupInfo" title="群聊信息" width="480px">
      <div v-if="groupProfile" class="group-info">
        <div class="group-info-header">
          <div class="group-avatar" @click="onGroupAvatarClick">
            <img v-if="groupForm.avatar" :src="groupForm.avatar" />
            <div v-else class="avatar-placeholder">群</div>
            <div class="upload-mask"><span>更换</span></div>
          </div>
          <input
            ref="groupAvatarInput"
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            style="display: none"
            @change="onGroupAvatarChange"
          />
          <div class="group-info-fields">
            <el-input v-model="groupForm.name" placeholder="群名称" />
            <el-input v-model="groupForm.notice" placeholder="群公告" />
          </div>
        </div>
        <div class="group-members">
          <div class="group-title">群成员</div>
          <div v-for="m in groupMembers" :key="m.userId" class="member-item">
            <div class="member-avatar">
              <img v-if="m.avatar" :src="m.avatar" />
              <span v-else>{{ m.nickname.slice(0, 1) }}</span>
            </div>
            <div class="member-info">
              <div class="member-name">{{ m.nickname }}</div>
              <div class="member-role">{{ m.role === 2 ? '群主' : m.role === 1 ? '管理员' : '成员' }}</div>
            </div>
            <div class="member-actions" v-if="groupProfile.role && groupProfile.role >= 1">
              <el-button size="small" @click="onKickMember(m.userId)" v-if="m.userId !== auth.user?.id">踢人</el-button>
              <el-button size="small" v-if="groupProfile.role === 2 && m.role === 0" @click="onSetAdmin(m.userId, 'set')">设为管理员</el-button>
              <el-button size="small" v-if="groupProfile.role === 2 && m.role === 1" @click="onSetAdmin(m.userId, 'unset')">取消管理员</el-button>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="showGroupInfo = false">取消</el-button>
        <el-button type="primary" @click="onSaveGroupProfile">保存</el-button>
      </template>
    </el-dialog>

    <!-- 通话浮层 -->
    <div v-if="callVisible" class="call-overlay">
      <div class="call-card">
        <div class="call-header">
          <div class="call-title">{{ callStatus === 'ringing' ? '来电' : callStatus === 'calling' ? '呼叫中' : '通话中' }}</div>
          <div class="call-type">{{ callType === 'video' ? '视频通话' : '语音通话' }}</div>
        </div>
        <div class="call-body" :class="{ audio: callType === 'audio' }">
          <video ref="remoteVideoRef" autoplay playsinline></video>
          <video ref="localVideoRef" autoplay muted playsinline class="local-video"></video>
        </div>
        <div class="call-actions">
          <button v-if="callStatus === 'ringing'" class="call-btn reject" @click="rejectCall">拒绝</button>
          <button v-if="callStatus === 'ringing'" class="call-btn accept" @click="acceptCall">接听</button>
          <button v-if="callStatus === 'calling' || callStatus === 'in-call'" class="call-btn ghost" @click="toggleMute">
            {{ isMuted ? '取消静音' : '静音' }}
          </button>
          <button v-if="(callStatus === 'calling' || callStatus === 'in-call') && callType === 'video'" class="call-btn ghost" @click="toggleCamera">
            {{ isCameraOff ? '开启摄像头' : '关闭摄像头' }}
          </button>
          <button v-if="callStatus === 'calling' || callStatus === 'in-call'" class="call-btn hangup" @click="hangupCall">挂断</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wechat-shell {
  display: grid;
  grid-template-columns: 68px 280px 1fr;
  height: 100vh;
  background: #f5f5f5;
  overflow: hidden;
}

.wechat-nav {
  background: #2b2b2b;
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
  border-radius: 6px;
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
  color: #9aa0a6;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
  position: relative;
}

.nav-icon:hover {
  color: #fff;
}

.nav-icon.active {
  color: #07c160;
}

.nav-badge {
  position: absolute;
  width: 8px;
  height: 8px;
  background: #ef4444;
  border-radius: 50%;
  top: 6px;
  right: 6px;
}

.icon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
}

.wechat-panel {
  background: #f7f7f7;
  border-right: 1px solid #e5e7eb;
  display: flex;
  flex-direction: column;
}

.panel-search {
  padding: 12px;
  background: #f7f7f7;
  -webkit-app-region: drag;
  border-bottom: 1px solid #ececec;
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

.group-section {
  margin-bottom: 16px;
}

.group-title {
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 8px;
}

.group-create {
  display: flex;
  gap: 8px;
}

.group-results {
  margin-top: 10px;
  max-height: 180px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 4px;
  border-bottom: 1px solid #f0f0f0;
}

.group-name {
  font-size: 14px;
  color: #111827;
}

.group-info {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.group-info-header {
  display: flex;
  gap: 12px;
  align-items: center;
}

.group-avatar {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  background: #e5e7eb;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.group-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.group-info-fields {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.group-members {
  border-top: 1px solid #f0f0f0;
  padding-top: 12px;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid #f3f4f6;
}

.member-avatar {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: #d1d5db;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.member-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.member-info {
  flex: 1;
}

.member-name {
  font-size: 13px;
  color: #111827;
}

.member-role {
  font-size: 12px;
  color: #9aa0a6;
}

.call-overlay {
  position: fixed;
  inset: 0;
  background: rgba(15, 23, 42, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.call-card {
  width: min(720px, 90vw);
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 16px;
  padding: 16px;
  box-shadow: 0 24px 50px rgba(15, 23, 42, 0.4);
}

.call-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 12px;
}

.call-title {
  font-size: 16px;
  font-weight: 600;
}

.call-type {
  font-size: 12px;
  color: #94a3b8;
}

.call-body {
  position: relative;
  background: #0b1220;
  border-radius: 12px;
  min-height: 320px;
  overflow: hidden;
}

.call-body video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.call-body.audio {
  min-height: 180px;
}

.local-video {
  position: absolute;
  right: 12px;
  bottom: 12px;
  width: 160px;
  height: 100px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.call-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 14px;
}

.call-btn {
  border: none;
  padding: 8px 18px;
  border-radius: 20px;
  font-size: 13px;
  cursor: pointer;
}

.call-btn.accept {
  background: #22c55e;
  color: #fff;
}

.call-btn.reject,
.call-btn.hangup {
  background: #ef4444;
  color: #fff;
}

.call-btn.ghost {
  background: #1f2937;
  color: #e5e7eb;
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
