<template>
  <div class="users-view">
    <div class="users-content card">
      <!-- 搜索和筛选 -->
      <div class="search-filter">
        <a-input-search
          :placeholder="t('users.search')"
          class="search-input"
          @search="handleSearch"
        />
        <div class="filter-buttons">
          <div class="sort-controls">
            <span class="sort-label">{{ t('users.sort') }}</span>
            <a-select v-model:value="orderby" class="sort-select" style="width: 120px" @change="handleSortChange">
              <a-select-option value="created_at">{{ t('users.createdAt') }}</a-select-option>
              <a-select-option value="updated_at">{{ t('users.updatedAt') }}</a-select-option>
              <a-select-option value="username">{{ t('users.username') }}</a-select-option>
              <a-select-option value="email">{{ t('users.email') }}</a-select-option>
              <a-select-option value="active">{{ t('users.status') }}</a-select-option>
              <a-select-option value="image_count">{{ t('users.imageCount') }}</a-select-option>
            </a-select>
            <a-select v-model:value="order" class="sort-select" style="width: 80px" @change="handleSortChange">
              <a-select-option value="desc">{{ t('admin.albums.sortOptions.desc') }}</a-select-option>
              <a-select-option value="asc">{{ t('admin.albums.sortOptions.asc') }}</a-select-option>
            </a-select>
          </div>
          <a-button type="text" @click="resetFilters">{{ t('users.resetFilters') }}</a-button>
          <a-button type="primary" @click="showCreateModal = true" class="create-button">
            <template #icon>
              <plus-outlined />
            </template>
            {{ t('users.createUser') }}
          </a-button>
        </div>
      </div>
      
      <!-- 删除中状态 -->
      <div class="deleting-status" v-if="deleting">
        <a-spin size="small">
          <span>{{ t('users.deleting') }}</span>
        </a-spin>
      </div>

      <!-- 用户列表 -->
      <a-table
        :columns="columns"
        :data-source="users"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #status="{ record }">
          <a-badge :status="record.active ? 'success' : 'default'" :text="record.active ? t('users.active') : t('users.inactive')" />
        </template>
        <template #role="{ record }">
          <a-tag>{{ record.role.name }}</a-tag>
        </template>
        <template #total_size="{ record }">
          <span>{{ formatFileSize(record.total_size) }} / {{ record.role.max_storage_size_mb == -1 ? t('dashboard.unlimited') : record.role.max_storage_size_mb + 'MB' }}</span>
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
          <a-button type="text" size="small" danger @click="handleDelete(record.id)">
            <DeleteOutlined />
          </a-button>
        </template>
      </a-table>
      <!-- 存储使用 -->
      <div class="total-storage">
        <span>{{ t('users.totalStorageUsed') }}：{{ formatFileSize(totalStorage) }}</span>
      </div>
    </div>

    <!-- 创建/编辑用户模态框 -->
    <a-modal
      v-model:open="showCreateModal"
      :title="editingUser ? t('users.edit') : t('users.createUser')"
      @ok="handleSaveUser"
      @cancel="handleCancel"
    >
      <a-form
        :model="userForm"
        :rules="userRules"
        ref="userFormRef"
      >
        <a-form-item name="username" :label="t('users.username')">
          <a-input v-model:value="userForm.username" :placeholder="t('users.enterUsername')" />
        </a-form-item>
        <a-form-item name="email" :label="t('users.email')">
          <a-input v-model:value="userForm.email" :placeholder="t('users.enterEmail')" />
        </a-form-item>
        <a-form-item name="password" :label="t('users.password')">
          <a-input-password v-model:value="userForm.password" :placeholder="t('users.enterPassword')" />
        </a-form-item>
        <a-form-item name="role" :label="t('users.role')">
          <a-select v-model:value="userForm.role" :placeholder="t('users.selectRole')" :loading="loadingRoles">
            <a-select-option v-for="role in roles" :key="role.id" :value="role.name">
              {{ role.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item name="active" :label="t('users.status')">
          <a-switch v-model:checked="userForm.active" />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import { userApi, roleApi } from '../../api/services';
import { uploadWebSocketService } from '../../api/websocket';
import type { User, AdminUserRequest, Role } from '../../types/api';
import { formatFileSize, formatDateTime } from '../../utils/index';
import { getErrorMessage } from '../../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 加载状态
const loading = ref(false);
// 删除中状态
const deleting = ref(false);
// 角色列表
const roles = ref<Role[]>([]);
// 角色加载状态
const loadingRoles = ref(false);

// 分页配置
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  pageSizeOptions: ['10', '20', '50', '100'],
  showTotal: (total: number) => `共 ${total} 条记录`,
});

// 用户列表
const users = ref<User[]>([]);

// 搜索关键词
const searchKey = ref('');

// 排序参数
const orderby = ref('created_at');
const order = ref('desc');

// 创建/编辑模态框
const showCreateModal = ref(false);
const userFormRef = ref();
const editingUser = ref<User | null>(null);

// 总使用空间
const totalStorage = ref(0);

// 用户表单
const userForm = reactive<AdminUserRequest>({
  username: '',
  email: '',
  password: '',
  role: 'user',
  active: true,
});

