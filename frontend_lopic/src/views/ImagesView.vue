<template>
  <div class="images-container">
    <!-- 顶部导航栏 -->
    <header class="images-header card">
      <div class="header-left">
        <div v-if="images!=null && images.length > 0" class="header-selector">
          <a-button 
            :type="multiSelectMode ? 'primary' : 'default'"
            @click="toggleMultiSelectMode"
            class="multi-select-button"
          >
          <SelectOutlined v-if="multiSelectMode" />
          <BorderOutlined v-else />
            <!-- {{ multiSelectMode ? '取消多选' : '多选' }} -->
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
      <div class="header-right">
        <div class="deleting-status" v-if="deleting">
          <a-spin size="small">
            <span>{{ t('images.deleting') }}</span>
          </a-spin>
        </div>
        <div class="album-filter-container">
          <a-select
            v-model:value="selectedFilterAlbumId"
            :placeholder="t('images.selectAlbum')"
            class="album-filter-select"
            @change="handleAlbumFilterChange"
          >
            <a-select-option value="">{{ t('images.allImages') }}</a-select-option>
            <a-select-option
              v-for="album in userAlbums"
              :key="album.id"
              :value="album.id.toString()"
            >
              {{ album.name }}
            </a-select-option>
            <a-select-option value="-1">{{ t('images.noAlbum') }}</a-select-option>
          </a-select>
        </div>
        <div class="search-container">
          <a-input
            v-model:value="searchQuery"
            :placeholder="t('images.search')"
            class="search-input"
            @keyup.enter="handleSearch"
          >
            <template #prefix>
              <search-outlined />
            </template>
          </a-input>
          <a-button type="default" @click="handleSearch" class="search-button">
            {{ t('images.searchButton') }}
          </a-button>
          <a-button type="text" @click="resetSearch" class="reset-button">
            {{ t('images.resetButton') }}
          </a-button>
        </div>
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
            type="default" 
            @click="showAddToAlbumModal = true" 
            class="bottom-add-to-album-button"
          >
            {{ t('images.addToAlbum') }}
          </a-button>
          <a-button 
            type="danger" 
            @click="handleBatchDelete" 
            class="bottom-batch-delete-button"
          >
            <template #icon>
              <delete-outlined />
            </template>
            {{ t('images.batchDelete') }}
          </a-button>
        </div>
      </div>
    </div>

    <!-- 主要内容 -->
    <main class="images-main">
      <!-- 图片列表 -->
      <div v-if="images!=null && images.length > 0" class="images-grid">
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
                @load="image.loaded = true"
                @error="image.loaded = true"
              />
              <div v-if="!image.loaded" class="image-loader">
                <div class="loader-spinner"></div>
              </div>
            </div>
            <div class="image-actions" @click.stop>
              <a-dropdown>
                <a-button type="text" class="action-button">
                  <!-- 透明 -->
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
                      {{ t('images.delete') }}
                    </a-menu-item>
                  </a-menu>
                </template>
              </a-dropdown>
            </div>
          </div>
          <div class="image-info">
             <p class="image-filename">{{ image.original_name }}</p>
            <p class="image-meta">
              {{ formatFileSize(image.file_size) }} • {{ image.width }} x {{ image.height }} • {{ image.mime_type.split('/')[1] }}
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
        <h3 class="empty-title">{{ t('images.noImages') }}</h3>
        <p class="empty-description">{{ t('images.goToUpload') }}</p>
        <router-link to="/manage/upload">
          <a-button type="primary" class="empty-button">
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

    <!-- 批量添加到相册模态框 -->
    <a-modal
      v-model:open="showAddToAlbumModal"
      :title="t('images.batchAddToAlbum')"
      @ok="handleBatchAddToAlbum"
      @cancel="showAddToAlbumModal = false"
    >
      <a-form
        :model="batchAddForm"
        :rules="batchAddRules"
        ref="batchAddFormRef"
      >
        <a-form-item name="album_id" :label="t('images.selectAlbum')">
          <a-select
            v-model:value="batchAddForm.album_id"
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
        <div class="batch-add-info">
          {{ t('images.addingToAlbum', { count: selectedImages.length }) }}
        </div>
      </a-form>
    </a-modal>

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
import { ref, reactive, onMounted, onUnmounted, computed } from 'vue';
import { message, Modal } from 'ant-design-vue';
import { useI18n } from 'vue-i18n';
import {
  SearchOutlined,
  CameraOutlined,
  EditOutlined,
  DeleteOutlined,
  UploadOutlined,
  MoreOutlined,
  AimOutlined,
  SelectOutlined,
  BorderOutlined,
  LinkOutlined,
} from '@ant-design/icons-vue';
import { imageApi, albumApi } from '../api/services';
import { uploadWebSocketService } from '../api/websocket';
import type { ImageResponse, AlbumResponse } from '../types/api';
import { formatDateTime, formatFileSize, getFileUrl } from '../utils/index';
import { getErrorMessage } from '../types/errorMessages';
import ImagePreviewModal from '../components/ImagePreviewModal.vue';

