<template>
  <div class="album-detail-container">
    <!-- 顶部导航栏 -->
    <header class="album-detail-header card">
      <div class="header-left">
        <a-button type="text" @click="goBack" class="back-button">
          <template #icon>
            <left-outlined />
          </template>
        </a-button>
        <div class="album-info">
          <div>
            <h1 class="header-title">{{ album?.name }}</h1>
            <span class="album-description">{{ album?.description || t('albums.noDescription') }}</span>
          </div>
          <div class="album-meta" v-if="album">
            <span class="album-date">{{ t('albums.createdAt') }} {{ formatDateTime(album.created_at) }}</span>
            <span class="album-count">{{ images.length }} {{ t('albums.images') }}</span>
          </div>
          <div v-if="images.length > 0" class="header-selector">
            <a-button 
              :type="multiSelectMode ? 'primary' : 'default'"
              @click="toggleMultiSelectMode"
              class="multi-select-button"
            >
              <SelectOutlined v-if="multiSelectMode" />
              <BorderOutlined v-else />
            </a-button>
            <a-checkbox 
              v-if="multiSelectMode"
              v-model:checked="selectAll" 
              @change="handleSelectAll" 
              class="header-checkbox"
            >
              {{ t('images.selectAll') }}
            </a-checkbox>
          </div>
        </div>
      </div>
      <div class="header-right">
        <router-link to="/manage/upload">
          <a-button type="primary" class="upload-button">
            <template #icon>
              <upload-outlined />
            </template>
            {{ t('images.uploadImage') }}
          </a-button>
        </router-link>
      </div>
    </header>

    <!-- 底部批量操作栏 -->
    <div v-if="selectedImages.length > 0" class="bottom-batch-actions">
      <div class="bottom-batch-actions-content">
        <div class="bottom-batch-actions-left">
          <span class="bottom-selected-count">
            {{ t('images.selectedCount', { count: selectedImages.length }) }}
          </span>
        </div>
        <div class="bottom-batch-actions-right">
          <a-button 
            type="default" 
            @click="showUpdateModal = true" 
            class="bottom-batch-update-button"
          >
            <template #icon>
              <edit-outlined />
            </template>
            {{ t('images.batchUpdate') }}
          </a-button>
          <a-button 
            type="danger" 
            @click="handleBatchRemove" 
            class="bottom-batch-remove-button"
          >
            <template #icon>
              <delete-outlined />
            </template>
            {{ t('albums.batchRemove') }}
          </a-button>
        </div>
      </div>
    </div>

    <!-- 批量更新图片模态框 -->
    <a-modal
      v-model:open="showUpdateModal"
      :title="t('images.batchUpdate')"
      @ok="handleBatchUpdate"
      @cancel="showUpdateModal = false"
    >
      <a-form
        :model="batchUpdateForm"
      >
        <a-form-item :label="t('images.imageName')">
          <a-input
            v-model:value="batchUpdateForm.original_name"
            :placeholder="t('images.enterImageNameOptional')"
          />
        </a-form-item>
        <a-form-item :label="t('images.tags')">
          <a-input
            v-model:value="batchUpdateForm.tags"
            :placeholder="t('images.enterTagsOptional')"
          />
        </a-form-item>
        <div class="batch-update-info">
          {{ t('images.updatingImages', { count: selectedImages.length }) }}
        </div>
      </a-form>
    </a-modal>

    <!-- 主要内容 -->
    <main class="album-detail-main">
      <!-- 图片列表 -->
      <div v-if="images.length > 0" class="images-grid">
        <div
          v-for="image in images"
          :key="image.id"
          class="image-card"
          @click="!multiSelectMode && handleImageClick(image)"
        >
          <div class="image-container">
            <div v-if="multiSelectMode" class="image-checkbox" @click.stop>
              <a-checkbox 
                :checked="selectedImages.includes(image.id)" 
                @change="(e: any) => handleImageSelect(image.id, e.target.checked)"
              />
            </div>
            <div class="image-loader-container">
              <img 
                :src="getFileUrl(image.thumbnail_url)" 
                :alt="image.original_name" 
                class="image-preview" 
                :class="{ 'image-loaded': image.loaded }"
                @load="image.loaded = true"
                @error="image.loaded = true"
              />
              <div v-if="!image.loaded" class="image-loading-spinner">
                <div class="spinner"></div>
              </div>
            </div>
            <div class="image-actions" @click.stop>
              <a-dropdown>
                <a-button type="text" class="action-button">
                  <template #icon>
                    <more-outlined />
                  </template>
                </a-button>
                <template #overlay>
                  <a-menu @click="(e: any) => handleImageAction(e.key, image)">
                    <a-menu-item key="update">
                      <template #icon>
                        <edit-outlined />
                      </template>
                      {{ t('images.update') }}
                    </a-menu-item>
                    <a-menu-item key="addToAlbum">
                      <template #icon>
                        <aim-outlined />
                      </template>
                      {{ t('images.addToAlbum') }}
                    </a-menu-item>
                    <a-menu-item key="copyLink">
                      <template #icon>
                        <link-outlined />
                      </template>
                      {{ t('images.copyLink') }}
                    </a-menu-item>
                    <a-menu-item key="delete" danger>
                      <template #icon>
                        <delete-outlined />
                      </template>
                      {{ t('albums.remove') }}
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
          </div>
          <div class="image-info">
            <p class="image-filename">{{ image.original_name || image.file_name }}</p>
            <p class="image-meta">
              {{ formatFileSize(image.file_size) }} • {{ image.width }} x {{ image.height }}
            </p>
            <p class="image-meta">{{ formatDateTime(image.created_at) }}</p>
            
            <div class="image-tags" v-if="image.tags && image.tags.length > 0">
              <a-tag 
                v-for="(tag, index) in image.tags" 
                :key="index"
                size="small"
                color="blue"
              >
                {{ tag }}
              </a-tag>
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
    <div v-else class="empty-state">
      <div class="empty-icon">
        <camera-outlined />
      </div>
      <h3 class="empty-title">{{ t('albums.noImagesInAlbum') }}</h3>
      <p class="empty-description">{{ t('albums.uploadImagesPrompt') }}</p>
      <router-link to="/manage/upload">
        <a-button
          type="primary"
          class="empty-button"
        >
          <template #icon>
            <upload-outlined />
          </template>
          {{ t('images.uploadImage') }}
        </a-button>
      </router-link>
    </div>

    <!-- 分页 -->
    <div v-if="total > 0" class="pagination-container">
      <a-pagination
        v-model:current="page"
        v-model:page-size="pageSize"
        :total="total"
        :show-size-changer="true"
        :page-size-options="['10', '20', '50', '100']"
        @change="handlePageChange"
        @showSizeChange="handlePageSizeChange"
      />
    </div>
  </main>

    <!-- 图片预览模态框 -->
    <ImagePreviewModal
      :visible="previewVisible"
      :image="previewImage"
      @update:visible="(value) => previewVisible = value"
      @close="handlePreviewClose"
    />

    <!-- 单个图片更新模态框 -->
    <a-modal
      v-model:open="showSingleUpdateModal"
      :title="t('images.updateImage')"
      @ok="handleSingleUpdate"
      @cancel="showSingleUpdateModal = false"
    >
      <a-form
        :model="singleUpdateForm"
      >
        <a-form-item :label="t('images.imageName')">
          <a-input
            v-model:value="singleUpdateForm.original_name"
            :placeholder="t('images.enterImageName')"
          />
        </a-form-item>
        <a-form-item :label="t('images.tags')">
          <a-input
            v-model:value="singleUpdateForm.tags"
            :placeholder="t('images.enterTagsOptional')"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 单个图片添加到相册模态框 -->
    <a-modal
      v-model:open="showSingleAddToAlbumModal"
      :title="t('images.addToAlbum')"
      @ok="handleSingleAddToAlbum"
      @cancel="showSingleAddToAlbumModal = false"
    >
      <a-form
        :model="singleAddForm"
      >
        <a-form-item :label="t('images.selectAlbum')">
          <a-select
            v-model:value="singleAddForm.album_id"
            :placeholder="t('images.selectAlbumPlaceholder')"
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
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { message, Modal } from 'ant-design-vue';
import { useI18n } from 'vue-i18n';
import {
  LeftOutlined,
  UploadOutlined,
  CameraOutlined,
  DeleteOutlined,
  EditOutlined,
  SelectOutlined,
  BorderOutlined,
  MoreOutlined,
  AimOutlined,
  LinkOutlined,
} from '@ant-design/icons-vue';
import { albumApi, imageApi } from '../api/services';
import type { AlbumResponse, ImageResponse } from '../types/api';
import { getFileUrl,formatDateTime,formatFileSize } from '../utils/index';
import { getErrorMessage } from '../types/errorMessages';
import ImagePreviewModal from '../components/ImagePreviewModal.vue';

