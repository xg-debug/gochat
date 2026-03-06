import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { loginRequest, profileRequest, registerRequest } from '../services/api'
import type { UserProfile } from '../types/user'

// 认证与用户信息
export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<UserProfile | null>(null)
  try {
    const savedUser = localStorage.getItem('user')
    if (savedUser) {
      user.value = JSON.parse(savedUser)
    }
  } catch (e) {
    console.error('Failed to parse user from localStorage', e)
  }

  const loading = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(username: string, password: string) {
    // 请求开始
    loading.value = true
    try {
      const result = await loginRequest(username, password)
      token.value = result.token
      user.value = result.user
      localStorage.setItem('token', result.token)
      localStorage.setItem('user', JSON.stringify(result.user))
    } finally {
      // 请求结束
      loading.value = false
    }
  }

  async function register(username: string, password: string, nickname: string) {
    loading.value = true
    try {
      const result = await registerRequest(username, password, nickname)
      token.value = result.token
      user.value = result.user
      localStorage.setItem('token', result.token)
      localStorage.setItem('user', JSON.stringify(result.user))
    } finally {
      loading.value = false
    }
  }

  // 获取用户信息
  async function fetchProfile() {
    if (!token.value) return
    try {
      const profile = await profileRequest()
      user.value = profile
      localStorage.setItem('user', JSON.stringify(profile))
    } catch (error) {
      console.error('Fetch profile failed', error)
      // 如果获取用户信息失败（可能是 token 过期），考虑是否要自动登出
      // 这里暂时只记录错误，交给调用方处理
      throw error
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // store对外暴露的接口
  return {
    token,
    user,
    loading,
    isAuthenticated,
    login,
    register,
    fetchProfile,
    logout,
  }
})
