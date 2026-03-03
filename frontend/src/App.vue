<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const handleUnauthorized = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  window.addEventListener('unauthorized', handleUnauthorized)
})

onUnmounted(() => {
  window.removeEventListener('unauthorized', handleUnauthorized)
})
</script>

<template>
  <router-view />
</template>
