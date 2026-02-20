<template>
  <a-result
    status="404"
    title="404 Not Found"
    sub-title="抱歉，您访问的页面不存在。"
  >
    <template #extra>
      <a-button type="primary" @click="$router.push('/console/dashboard')">
        返回首页
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
