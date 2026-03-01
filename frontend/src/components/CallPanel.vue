<script setup lang="ts">
import { ref } from 'vue'

// WebRTC 面板状态（仅展示 UI，占位信令逻辑）
const callStatus = ref('空闲')
</script>

<template>
  <div class="call-panel">
    <div class="panel-title">音视频通话</div>
    <div class="panel-status">当前状态：{{ callStatus }}</div>
    <div class="panel-actions">
      <el-button size="small" type="primary">发起呼叫</el-button>
      <el-button size="small">挂断</el-button>
    </div>
    <div class="panel-tip">WebRTC 信令服务将在后端完善后接入</div>
  </div>
</template>

<style scoped>
.call-panel {
  padding: 14px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.6);
}

.panel-title {
  font-size: 14px;
  margin-bottom: 8px;
}

.panel-status {
  font-size: 12px;
  color: #cbd5f5;
  margin-bottom: 10px;
}

.panel-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 8px;
}

.panel-tip {
  font-size: 12px;
  color: #94a3b8;
}
</style>
