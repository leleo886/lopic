<template>
  <div class="upload-container">
    <div class="upload-header card">
      <h2 class="upload-title">{{ t('upload.title') }}</h2>
      <p class="upload-description">{{ t('upload.description') }}</p>
      <p class="upload-description">{{ t('upload.multipleAlbums') }}</p>
    </div>

    <div class="upload-content" v-if="!loading">
      <div class="upload-main card">
        <div class="upload-section">
          <h3 class="section-title">{{ t('upload.selectFiles') }}</h3>
          <a-upload
            action=""
            :multiple="true"
            :custom-request="handleCustomUpload"
            :file-list="fileList"
            :before-upload="beforeUpload"
            :on-remove="handleRemoveFile"
            class="upload-component"
          >
            <a-button type="primary" class="select-button">
              <template #icon>
                <upload-outlined />
              </template>
              {{ t('upload.selectFiles') }}
            </a-button>
          </a-upload>
        </div>

        <!-- 上传进度 -->
        <div class="upload-progress-section" v-if="showProgress">
          <h3 class="section-title">{{ t('upload.uploadProgress') }}</h3>
          <div class="progress-container">
            <div class="progress-header">
              <span class="progress-file-name">{{ progressStatus.fileCount }} {{ t('upload.files') }}</span>
              <span class="progress-status" :class="getProgressStatusClass(progressStatus.status)">
                {{ getProgressStatusText(progressStatus.status) }}
              </span>
            </div>
            <div class="progress-bar-container">
              <a-progress 
                :percent="Math.round(progressStatus.progress || 0)" 
                :status="getProgressStatus(progressStatus.status)"
                :format="() => `${Math.round(progressStatus.progress || 0)}%`"
              />
            </div>
            <div class="progress-footer" v-if="progressStatus.error">
              <span class="progress-error">{{ progressStatus.error }}</span>
            </div>
          </div>
        </div>

        <div class="upload-form">
          <h3 class="section-title">{{ t('upload.uploadSettings') }}</h3>
          <a-form
            :model="uploadForm"
            :rules="uploadRules"
          >
            <a-form-item :label="t('upload.tags')">
              <a-input
                v-model:value="uploadForm.tags"
                :placeholder="t('upload.enterTags')"
              />
            </a-form-item>
            <a-form-item :label="t('upload.albums')">
              <a-select
                v-model:value="uploadForm.album_ids"
                mode="multiple"
                :placeholder="t('upload.selectAlbumsPlaceholder')"
              >
                <a-select-option
                  v-for="album in userAlbums"
                  :key="album.id"
                  :value="album.id"
                >
                  {{ album.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-form>
        </div>

        <div class="upload-actions">
          <a-button type="primary" @click="confirmUpload" class="confirm-button" :loading="uploading">
            {{ uploading ? t('upload.uploading') : t('upload.confirmUpload') }}
          </a-button>
          <a-button type="default" @click="resetForm" class="reset-button">
            {{ t('upload.reset') }}
          </a-button>
        </div>
      </div>

      <div class="upload-limits card">
        <h3 class="section-title">{{ t('upload.uploadLimits') }}</h3>
        <div class="limits-list">
          <div class="limit-item">
            <span class="limit-label">{{ t('upload.maxFileSize') }}：</span>
            <span class="limit-value">{{ maxFileSize == -1 ? t('upload.unlimited') : maxFileSize + 'MB' }}</span>
          </div>
          <div class="limit-item">
            <span class="limit-label">{{ t('upload.filesPerUpload') }}：</span>
            <span class="limit-value">{{ maxFilesPerUpload == -1 ? t('upload.unlimited') : maxFilesPerUpload }}</span>
          </div>
          <div class="limit-item">
            <span class="limit-label">{{ t('upload.storageLimit') }}：</span>
            <span class="limit-value">{{ maxStorageSize == -1 ? t('upload.unlimited') : maxStorageSize + 'MB' }}</span>
          </div>
          <div class="limit-item">
            <span class="limit-label">{{ t('upload.albumsLimit') }}：</span>
            <span class="limit-value">{{ maxAlbumsPerUser == -1 ? t('upload.unlimited') : maxAlbumsPerUser }}</span>
          </div>
          <div class="limit-item">
            <span class="limit-label">{{ t('upload.allowedExtensions') }}：</span>
            <span class="limit-value">{{ allowedExtensions.join(', ') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div class="loading-container" v-else>
      <a-spin :tip="t('upload.loading')">
        <div class="loading-content"></div>
      </a-spin>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue';
import { message } from 'ant-design-vue';
import { UploadOutlined } from '@ant-design/icons-vue';
import { imageApi, albumApi, userApi } from '../api/services';
import { uploadWebSocketService } from '../api/websocket';
import type { AlbumResponse, User } from '../types/api';
import { getErrorMessage } from '../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();


// 加载状态
const loading = ref(true);
// 用户数据
const currentUser = ref<User | null>(null);
// 用户相册列表
const userAlbums = ref<AlbumResponse[]>([]);
// 文件列表
const fileList = ref<any[]>([]);
// 上传状态
const uploading = ref(false);

// 上传进度状态
const showProgress = ref(false);
const progressStatus = reactive({
  progress: 0,
  status: 'uploading' as 'uploading' | 'processing' | 'completed' | 'error',
  error: '',
  fileCount: 0,
});

// 上传表单
const uploadForm = reactive({
  tags: '',
  album_ids: [] as number[],
});

// 表单验证规则
const uploadRules = {
  tags: [],
  album_ids: [],
};

// 从用户角色获取上传限制
const maxFileSize = computed(() => currentUser.value?.role.max_file_size_mb || 10);
const maxFilesPerUpload = computed(() => currentUser.value?.role.max_files_per_upload || 10);
const maxStorageSize = computed(() => currentUser.value?.role.max_storage_size_mb || 1024);
const maxAlbumsPerUser = computed(() => currentUser.value?.role.max_albums_per_user || 50);
const allowedExtensions = computed(() => currentUser.value?.role.allowed_extensions || ['jpg', 'jpeg', 'png', 'gif']);

// 获取当前用户信息
const fetchCurrentUser = async () => {
  try {
    const response = await userApi.getCurrentUser();
    currentUser.value = response.data;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('upload.fetchUserFailed');
    message.error(errorMessage);
  }
};

// 获取用户相册
const fetchUserAlbums = async () => {
  try {
    const response = await albumApi.getUserAlbums({ page: 1, page_size: 100 });
    userAlbums.value = response.data.albums;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('upload.fetchAlbumsFailed');
    message.error(errorMessage);
  }
};

// 处理上传前验证
const beforeUpload = (file: any) => {
  // path allowedExtensions
  const fileName = file.name.toLowerCase();
  const ext = fileName.split('.').pop(); // 获取扩展名
  const isAllowed = allowedExtensions.value.includes("." + ext);
  if (!isAllowed) {
    message.error(t('upload.invalidFileExtension', { extensions: allowedExtensions.value.join(', ') }));
    return false;
  }

  const isImage = file.type.startsWith('image/');
  if (!isImage) {
    message.error(t('upload.onlyImagesAllowed'));
    return false;
  }
  const isLtMaxSize = file.size / 1024 / 1024 < maxFileSize.value;
  if (!isLtMaxSize && maxFileSize.value !== -1) {
    message.error(t('upload.fileSizeExceeded', { size: maxFileSize.value }));
    return false;
  }
  if (fileList.value.length >= maxFilesPerUpload.value && maxFilesPerUpload.value !== -1) {
    message.error(t('upload.maxFilesExceeded', { count: maxFilesPerUpload.value }));
    return false;
  }
  fileList.value = [...fileList.value, file];
  return false; // 阻止自动上传，使用自定义上传
};

// 处理移除文件
const handleRemoveFile = (file: any) => {
  const index = fileList.value.indexOf(file);
  if (index > -1) {
    fileList.value.splice(index, 1);
  }
};

// 处理自定义上传
const handleCustomUpload = () => {
  // 自定义上传由 confirmUpload 处理
};

// 确认上传
const confirmUpload = async () => {
  if (fileList.value.length === 0) {
    message.error(t('upload.selectFilesFirst'));
    return;
  }

  try {
    uploading.value = true;
    showProgress.value = true;
    
    // 初始化进度状态
    progressStatus.progress = 0;
    progressStatus.status = 'uploading';
    progressStatus.error = '';
    progressStatus.fileCount = fileList.value.length;

    // 准备 FormData
    const formData = new FormData();
    fileList.value.forEach(file => {
      formData.append('file', file);
    });

    // 添加标签
    if (uploadForm.tags) {
      // 对Tags进行去重
      const uniqueTags = Array.from(new Set(uploadForm.tags.split(',')));
      const tags = uniqueTags.map(tag => tag.trim()).filter(Boolean);
      tags.forEach(tag => {
        formData.append('tags', tag);
      });
    }

    // 添加相册ID列表
    if (uploadForm.album_ids && uploadForm.album_ids.length > 0) {
      uploadForm.album_ids.forEach(albumId => {
        formData.append('album_ids', albumId.toString());
      });
    }

    // 调用上传API
    message.success(t('upload.uploadStarted'));
    await imageApi.uploadImages(formData);
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('upload.uploadFailed');
    message.error(errorMessage);
    
    // 标记为错误状态
    progressStatus.status = 'error';
    progressStatus.error = errorMessage;
  } finally {
    uploading.value = false;
  }
};

// 重置表单
const resetForm = () => {
  fileList.value = [];
  uploadForm.tags = '';
  uploadForm.album_ids = [];
  showProgress.value = false;
  progressStatus.progress = 0;
  progressStatus.status = 'uploading';
  progressStatus.error = '';
  progressStatus.fileCount = 0;
};

// 处理上传开始
const handleUploadStart = (_data: { upload_id: string; file_name: string; file_size: number }) => {
  // 上传开始，保持当前状态
};

// 处理上传进度
const handleUploadProgress = (data: { upload_id: string; file_name: string; progress: number; read: number; total: number }) => {
  // 更新总进度
  if (data.upload_id === 'total') {
    progressStatus.progress = data.progress;
  }
};

// 处理上传错误
const handleUploadError = (data: { upload_id: string; file_name: string; error: string }) => {
  progressStatus.status = 'error';
  progressStatus.error = `${data.file_name}: ${data.error}`;
  message.error(`${t('upload.uploadFailed')}: ${data.file_name} - ${data.error}`);
};

// 处理上传完成
const handleUploadComplete = (_data: { upload_id: string; file_name: string; file_url: string; thumbnail_url: string; image_id: number }) => {
  // 单个文件上传完成，更新进度为100%
  progressStatus.progress = 100;
  progressStatus.status = 'processing';
};

// 处理内部处理开始
const handleProcessingStart = (data: { message: string; file_count: number }) => {
  console.log(data.message);
  progressStatus.status = 'processing';
  progressStatus.progress = 100;
};

// 处理内部处理错误
const handleProcessingError = (data: { message: string; error: string, code: string }) => {
  progressStatus.status = 'error';
  progressStatus.error = getErrorMessage(data.code);
  message.error(`${t('upload.processingFailed')}: ${getErrorMessage(data.code)}`);
};

// 处理内部处理完成
const handleProcessingComplete = (data: { message: string; file_count: number }) => {
  progressStatus.status = 'completed';
  progressStatus.progress = 100;
  message.success(`${t('upload.uploadSuccess')} (${data.file_count} ${t('upload.files')})`);
};

// 获取进度状态类
const getProgressStatusClass = (status: string) => {
  switch (status) {
    case 'uploading':
      return 'status-uploading';
    case 'completed':
      return 'status-completed';
    case 'processing':
      return 'status-processing';
    case 'error':
      return 'status-error';
    default:
      return '';
  }
};

// 获取进度状态文本
const getProgressStatusText = (status: string) => {
  switch (status) {
    case 'uploading':
      return t('upload.uploading');
    case 'completed':
      return t('upload.completed');
    case 'processing':
      return t('upload.processing');
    case 'error':
      return t('upload.uploadFailed');
    default:
      return status;
  }
};

// 获取进度条状态
const getProgressStatus = (status: string) => {
  switch (status) {
    case 'completed':
      return 'success';
    case 'error':
      return 'exception';
    case 'processing':
      return 'active';
    default:
      return undefined;
  }
};

// 初始化
onMounted(async () => {
  try {
    loading.value = true;
    // 并行获取用户信息和相册列表
    await Promise.all([
      fetchCurrentUser(),
      fetchUserAlbums()
    ]);
    
    // 注册 WebSocket 事件监听器
    uploadWebSocketService.on('start', handleUploadStart);
    uploadWebSocketService.on('progress', handleUploadProgress);
    uploadWebSocketService.on('error', handleUploadError);
    uploadWebSocketService.on('complete', handleUploadComplete);
    uploadWebSocketService.on('processingStart', handleProcessingStart);
    uploadWebSocketService.on('processingError', handleProcessingError);
    uploadWebSocketService.on('processingComplete', handleProcessingComplete);
    
    // 建立 WebSocket 连接
    uploadWebSocketService.connect();
  } catch (error) {
    console.error('Failed to initialize:', error);
  } finally {
    loading.value = false;
  }
});

// 组件卸载时关闭 WebSocket 连接并移除事件监听器
onUnmounted(() => {
  uploadWebSocketService.off('start', handleUploadStart);
  uploadWebSocketService.off('progress', handleUploadProgress);
  uploadWebSocketService.off('error', handleUploadError);
  uploadWebSocketService.off('complete', handleUploadComplete);
  uploadWebSocketService.off('processingStart', handleProcessingStart);
  uploadWebSocketService.off('processingError', handleProcessingError);
  uploadWebSocketService.off('processingComplete', handleProcessingComplete);
  uploadWebSocketService.disconnect();
});
</script>

<style scoped>
/* 样式部分保持不变 */
.upload-container {
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 上传头部 */
.upload-header {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-xl);
  margin-bottom: var(--spacing-xl);
  border-radius: var(--border-radius-lg);
}

.upload-title {
  font-size: var(--font-size-2xl);
  font-weight: 600;
  color: var(--primary-color);
  margin: 0 0 var(--spacing-sm) 0;
}

.upload-description {
  font-size: var(--font-size-base);
  color: var(--text-secondary);
  margin: 0;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 500px;
}

.loading-content {
  width: 200px;
  height: 200px;
}

/* 上传内容 */
.upload-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
}

/* 上传主区域 */
.upload-main {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-xl);
  border-radius: var(--border-radius-lg);
}

/* 上传部分 */
.upload-section {
  margin-bottom: var(--spacing-xl);
}

.section-title {
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-primary);
  margin: 0 0 var(--spacing-lg) 0;
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--border-color);
}

