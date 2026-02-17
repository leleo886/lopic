<template>
  <div class="roles-view">
    <div class="roles-content card">
      <!-- 搜索和筛选 -->
      <div class="search-filter">
        <a-input-search
          :placeholder="t('roles.search')"
          class="search-input"
          @search="handleSearch"
        />
        <div class="filter-buttons">
          <div class="sort-controls">
            <span class="sort-label">{{ t('roles.sort') }}</span>
            <a-select v-model:value="orderby" class="sort-select" style="width: 120px" @change="handleSortChange">
              <a-select-option value="created_at">{{ t('roles.createdAt') }}</a-select-option>
              <a-select-option value="updated_at">{{ t('roles.updatedAt') }}</a-select-option>
              <a-select-option value="name">{{ t('roles.roleName') }}</a-select-option>
            </a-select>
            <a-select v-model:value="order" class="sort-select" style="width: 80px" @change="handleSortChange">
              <a-select-option value="desc">{{ t('admin.albums.sortOptions.desc') }}</a-select-option>
              <a-select-option value="asc">{{ t('admin.albums.sortOptions.asc') }}</a-select-option>
            </a-select>
          </div>
          <a-button type="text" @click="resetFilters">{{ t('roles.resetFilters') }}</a-button>
          <a-button type="primary" @click="showCreateModal = true" class="create-button">
            <template #icon>
              <plus-outlined />
            </template>
            {{ t('roles.createRole') }}
          </a-button>
        </div>
      </div>

      <!-- 角色列表 -->
      <a-table
        :columns="columns"
        :data-source="roles"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #usersCount="{ record }">
          <span>{{ usersCount[record.name] }}</span>
        </template>
        <template #created_at="{ record }">
          <span>{{ formatDateTime(record.created_at) }}</span>
        </template>
        <template #updated_at="{ record }">
          <span>{{ formatDateTime(record.updated_at) }}</span>
        </template>
        <template #actions="{ record }">
          <a-button type="text" size="small" @click="handleEdit(record)">
            <EditOutlined />
          </a-button>
          <a-button type="text" size="small" danger @click="handleDelete(record.id, record.name)">
            <DeleteOutlined />
          </a-button>
        </template>
      </a-table>
       <span class="tip">'-1' 表示无限制</span>
    </div>

    <!-- 创建/编辑角色模态框 -->
    <a-modal
      v-model:open="showCreateModal"
      :title="editingRole ? t('roles.edit') : t('roles.createRole')"
      @ok="handleSaveRole"
      @cancel="handleCancel"
    >
      <a-form
        :model="roleForm"
        :rules="roleRules"
        ref="roleFormRef"
      >
        <a-form-item name="name" :label="t('roles.roleName')">
          <a-input v-model:value="roleForm.name" :placeholder="t('roles.enterRoleName')" />
        </a-form-item>
        <a-form-item name="description" :label="t('roles.roleDescription')">
          <a-textarea v-model:value="roleForm.description" :placeholder="t('roles.enterRoleDescription')" rows="3" />
        </a-form-item>
        <a-form-item name="allowed_extensions" :label="t('roles.allowedExtensions')">
          <a-select
            v-model:value="roleForm.allowed_extensions"
            mode="multiple"
            :placeholder="t('roles.selectAllowedExtensions')"
            style="width: 100%"
          >
            <a-select-option
              v-for="ext in allowedExtensionsOptions"
              :key="ext"
              :value="ext"
            >
              {{ ext }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item name="gallery_open" :label="t('roles.galleryOpen')">
          <a-switch v-model:checked="roleForm.gallery_open" />
          <a-tooltip class="gallery-tooltip" color="var(--primary-color)"
             placement="top"
             :title="t('albums.galleryTooltip')">
            <InfoCircleOutlined />
          </a-tooltip>
        </a-form-item>
        <a-form-item name="max_albums_per_user" :label="t('roles.maxAlbumsPerUser')">
          <a-input-number v-model:value="roleForm.max_albums_per_user" :placeholder="t('roles.enterMaxAlbums')" />
        </a-form-item>
        <a-form-item name="max_file_size_mb" :label="t('roles.maxFileSizeMb')">
          <a-input-number v-model:value="roleForm.max_file_size_mb" :placeholder="t('roles.enterMaxFileSize')" />
        </a-form-item>
        <a-form-item name="max_files_per_upload" :label="t('roles.maxFilesPerUpload')">
          <a-input-number v-model:value="roleForm.max_files_per_upload" :placeholder="t('roles.enterMaxFilesPerUpload')" />
        </a-form-item>
        <a-form-item name="max_storage_size_mb" :label="t('roles.storageLimit')">
          <a-input-number v-model:value="roleForm.max_storage_size_mb" :placeholder="t('roles.enterStorageLimit')" />
        </a-form-item>
        <a-form-item name="storage_name" :label="t('roles.storageName')">
          <a-select v-model:value="roleForm.storage_name" :placeholder="t('roles.selectStorage')" :loading="storagesLoading">
            <a-select-option v-for="storage in storages" :key="storage.name" :value="storage.name">
              {{ storage.name }} ({{ storage.type }})
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons-vue';
import { roleApi, storageApi } from '../../api/services';
import type { Role, Storage } from '../../types/api';
import { formatDateTime } from '../../utils/index';
import { getErrorMessage } from '../../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);

// 存储配置加载状态
const storagesLoading = ref(false);

// 存储配置列表
const storages = ref<Storage[]>([]);

// 允许的文件扩展名选项（从环境变量读取）
const allowedExtensionsOptions = computed(() => {
  const envExtensions = import.meta.env.VITE_ALLOWED_EXTENSIONS || '.jpg,.jpeg,.png,.bmp,.tiff,.tif,.webp,.gif';
  return envExtensions
    .split(',')
    .map((ext: string) => ext.trim())
    .filter((ext: string) => ext !== '');
});

// 分页配置
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showTotal: (total: number) => `共 ${total} 条记录`,
});

