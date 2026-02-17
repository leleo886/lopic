<template>
  <div class="storage-view">
    <div class="storage-content card">
      <!-- 搜索和筛选 -->
      <div class="search-filter">
        <a-input-search
          :placeholder="t('storages.search')"
          class="search-input"
          @search="handleSearch"
        />
        <div class="filter-buttons">
          <div class="sort-controls">
            <span class="sort-label">{{ t('storages.sort') }}</span>
            <a-select v-model:value="orderby" class="sort-select" style="width: 120px" @change="handleSortChange">
              <a-select-option value="created_at">{{ t('storages.createdAt') }}</a-select-option>
              <a-select-option value="updated_at">{{ t('storages.updatedAt') }}</a-select-option>
              <a-select-option value="name">{{ t('storages.storageName') }}</a-select-option>
            </a-select>
            <a-select v-model:value="order" class="sort-select" style="width: 80px" @change="handleSortChange">
              <a-select-option value="desc">{{ t('admin.albums.sortOptions.desc') }}</a-select-option>
              <a-select-option value="asc">{{ t('admin.albums.sortOptions.asc') }}</a-select-option>
            </a-select>
          </div>
          <a-button type="text" @click="resetFilters">{{ t('storages.resetFilters') }}</a-button>
          <a-button type="primary" @click="showCreateModal = true" class="create-button">
            <template #icon>
              <plus-outlined />
            </template>
            {{ t('storages.createStorage') }}
          </a-button>
        </div>
      </div>

      <!-- 存储配置列表 -->
      <a-table
        :columns="columns"
        :data-source="storages"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #type="{ record }">
          <span>{{ record.type }}</span>
        </template>
        <template #config="{ record }">
          <span class="config-content">{{ getConfigSummary(record.config) }}</span>
        </template>
        <template #created_at="{ record }">
          <span>{{ formatDateTime(record.created_at) }}</span>
        </template>
        <template #updated_at="{ record }">
          <span>{{ formatDateTime(record.updated_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" @click="handleEdit(record)" :disabled="record.name === 'local'">
            <EditOutlined />
          </a-button>
          <a-button type="text" size="small" @click="handleTestConnection(record)" :disabled="record.name === 'local'">
            <LinkOutlined />
          </a-button>
          <a-button type="text" size="small" danger @click="handleDelete(record.id, record.name)" :disabled="record.name === 'local'">
            <DeleteOutlined />
          </a-button>
        </template>
      </a-table>
      <span class="tip">{{ t('storages.localStorageTip') }}</span>
    </div>

    <!-- 创建/编辑存储配置模态框 -->
    <a-modal
      v-model:open="showCreateModal"
      :title="editingStorage ? t('storages.edit') : t('storages.createStorage')"
      :confirm-loading="saveLoading"
      @ok="handleSaveStorage"
      @cancel="handleCancel"
    >
      <a-form
        :model="storageForm"
        :rules="storageRules"
        ref="storageFormRef"
      >
        <a-form-item name="name" :label="t('storages.storageName')">
          <a-input v-model:value="storageForm.name" :placeholder="t('storages.enterStorageName')" :disabled="editingStorage && editingStorage.name === 'local'" />
        </a-form-item>
        <a-form-item name="type" :label="t('storages.storageType')">
          <a-radio-group v-model:value="storageForm.type" :disabled="editingStorage && editingStorage.name === 'local'">
            <a-radio value="webdav">{{ t('storages.webdavStorage') }}</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item name="config.base_path" :label="t('storages.basePath')">
          <a-input v-model:value="storageForm.config.base_path" :placeholder="t('storages.enterBasePath')" />
        </a-form-item>
        <a-form-item name="config.static_url" :label="t('storages.staticUrl')">
          <a-input v-model:value="storageForm.config.static_url" :placeholder="t('storages.enterStaticUrl')" />
        </a-form-item>
        <a-form-item name="config.base_url" :label="t('storages.webdavBaseUrl')" v-show="storageForm.type === 'webdav'">
          <a-input v-model:value="storageForm.config.base_url" :placeholder="t('storages.enterWebdavBaseUrl')" />
        </a-form-item>
        <a-form-item name="config.username" :label="t('storages.webdavUsername')" v-show="storageForm.type === 'webdav'">
          <a-input v-model:value="storageForm.config.username" :placeholder="t('storages.enterWebdavUsername')" />
        </a-form-item>
        <a-form-item name="config.password" :label="t('storages.webdavPassword')" v-show="storageForm.type === 'webdav'">
          <a-input 
            v-model:value="storageForm.config.password" 
            :placeholder="t('storages.enterWebdavPassword') + ' (留空保持原密码)'" 
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined, LinkOutlined } from '@ant-design/icons-vue';
import { storageApi } from '../../api/services';
import type { Storage, StorageRequest, StorageConfig } from '../../types/api';
import { formatDateTime } from '../../utils/index';
import { getErrorMessage } from '../../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);
const saveLoading = ref(false);

// 分页配置
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showTotal: (total: number) => `共 ${total} 条记录`,
});

