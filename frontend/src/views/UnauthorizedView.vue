<template>
  <a-result
    status="403"
    title="403 Unauthorized"
    sub-title="抱歉，您没有权限访问此页面。"
  >
    <template #extra>
      <a-button type="primary" @click="$router.push('/console/dashboard')">
        返回控制台
      </a-button>
    </template>
  </a-result>
</template>

<script setup lang="ts">
</script>

<style scoped>
:deep(.ant-result) {
  padding: 48px 32px;
}
</style>