// 角色列表
const roles = ref<Role[]>([]);

// 角色用户数量
const usersCount = ref<Record<number, number>>({});

// 搜索关键词
const searchKey = ref('');

// 排序参数
const orderby = ref('created_at');
const order = ref('desc');

// 创建/编辑模态框
const showCreateModal = ref(false);
const roleFormRef = ref();
const editingRole = ref<Role | null>(null);

// 角色表单
const roleForm = reactive({
  name: '',
  description: '',
  allowed_extensions: ['.jpg','.png'],
  gallery_open: false,
  max_albums_per_user: 0,
  max_file_size_mb: 0,
  max_files_per_upload: 0,
  max_storage_size_mb: 0,
  storage_name: 'local',
});



// 表单验证规则
const roleRules = {
  name: [
    { required: true, message: () => t('roles.enterRoleName'), trigger: 'blur' },
  ],
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
    title: t('roles.roleName'),
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: t('roles.roleDescription'),
    dataIndex: 'description',
    key: 'description',
    ellipsis: true,
  },
  {
    title: t('roles.userCount'),
    key: 'usersCount',
    slots: { customRender: 'usersCount' },
  },
  {
    title: t('roles.maxFilesPerUpload'),
    dataIndex: 'max_files_per_upload',
    key: 'max_files_per_upload',
  },
  {
    title: t('roles.maxFileSizeMb'),
    dataIndex: 'max_file_size_mb',
    key: 'max_file_size_mb',
  },
  {
    title: t('roles.maxAlbumsPerUser'),
    dataIndex: 'max_albums_per_user',
    key: 'max_albums_per_user',
  },
  {
    title: t('roles.allowedExtensions'),
    dataIndex: 'allowed_extensions',
    key: 'allowed_extensions',
    ellipsis: true,
    width: 120,
  },
  {
    title: t('roles.storageLimit'),
    dataIndex: 'max_storage_size_mb',
    key: 'max_storage_size_mb',
  },
  {
    title: t('roles.storageName'),
    dataIndex: 'storage_name',
    key: 'storage_name',
  },
  {
    title: t('roles.createdAt'),
    key: 'created_at',
    slots: { customRender: 'created_at' },
  },
  {
    title: t('roles.updatedAt'),
    key: 'updated_at',
    slots: { customRender: 'updated_at' },
  },
  {
    title: '操作',
    key: 'actions',
    slots: { customRender: 'actions' },
  },
];

