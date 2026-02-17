<template>
  <div class="albums-container">
    <!-- 顶部导航栏 -->
    <header class="albums-header card">
        <a-button type="primary" @click="showCreateModal = true" class="create-button">
          <template #icon>
            <plus-outlined />
          </template>
          {{ t('albums.createAlbum') }}
        </a-button>
        <a-button type="primary" @click="openGallery" class="create-button">
          <template #icon>
            <SwapRightOutlined />
          </template>
          {{ t('albums.openGallery') }}
        </a-button>
    </header>

    <!-- 主要内容 -->
    <main class="albums-main">
      <!-- 相册列表 -->
      <div v-if="albums.length > 0" class="albums-grid">
        <div
          v-for="album in albums"
          :key="album.id"
          class="album-card card"
          @click="navigateToAlbum(album.id)"
        >
          <div class="album-cover">
            <camera-outlined class="cover-icon" />
            <span class="cover-text">{{ album.name }}</span>
          </div>
          <div class="album-info">
            <h3 class="album-title">{{ album.name }}</h3>
            <p class="album-description">{{ album.description }}</p>
            <div class="album-meta">
              <span class="album-date">{{ formatDateTime(album.created_at) }}</span>
              <span class="album-images">{{ album.image_count }} {{ t('albums.images') }}</span>
            </div>
          </div>
          <div class="album-actions">
            <div style="width: 100%;">{{ t('albums.galleryEnabled') }}：
              <span v-if="album.gallery_enabled">
                <CheckCircleTwoTone twoToneColor="#52c41a" />
              </span>
              <span v-else>
                <CloseCircleTwoTone twoToneColor="#eb2f96" />
              </span>
            </div>
            <a-button
              type="text"
              size="small"
              @click.stop="handleEdit(album)"
            >
              <EditOutlined />
              {{ t('albums.edit') }}
            </a-button>
            <a-button
              type="text"
              size="small"
              danger
              @click.stop="handleDelete(album.id)"
            >
              <DeleteOutlined />
              {{ t('albums.delete') }}
            </a-button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-state">
        <div class="empty-icon">
          <aim-outlined />
        </div>
        <h3 class="empty-title">{{ t('albums.noAlbums') }}</h3>
        <p class="empty-description">{{ t('albums.createFirstAlbum') }}</p>
        <a-button
          type="primary"
          @click="showCreateModal = true"
          class="empty-button"
        >
          <template #icon>
            <plus-outlined />
          </template>
          {{ t('albums.createAlbum') }}
        </a-button>
      </div>
    </main>

    <!-- 创建/编辑相册模态框 -->
    <a-modal
      v-model:open="showCreateModal"
      :title="editingAlbum ? t('albums.editAlbum') : t('albums.createAlbum')"
      :confirmLoading="saveLoading"
      @ok="handleSaveAlbum"
      @cancel="handleCancel"
    >
      <a-form
        :model="albumForm"
        :rules="albumRules"
        ref="albumFormRef"
      >
        <a-form-item name="serial_number" :label="t('albums.serialNumber')">
          <a-input-number 
            v-model:value="albumForm.serial_number" 
            :placeholder="t('albums.enterSerialNumber')" 
            min="0"
            style="width: 100%;"
          />
        </a-form-item>
        <a-form-item name="name" :label="t('albums.albumName')">
          <a-input
            v-model:value="albumForm.name"
            :placeholder="t('albums.enterAlbumName')"
          />
        </a-form-item>
        <a-form-item name="description" :label="t('albums.albumDescription')">
          <a-textarea
            v-model:value="albumForm.description"
            :placeholder="t('albums.enterAlbumDescription')"
            rows="3"
          />
        </a-form-item>
        <a-form-item name="gallery_enabled" :label="t('albums.galleryEnabled')">
          <a-switch v-model:checked="albumForm.gallery_enabled" />
          <a-tooltip class="gallery-tooltip" color="var(--primary-color)"
             placement="top"
             :title="t('albums.galleryTooltip')">
            <InfoCircleOutlined />
          </a-tooltip>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { message, Modal } from 'ant-design-vue';
