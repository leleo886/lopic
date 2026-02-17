<template>
  <div class="albums-view">
    <div class="albums-content card">
      <!-- 搜索和筛选 -->
      <div class="search-filter">
        <a-input-search
          :placeholder="t('admin.albums.searchPlaceholder')"
          class="search-input"
          @search="handleSearch"
        />
        <div class="filter-buttons">
          <div class="sort-controls">
            <span class="sort-label">{{ t('admin.albums.sortLabel') }}</span>
            <a-select v-model:value="orderby" class="sort-select" style="width: 120px" @change="handleSortChange">
              <a-select-option value="created_at">{{ t('admin.albums.sortOptions.createdAt') }}</a-select-option>
              <a-select-option value="updated_at">{{ t('admin.albums.sortOptions.updatedAt') }}</a-select-option>
              <a-select-option value="name">{{ t('admin.albums.sortOptions.name') }}</a-select-option>
              <a-select-option value="image_count">{{ t('admin.albums.sortOptions.imageCount') }}</a-select-option>
            </a-select>
            <a-select v-model:value="order" class="sort-select" style="width: 80px" @change="handleSortChange">
              <a-select-option value="desc">{{ t('admin.albums.sortOptions.desc') }}</a-select-option>
              <a-select-option value="asc">{{ t('admin.albums.sortOptions.asc') }}</a-select-option>
            </a-select>
          </div>
          <a-button type="text" @click="resetFilters">{{ t('admin.albums.resetFilters') }}</a-button>
        </div>
      </div>

       <!-- 批量操作栏 -->
      <div v-if="selectedRowKeys.length > 0" class="batch-operations">
        <span>{{ selectedRowKeys.length }} {{ t('admin.albums.selectedCount') }}</span>
        <a-button type="primary" danger @click="handleBatchDelete">
          <DeleteOutlined />
          {{ t('admin.albums.batchDelete') }}
        </a-button>
      </div>

      <!-- 相册列表 -->
      <a-table
        :columns="columns"
        :data-source="albums"
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
              text: t('admin.albums.selectAll'),
              onSelect: () => {
                selectedRowKeys = albums.map(item => item.id);
              }
            },
            {
              key: 'clear',
              text: t('admin.albums.clearSelection'),
              onSelect: () => {
                selectedRowKeys = [];
              }
            }
          ]
        }"
      >
        <template #user_id="{ record }">
          <a size="small" @click="handleViewUser(record.user_id)">
            {{ record.user_id }}
          </a>
        </template>
        <template #created_at="{ record }">
          <span>{{ formatDateTime(record.created_at) }}</span>
        </template>
        <template #updated_at="{ record }">
          <span>{{ formatDateTime(record.updated_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" danger @click="handleDelete(record.id)">
            <DeleteOutlined />
          </a-button>
        </template>
      </a-table>
    </div>
    
    <a-modal
      v-model:open="showUserPreviewModal"
      :title="t('admin.albums.userPreviewTitle')"
      width="40%"
      @cancel="handleCancelUserPreview"
      @ok="handleCancelUserPreview"
    >
        <a-descriptions :column="1" bordered size="small">
          <a-descriptions-item :label="t('admin.albums.userId')">{{ userPreview.id }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.albums.username')">{{ userPreview.username }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.albums.email')">{{ userPreview.email }}</a-descriptions-item>
          <a-descriptions-item :label="t('admin.albums.registerTime')">{{ formatDateTime(userPreview.created_at) }}</a-descriptions-item>
        </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { DeleteOutlined } from '@ant-design/icons-vue';
import { albumApi, userApi } from '../../api/services';
import type { AlbumResponse, User } from '../../types/api';
import { formatDateTime } from '../../utils/index';
import { getErrorMessage } from '../../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);

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

// 相册列表
const albums = ref<AlbumResponse[]>([]);

// 用户预览
const showUserPreviewModal = ref(false);
const userPreview = ref<User>({} as User);

// 搜索关键词
const searchKey = ref('');

// 排序参数
const orderby = ref('created_at');
const order = ref('desc');

// 表格列配置
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
  },
  {
    title: t('admin.albums.albumName'),
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: t('admin.albums.albumDescription'),
    dataIndex: 'description',
    key: 'description',
  },
  {
    title: t('admin.albums.user'),
    key: 'user_id',
    slots: { customRender: 'user_id' },
  },
  {
    title: t('admin.albums.imageCount'),
    key: 'image_count',
    dataIndex: 'image_count',
  },
  {
    title: t('admin.albums.createdAt'),
    key: 'created_at',
    slots: { customRender: 'created_at' },
  },
  {
    title: t('admin.albums.updatedAt'),
    key: 'updated_at',
    slots: { customRender: 'updated_at' },
  },
  {
    title: t('admin.albums.actions'),
    key: 'actions',
    slots: { customRender: 'actions' },
  },
];

// 获取相册列表
const fetchAlbums = async () => {
  try {
    loading.value = true;
    const response = await albumApi.getAllAlbums({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      searchkey: searchKey.value,
      orderby: orderby.value,
      order: order.value,
    });
    albums.value = response.data.albums;
    pagination.value.total = response.data.total;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.albums.fetchFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理搜索
const handleSearch = (value: string) => {
  searchKey.value = value;
  pagination.value.current = 1;
  fetchAlbums();
};

// 处理排序变化
const handleSortChange = () => {
  pagination.value.current = 1;
  fetchAlbums();
};

// 重置筛选
const resetFilters = () => {
  searchKey.value = '';
  orderby.value = 'created_at';
  order.value = 'desc';
  pagination.value.current = 1;
  fetchAlbums();
};

// 处理表格变化
const handleTableChange = (newPagination: any) => {
  pagination.value = newPagination;
  fetchAlbums();
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
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.albums.fetchUserFailed');
    message.error(errorMessage);
  }
};

// 处理取消用户预览
const handleCancelUserPreview = () => {
  showUserPreviewModal.value = false;
  userPreview.value = {} as User;
};

// 处理删除相册
const handleDelete = (id: number) => {
  Modal.confirm({
    title: t('admin.albums.confirmDeleteTitle'),
    content: t('admin.albums.confirmDeleteContent'),
    onOk: async () => {
      try {
        await albumApi.deleteAdminAlbums([id]);
        message.success(t('admin.albums.deleteSuccess'));
        fetchAlbums();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.albums.deleteFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理行选择变化
const handleRowSelectionChange = (selectedKeys: number[]) => {
  selectedRowKeys.value = selectedKeys;
};

// 处理批量删除
const handleBatchDelete = () => {
  if (selectedRowKeys.value.length === 0) return;
  
  Modal.confirm({
    title: t('admin.albums.confirmBatchDeleteTitle'),
    content: t('admin.albums.confirmBatchDeleteContent', { count: selectedRowKeys.value.length }),
    onOk: async () => {
      try {
        await albumApi.deleteAdminAlbums(selectedRowKeys.value);
        message.success(t('admin.albums.batchDeleteSuccess'));
        selectedRowKeys.value = [];
        fetchAlbums();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('admin.albums.batchDeleteFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 初始化
onMounted(() => {
  fetchAlbums();
});
</script>

<style scoped>
.albums-view {
  width: 100%;
}

.albums-content {
  padding: var(--spacing-lg);
}

.search-filter {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-lg);
}

.search-input {
  width: 300px;
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

/* 响应式设计 */
@media (max-width: 768px) {
  .albums-view {
    width: 100%;
  }
  
  .albums-content {
    padding: var(--spacing-md);
  }
  
  .search-filter {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
  }
  
  .search-input {
    width: 100%;
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
  
  :deep(.ant-table-wrapper) {
    overflow-x: auto;
  }
  
  :deep(.ant-table) {
    min-width: 800px;
  }
  
  :deep(.ant-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
  
  :deep(.ant-descriptions) {
    font-size: var(--font-size-sm);
  }
}
</style>