// 获取角色列表
const fetchRoles = async () => {
  try {
    loading.value = true;
    const response = await roleApi.getRoles({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      searchkey: searchKey.value,
      orderby: orderby.value,
      order: order.value,
    });
    roles.value = response.data;
    pagination.value.total = response.data.length;
    fetchRolesUsersCount();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('roles.fetchFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 获取角色用户数量
const fetchRolesUsersCount = async () => {
  try {
    const response = await roleApi.getRolesUsersCount();
    usersCount.value = response.data;
  } catch (error) {
    console.error('Failed to fetch roles users count:', error);
  }
};

// 获取存储配置列表
const fetchStorages = async () => {
  try {
    storagesLoading.value = true;
    const response = await storageApi.getStorages({
      page_size: 100,
    });
    storages.value = response.data;
  } catch (error: any) {
    console.error('Failed to fetch storages:', error);
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('storages.fetchFailed');
    message.error(errorMessage);
  } finally {
    storagesLoading.value = false;
  }
};

// 处理搜索
const handleSearch = (value: string) => {
  searchKey.value = value;
  pagination.value.current = 1;
  fetchRoles();
};

// 处理排序变化
const handleSortChange = () => {
  pagination.value.current = 1;
  fetchRoles();
};

// 重置筛选
const resetFilters = () => {
  searchKey.value = '';
  orderby.value = 'created_at';
  order.value = 'desc';
  pagination.value.current = 1;
  fetchRoles();
};

// 处理表格变化
const handleTableChange = (newPagination: any) => {
  pagination.value = newPagination;
  fetchRoles();
};

// 处理编辑角色
const handleEdit = (role: Role) => {
  editingRole.value = role;
  roleForm.name = role.name;
  roleForm.description = role.description;
  roleForm.allowed_extensions = role.allowed_extensions;
  roleForm.gallery_open = role.gallery_open;
  roleForm.max_albums_per_user = role.max_albums_per_user;
  roleForm.max_file_size_mb = role.max_file_size_mb;
  roleForm.max_files_per_upload = role.max_files_per_upload;
  roleForm.max_storage_size_mb = role.max_storage_size_mb;
  roleForm.storage_name = role.storage_name;
  
  showCreateModal.value = true;
};

// 处理删除角色
const handleDelete = (id: number, name: string) => {
  Modal.confirm({
    title: t('roles.confirmDelete'),
    content: t('roles.deleteContent', { name }),
    onOk: async () => {
      try {
        await roleApi.deleteRole(id);
        message.success(t('roles.deleteSuccess'));
        fetchRoles();
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('roles.deleteFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理保存角色
const handleSaveRole = async () => {
  if (!roleFormRef.value) return;
  
  try {
    await roleFormRef.value.validate();
    
    if (editingRole.value) {
      // 更新角色
      await roleApi.updateRole(editingRole.value.id, roleForm);
      message.success(t('roles.updateSuccess'));
    } else {
      // 创建角色
      await roleApi.createRole(roleForm);
      message.success(t('roles.createSuccess'));
    }
    
    handleCancel();
    fetchRoles();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('roles.saveFailed');
    message.error(errorMessage);
  }
};

// 处理取消
const handleCancel = () => {
  showCreateModal.value = false;
  editingRole.value = null;
  roleForm.name = '';
  roleForm.description = '';
  roleForm.allowed_extensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp'];
  roleForm.gallery_open = false;
  roleForm.max_albums_per_user = 100;
  roleForm.max_file_size_mb = 10;
  roleForm.max_files_per_upload = 10;
  roleForm.max_storage_size_mb = 1024;
  roleForm.storage_name = 'local';
  if (roleFormRef.value) {
    roleFormRef.value.resetFields();
  }
};

// 初始化
onMounted(() => {
  fetchRoles();
  fetchStorages();
});
</script>

<style scoped>
.roles-view {
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

.roles-content {
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

.gallery-tooltip {
    position: relative;
    top: 2px;
    left: 5px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .roles-view {
    width: 100%;
  }
  
  .roles-content {
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
    min-width: 1000px;
  }
  
  :deep(.ant-pagination) {
    flex-wrap: wrap;
    justify-content: center;
  }
}
</style>