const { t } = useI18n();

// 路由
const router = useRouter();
const route = useRoute();

// 相册ID
const albumId = computed(() => parseInt(route.params.id as string));

// 相册信息
const album = ref<AlbumResponse | null>(null);

// 图片列表
const images = ref<ImageResponse[]>([]);

// 选中的图片ID列表
const selectedImages = ref<number[]>([]);

// 加载状态
const loading = ref(false);

// 分页信息
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

// 模态框状态
const showUpdateModal = ref(false);
const showSingleUpdateModal = ref(false);
const showSingleAddToAlbumModal = ref(false);
// 预览模态框状态
const previewVisible = ref(false);
const previewImage = ref<ImageResponse | null>(null);
// 多选模式
const multiSelectMode = ref(false);
// 单个图片操作相关
const currentImage = ref<ImageResponse | null>(null);
const singleAddForm = reactive({
  album_id: '',
});
const singleUpdateForm = reactive({
  original_name: '',
  tags: '',
});
// 相册列表
const userAlbums = ref<AlbumResponse[]>([]);

// 批量更新表单
const batchUpdateForm = reactive({
  original_name: '',
  tags: '',
});

// 全选状态
const selectAll = computed({
  get: () => {
    return images.value.length > 0 && selectedImages.value.length === images.value.length;
  },
  set: (value) => {
    if (value) {
      selectedImages.value = images.value.map(image => image.id);
    } else {
      selectedImages.value = [];
    }
  }
});

