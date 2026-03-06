import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import LoginView from '../views/LoginView.vue'
import ChatView from '../views/ChatView.vue'
import AdminView from '../views/AdminView.vue'

// 前端路由配置：登录、聊天、后台管理
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login',
    },
    {
      path: '/login',
      name: 'Login',
      component: LoginView,
      meta: { requiresAuth: false },
    },
    {
      path: '/chat',
      name: 'Chat',
      component: ChatView,
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      name: 'Admin',
      component: AdminView,
      meta: { requiresAuth: true },
    },
  ],
})

// 全局前置守卫
router.beforeEach(async (to: RouteLocationNormalized, _from, next) => {
  const authStore = useAuthStore()

  // 尝试恢复登录状态：用户信息为空，但token存在，说明浏览器刷新了，因为Pinia内存丢失，但localStorage token还存在
  if (!authStore.user && authStore.token) {
    try {
      await authStore.fetchProfile()
    } catch (error) {
      console.error('Failed to fetch profile:', error)
      authStore.logout()
      next('/login')
      return
    }
  }

  const requiresAuth = to.meta.requiresAuth

  // 需要登录，但是未登录
  if (requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    // 已登录，但是访问登录页，重定向到聊天页
    next('/chat')
  } else {
    next()
  }
})

export default router
