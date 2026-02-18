<template>
  <div class="backup-view">
    <div class="backup-content card">
      <div class="backup-header">
        <h2 class="backup-title">{{ t('admin.backup.title') }}</h2>
        <div class="backup-actions">
          <a-button 
            type="primary" 
            @click="handleCreateBackup" 
            :loading="creatingBackup"
            class="create-backup-button"
          >
            {{ t('admin.backup.createBackup') }}
          </a-button>
          <a-upload
            :custom-request="handleUploadBackup"
            :show-upload-list="false"
            accept=".zip"
            :disabled="uploadingBackup"
          >
            <a-button 
              :type="'default'"
              :loading="uploadingBackup"
              class="upload-backup-button"
            >
              {{ t('admin.backup.uploadBackup') }}
            </a-button>
          </a-upload>
          <a-button 
            type="default" 
            @click="handleRefreshAll"
            :loading="refreshing"
            class="refresh-backup-button"
          >
            {{ t('admin.backup.refresh') }}
          </a-button>
        </div>
      </div>

      <!-- 上传进度 -->
      <div class="upload-progress-section" v-if="showProgress">
        <h3 class="section-title">{{ t('admin.backup.uploadProgress') }}</h3>
        <div class="progress-container">
          <div class="progress-header">
            <span class="progress-file-name">{{ progressStatus.fileName || t('admin.backup.backupFile') }}</span>
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

      <a-tabs v-model:activeKey="activeTab" type="card" size="large">
        <!-- 备份记录 -->
        <a-tab-pane key="backups" :tab="t('admin.backup.backupRecords')">
          <div class="tab-content">
            <a-spin :spinning="loading.backups" :tip="t('admin.backup.loadingBackups')">
              <a-table
                :columns="backupColumns"
                :data-source="backupList"
                :row-key="(record: { id: number }) => record.id"
                :pagination="false"
                :loading="loading.backups"
              >
                <template #size="{ record }">
                  {{ formatFileSize(record.size) }}
                </template>
                <template #start_time="{ record }">
                  {{ formatDateTime(record.start_time) }}
                </template>
                <template #end_time="{ record }">
                  {{ formatDateTime(record.end_time) }}
                </template>
                <template #error="{ record }">
                  {{ record.error || t('admin.backup.none') }}
                </template>
                <template #bodyCell="{ record, column }">
                  <template v-if="column.key === 'actions'">
                    <a-button 
                      type="link" 
                      @click="handleDownloadBackup(record.id)"
                      :loading="downloadingBackup === record.id"
                    >
                      {{ t('admin.backup.download') }}
                    </a-button>
                    <a-button 
                      type="link" 
                      @click="handleRestoreBackup(record.id)"
                      :loading="restoringBackup === record.id"
                    >
                      {{ t('admin.backup.restore') }}
                    </a-button>
                    <a-button 
                      type="link" 
                      danger 
                      @click="handleDeleteBackup(record.id)"
                      :loading="deletingBackup === record.id"
                    >
                      {{ t('admin.backup.delete') }}
                    </a-button>
                  </template>
                  <template v-else-if="column.key === 'status'">
                    <a-tag :color="getStatusColor(record.status)">
                      {{ getStatusText(record.status) }}
                    </a-tag>
                  </template>
                </template>
              </a-table>
              <div v-if="backupList.length === 0 && !loading.backups" class="empty-state">
                <p>{{ t('admin.backup.noBackupRecords') }}</p>
              </div>
            </a-spin>
          </div>
        </a-tab-pane>

        <!-- 恢复记录 -->
        <a-tab-pane key="restores" :tab="t('admin.backup.restoreRecords')">
          <div class="tab-content">
            <a-spin :spinning="loading.restores" :tip="t('admin.backup.loadingRestores')">
              <a-table
                :columns="restoreColumns"
                :data-source="restoreList"
                :row-key="(record: { id: number }) => record.id"
                :pagination="false"
                :loading="loading.restores"
              >
                <template #start_time="{ record }">
                  {{ formatDateTime(record.start_time) }}
                </template>
                <template #end_time="{ record }">
                  {{ formatDateTime(record.end_time) }}
                </template>
                <template #error="{ record }">
                  {{ record.error || t('admin.backup.none') }}
                </template>
                <template #bodyCell="{ record, column }">
                  <template v-if="column.key === 'status'">
                    <a-tag :color="getStatusColor(record.status)">
                      {{ getStatusText(record.status) }}
                    </a-tag>
                  </template>
                  <template v-else-if="column.key === 'backup_task_id'">
                    <a @click="handlePreviewBackup(record.backup_task)">
                      {{ record.backup_task_id }}
                    </a>
                  </template>
                  <template v-else-if="column.key === 'actions'">
                    <a-button 
                      type="link" 
                      danger 
                      @click="handleDeleteRestore(record.id)"
                      :loading="deletingRestore === record.id"
                    >
                      {{ t('admin.backup.delete') }}
                    </a-button>
                  </template>
                </template>
              </a-table>
              <div v-if="restoreList.length === 0 && !loading.restores" class="empty-state">
                <p>{{ t('admin.backup.noRestoreRecords') }}</p>
              </div>
            </a-spin>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>
  </div>

  <!-- 备份文件预览模态框 -->
  <a-modal
    v-model:open="showBackupPreviewModal"
    :title="t('admin.backup.backupFile')"
    width="50%"
    @cancel="handleCancelBackupPreview"
    @ok="handleCancelBackupPreview"
  >
    <a-descriptions :column="1" bordered size="small">
      <a-descriptions-item :label="t('admin.backup.backupId')">{{ backupPreview.id }}</a-descriptions-item>
      <a-descriptions-item :label="t('admin.backup.status')">{{ getStatusText(backupPreview.status) }}</a-descriptions-item>
      <a-descriptions-item :label="t('admin.backup.backupSize')">{{ formatFileSize(backupPreview.size) }}</a-descriptions-item>
     <a-descriptions-item :label="t('admin.backup.storagePath')">{{ backupPreview.storage_path }}</a-descriptions-item>
      <a-descriptions-item :label="t('admin.backup.startTime')">{{ formatDateTime(backupPreview.start_time) }}</a-descriptions-item>
      <a-descriptions-item :label="t('admin.backup.endTime')">{{ formatDateTime(backupPreview.end_time) }}</a-descriptions-item>
      <a-descriptions-item :label="t('admin.backup.errorMessage')" v-if="backupPreview.error">{{ backupPreview.error }}</a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { systemApi } from '../../api/services';