// 存储配置列表
const storages = ref<Storage[]>([]);

// 搜索关键词
const searchKey = ref('');

// 排序参数
const orderby = ref('created_at');
const order = ref('desc');

// 创建/编辑模态框
const showCreateModal = ref(false);
const storageFormRef = ref<any>();
const editingStorage = ref<Storage | null>(null);

// 存储配置表单
const storageForm = reactive({
  name: '',
  type: 'webdav',
  config: {
    base_path: '/webdav/base/path',
    static_url: 'https://dav.example.com/d',
    base_url: 'https://dav.example.com',
    username: 'example',
    password: '123456',
  },
});

// 表单验证规则
const storageRules = {
  name: [
    { required: true, message: () => t('storages.enterStorageName'), trigger: 'blur' },
  ],
  type: [
    { required: true, message: () => t('storages.selectStorageType'), trigger: ['blur', 'change'] },
  ],
  'config': {
    'base_path': [
      { required: true, message: () => t('storages.enterBasePath'), trigger: 'blur' },
    ],
    'static_url': [
      { required: true, message: () => t('storages.enterStaticUrl'), trigger: 'blur' },
    ],
    'base_url': [
      {
        required: () => storageForm.type === 'webdav',
        message: () => t('storages.enterWebdavBaseUrl'),
        trigger: ['blur', 'change'],
      },
    ],
    'username': [
      {
        required: () => storageForm.type === 'webdav',
        message: () => t('storages.enterWebdavUsername'),
        trigger: ['blur', 'change'],
      },
    ],
    'password': [
      {
        required: false, // 密码不再必填，留空保持原密码
        message: () => t('storages.enterWebdavPassword'),
        trigger: ['blur', 'change'],
      },
    ],
  },
};

// 表格列配置
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    fixed: 'start',
  },
  {
    title: t('storages.storageName'),
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: t('storages.storageType'),
    key: 'type',
    slots: { customRender: 'type' },
  },
  {
    title: t('storages.config'),
    key: 'config',
    slots: { customRender: 'config' },
    ellipsis: true,
    width: 300,
  },
  {
    title: t('storages.createdAt'),
    key: 'created_at',
    slots: { customRender: 'created_at' },
  },
  {
    title: t('storages.updatedAt'),
    key: 'updated_at',
    slots: { customRender: 'updated_at' },
  },
  {
    title: '操作',
    key: 'actions',
    slots: { customRender: 'actions' },
  },
];