const { t } = useI18n();

// 图片列表
const images = ref<ImageResponse[]>([]);
// 相册列表
const userAlbums = ref<AlbumResponse[]>([]);
// 选中的图片ID列表
const selectedImages = ref<number[]>([]);
// 加载状态
const loading = ref(false);
// 删除中状态
const deleting = ref(false);
// 分页信息
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
// 搜索信息
const searchQuery = ref('');
const isSearching = ref(false);
// 模态框状态
const showAddToAlbumModal = ref(false);
const showUpdateModal = ref(false);
// 预览模态框状态
const previewVisible = ref(false);
const previewImage = ref<ImageResponse | null>(null);
// 多选模式
const multiSelectMode = ref(false);
// 单个图片操作相关
const showSingleUpdateModal = ref(false);
const showSingleAddToAlbumModal = ref(false);
const currentImage = ref<ImageResponse | null>(null);
const singleAddForm = reactive({
  album_id: '',
});
const singleUpdateForm = reactive({
  original_name: '',
  tags: '',
});
// 相册过滤
const selectedFilterAlbumId = ref<string>('');
// 批量添加到相册表单
const batchAddFormRef = ref();
const batchAddForm = reactive({
  album_id: '',
});
// 批量更新表单
const batchUpdateForm = reactive({
  original_name: '',
  tags: '',
});
// 批量添加到相册验证规则
const batchAddRules = {
  album_id: [
    { required: true, message: () => t('images.selectAlbum'), trigger: 'blur' },
  ],
};
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

// 切换多选模式
const toggleMultiSelectMode = () => {
  multiSelectMode.value = !multiSelectMode.value;
  if (!multiSelectMode.value) {
    selectedImages.value = [];
  }
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

// 处理单个图片操作
const handleImageAction = (action: string, image: ImageResponse) => {
  currentImage.value = image;
  switch (action) {
    case 'update':
      // 填充更新表单
      singleUpdateForm.original_name = image.original_name;
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
    title: t('images.confirmDelete'),
    content: t('images.deleteContent'),
    onOk: async () => {
      try {
        deleting.value = true;
        imageApi.deleteUserImages([id]);
        message.warning(t('images.deleteStarted'));
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.deleteFailed');
        message.error(errorMessage);
        deleting.value = false;
      }
    },
  });
};