// 返回相册列表
const goBack = () => {
  router.push('/manage/albums');
};

// 切换多选模式
const toggleMultiSelectMode = () => {
  multiSelectMode.value = !multiSelectMode.value;
  if (!multiSelectMode.value) {
    selectedImages.value = [];
  }
};

// 处理图片点击，显示预览模态框
const handleImageClick = (image: ImageResponse) => {
  previewImage.value = image;
  previewVisible.value = true;
};

// 处理预览模态框关闭
const handlePreviewClose = () => {
  previewVisible.value = false;
  previewImage.value = null;
};

// 处理图片选择
const handleImageSelect = (imageId: number, checked: boolean) => {
  if (checked) {
    selectedImages.value.push(imageId);
  } else {
    selectedImages.value = selectedImages.value.filter(id => id !== imageId);
  }
};

// 处理全选
const handleSelectAll = (e: any) => {
  selectAll.value = e.target.checked;
};

// 获取相册信息
const fetchAlbumInfo = async () => {
  try {
    const response = await albumApi.getUserAlbum(albumId.value);
    album.value = response.data;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.fetchAlbumFailed');
    message.error(errorMessage);
  }
};

// 获取相册图片
const fetchAlbumImages = async () => {
  try {
    loading.value = true;
    
    // 保存当前已加载的图片ID集合
    const loadedImageIds = new Set(images.value?.map(img => img.id) || []);
    
    const response = await albumApi.getUserAlbumImages(albumId.value, { 
      page: page.value, 
      page_size: pageSize.value 
    });
    
    // 处理图片加载状态：只有新图片设置为未加载
    images.value = (response.data.images || []).map((image: any) => ({
      ...image,
      loaded: loadedImageIds.has(image.id) // 如果图片之前已加载，则保持加载状态
    }));
    total.value = response.data.total;
    // 重置选中状态
    selectedImages.value = [];
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.fetchImagesFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理批量移除图片
const handleBatchRemove = () => {
  if (selectedImages.value.length === 0) return;

  Modal.confirm({
    title: t('albums.confirmBatchRemove'),
    content: t('albums.batchRemoveContent', { count: selectedImages.value.length }),
    onOk: async () => {
      try {
        // 调用从相册移除图片的API
        await imageApi.removeImagesFromAlbum({
          album_id: albumId.value,
          ids: selectedImages.value
        });
        message.success(t('albums.batchRemoveSuccess'));
        // 从本地列表中移除
        images.value = images.value.filter(image => !selectedImages.value.includes(image.id));
        // 清空选中列表
        selectedImages.value = [];
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.batchRemoveFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 复制到剪贴板
const copyToClipboard = async (text: string) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
      message.success(t('images.linkCopied'));
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.left = '-9999px';
      textArea.style.top = '0';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      const successful = document.execCommand('copy');
      document.body.removeChild(textArea);
      if (successful) {
        message.success(t('images.linkCopied'));
      } else {
        message.error(t('images.copyLinkFailed'));
      }
    }
  } catch {
    message.error(t('images.copyLinkFailed'));
  }
};

// 处理图片操作
const handleImageAction = (action: string, image: ImageResponse) => {
  currentImage.value = image;
  switch (action) {
    case 'update':
      // 填充更新表单
      singleUpdateForm.original_name = image.original_name || '';
      singleUpdateForm.tags = image.tags?.join(',') || '';
      showSingleUpdateModal.value = true;
      break;
    case 'addToAlbum':
      showSingleAddToAlbumModal.value = true;
      break;
    case 'copyLink':
      // 复制图片链接到剪贴板
      const imageUrl = getFileUrl(image.file_url);
      copyToClipboard(imageUrl);
      break;
    case 'delete':
      handleSingleDelete(image.id);
      break;
  }
};

// 处理单个图片删除
const handleSingleDelete = (id: number) => {
  Modal.confirm({
    title: t('albums.confirmRemove'),
    content: t('albums.removeContent'),
    onOk: async () => {
      try {
        // 调用从相册移除图片的API
        await imageApi.removeImagesFromAlbum({
          album_id: albumId.value,
          ids: [id]
        });
        message.success(t('albums.removeSuccess'));
        // 从本地列表中移除
        images.value = images.value.filter(image => image.id !== id);
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.removeFailed');
        message.error(errorMessage);
      }
    },
  });
};

// 处理单个图片更新
const handleSingleUpdate = async () => {
  if (!currentImage.value) return;

  try {
    // 准备更新数据
    const tags = singleUpdateForm.tags.split(',').map((tag: string) => tag.trim()).filter(Boolean);
    const updateData: any = {
      ids: [currentImage.value.id],
      tags,
    };

    // 如果填写了图片名称，则添加到更新数据中
    if (singleUpdateForm.original_name.trim()) {
      updateData.original_name = singleUpdateForm.original_name.trim();
    }

    // 调用更新API
    await imageApi.updateUserImage(updateData);
    message.success(t('images.updateSuccess'));

    // 更新本地列表
    images.value = images.value.map(image => {
      if (image.id === currentImage.value?.id) {
        const updatedImage = {
          ...image,
          tags,
        };
        
        // 如果填写了图片名称，则更新本地数据
        if (singleUpdateForm.original_name.trim()) {
          updatedImage.original_name = singleUpdateForm.original_name.trim();
        }
        
        return updatedImage;
      }
      return image;
    });

    // 关闭模态框
    showSingleUpdateModal.value = false;
    // 重置表单
    singleUpdateForm.original_name = '';
    singleUpdateForm.tags = '';
    currentImage.value = null;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.updateFailed');
    message.error(errorMessage);
  }
};

// 处理单个图片添加到相册
const handleSingleAddToAlbum = async () => {
  if (!currentImage.value || !singleAddForm.album_id) return;

  try {
    // 调用添加图片到相册的API
    await imageApi.addImagesToAlbum({
      album_id: parseInt(singleAddForm.album_id.toString()),
      ids: [currentImage.value.id]
    });
    message.success(t('images.addToAlbumSuccess'));

    // 关闭模态框
    showSingleAddToAlbumModal.value = false;
    // 重置表单
    singleAddForm.album_id = '';
    currentImage.value = null;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.addToAlbumFailed');
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
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('albums.fetchAlbumsFailed');
    message.error(errorMessage);
  }
};

// 处理批量更新图片
const handleBatchUpdate = async () => {
  if (!selectedImages.value.length) return;

  try {
    // 准备更新数据
    const tags = batchUpdateForm.tags.split(',').map((tag: string) => tag.trim()).filter(Boolean);
    const updateData: any = {
      ids: selectedImages.value,
      tags,
    };

    // 如果填写了图片名称，则添加到更新数据中
    if (batchUpdateForm.original_name.trim()) {
      updateData.original_name = batchUpdateForm.original_name.trim();
    }

    // 调用更新API
    await imageApi.updateUserImage(updateData);
    message.success(t('images.batchUpdateSuccess'));

    // 更新本地列表
    images.value = images.value.map(image => {
      if (selectedImages.value.includes(image.id)) {
        const updatedImage = {
          ...image,
          tags,
        };
        
        // 如果填写了图片名称，则更新本地数据
        if (batchUpdateForm.original_name.trim()) {
          updatedImage.original_name = batchUpdateForm.original_name.trim();
        }
        
        return updatedImage;
      }
      return image;
    });

    // 关闭模态框
    showUpdateModal.value = false;
    // 重置表单
    batchUpdateForm.original_name = '';
    batchUpdateForm.tags = '';
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.batchUpdateFailed');
    message.error(errorMessage);
  }
};

// 处理分页变化
const handlePageChange = (currentPage: number) => {
  page.value = currentPage;
  fetchAlbumImages();
};

// 处理每页大小变化
const handlePageSizeChange = (_current: number, size: number) => {
  pageSize.value = size;
  page.value = 1;
  fetchAlbumImages();
};

// 初始化
onMounted(() => {
  fetchAlbumInfo();
  fetchAlbumImages();
  fetchUserAlbums();
});
</script>

<style scoped>
.album-detail-container {
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 顶部导航栏 */
.album-detail-header {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: 0 var(--spacing-xl);
  min-height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-lg);
  flex: 1;
  min-width: 300px;
}

.album-info {
  width: 100%;
  display: flex;
  justify-content: space-between;
}

.back-button {
  color: var(--text-secondary);
  margin-top: 2px;
}

.back-button:hover {
  color: var(--text-primary);
}

.header-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 var(--spacing-xs) 0;
}

.album-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-lg);
  font-size: var(--font-size-sm);
  color: var(--text-tertiary);
  align-items: center;
}

.album-description {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  flex: 1;
  min-width: 200px;
}

.upload-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
  white-space: nowrap;
}