import type { BackupTask, RestoreTask } from '../../types/api';
import { getErrorMessage } from '../../types/errorMessages';
import { formatFileSize, formatDateTime} from '../../utils/index';
import { useI18n } from 'vue-i18n';
import { uploadWebSocketService } from '../../api/websocket';

const { t } = useI18n();

// 加载状态
const loading = reactive({
  backups: false,
  restores: false,
});

// 创建备份状态
const creatingBackup = ref(false);

// 恢复备份状态
const restoringBackup = ref<number | null>(null);

// 删除备份状态
const deletingBackup = ref<number | null>(null);

// 下载备份状态
const downloadingBackup = ref<number | null>(null);

// 上传备份状态
const uploadingBackup = ref(false);

// 刷新状态
const refreshing = ref(false);

// 激活的标签
const activeTab = ref('backups');

// 上传进度状态
const showProgress = ref(false);
const progressStatus = reactive({
  progress: 0,
  status: 'uploading' as 'uploading' | 'completed' | 'error',
  error: '',
  fileName: '',
});

// 备份列表
const backupList = ref<BackupTask[]>([]);

// 恢复列表
const restoreList = ref<RestoreTask[]>([]);

// 备份任务预览
const showBackupPreviewModal = ref(false);
const backupPreview = ref<BackupTask>({
  id: 0,
  status: '',
  size: 0,
  created_at: '',
  updated_at: '',
  start_time: '',
  end_time: '',
  error: '',
  storage_path: '',
});

