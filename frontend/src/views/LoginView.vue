<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const mode = ref<'login' | 'register'>('login')
const form = reactive({
  username: '',
  password: '',
  nickname: '',
  confirmPassword: '',
})

const auth = useAuthStore()
const router = useRouter()

async function onSubmit() {
  if (!form.username || !form.password) {
    ElMessage.error('请输入账号和密码')
    return
  }
  try {
    await auth.login(form.username, form.password)
    if (auth.isAuthenticated) {
      router.push('/chat')
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '登录失败'
    ElMessage.error(message)
  }
}

async function onRegister() {
  if (!form.username || !form.password) {
    ElMessage.error('请输入账号和密码')
    return
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.error('两次密码不一致')
    return
  }
  try {
    await auth.register(form.username, form.password, form.nickname)
    if (auth.isAuthenticated) {
      ElMessage.success('注册成功')
      auth.logout()
      mode.value = 'login'
      form.password = ''
      form.confirmPassword = ''
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : '注册失败'
    ElMessage.error(message)
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-title">GoChat</div>
      <el-tabs v-model="mode">
        <el-tab-pane label="登录" name="login">
          <el-form label-position="top">
            <el-form-item label="账号">
              <el-input v-model="form.username" placeholder="请输入用户名" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" placeholder="请输入密码" />
            </el-form-item>
            <el-button type="primary" :loading="auth.loading" @click="onSubmit">
              登录
            </el-button>
          </el-form>
        </el-tab-pane>
        <el-tab-pane label="注册" name="register">
          <el-form label-position="top">
            <el-form-item label="账号">
              <el-input v-model="form.username" placeholder="请输入用户名" />
            </el-form-item>
            <el-form-item label="昵称">
              <el-input v-model="form.nickname" placeholder="可选" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="form.password" type="password" placeholder="请输入密码" />
            </el-form-item>
            <el-form-item label="确认密码">
              <el-input v-model="form.confirmPassword" type="password" placeholder="再次输入密码" />
            </el-form-item>
            <el-button type="primary" :loading="auth.loading" @click="onRegister">
              注册
            </el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #eef1f5;
}

.login-card {
  width: 360px;
  padding: 32px;
  background: #fff;
  border-radius: 14px;
  color: #111827;
  box-shadow: 0 16px 30px rgba(15, 23, 42, 0.1);
}

.login-title {
  font-size: 20px;
  margin-bottom: 20px;
  text-align: center;
  color: #07c160;
  font-weight: 600;
}
</style>
