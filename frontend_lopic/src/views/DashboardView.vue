<template>
  <div class="dashboard-container">
    <!-- 欢迎卡片 -->
    <div class="welcome-card card">
      <h2 class="card-title">{{ t('dashboard.welcome') }}，{{ currentUser?.username }}！</h2>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card card">
        <div class="stat-icon">
          <aim-outlined />
        </div>
        <div class="stat-content">
          <h3 class="stat-title">{{ t('dashboard.myAlbums') }}</h3>
          <p class="stat-value">{{ albumCount }} / {{ maxAlbums == -1 ? t('dashboard.unlimited') : maxAlbums }}</p>
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-icon">
          <camera-outlined />
        </div>
        <div class="stat-content">
          <h3 class="stat-title">{{ t('dashboard.myImages') }}</h3>
          <p class="stat-value">{{ imageCount }}</p>
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-icon">
          <tag-outlined />
        </div>
        <div class="stat-content">
          <h3 class="stat-title">{{ t('dashboard.usedStorage') }}</h3>
          <p class="stat-value">{{ formatFileSize(totalSize) }}</p>
        </div>
      </div>
      <div class="stat-card card">
        <div class="stat-icon">
          <clock-circle-outlined />
        </div>
        <div class="stat-content">
          <h3 class="stat-title">{{ t('dashboard.storageSpace') }}</h3>
          <p class="stat-value">{{ maxStorageSize == -1 ? t('dashboard.unlimited') : maxStorageSize + ' MB' }}</p>
        </div>
      </div>
    </div>

    <!-- 标签云 -->
    <div class="tag-cloud-card card">
      <h3 class="card-title">{{ t('dashboard.myTags') }}</h3>
      <div class="tag-cloud-container">
        <div v-if="tagCloud.length > 0" class="tag-cloud-wrapper">
          <sphere-tag-cloud :tags="tagCloud" @tag-click="handleTagClick"></sphere-tag-cloud>
        </div>
        <div v-else class="tag-cloud-empty">
          <p>{{ t('dashboard.noTags') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { AimOutlined, CameraOutlined, TagOutlined, ClockCircleOutlined } from '@ant-design/icons-vue';
import { albumApi, userApi } from '../api/services';
import { formatFileSize } from '../utils/index';
import type { User, TagCloudItem } from '../types/api';
import SphereTagCloud from '../components/SphereTagCloud.vue';

const { t } = useI18n();

// 当前用户
const currentUser = ref<User | null>(null);

// 统计数据
const albumCount = ref(0);
const imageCount = ref(0);
const totalSize = ref(0);
const maxStorageSize = ref(0);
const maxAlbums = ref(0);

// 标签云数据
const tagCloud = ref<TagCloudItem[]>([]);

// 获取用户信息
const fetchUserInfo = async () => {
  try {
    const userResponse = await userApi.getCurrentUser();
    if (userResponse.data) {
      currentUser.value = userResponse.data || null;
      if (currentUser.value) {
        imageCount.value = currentUser.value.image_count || 0;
        totalSize.value = currentUser.value.total_size || 0;
        maxStorageSize.value = currentUser.value.role?.max_storage_size_mb || 0;
        maxAlbums.value = currentUser.value.role?.max_albums_per_user || 0;
      }
    }
  } catch (error) {
    console.error('Failed to fetch user info:', error);
  }
};

// 获取统计数据
const fetchStats = async () => {
  try {
    // 获取相册数量
    const albumsResponse = await albumApi.getUserAlbums({ page: 1, page_size: 1 });
    albumCount.value = albumsResponse.data.total;
  } catch (error) {
    console.error('Failed to fetch stats:', error);
  }
};

// 获取标签云数据
const fetchTagCloud = async () => {
  try {
    const tagCloudResponse = await userApi.getCurrentUserTagsCloud();
    tagCloud.value = tagCloudResponse.data || [];
  } catch (error) {
    console.error('Failed to fetch tag cloud:', error);
  }
};

// 处理标签点击事件
const handleTagClick = (tag: TagCloudItem) => {
  console.log('Clicked tag:', tag.tag);
  // 可以在这里添加标签点击后的逻辑，比如搜索该标签的图片
};

// 初始化
onMounted(async () => {
  await fetchUserInfo();
  await fetchStats();
  await fetchTagCloud();
});
</script>

<style scoped>
.dashboard-container {
  width: 100%;
}

/* 欢迎卡片 */
.welcome-card {
  margin-bottom: var(--spacing-xl);
}

.card-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
}

.card-subtitle {
  font-size: var(--font-size-base);
  color: var(--text-secondary);
  margin: 0;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.stat-icon {
  width: 64px;
  height: 64px;
  border-radius: var(--border-radius-md);
  background-color: var(--primary-light);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: var(--primary-color);
}

.stat-content {
  flex: 1;
}

.stat-title {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--text-secondary);
  margin-bottom: var(--spacing-xs);
}

.stat-value {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}



/* 标签云 */
.tag-cloud-card {
  margin-bottom: var(--spacing-xl);
}

.tag-cloud-container {
  margin-top: var(--spacing-md);
}

.tag-cloud-wrapper {
  border-radius: var(--border-radius-md);
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tag-cloud-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  background-color: var(--bg-light);
  border-radius: var(--border-radius-md);
  min-height: 300px;
  color: var(--text-tertiary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-container {
    padding: 0;
  }
  
  .welcome-card {
    margin-bottom: var(--spacing-md);
  }
  
  .card-title {
    font-size: var(--font-size-base);
  }
  
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: var(--spacing-md);
    margin-bottom: var(--spacing-md);
  }
  
  .stat-card {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
    padding: var(--spacing-md);
  }
  
  .stat-icon {
    width: 48px;
    height: 48px;
    font-size: 20px;
  }
  
  .stat-title {
    font-size: var(--font-size-sm);
  }
  
  .stat-value {
    font-size: var(--font-size-lg);
  }
  
  .tag-cloud-card {
    margin-bottom: var(--spacing-md);
  }
  
  .tag-cloud-wrapper {
    min-height: 200px;
  }
  
  .tag-cloud-empty {
    min-height: 200px;
    padding: var(--spacing-lg);
  }
}
</style>
