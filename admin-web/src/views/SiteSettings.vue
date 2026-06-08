<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { fetchSiteSettings, updateSiteSettings } from '../api/platform'
import ImageUpload from '../components/ImageUpload.vue'

const form = ref({ logo_url: '', site_name: '搜店网' })
const saving = ref(false)

async function load() {
  form.value = await fetchSiteSettings()
}

async function handleSave() {
  saving.value = true
  try {
    await updateSiteSettings(form.value)
    ElMessage.success('已保存')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <h2>站点设置</h2>
    <el-card style="max-width:560px">
      <el-form label-width="90px">
        <el-form-item label="站点名称">
          <el-input v-model="form.site_name" placeholder="搜店网" />
        </el-form-item>
        <el-form-item label="Logo">
          <ImageUpload v-model="form.logo_url" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
h2 { margin: 0 0 16px; font-size: 18px; }
</style>