import { useI18n } from 'vue-i18n';
import {
  PlusOutlined,
  AimOutlined,
  CameraOutlined,
  DeleteOutlined,
  EditOutlined,
  CheckCircleTwoTone,
  CloseCircleTwoTone,
  InfoCircleOutlined,
  SwapRightOutlined,
} from '@ant-design/icons-vue';
import { albumApi, userApi } from '../api/services';
import { useAuthStore } from '../stores/auth';
import { formatDateTime } from '../utils/index'
import type { AlbumResponse, AlbumRequest, User } from '../types/api';
import { getErrorMessage } from '../types/errorMessages';

const { t } = useI18n();

const authStore = useAuthStore();

// 路由
const router = useRouter();

// 相册列表
const albums = ref<Array<AlbumResponse & { image_count?: number }>>([]);

  
// 当前用户
const currentUser = ref<User | null>(null);

// 加载状态
const loading = ref(false);

// 创建/编辑模态框
const showCreateModal = ref(false);
const saveLoading = ref(false);
const albumFormRef = ref();
const editingAlbum = ref<AlbumResponse | null>(null);

// 相册表单
const albumForm = reactive<AlbumRequest>({
  name: '',
  description: '',
  gallery_enabled: false,
  serial_number: 0,
});

// 表单验证规则
const albumRules = {
  name: [
    { required: true, message: () => t('albums.enterAlbumName'), trigger: 'blur' },
    { max: 100, message: () => t('albums.albumNameLimit'), trigger: 'blur' },
  ],
  description: [
    { max: 500, message: () => t('albums.albumDescriptionLimit'), trigger: 'blur' },
  ],
};

// 获取用户信息
const fetchUserInfo = async () => {
  try {
    const userResponse = await userApi.getCurrentUser();
    if (userResponse.data) {
      currentUser.value = userResponse.data || null;
    }
  } catch (error) {
    console.error('Failed to fetch user info:', error);
  }
};

const openGallery = () => {
  const username = authStore.user?.username;
  router.push(`/gallery/${username}`);
}