.upload-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

.header-selector {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-top: var(--spacing-sm);
}

.header-checkbox {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.header-selected-count {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  font-weight: 500;
  background-color: var(--bg-light);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius-full);
  border: 1px solid var(--border-color);
}

/* 底部批量操作栏 */
.bottom-batch-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: var(--bg-color);
  border-top: 1px solid var(--border-color);
  box-shadow: var(--shadow-lg);
  z-index: 1000;
  animation: slide-up 0.3s ease-out;
}

@keyframes slide-up {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.bottom-batch-actions-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
  gap: var(--spacing-lg);
}

.bottom-batch-actions-left {
  flex: 1;
}

.bottom-selected-count {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-primary);
}

.bottom-batch-actions-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.bottom-batch-update-button,
.bottom-batch-remove-button {
  white-space: nowrap;
  border-radius: var(--border-radius-md);
  transition: all var(--transition-fast);
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-sm);
}

.bottom-batch-update-button:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.bottom-batch-remove-button:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 主要内容 */
.album-detail-main {
  padding: var(--spacing-xl);
  max-width: 1200px;
  margin: 0 auto;
}

/* 图片网格 */
.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

/* 图片卡片 */
.image-card {
  background-color: var(--bg-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  transition: transform var(--transition-normal), box-shadow var(--transition-normal);
}

.image-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-md);
}

