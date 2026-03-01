import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { loginRequest, profileRequest, registerRequest } from '../services/api'

export type UserProfile = {
  id: number
  username: string
  nickname: string
  avatar: string
}

// 认证与用户信息状态
export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<UserProfile | null>(null)
  const loading = ref(false)

  const isAuthenticated = computed(() => Boolean(token.value))

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const result = await loginRequest(username, password)
      token.value = result.token
      user.value = result.user
      localStorage.setItem('token', result.token)
    } finally {
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
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    if (!token.value) return
    user.value = await profileRequest()
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

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