// 备份列配置
const backupColumns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: t('admin.backup.status'),
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: t('admin.backup.backupSize'),
    dataIndex: 'size',
    key: 'size',
    slots: { customRender: 'size' },
  },
  {
    title: t('admin.backup.backupFile'),
    dataIndex: 'storage_path',
    key: 'storage_path',
  },
  {
    title: t('admin.backup.startTime'),
    dataIndex: 'start_time',
    key: 'start_time',
    slots: { customRender: 'start_time' },
  },
  {
    title: t('admin.backup.endTime'),
    dataIndex: 'end_time',
    key: 'end_time',
    slots: { customRender: 'end_time' },
  },
   {
    title: t('admin.backup.errorMessage'),
    dataIndex: 'error',
    key: 'error',
    slots: { customRender: 'error' },
  },
  {
    title: t('admin.backup.actions'),
    key: 'actions',
  },
];

// 恢复列配置
const restoreColumns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: t('admin.backup.restoreSource'),
    dataIndex: 'backup_task_id',
    key: 'backup_task_id',
  },
  {
    title: t('admin.backup.status'),
    dataIndex: 'status',
    key: 'status',
  },
  {
    title: t('admin.backup.startTime'),
    dataIndex: 'start_time',
    key: 'start_time',
    slots: { customRender: 'start_time' },
  },
  {
    title: t('admin.backup.endTime'),
    dataIndex: 'end_time',
    key: 'end_time',
    slots: { customRender: 'end_time' },
  },
  {
    title: t('admin.backup.errorMessage'),
    dataIndex: 'error',
    key: 'error',
    slots: { customRender: 'error' },
  },
  {
    title: t('admin.backup.actions'),
    key: 'actions',
    fixed: 'right' as const,
  },
];

// 删除备份状态
const deletingRestore = ref<number | null>(null);


// 获取状态颜色
const getStatusColor = (status: string): string => {
  switch (status) {
    case 'completed':
      return 'green';
    case 'failed':
      return 'red';
    case 'pending':
    case 'running':
      return 'blue';
    default:
      return 'default';
  }
};

// 获取状态文本
const getStatusText = (status: string): string => {
  switch (status) {
    case 'completed':
      return t('admin.backup.statusCompleted');
    case 'failed':
      return t('admin.backup.statusFailed');
    case 'pending':
      return t('admin.backup.statusPending');
    case 'running':
      return t('admin.backup.statusRunning');
    default:
      return status;
  }
};

// 获取进度状态类
const getProgressStatusClass = (status: string) => {
  switch (status) {
    case 'uploading':
      return 'status-uploading';
    case 'completed':
      return 'status-completed';
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
      return t('admin.backup.uploading');
    case 'completed':
      return t('admin.backup.statusCompleted');
    case 'error':
      return t('admin.backup.uploadFailed');
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
    default:
      return undefined;
  }
};

// 获取备份列表
const fetchBackups = async () => {
  try {
    loading.backups = true;
    const response = await systemApi.getBackups();
    // 检查响应结构
    if (response.data && Array.isArray(response.data)) {
      backupList.value = response.data;
    } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
      backupList.value = response.data.data;
    } else {
      backupList.value = [];
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.fetchBackupsFailed');
    message.error(errorMessage);
  } finally {
    loading.backups = false;
  }
};