// 表单验证规则
const userRules = {
  username: [
    { required: true, message: () => t('users.enterUsername'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: () => t('users.enterEmail'), trigger: 'blur' },
    { type: 'email', message: () => t('users.enterValidEmail'), trigger: 'blur' },
  ],
  password: [
    { required: false, message: () => t('users.enterPassword'), trigger: 'blur' },
    { min: 6, message: () => t('users.passwordLength'), trigger: 'blur' },
  ],
  role: [
    { required: true, message: () => t('users.selectRole'), trigger: 'change' },
  ],
};

// 表格列配置
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
  },
  {
    title: t('users.username'),
    dataIndex: 'username',
    key: 'username',
  },
  {
    title: t('users.email'),
    dataIndex: 'email',
    key: 'email',
  },
  {
    title: t('users.role'),
    dataIndex: 'role',
    key: 'role',
    slots: { customRender: 'role' },
  },
  {
    title: t('users.status'),
    dataIndex: 'active',
    key: 'active',
    slots: { customRender: 'status' },
  },
  {
    title: t('users.imageCount'),
    dataIndex: 'image_count',
    key: 'image_count',
  },
  {
    title: t('users.usedStorage'),
    key: 'total_size',
    slots: { customRender: 'total_size' },
  },
  {
    title: t('users.createdAt'),
    key: 'created_at',
    slots: { customRender: 'created_at' },
  },
  {
    title: t('users.updatedAt'),
    key: 'updated_at',
    slots: { customRender: 'updated_at' },
  },
  {
    title: '操作',
    key: 'actions',
    slots: { customRender: 'actions' },
  },
];

// 获取用户列表
const fetchUsers = async () => {
  try {
    loading.value = true;
    const response = await userApi.getUsers({
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      searchkey: searchKey.value,
      orderby: orderby.value,
      order: order.value,
    });
    users.value = response.data.users;
    pagination.value.total = response.data.total;
    for (const user of response.data.users) {
      totalStorage.value += user.total_size || 0;
    }
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('users.fetchFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理搜索
const handleSearch = (value: string) => {
  searchKey.value = value;
  pagination.value.current = 1;
  fetchUsers();
};

// 重置筛选
const resetFilters = () => {
  searchKey.value = '';
  orderby.value = 'created_at';
  order.value = 'desc';
  pagination.value.current = 1;
  fetchUsers();
};

// 处理排序变化
const handleSortChange = () => {
  pagination.value.current = 1;
  fetchUsers();
};

// 处理表格变化
const handleTableChange = (newPagination: any) => {
  pagination.value = newPagination;
  fetchUsers();
};

// 处理编辑用户
const handleEdit = (user: User) => {
  editingUser.value = user;
  userForm.username = user.username;
  userForm.email = user.email;
  userForm.role = user.role.name;
  userForm.active = user.active;
  showCreateModal.value = true;
};

// 处理删除用户
const handleDelete = (id: number) => {
  Modal.confirm({
    title: t('users.confirmDelete'),
    content: t('users.deleteContent'),
    onOk: async () => {
      try {
        deleting.value = true;
        userApi.deleteUser(id);
        message.warning(t('users.deleteStarted'));
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('users.deleteFailed');
        message.error(errorMessage);
        deleting.value = false;
      }
    },
  });
};

// 处理保存用户
const handleSaveUser = async () => {
  if (!userFormRef.value) return;
  
  try {
    await userFormRef.value.validate();
    
    if (editingUser.value) {
      // 更新用户
      await userApi.updateUser(editingUser.value.id, userForm);
      message.success(t('users.updateSuccess'));
    } else {
      // 创建用户
      await userApi.createUser(userForm);
      message.success(t('users.createSuccess'));
    }
    
    handleCancel();
    fetchUsers();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('users.saveFailed');
    message.error(errorMessage);
  }
};

// 处理取消
const handleCancel = () => {
  showCreateModal.value = false;
  editingUser.value = null;
  userForm.username = '';
  userForm.email = '';
  userForm.password = '';
  userForm.role = 'user';
  userForm.active = true;
  if (userFormRef.value) {
    userFormRef.value.resetFields();
  }
};

// 获取角色列表
const fetchRoles = async () => {
  try {
    loadingRoles.value = true;
    const response = await roleApi.getRoles({ page: 1, page_size: 100 });
    roles.value = response.data;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('roles.fetchFailed');
    message.error(errorMessage);
  } finally {
    loadingRoles.value = false;
  }
};

// 初始化
onMounted(() => {
  fetchUsers();
  fetchRoles();
  
  // 连接 WebSocket
  uploadWebSocketService.connect();
  
  // 注册删除用户成功监听器
  uploadWebSocketService.on('deleteUserSuccess', () => {
    deleting.value = false;
    message.success(t('users.deleteSuccess'));
    fetchUsers();
  });
  
  // 注册删除用户错误监听器
  uploadWebSocketService.on('deleteUserError', (data: { message: string; error: string; code: string }) => {
    deleting.value = false;
    message.error(t('users.deleteFailed') + `: ${getErrorMessage(data.code)}`);
    fetchUsers();
  });
});

// 组件卸载
onUnmounted(() => {
  // 断开 WebSocket 连接
  uploadWebSocketService.disconnect();
});
</script>

<style scoped>
.users-view {
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

.users-content {
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

.deleting-status {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  margin-bottom: var(--spacing-lg);
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  font-size: 14px;
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

.total-storage {
  color: var(--primary-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .users-view {
    width: 100%;
  }
  
  .users-content {
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
  
  .total-storage {
    margin-top: var(--spacing-md);
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