.select-button {
  margin-bottom: var(--spacing-md);
}

.upload-hint {
  margin-top: var(--spacing-sm);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  line-height: 1.5;
}

/* 上传表单 */
.upload-form {
  margin-bottom: var(--spacing-xl);
}

/* 上传操作 */
.upload-actions {
  display: flex;
  gap: var(--spacing-md);
  justify-content: flex-end;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--border-color);
}

.confirm-button {
  min-width: 120px;
}

/* 上传进度 */
.upload-progress-section {
  margin-bottom: var(--spacing-xl);
}

.progress-container {
  padding: var(--spacing-md);
  background-color: var(--bg-light);
  border-radius: var(--border-radius-md);
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xs);
}

.progress-file-name {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
}

.progress-status {
  font-size: var(--font-size-xs);
  color: var(--text-secondary);
}

.status-processing {
  color: var(--warning-color);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
  100% {
    opacity: 1;
  }
}

.progress-bar-container {
  margin-bottom: var(--spacing-xs);
}

.progress-footer {
  margin-top: var(--spacing-xs);
}

.progress-error {
  font-size: var(--font-size-xs);
  color: var(--error-color);
}

/* 上传限制 */
.upload-limits {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-xl);
  border-radius: var(--border-radius-lg);
  height: fit-content;
}

