<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { uploadImage } from '../api/upload'

const props = withDefaults(defineProps<{
  modelValue: string | string[]
  multiple?: boolean
  limit?: number
  size?: 'default' | 'small'
}>(), {
  multiple: false,
  limit: 9,
  size: 'default',
})

const emit = defineEmits<{
  'update:modelValue': [string | string[]]
}>()

const uploading = ref(false)

const fileList = computed(() => {
  if (props.multiple) {
    const urls = Array.isArray(props.modelValue) ? props.modelValue : []
    return urls.filter(Boolean)
  }
  return props.modelValue ? [props.modelValue as string] : []
})

watch(
  () => props.modelValue,
  () => {},
  { deep: true },
)

async function handleUpload(options: { file: File }) {
  uploading.value = true
  try {
    const res = await uploadImage(options.file)
    if (props.multiple) {
      const urls = Array.isArray(props.modelValue) ? [...props.modelValue] : []
      urls.push(res.url)
      emit('update:modelValue', urls)
    } else {
      emit('update:modelValue', res.url)
    }
  } catch {
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
  }
}

function removeAt(index: number) {
  if (props.multiple) {
    const urls = [...(Array.isArray(props.modelValue) ? props.modelValue : [])]
    urls.splice(index, 1)
    emit('update:modelValue', urls)
  } else {
    emit('update:modelValue', '')
  }
}

const boxSize = computed(() => (props.size === 'small' ? 64 : 100))

const canAdd = computed(() => {
  if (!props.multiple) return fileList.value.length === 0
  return fileList.value.length < props.limit
})
</script>

<template>
  <div class="image-upload" :class="{ small: size === 'small' }">
    <div
      v-for="(url, index) in fileList"
      :key="url + index"
      class="thumb"
      :style="{ width: boxSize + 'px', height: boxSize + 'px' }"
    >
      <el-image :src="url" fit="cover" class="img" />
      <span class="remove" @click="removeAt(index)">×</span>
    </div>
    <el-upload
      v-if="canAdd"
      :show-file-list="false"
      :http-request="handleUpload"
      accept="image/jpeg,image/png,image/webp"
      :disabled="uploading"
    >
      <div class="upload-trigger" :style="{ width: boxSize + 'px', height: boxSize + 'px' }">
        <el-icon><Plus /></el-icon>
        <span v-if="size !== 'small'">{{ uploading ? '上传中' : '上传' }}</span>
      </div>
    </el-upload>
  </div>
</template>

<style scoped>
.image-upload {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.image-upload.small {
  gap: 0;
}
.thumb {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #eee;
}
.img {
  width: 100%;
  height: 100%;
}
.remove {
  position: absolute;
  top: 2px;
  right: 6px;
  color: #fff;
  background: rgba(0, 0, 0, 0.45);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  line-height: 18px;
  text-align: center;
  cursor: pointer;
  font-size: 16px;
}
.upload-trigger {
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #909399;
  cursor: pointer;
  font-size: 12px;
  gap: 4px;
}
.upload-trigger:hover {
  border-color: #409eff;
  color: #409eff;
}
.image-upload.small .remove {
  width: 16px;
  height: 16px;
  line-height: 14px;
  font-size: 12px;
  right: 2px;
}
</style>