// 获取恢复列表
const fetchRestores = async () => {
  try {
    loading.restores = true;
    const response = await systemApi.getRestoreTasks();
    // 检查响应结构
    if (response.data && Array.isArray(response.data)) {
      restoreList.value = response.data;
    } else if (response.data && response.data.data && Array.isArray(response.data.data)) {
      restoreList.value = response.data.data;
    } else {
      restoreList.value = [];
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.fetchRestoresFailed');
    message.error(errorMessage);
  } finally {
    loading.restores = false;
  }
};

// 创建备份
const handleCreateBackup = async () => {
  Modal.confirm({
    title: t('admin.backup.confirmCreateTitle'),
    content: t('admin.backup.confirmCreateContent'),
    okText: t('admin.backup.confirmCreate'),
    cancelText: t('admin.backup.cancel'),
    onOk: async () => {
      try {
        creatingBackup.value = true;
        await systemApi.createBackup();
        message.success(t('admin.backup.createBackupStarted'));
        // 刷新备份列表
        fetchBackups();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.createBackupFailed');
        message.error(errorMessage);
      } finally {
        creatingBackup.value = false;
      }
    },
  });
};

// 恢复备份
const handleRestoreBackup = (backupId: number) => {
  Modal.confirm({
    title: t('admin.backup.confirmRestoreTitle'),
    content: t('admin.backup.confirmRestoreContent', { id: backupId }),
    okText: t('admin.backup.confirmRestore'),
    okType: 'danger',
    cancelText: t('admin.backup.cancel'),
    onOk: async () => {
      try {
        restoringBackup.value = backupId;
        await systemApi.restoreBackup(backupId);
        message.success(t('admin.backup.restoreStarted'));
        // 刷新恢复列表
        fetchRestores();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.restoreBackupFailed');
        message.error(errorMessage);
      } finally {
        restoringBackup.value = null;
      }
    },
  });
};

// 删除备份
const handleDeleteBackup = (backupId: number) => {
  Modal.confirm({
    title: t('admin.backup.confirmDeleteTitle'),
    content: t('admin.backup.confirmDeleteContent', { id: backupId }),
    okText: t('admin.backup.confirmDelete'),
    okType: 'danger',
    cancelText: t('admin.backup.cancel'),
    onOk: async () => {
      try {
        deletingBackup.value = backupId;
        await systemApi.deleteBackup(backupId);
        message.success(t('admin.backup.deleteBackupSuccess'));
        // 刷新备份列表
        fetchBackups();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.deleteBackupFailed');
        message.error(errorMessage);
      } finally {
        deletingBackup.value = null;
      }
    },
  });
};

// 下载备份
const handleDownloadBackup = async (backupId: number) => {
  try {
    downloadingBackup.value = backupId;
    const response = await systemApi.downloadBackup(backupId);
    
    // 创建下载链接
    const url = window.URL.createObjectURL(new Blob([response.data as BlobPart]));
    const link = document.createElement('a');
    link.href = url;
    
    // 设置文件名
    const timestamp = new Date().getTime();
    link.setAttribute('download', `backup_${backupId}_${timestamp}.zip`);
    
    // 触发下载
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    
    message.success(t('admin.backup.downloadBackupSuccess'));
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.downloadBackupFailed');
    message.error(errorMessage);
  } finally {
    downloadingBackup.value = null;
  }
};

// 上传备份
const handleUploadBackup = async (options: any) => {
  const { file } = options;
  
  try {
    uploadingBackup.value = true;
    showProgress.value = true;
    
    // 初始化进度状态
    progressStatus.progress = 0;
    progressStatus.status = 'uploading';
    progressStatus.error = '';
    progressStatus.fileName = file.name || t('admin.backup.backupFile');
    
    // 创建 FormData 对象
    const formData = new FormData();
    formData.append('file', file);
    
    // 调用上传 API
    message.info(t('admin.backup.uploadStarted'));
    const response = await systemApi.uploadBackup(formData);
    
    // 调用成功回调
    if (options.onSuccess) {
      options.onSuccess(response);
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.uploadBackupFailed');
    message.error(errorMessage);
    
    // 标记为错误状态
    progressStatus.status = 'error';
    progressStatus.error = errorMessage;
    
    // 调用失败回调
    if (options.onError) {
      options.onError(error);
    }
  } finally {
    uploadingBackup.value = false;
  }
};

// 刷新所有数据
const handleRefreshAll = async () => {
  try {
    refreshing.value = true;
    
    // 并行刷新备份列表和恢复任务列表
    await Promise.all([
      fetchBackups(),
      fetchRestores()
    ]);
    
    message.success(t('admin.backup.refreshSuccess'));
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.refreshFailed');
    message.error(errorMessage);
  } finally {
    refreshing.value = false;
  }
};

// 删除恢复任务
const handleDeleteRestore = (restoreId: number) => {
  Modal.confirm({
    title: t('admin.backup.confirmDeleteTitle'),
    content: t('admin.backup.confirmDeleteRestoreContent', { id: restoreId }),
    okText: t('admin.backup.confirmDelete'),
    okType: 'danger',
    cancelText: t('admin.backup.cancel'),
    onOk: async () => {
      try {
        deletingRestore.value = restoreId;
        await systemApi.deleteRestoreTask(restoreId);
        message.success(t('admin.backup.deleteRestoreSuccess'));
        // 刷新恢复列表
        fetchRestores();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.backup.deleteRestoreFailed');
        message.error(errorMessage);
      } finally {
        deletingRestore.value = null;
      }
    },
  });
};

// 处理备份任务预览
const handlePreviewBackup = (backupTask: BackupTask) => {
  if (backupTask) {
    backupPreview.value = backupTask;
    showBackupPreviewModal.value = true;
  }
};

// 处理取消备份任务预览
const handleCancelBackupPreview = () => {
  showBackupPreviewModal.value = false;
  backupPreview.value = {
    id: 0,
    status: '',
    size: 0,
    created_at: '',
    updated_at: '',
    start_time: '',
    end_time: '',
    error: '',
    storage_path: '',
  };
};

// 监听标签切换，加载对应数据
watch(activeTab, (newTab) => {
  if (newTab === 'backups' && backupList.value.length === 0) {
    fetchBackups();
  } else if (newTab === 'restores' && restoreList.value.length === 0) {
    fetchRestores();
  }
});

// 处理上传进度
const handleUploadProgress = (data: { upload_id: string; file_name: string; progress: number; read: number; total: number }) => {
  if (data.upload_id === 'total') {
    progressStatus.progress = data.progress;
    // 进度到 100% 时标记为完成
    if (data.progress >= 100) {
      progressStatus.status = 'completed';
      message.success(t('admin.backup.uploadBackupSuccess'));
      // 刷新备份列表
      fetchBackups();
    }
  }
};

// 处理上传错误
const handleUploadError = (data: { upload_id: string; file_name: string; error: string }) => {
  progressStatus.status = 'error';
  progressStatus.error = data.error;
  message.error(`${t('admin.backup.uploadFailed')}: ${data.error}`);
};

// 初始化
onMounted(() => {
  fetchBackups();
  
  // 注册 WebSocket 事件监听器
  uploadWebSocketService.on('progress', handleUploadProgress);
  uploadWebSocketService.on('error', handleUploadError);
  
  // 建立 WebSocket 连接
  uploadWebSocketService.connect();
});

// 组件卸载时关闭 WebSocket 连接并移除事件监听器
onUnmounted(() => {
  uploadWebSocketService.off('progress', handleUploadProgress);
  uploadWebSocketService.off('error', handleUploadError);
  uploadWebSocketService.disconnect();
});
</script>

<style scoped>
.backup-view {
  width: 100%;
}

.backup-content {
  padding: var(--spacing-lg);
}

.backup-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-xl);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--border-color);
}