// 获取存储配置列表
const fetchStorages = async () => {
  try {
    loading.value = true;
    const response = await storageApi.getStorages({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      searchkey: searchKey.value,
      orderby: orderby.value,
      order: order.value,
    });
    storages.value = response.data;
    pagination.value.total = response.data.length;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('storages.fetchFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理搜索
const handleSearch = (value: string) => {
  searchKey.value = value;
  pagination.value.current = 1;
  fetchStorages();
};

// 处理排序变化
const handleSortChange = () => {
  pagination.value.current = 1;
  fetchStorages();
};

// 重置筛选
const resetFilters = () => {
  searchKey.value = '';
  orderby.value = 'created_at';
  order.value = 'desc';
  pagination.value.current = 1;
  fetchStorages();
};

// 处理表格变化
const handleTableChange = (newPagination: any) => {
  pagination.value = newPagination;
  fetchStorages();
};

// 处理编辑存储配置
const handleEdit = (storage: Storage) => {
  editingStorage.value = storage;
  storageForm.name = storage.name;
  storageForm.type = storage.type;
  storageForm.config = JSON.parse(JSON.stringify(storage.config));
  
  // 处理密码占位符
  if (storageForm.config.password === "[SET]") {
    storageForm.config.password = ""; // 清空密码字段，用户可选择留空保持原密码
  }
  
  showCreateModal.value = true;
};

// 处理删除存储配置
const handleDelete = (id: number, name: string) => {
  if (name === 'local') {
    message.warning(t('storages.cannotDeleteLocalStorage'));
    return;
  }
  
  Modal.confirm({
    title: t('storages.confirmDelete'),
    content: t('storages.deleteContent', { name }),
    onOk: async () => {
      try {
        await storageApi.deleteStorage(id);
        message.success(t('storages.deleteSuccess'));
        fetchStorages();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('storages.deleteFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理测试存储连接
const handleTestConnection = (storage: Storage) => {
  Modal.confirm({
    title: t('storages.confirmTestConnection'),
    content: t('storages.testConnectionContent', { name: storage.name }),
    onOk: async () => {
      try {
        const storageRequest: StorageRequest = {
          name: storage.name,
          type: storage.type,
          config: storage.config,
        };
        
        // 处理密码占位符
        if (storageRequest.config.password === "[SET]") {
          // 密码已设置，使用空字符串表示保持原密码
          storageRequest.config.password = "";
        }
        
        await storageApi.testStorageConnection(storageRequest);
        message.success(t('storages.testConnectionSuccess'));
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('storages.testConnectionFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理保存存储配置
const handleSaveStorage = async () => {
  if (!storageFormRef.value) return;
  
  try {
    // 手动验证必填字段
    if (!storageForm.name) {
      message.error(t('storages.enterStorageName'));
      return;
    }
    if (!storageForm.type) {
      message.error(t('storages.selectStorageType'));
      return;
    }
    if (!storageForm.config.base_path) {
      message.error(t('storages.enterBasePath'));
      return;
    }
    if (!storageForm.config.static_url) {
      message.error(t('storages.enterStaticUrl'));
      return;
    }
    if (storageForm.type === 'webdav') {
      if (!storageForm.config.base_url) {
        message.error(t('storages.enterWebdavBaseUrl'));
        return;
      }
      if (!storageForm.config.username) {
        message.error(t('storages.enterWebdavUsername'));
        return;
      }
      // 移除密码的必填验证，允许留空保持原密码
    }
    
    saveLoading.value = true;
    
    const storageRequest: StorageRequest = {
      name: storageForm.name,
      type: storageForm.type,
      config: storageForm.config,
    };
    
    if (editingStorage.value) {
      // 更新存储配置
      await storageApi.updateStorage(editingStorage.value.id, storageRequest);
      message.success(t('storages.updateSuccess'));
    } else {
      // 创建存储配置
      await storageApi.createStorage(storageRequest);
      message.success(t('storages.createSuccess'));
    }
    
    handleCancel();
    fetchStorages();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('storages.saveFailed');
    message.error(errorMessage);
  } finally {
    saveLoading.value = false;
  }
};

// 处理取消
const handleCancel = () => {
  showCreateModal.value = false;
  editingStorage.value = null;
  storageForm.name = '';
  storageForm.type = 'webdav';
  storageForm.config = {
    base_path: '/webdav/base/path',
    static_url: 'https://dav.example.com/d',
    base_url: 'https://dav.example.com',
    username: 'example',
    password: '123456',
  };
  if (storageFormRef.value) {
    storageFormRef.value.resetFields();
  }
};

// 获取配置摘要
const getConfigSummary = (config: StorageConfig): string => {
  if (config.base_url) {
    return `WebDAV: ${config.base_url} (${config.username})`;
  }
  return `Local: ${config.base_path} (${config.static_url})`;
};

// 初始化
onMounted(() => {
  fetchStorages();
});
</script>

<style scoped>
.storage-view {
  width: 100%;
}

.create-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.create-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

.storage-content {
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

.tip {
  font-size: var( --font-size-base);
  color: var(--primary-dark);
}

.config-content {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .storage-view {
    width: 100%;
  }
  
  .storage-content {
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
  
  .create-button {
    flex: 1;
    min-width: 120px;
  }
  
  .tip {
    font-size: var(--font-size-sm);
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
}
</style>