// 处理单个图片更新
const handleSingleUpdate = async () => {
  if (!currentImage.value) return;

  try {
    // 准备更新数据
    // 对Tags进行去重
    const uniqueTags = Array.from(new Set(singleUpdateForm.tags.split(',')));
    const tags = uniqueTags.map(tag => tag.trim()).filter(Boolean);
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

// 获取图片列表
const fetchImages = async () => {
  try {
    loading.value = true;
    
    // 保存当前已加载的图片ID集合
    const loadedImageIds = new Set(images.value?.map(img => img.id) || []);
    
    // 根据选中的相册过滤条件获取图片
    let newImages: any[] = [];
    let totalCount = 0;
    
    if (selectedFilterAlbumId.value === '') {
      // 获取所有图片
      const response = await imageApi.getUserImages({ 
        page: page.value, 
        page_size: pageSize.value 
      });
      newImages = response.data.images || [];
      totalCount = response.data.total;
    } else if (selectedFilterAlbumId.value === '-1') {
      // 获取无所属相册的图片
      const response = await albumApi.getImagesNotInAnyAlbum({
        page: page.value,
        page_size: pageSize.value
      });
      newImages = response.data.images || [];
      totalCount = response.data.total;
    } else {
      // 获取指定相册中的图片
      const albumId = parseInt(selectedFilterAlbumId.value);
      const response = await albumApi.getUserAlbumImages(albumId, {
        page: page.value,
        page_size: pageSize.value
      });
      newImages = response.data.images || [];
      totalCount = response.data.total;
    }
    
    // 处理图片加载状态：只有新图片设置为未加载
    images.value = newImages.map((image: any) => ({
      ...image,
      loaded: loadedImageIds.has(image.id) // 如果图片之前已加载，则保持加载状态
    }));
    
    total.value = totalCount;
    
    // 重置选中状态
    selectedImages.value = [];
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.fetchImagesFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 处理相册过滤变化
const handleAlbumFilterChange = () => {
  page.value = 1;
  fetchImages();
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

// 搜索图片
const handleSearch = async () => {
  if (!searchQuery.value.trim()) {
    resetSearch();
    return;
  }

  try {
    loading.value = true;
    isSearching.value = true;
    page.value = 1;

    // 保存当前已加载的图片ID集合
    const loadedImageIds = new Set(images.value?.map(img => img.id) || []);

    // 尝试按标签和标题搜索
    const response = await imageApi.searchImages({
      search_key: searchQuery.value,
      page: page.value,
      page_size: pageSize.value
    });

    // 处理图片加载状态：只有新图片设置为未加载
    images.value = response.data.images.map((image: any) => ({
      ...image,
      loaded: loadedImageIds.has(image.id) // 如果图片之前已加载，则保持加载状态
    }));
    total.value = response.data.total;
    // 重置选中状态
    selectedImages.value = [];
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.searchFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 重置搜索
const resetSearch = () => {
  searchQuery.value = '';
  isSearching.value = false;
  page.value = 1;
  fetchImages();
};

// 处理分页变化
const handlePageChange = (currentPage: number) => {
  page.value = currentPage;
  if (isSearching.value) {
    handleSearch();
  } else {
    fetchImages();
  }
};

// 处理每页大小变化
const handlePageSizeChange = (_current: number, size: number) => {
  pageSize.value = size;
  page.value = 1;
  if (isSearching.value) {
    handleSearch();
  } else {
    fetchImages();
  }
};

// 处理批量删除图片
const handleBatchDelete = () => {
  if (selectedImages.value.length === 0) return;

  Modal.confirm({
    title: t('images.confirmBatchDelete'),
    content: t('images.batchDeleteContent', { count: selectedImages.value.length }),
    onOk: async () => {
      try {
        deleting.value = true;
        imageApi.deleteUserImages(selectedImages.value);
        message.warning(t('images.batchDeleteStarted'));
        // 清空选中列表
        selectedImages.value = [];
      } catch (error: any) {
        const errorCode = error.response?.data?.code;
        const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.batchDeleteFailed');
        message.error(errorMessage);
        deleting.value = false;
      }
    },
  });
};

// 处理批量添加到相册
const handleBatchAddToAlbum = async () => {
  if (!batchAddFormRef.value) return;

  try {
    await batchAddFormRef.value.validate();

    // 调用添加图片到相册的API
    await imageApi.addImagesToAlbum({
      album_id: parseInt(batchAddForm.album_id.toString()),
      ids: selectedImages.value
    });
    message.success(t('images.batchAddToAlbumSuccess'));

    // 关闭模态框
    showAddToAlbumModal.value = false;
    // 重置表单
    batchAddForm.album_id = '';
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('images.batchAddToAlbumFailed');
    message.error(errorMessage);
  }
};

// 处理批量更新图片
const handleBatchUpdate = async () => {
  if (!selectedImages.value.length) return;

  try {
    // 准备更新数据
    // 对Tags进行去重
    const uniqueTags = Array.from(new Set(batchUpdateForm.tags.split(',')));
    const tags = uniqueTags.map(tag => tag.trim()).filter(Boolean);
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

// 组件挂载
onMounted(() => {
  fetchImages();
  fetchUserAlbums();
  
  // 连接 WebSocket
  uploadWebSocketService.connect();
  
  // 注册删除成功监听器
  uploadWebSocketService.on('deleteSuccess', () => {
    deleting.value = false;
    message.success(t('images.deleteSuccess'));
    fetchImages();
  });
  
  // 注册删除错误监听器
  uploadWebSocketService.on('deleteError', (data: { message: string, error: string, code: string }) => {
    deleting.value = false;
    const errorMessage = data.error ? getErrorMessage(data.error) : t('images.deleteFailed');
    message.error(errorMessage);
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
.images-container {
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 顶部导航栏 */
.images-header {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: 0 var(--spacing-xl);
  min-height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
  flex-wrap: wrap;
  gap: var(--spacing-md);
}

.header-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--primary-color);
  margin: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.header-selector {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.header-checkbox {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.deleting-status {
  display: flex;
  align-items: center;
  margin-right: var(--spacing-md);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius);
  background-color: var(--bg-light);
  font-size: var(--font-size-sm);
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

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  flex-wrap: wrap;
}

.search-container {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-wrap: wrap;
}

.album-filter-container {
  display: flex;
  align-items: center;
}

.album-filter-select {
  width: 200px;
  border-radius: var(--border-radius-md);
}

.search-input {
  width: 300px;
}

.upload-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
  white-space: nowrap;
  border-radius: var(--border-radius-md);
  transition: all var(--transition-fast);
}

.upload-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
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
.bottom-add-to-album-button,
.bottom-batch-delete-button {
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

.bottom-add-to-album-button:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

.bottom-batch-delete-button:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-sm);
}

/* 主要内容 */
.images-main {
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

.image-loader-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.image-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--transition-normal);
}

.image-loader {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-light);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1;
}

.loader-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border-color);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.image-card:hover .image-preview {
  transform: scale(1.05);
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

/* 分页容器 */
.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: var(--spacing-xl);
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



/* 批量添加到相册模态框 */
.batch-add-info {
  margin-top: var(--spacing-md);
  padding: var(--spacing-md);
  background-color: var(--bg-light);
  border-radius: var(--border-radius-md);
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
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

.preview-image {
  max-width: 100%;
  max-height: 50vh;
  object-fit: contain;
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-md);
  margin-bottom: var(--spacing-lg);
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
  .images-container {
    min-height: auto;
  }
  
  .images-header {
    height: auto;
    padding: var(--spacing-md);
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
  }
  
  .header-left {
    justify-content: flex-start;
  }
  
  .header-right {
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-sm);
  }
  
  .header-selector {
    flex-wrap: wrap;
    gap: var(--spacing-sm);
  }
  
  .search-container {
    flex-wrap: wrap;
    gap: var(--spacing-sm);
  }
  
  .search-input {
    flex: 1;
    min-width: 0;
  }
  
  .search-button,
  .reset-button {
    flex: 0 0 auto;
  }
  
  .album-filter-select {
    width: 100%;
  }
  
  .images-main {
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
</style>