.backup-actions {
  display: flex;
  gap: var(--spacing-md);
  align-items: center;
}

.backup-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.create-backup-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.create-backup-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

/* 上传进度 */
.upload-progress-section {
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

.status-uploading {
  color: var(--primary-color);
}

.status-completed {
  color: var(--success-color);
}

.status-error {
  color: var(--error-color);
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

.tab-content {
  margin-top: var(--spacing-lg);
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  color: var(--text-secondary);
  background-color: var(--bg-light);
  border-radius: var(--border-radius-md);
  margin-top: var(--spacing-md);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .backup-view {
    width: 100%;
  }
  
  .backup-content {
    padding: var(--spacing-md);
  }
  
  .backup-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-md);
    padding-bottom: var(--spacing-md);
    margin-bottom: var(--spacing-md);
  }
  
  .backup-title {
    font-size: var(--font-size-lg);
  }
  
  .backup-actions {
    flex-wrap: wrap;
    width: 100%;
    gap: var(--spacing-sm);
  }
  
  .backup-actions .ant-btn {
    flex: 1;
    min-width: 100px;
  }
  
  .tab-content {
    padding: 0;
  }
  
  .empty-state {
    padding: var(--spacing-lg);
    font-size: var(--font-size-sm);
  }
  
  :deep(.ant-table-wrapper) {
    overflow-x: auto;
  }
  
  :deep(.ant-table) {
    min-width: 800px;
  }
  
  :deep(.ant-tabs-nav) {
    margin-bottom: var(--spacing-md);
  }
  
  :deep(.ant-descriptions) {
    font-size: var(--font-size-sm);
  }
}
</style>