<template>
  <div class="page">
    <h3>无权限访问</h3>
    <p>您没有访问该页面的权限，请联系管理员。</p>
    <el-button type="primary" @click="$router.push('/dashboard')">返回首页</el-button>
  </div>
</template>

<style scoped>
.page { padding: 40px; text-align: center; }
p { color: #999; margin: 16px 0 24px; }
</style>