/* 图片容器 */
.image-container {
  position: relative;
  width: 100%;
  height: 160px;
  overflow: hidden;
}

.image-checkbox {
  position: absolute;
  top: var(--spacing-sm);
  left: var(--spacing-sm);
  z-index: 10;
}

.image-checkbox .ant-checkbox {
  background-color: rgba(255, 255, 255, 0.9);
  border-radius: 2px;
}

.image-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.image-card:hover .image-preview {
  transform: scale(1.05);
}

/* 图片加载容器 */
.image-loader-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

/* 图片加载状态 */
.image-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: opacity var(--transition-normal), transform var(--transition-normal);
  opacity: 0;
}

.image-preview.image-loaded {
  opacity: 1;
}

/* 加载中动画 */
.image-loading-spinner {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(255, 255, 255, 0.1);
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-radius: 50%;
  border-top-color: var(--primary-color);
  animation: spin 1s ease-in-out infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}



/* 图片信息 */
.image-info {
  padding: var(--spacing-md);
}

.image-filename {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: var(--spacing-xs);
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.image-meta {
  font-size: var(--font-size-xs);
  color: var(--text-tertiary);
  margin-bottom: var(--spacing-sm);
}

.image-tags {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-xs);
  max-height: 32px;
  overflow: hidden;
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

/* 分页容器 */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: var(--spacing-xl);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .album-detail-container {
    min-height: auto;
  }
  
  .album-detail-header {
    height: auto;
    padding: var(--spacing-md);
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
  }
  
  .header-left {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .back-button {
    padding: 0;
    margin-bottom: var(--spacing-xs);
  }
  
  .album-info {
    width: 100%;
  }
  
  .header-title {
    font-size: var(--font-size-lg);
  }
  
  .album-description {
    font-size: var(--font-size-sm);
  }
  
  .album-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
    font-size: var(--font-size-xs);
  }
  
  .header-selector {
    margin-top: var(--spacing-sm);
  }
  
  .header-right {
    width: 100%;
  }
  
  .upload-button {
    width: 100%;
  }
  
  .album-detail-main {
    padding: var(--spacing-md);
  }
  
  .images-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-sm);
  }
  
  .image-card {
    margin-bottom: 0;
  }
  
  .image-container {
    height: 120px;
  }
  
  .image-preview {
    height: 120px;
  }
  
  .image-info {
    padding: var(--spacing-sm);
  }
  
  .image-filename {
    font-size: var(--font-size-xs);
  }
  
  .image-meta {
    font-size: 10px;
    margin-bottom: var(--spacing-xs);
  }
  
  .image-tags {
    max-height: 24px;
  }
  
  .image-tags .ant-tag {
    font-size: 10px;
    padding: 0 4px;
    margin: 0;
  }
  
  .bottom-batch-actions {
    padding: var(--spacing-sm);
  }
  
  .bottom-batch-actions-content {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .bottom-batch-actions-left {
    text-align: center;
  }
  
  .bottom-batch-actions-right {
    flex-wrap: wrap;
    justify-content: center;
    gap: var(--spacing-xs);
  }
  
  .bottom-batch-actions-right .ant-btn {
    flex: 1;
    min-width: 100px;
    font-size: var(--font-size-xs);
    padding: var(--spacing-xs) var(--spacing-sm);
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
  
  .pagination-container {
    margin-top: var(--spacing-md);
  }
  
  .pagination-container .ant-pagination {
    flex-wrap: wrap;
    justify-content: center;
  }
}

/* 多选按钮样式 */
.multi-select-button {
  border: none;
  box-shadow: var(--shadow-md);
  border-radius: var(--border-radius-md);
  transition: all var(--transition-fast);
}

.multi-select-button:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 图片操作按钮 */
.image-actions {
  position: absolute;
  bottom: var(--spacing-sm);
  right: var(--spacing-sm);
  z-index: 10;
}

.action-button {
  width: 32px;
  height: 32px;
  min-width: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  box-shadow: var(--shadow-sm);
  transition: all var(--transition-fast);
}

.action-button:hover {
  background-color: var(--primary-color);
  color: white;
  transform: scale(1.1);
}

.action-button .anticon {
  font-size: 14px;
}
</style>