// 获取相册列表
const fetchAlbums = async () => {
  try {
    loading.value = true;
    const response = await albumApi.getUserAlbums({ page: 1, page_size: 100 });
    albums.value = response.data.albums;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.fetchAlbumsFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 导航到相册详情
const navigateToAlbum = (id: number) => {
  router.push(`/manage/album/${id}`);
};


// 处理编辑相册
const handleEdit = (album: AlbumResponse) => {
  editingAlbum.value = album;
  albumForm.name = album.name;
  albumForm.description = album.description;
  albumForm.gallery_enabled = album.gallery_enabled;
  albumForm.serial_number = (album as any).serial_number || 0;
  showCreateModal.value = true;
};

// 处理保存相册
const handleSaveAlbum = async () => {
  if (!albumFormRef.value) return;

  saveLoading.value = true;
  
  await fetchUserInfo();
  if (!currentUser.value?.role.gallery_open && albumForm.gallery_enabled) {
    message.error(t('albums.galleryPermissionDenied'));
    saveLoading.value = false;
    return;
  }
  
  try {
    await albumFormRef.value.validate();
    
    if (editingAlbum.value) {
        // 更新相册
        await albumApi.updateUserAlbum(editingAlbum.value.id, albumForm);
        message.success(t('albums.updateSuccess'));
        // 更新本地列表
      const index = albums.value.findIndex(a => a.id === editingAlbum.value?.id);
      if (index !== -1) {
        const existingAlbum = albums.value[index];
        if (existingAlbum) {
          albums.value[index] = {
            ...existingAlbum,
            ...albumForm,
            id: editingAlbum.value.id,
            user_id: existingAlbum.user_id,
            created_at: existingAlbum.created_at,
            updated_at: existingAlbum.updated_at,
          };
        }
      }
    } else {
        // 创建相册
        const response = await albumApi.createAlbum(albumForm);
        message.success(t('albums.createSuccess'));
        // 添加到本地列表
      albums.value.unshift({
        ...response.data,
        image_count: 0,
      });
    }
    
    handleCancel();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.saveFailed');
    message.error(errorMessage);
  } finally {
    saveLoading.value = false;
  }
};

// 处理删除相册
const handleDelete = (id: number) => {
  Modal.confirm({
    title: t('albums.confirmDelete'),
    content: t('albums.deleteConfirmContent'),
    onOk: async () => {
      try {
        await albumApi.deleteUserAlbum(id);
        message.success(t('albums.deleteSuccess'));
        // 从本地列表中移除
        albums.value = albums.value.filter(album => album.id !== id);
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.deleteFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理取消
const handleCancel = () => {
  showCreateModal.value = false;
  editingAlbum.value = null;
  albumForm.name = '';
  albumForm.description = '';
  albumForm.gallery_enabled = false;
  if (albumFormRef.value) {
    albumFormRef.value.resetFields();
  }
};

// 初始化
onMounted(() => {
  fetchAlbums();
});
</script>

<style scoped>
.albums-container {
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 顶部导航栏 */
.albums-header {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: 0 var(--spacing-xl);
  height: 64px; 
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--primary-color);
  margin: 0;
}

.create-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.create-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

/* 主要内容 */
.albums-main {
  padding: var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
}

/* 相册网格 */
.albums-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--spacing-lg);
}

/* 相册卡片 */
.album-card {
  cursor: pointer;
  transition: transform var(--transition-normal), box-shadow var(--transition-normal);
  overflow: hidden;
}

.album-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

/* 相册封面 */
.album-cover {
  width: 100%;
  height: 200px;
  border-radius: var(--border-radius-md);
  background: linear-gradient(135deg, var(--primary-light), var(--primary-color));
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--spacing-md);
  position: relative;
  overflow: hidden;
}

.cover-icon {
  font-size: 48px;
  color: white;
  opacity: 0.8;
}

.cover-text {
  position: absolute;
  bottom: var(--spacing-md);
  left: var(--spacing-md);
  right: var(--spacing-md);
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: white;
  text-align: center;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* 相册信息 */
.album-info {
  margin-bottom: var(--spacing-md);
}

.album-title {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-xs);
}

.album-description {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-sm);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.album-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
}

/* 相册操作 */
.album-actions {
  display: flex;
  gap: var(--spacing-sm);
  justify-content: flex-end;
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--border-color);
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xxl) 0;
  text-align: center;
}

.empty-icon {
  width: 96px;
  height: 96px;
  border-radius: var(--border-radius-full);
  background-color: var(--bg-dark);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  color: var(--text-tertiary);
  margin-bottom: var(--spacing-lg);
}

.empty-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.empty-description {
  font-size: var(--font-size-base);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-lg);
}

.empty-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.empty-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

.gallery-tooltip {
    position: relative;
    top: 2px;
    left: 5px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .albums-container {
    min-height: auto;
  }
  
  .albums-header {
    height: auto;
    padding: var(--spacing-md);
    flex-wrap: wrap;
    gap: var(--spacing-sm);
  }
  
  .create-button {
    flex: 1;
    min-width: 140px;
  }
  
  .albums-main {
    padding: var(--spacing-md);
  }
  
  .albums-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }
  
  .album-cover {
    height: 150px;
  }
  
  .cover-icon {
    font-size: 36px;
  }
  
  .album-info {
    margin-bottom: var(--spacing-sm);
  }
  
  .album-title {
    font-size: var(--font-size-base);
  }
  
  .album-description {
    font-size: var(--font-size-sm);
    -webkit-line-clamp: 1;
  }
  
  .album-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
  }
  
  .album-actions {
    flex-wrap: wrap;
    gap: var(--spacing-xs);
    padding-top: var(--spacing-sm);
  }
  
  .album-actions .ant-btn {
    flex: 1;
    min-width: 80px;
  }
  
  .empty-state {
    padding: var(--spacing-xl) 0;
  }
  
  .empty-icon {
    width: 72px;
    height: 72px;
    font-size: 36px;
  }
  
  .empty-title {
    font-size: var(--font-size-base);
  }
  
  .empty-description {
    font-size: var(--font-size-sm);
  }
}
</style>
