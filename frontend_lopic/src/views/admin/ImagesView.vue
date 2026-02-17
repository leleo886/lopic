<template>
  <div class="images-view">
    <div class="images-content card">
      <!-- 搜索和筛选 -->
      <div class="search-filter">
        <div class="search-fields">
          <a-input-search
            v-model:value="searchKey"
            :placeholder="t('admin.images.searchPlaceholder')"
            class="search-input"
            @search="handleSearch"
          />
          <div class="field-filter">
            <a-select v-model:value="selectedField" class="field-select" style="width: 150px" :placeholder="t('admin.images.selectField')">
              <a-select-option value="storage_name">{{ t('admin.images.storageName') }}</a-select-option>
              <a-select-option value="user_id">{{ t('admin.images.userId') }}</a-select-option>
              <a-select-option value="album_id">{{ t('admin.images.albumId') }}</a-select-option>
            </a-select>
            <a-input
              v-model:value="fieldValue"
              class="field-value-input"
              :placeholder="t('admin.images.enterValue')"
              style="width: 150px"
              @pressEnter="handleSearch"
            />
            <a-button type="primary" @click="handleSearch" :disabled="!selectedField || !fieldValue">
              {{ t('admin.images.applyFilter') }}
            </a-button>
          </div>
        </div>
        <div class="filter-buttons">
          <div class="sort-controls">
            <span class="sort-label">{{ t('admin.images.sortLabel') }}</span>
            <a-select v-model:value="orderby" class="sort-select" style="width: 120px" @change="handleSortChange">
              <a-select-option value="created_at">{{ t('admin.images.sortOptions.createdAt') }}</a-select-option>
              <a-select-option value="updated_at">{{ t('admin.images.sortOptions.updatedAt') }}</a-select-option>
              <a-select-option value="file_size">{{ t('admin.images.sortOptions.fileSize') }}</a-select-option>
              <a-select-option value="file_name">{{ t('admin.images.sortOptions.fileName') }}</a-select-option>
            </a-select>
            <a-select v-model:value="order" class="sort-select" style="width: 80px" @change="handleSortChange">
              <a-select-option value="desc">{{ t('admin.images.sortOptions.desc') }}</a-select-option>
              <a-select-option value="asc">{{ t('admin.images.sortOptions.asc') }}</a-select-option>
            </a-select>
          </div>
          <a-button type="text" @click="resetFilters">{{ t('admin.images.resetFilters') }}</a-button>
        </div>
      </div>

      <!-- 批量操作栏 -->
      <div v-if="selectedRowKeys.length > 0" class="batch-operations">
        <span>{{selectedRowKeys.length }} {{ t('admin.images.selectedCount') }}</span>
        <a-button type="primary" @click="showUpdateStorageModal = true">
          {{ t('admin.images.batchUpdateStorage') }}
        </a-button>
        <a-button type="primary" danger @click="handleBatchDelete">
          <DeleteOutlined />
          {{ t('admin.images.batchDelete') }}
        </a-button>
      </div>
      
      <!-- 删除中状态 -->
      <div class="deleting-status" v-if="deleting">
        <a-spin size="small">
          <span>{{ t('admin.images.deleting') }}</span>
        </a-spin>
      </div>

      <!-- 图片列表 -->
      <a-table
        :columns="columns"
        :data-source="images"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
        :row-selection="{
          selectedRowKeys: selectedRowKeys,
          onChange: handleRowSelectionChange,
          selections: [
            {
              key: 'all-data',
              text: t('admin.images.selectAll'),
              onSelect: () => {
                selectedRowKeys = images.map(item => item.id);
              }
            },
            {
              key: 'clear',
              text: t('admin.images.clearSelection'),
              onSelect: () => {
                selectedRowKeys = [];
              }
            }
          ]
        }"
      >
        <template #image="{ record }">
          <a-avatar :src= "getFileUrl(record.thumbnail_url)" :alt="record.filename" size="large" />
        </template>
        <template #size="{ record }">
          <span>{{ formatFileSize(record.file_size) }}</span>
        </template>
        <template #user_id="{ record }">
          <a @click="handleViewUser(record.user_id)">{{ record.user_id }}</a>
        </template>
        <template #albums="{ record }">
          <span v-for="(album, index) in record.albums" :key="album.id">
            <a @click="handleViewAlbum(album.id)">{{ album.id }}</a>
            <span v-if="index < record.albums.length - 1">, </span>
          </span>
        </template>
        <template #storage_name="{ record }">
          <a @click="handleViewStorage(record.storage_name)">{{ record.storage_name }}</a>
        </template>
        <template #tags="{ record }">
          <span v-for="(tag, index) in record.tags" :key="tag.id">
            {{ tag }}
            <span v-if="index < record.tags.length - 1">, </span>
          </span>
        </template>
        <template #created_at="{ record }">
          <span>{{ formatDateTime(record.created_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" @click="handleView(record)">
            <ExpandOutlined />
          </a-button>
          <a-button type="text" size="small" danger @click="handleDelete(record.id)">
            <DeleteOutlined />
          </a-button>
        </template>
      </a-table>
    </div>

    <!-- 图片预览模态框 -->
    <ImagePreviewModal
      :visible="showPreviewModal"
      :image="previewImage"
      @update:visible="(value) => showPreviewModal = value"
      @close="handleCancelPreview"
    />

    <!-- 用户预览模态框 -->
    <a-modal
      v-model:open="showUserPreviewModal"
      :title="t('admin.images.userPreviewTitle')"
      width="40%"
      @cancel="handleCancelUserPreview"
      @ok="handleCancelUserPreview"
    >
        <a-descriptions :column="1" bordered size="small">
          <a-descriptions-item :label="t('admin.images.userId')">{{ userPreview.id }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.username')">{{ userPreview.username }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.email')">{{ userPreview.email }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.registerTime')">{{ formatDateTime(userPreview.created_at) }}</a-descriptions-item>
        </a-descriptions>
    </a-modal>

    <!-- 相册预览模态框 -->
    <a-modal
      v-model:open="showAlbumPreviewModal"
      :title="t('admin.images.albumPreviewTitle')"
      width="40%"
      @cancel="handleCancelAlbumPreview"
      @ok="handleCancelAlbumPreview"
    >
        <a-descriptions :column="1" bordered size="small">
          <a-descriptions-item :label="t('admin.images.albumName')">{{ album.name }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.albumDescription')">{{ album.description || t('admin.images.none') }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.createdAt')">{{ formatDateTime(album.created_at) }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.updatedAt')">{{ formatDateTime(album.updated_at) }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.coverImage')">
            <a v-if="album.cover_image" :href="album.cover_image" target="_blank">{{ t('admin.images.view') }}</a>
            <span v-else>{{ t('admin.images.none') }}</span>
          </a-descriptions-item>
          <a-descriptions-item :label="t('admin.images.imageCount')">{{ album.image_count }} {{ t('admin.images.images') }}</a-descriptions-item>
        </a-descriptions>
    </a-modal>

    <!-- 存储配置预览模态框 -->
    <a-modal
      v-model:open="showStoragePreviewModal"
      :title="t('admin.images.storagePreviewTitle')"
      width="40%"
      @cancel="handleCancelStoragePreview"
      @ok="handleCancelStoragePreview"
    >
        <a-descriptions :column="1" bordered size="small">
          <a-descriptions-item :label="t('storages.storageName')">{{ storage.name }}</a-descriptions-item>
          <a-descriptions-item :label="t('storages.storageType')">{{ storage.type }}</a-descriptions-item>
          <a-descriptions-item :label="t('storages.storageConfig')">
            <pre>{{ JSON.stringify(storage.config, null, 2) }}</pre>
          </a-descriptions-item>
        </a-descriptions>
    </a-modal>

    <!-- 批量更新存储名称模态框 -->
    <a-modal
      v-model:open="showUpdateStorageModal"
      :title="t('admin.images.batchUpdateStorageTitle')"
      width="40%"
      @cancel="handleCancelUpdateStorage"
      @ok="handleBatchUpdateStorage"
    >
      <a-form>
        <a-form-item :label="t('admin.images.storageName')">
          <a-select
            v-model:value="updateStorageForm.storage_name"
            :placeholder="t('admin.images.selectStoragePlaceholder')"
            style="width: 100%"
          >
            <a-select-option
              v-for="storage in storages"
              :key="storage.name"
              :value="storage.name"
            >
              {{ storage.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <div class="batch-update-info">
          {{ t('admin.images.updatingImages', { count: selectedRowKeys.length }) }}
        </div>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { imageApi, userApi, storageApi } from '../../api/services';
import { uploadWebSocketService } from '../../api/websocket';
import type { ImageResponse, AlbumResponse, User, Storage } from '../../types/api';
import { formatDateTime, formatFileSize, getFileUrl } from '../../utils/index';
import { DeleteOutlined, ExpandOutlined } from '@ant-design/icons-vue';
import { getErrorMessage } from '../../types/errorMessages';
import ImagePreviewModal from '../../components/ImagePreviewModal.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);
// 删除中状态
const deleting = ref(false);

// 分页配置
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showTotal: (total: number) => t('pagination.total', { total }),
});

// 选中的行 keys
const selectedRowKeys = ref<number[]>([]);

// 图片列表
const images = ref<ImageResponse[]>([]);

// 搜索关键词
const searchKey = ref('');

// 字段筛选
const selectedField = ref('');
const fieldValue = ref('');

// 排序参数
const orderby = ref('created_at');
const order = ref('desc');

// 图片预览
const showPreviewModal = ref(false);
const previewImageUrl = ref('');
const previewImage = ref<ImageResponse | null>(null);

// 用户预览
const showUserPreviewModal = ref(false);
const userPreview = ref<User>({} as User);

// 相册预览
const showAlbumPreviewModal = ref(false);
const album = ref<AlbumResponse>({} as AlbumResponse);

// 存储配置预览
const showStoragePreviewModal = ref(false);
const storage = ref<Storage>({} as Storage);

// 批量更新存储名称
const showUpdateStorageModal = ref(false);
const updateStorageForm = ref({
  storage_name: '',
});
const storages = ref<Storage[]>([]);

// 表格列配置
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: t('admin.images.image'),
    key: 'image',
    slots: { customRender: 'image' },
  },
  {
    title: t('admin.images.filename'),
    dataIndex: 'original_name',
    key: 'original_name',
    ellipsis: true,
  },
  {
    title: t('admin.images.fileSize'),
    key: 'size',
    slots: { customRender: 'size' },
  },
  {
    title: t('admin.images.mimeType'),
    dataIndex: 'mime_type',
    key: 'mime_type',
  },
  {
    title: t('admin.images.user'),
    key: 'user_id',
    slots: { customRender: 'user_id' },
  },
  {
    title: t('admin.images.tags'),
    key: 'tags',
    slots: { customRender: 'tags' },
    ellipsis: true, 
    width: 150,
  },
  {
    title: t('admin.images.albums'),
    key: 'albums',
    slots: { customRender: 'albums' },
  },
  {
    title: t('admin.images.storageName'),
    key: 'storage_name',
    slots: { customRender: 'storage_name' },
  },
  {
    title: t('admin.images.createdAt'),
    key: 'created_at',
    slots: { customRender: 'created_at' },
  },
  {
    title: t('admin.images.actions'),
    key: 'actions',
    slots: { customRender: 'actions' },
  },
];

// 获取图片列表
const fetchImages = async () => {
  try {
    loading.value = true;
    const response = await imageApi.getImages({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      searchkey: searchKey.value,
      orderby: orderby.value,
      order: order.value,
      field: selectedField.value,
      value: fieldValue.value,
    });
    images.value = response.data.images;
    pagination.value.total = response.data.total;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.fetchImagesFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理搜索
const handleSearch = (value: string | Event) => {
  // 检查 value 是否为事件对象
  if (value instanceof Event) {
    // 事件对象，不更新 searchKey
  } else {
    // 字符串值，更新 searchKey
    searchKey.value = value;
  }
  pagination.value.current = 1;
  fetchImages();
};

// 处理排序变化
const handleSortChange = () => {
  pagination.value.current = 1;
  fetchImages();
};

// 重置筛选
const resetFilters = () => {
  searchKey.value = '';
  selectedField.value = '';
  fieldValue.value = '';
  orderby.value = 'created_at';
  order.value = 'desc';
  pagination.value.current = 1;
  fetchImages();
};

// 处理表格变化
const handleTableChange = (newPagination: any) => {
  pagination.value = newPagination;
  // 切换页时清空选择
  selectedRowKeys.value = [];
  fetchImages();
};

// 处理行选择变化
const handleRowSelectionChange = (selectedKeys: number[]) => {
  selectedRowKeys.value = selectedKeys;
};

// 处理批量删除
const handleBatchDelete = () => {
  if (selectedRowKeys.value.length === 0) return;
  
  Modal.confirm({
    title: t('admin.images.confirmBatchDeleteTitle'),
    content: t('admin.images.confirmBatchDeleteContent', { count: selectedRowKeys.value.length }),
    onOk: async () => {
      try {
        deleting.value = true;
        imageApi.deleteImages(selectedRowKeys.value);
        message.warning(t('admin.images.batchDeleteStarted'));
        selectedRowKeys.value = [];
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.batchDeleteFailed');
        message.error(errorMessage);
        deleting.value = false;
      }
    },
  });
};

// 处理查看用户
const handleViewUser = async (id: number) => {
  try {
    const user = await userApi.getUser(id);
    if (user) {
      userPreview.value = user.data || {} as User;
      showUserPreviewModal.value = true;
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.fetchUserFailed');
    message.error(errorMessage);
  }
};

// 处理取消用户预览
const handleCancelUserPreview = () => {
  showUserPreviewModal.value = false;
  userPreview.value = {} as User;
};

// 处理查看相册
const handleViewAlbum = (id: number) => {
  const calbum = images.value.find((item) => item.albums?.some((album) => album.id === id));
  if (calbum) {
    album.value = calbum.albums?.find((item) => item.id === id) || {} as AlbumResponse;
    showAlbumPreviewModal.value = true;
  }
};

// 处理取消相册预览
const handleCancelAlbumPreview = () => {
  showAlbumPreviewModal.value = false;
  album.value = {} as AlbumResponse;
};

// 处理查看存储配置
const handleViewStorage = async (name: string) => {
  try {
    const response = await storageApi.getStorageByName(name);
    if (response) {
      storage.value = response.data || {} as Storage;
      showStoragePreviewModal.value = true;
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.fetchStorageFailed');
    message.error(errorMessage);
  }
};

// 处理取消存储配置预览
const handleCancelStoragePreview = () => {
  showStoragePreviewModal.value = false;
  storage.value = {} as Storage;
};

// 获取存储配置列表
const fetchStorages = async () => {
  try {
    const response = await storageApi.getStorages({ page: 1, page_size: 100 });
    if (response) {
      storages.value = response.data;
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.fetchStoragesFailed');
    message.error(errorMessage);
  }
};

// 处理批量更新存储名称
const handleBatchUpdateStorage = async () => {
  if (selectedRowKeys.value.length === 0) return;
  if (!updateStorageForm.value.storage_name) {
    message.error(t('admin.images.selectStorageRequired'));
    return;
  }

  try {
    await imageApi.updateImageStorage({
      ids: selectedRowKeys.value,
      storage_name: updateStorageForm.value.storage_name,
    });
    message.success(t('admin.images.batchUpdateStorageSuccess'));
    selectedRowKeys.value = [];
    showUpdateStorageModal.value = false;
    updateStorageForm.value.storage_name = '';
    fetchImages();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.batchUpdateStorageFailed');
    message.error(errorMessage);
  }
};

// 处理取消更新存储名称
const handleCancelUpdateStorage = () => {
  showUpdateStorageModal.value = false;
  updateStorageForm.value.storage_name = '';
};

// 处理查看图片
const handleView = (image: ImageResponse) => {
  previewImage.value = image;
  showPreviewModal.value = true;
};

// 处理删除图片
const handleDelete = (id: number) => {
  Modal.confirm({
    title: t('admin.images.confirmDeleteTitle'),
    content: t('admin.images.confirmDeleteContent'),
    onOk: async () => {
      try {
        deleting.value = true;
        imageApi.deleteImages([id]);
        message.warning(t('admin.images.deleteStarted'));
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.images.deleteImageFailed');
        message.error(errorMessage);
        deleting.value = false;
      }
    },
  });
};

// 处理取消预览
const handleCancelPreview = () => {
  showPreviewModal.value = false;
  previewImageUrl.value = '';
  previewImage.value = null;
};

// 初始化
onMounted(() => {
  fetchImages();
  fetchStorages();
  
  // 连接 WebSocket
  uploadWebSocketService.connect();
  
  // 注册删除成功监听器
  uploadWebSocketService.on('deleteSuccess', () => {
    deleting.value = false;
    message.success(t('admin.images.deleteImageSuccess'));
    fetchImages();
  });
  
  // 注册删除错误监听器
  uploadWebSocketService.on('deleteError', (data: { message: string; error: string; code: string }) => {
    deleting.value = false;
    message.error(t('admin.images.deleteImageFailed') + `: ${getErrorMessage(data.code)})`);
    fetchImages();
  });
});

// 组件卸载
onUnmounted(() => {
  // 断开 WebSocket 连接
  uploadWebSocketService.disconnect();
});
</script>

<style scoped>
.images-view {
  width: 100%;
}

.images-content {
  padding: var(--spacing-lg);
}

.search-filter {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-lg);
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.search-fields {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.search-input {
  width: 300px;
}

.field-filter {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.field-select {
  margin-right: var(--spacing-xs);
}

.field-value-input {
  margin-right: var(--spacing-xs);
}

.batch-operations {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  margin-bottom: 20px;
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.batch-operations span {
  font-size: 14px;
  color: #666;
  font-weight: 500;
}

.deleting-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  margin-bottom: 20px;
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  font-size: 14px;
}

.filter-buttons {
  display: flex;
  gap: var(--spacing-sm);
  align-items: center;
}

.sort-controls {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.sort-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.sort-select {
  margin-right: var(--spacing-xs);
}

.preview-image {
  max-width: 100%;
  max-height: 50vh;
  object-fit: contain;
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-md);
  margin-bottom: var(--spacing-lg);
}

/* 预览模态框 */
.image-preview-modal .ant-modal-content {
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-xl);
}

.preview-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: var(--spacing-lg);
}

.preview-info {
  width: 100%;
  text-align: center;
}

.preview-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-md);
}

.preview-meta {
  display: flex;
  justify-content: center;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  flex-wrap: wrap;
}

.preview-tags {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  margin-top: var(--spacing-md);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .images-view {
    width: 100%;
  }
  
  .images-content {
    padding: var(--spacing-md);
  }
  
  .search-filter {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
  }
  
  .search-fields {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .search-input {
    width: 100%;
  }
  
  .field-filter {
    flex-wrap: wrap;
    gap: var(--spacing-xs);
  }
  
  .field-select {
    width: 100% !important;
  }
  
  .field-value-input {
    flex: 1;
    min-width: 100px;
  }
  
  .filter-buttons {
    flex-wrap: wrap;
    justify-content: flex-start;
    gap: var(--spacing-xs);
  }
  
  .sort-controls {
    flex-wrap: wrap;
  }
  
  .sort-select {
    width: 100px !important;
  }
  
  .batch-operations {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-sm);
    padding: var(--spacing-md);
  }
  
  .batch-operations span {
    text-align: center;
  }
  
  .batch-operations .ant-btn {
    width: 100%;
  }
  
  .deleting-status {
    padding: var(--spacing-sm);
    font-size: var(--font-size-sm);
  }
  
  :deep(.ant-table-wrapper) {
    overflow-x: auto;
  }
  
  :deep(.ant-table) {
    min-width: 1000px;
  }
  
  :deep(.ant-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
  
  :deep(.ant-descriptions) {
    font-size: var(--font-size-sm);
  }
  
  .batch-update-info {
    font-size: var(--font-size-sm);
  }
}
</style>
