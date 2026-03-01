<script setup lang="ts">
// 后台管理入口占位，后续可接入角色权限与管理功能
</script>

<template>
  <div class="admin-page">
    <div class="admin-card">
      <div class="admin-title">后台管理中心</div>
      <div class="admin-desc">这里将接入用户、群组、消息、文件、通话记录管理能力。</div>
    </div>
  </div>
</template>

<style scoped>
.admin-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
}

.admin-card {
  padding: 32px;
  border-radius: 16px;
  background: #0b1220;
}

.admin-title {
  font-size: 18px;
  margin-bottom: 12px;
}

.admin-desc {
  font-size: 14px;
  color: #94a3b8;
}
</style>
