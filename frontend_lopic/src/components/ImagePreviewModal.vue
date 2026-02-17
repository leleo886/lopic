<template>
  <a-modal
    v-model:open="props.visible"
    :title="props.image?.original_name || '图片预览'"
    :width="800"
    :footer="null"
    @cancel="handleCancel"
    class="image-preview-modal"
  >
    <div v-if="props.image" class="preview-content">
      <div class="preview-image-container">
        <a-image
          :src="getFileUrl(props.image.file_url)"
          :alt="props.image.original_name"
          placeholder="加载中..."
          class="preview-image"
        />
      </div>
      <div class="preview-info">
        <div class="preview-details">
          <p>文件大小：{{ formatFileSize(props.image.file_size) }}</p>
          <p>图片尺寸：{{ props.image.width }} x {{ props.image.height }}</p>
          <p>上传时间：{{ formatDateTime(props.image.created_at) }}</p>
        </div>
        <div class="preview-tags" v-if="props.image.tags && props.image.tags.length > 0">
          <span class="tags-label">标签：</span>
          <a-tag
            v-for="(tag, index) in props.image.tags"
            :key="index"
            class="preview-tag"
          >
            {{ tag }}
          </a-tag>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script setup lang="ts">
import type { ImageResponse } from '../types/api';
import { serverUrl } from '../api/axios';

interface Props {
  visible: boolean;
  image: ImageResponse | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'close'): void;
}>();

// 获取文件完整 URL
const getFileUrl = (path: string): string => {
  if (path.startsWith('http')) {
    return path;
  }
  return `${serverUrl}${path}`;
};

// 格式化文件大小
const formatFileSize = (size: number): string => {
  if (size < 1024) {
    return `${size} B`;
  } else if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  } else {
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }
};

// 格式化日期时间
const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

// 处理取消
const handleCancel = () => {
  emit('update:visible', false);
  emit('close');
};
</script>

<style scoped>
.image-preview-modal {
  .ant-modal-content {
    border-radius: var(--border-radius-md);
  }
}

.preview-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  max-height: 500px;
  overflow-y: auto;
}

.preview-image-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--spacing-lg);
  background-color: var(--bg-light);
  border-radius: var(--border-radius-md);
  min-height: 300px;
  max-height: 450px;
  overflow: hidden;
}

.preview-image {
  max-width: 100%;
  max-height: 400px;
  border-radius: var(--border-radius-md);
}

.preview-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.preview-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.preview-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.preview-details p {
  margin: 0;
}

.preview-tags {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.tags-label {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-secondary);
  white-space: nowrap;
  margin-top: 2px;
}

.preview-tag {
  margin-bottom: var(--spacing-xs);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .preview-details {
    grid-template-columns: 1fr;
  }
  
  .preview-image {
    max-height: 300px;
  }
  
  .preview-image-container {
    min-height: 200px;
  }
}

@media (max-width: 480px) {
  .preview-image {
    max-height: 200px;
  }
  
  .preview-image-container {
    min-height: 150px;
  }
}
</style>