.limits-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.limit-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.limit-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  font-weight: 500;
}

.limit-value {
  font-size: var(--font-size-sm);
  color: var(--text-primary);
  background-color: var(--bg-light);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius-sm);
  border: 1px solid var(--border-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .upload-container {
    min-height: auto;
  }
  
  .upload-header {
    padding: var(--spacing-md);
    margin-bottom: var(--spacing-md);
  }
  
  .upload-title {
    font-size: var(--font-size-xl);
  }
  
  .upload-description {
    font-size: var(--font-size-sm);
  }
  
  .upload-content {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
    padding: 0 var(--spacing-md);
  }
  
  .upload-main {
    padding: var(--spacing-md);
  }
  
  .upload-section {
    margin-bottom: var(--spacing-md);
  }
  
  .section-title {
    font-size: var(--font-size-base);
    margin-bottom: var(--spacing-md);
  }
  
  .upload-form {
    margin-bottom: var(--spacing-md);
  }
  
  .upload-actions {
    flex-direction: column;
    gap: var(--spacing-sm);
    padding-top: var(--spacing-md);
  }
  
  .upload-actions .ant-btn {
    width: 100%;
  }
  
  .confirm-button {
    min-width: auto;
    order: -1;
  }
  
  .upload-progress-section {
    margin-bottom: var(--spacing-md);
  }
  
  .progress-list {
    max-height: 150px;
  }
  
  .progress-file-name {
    font-size: var(--font-size-xs);
  }
  
  .upload-limits {
    padding: var(--spacing-md);
  }
  
  .limits-list {
    gap: var(--spacing-sm);
  }
  
  .limit-label {
    font-size: var(--font-size-xs);
  }
  
  .limit-value {
    font-size: var(--font-size-xs);
  }
  
  .loading-container {
    min-height: 300px;
  }
}
</style>