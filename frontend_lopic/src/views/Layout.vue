<template>
  <BaseLayout layoutType="user" :pageTitle="t('layout.userPanel')">
    <template #sidebar>
      <ul class="nav-menu">
        <li class="nav-item">
          <router-link to="/manage/dashboard" class="nav-link">
            <dashboard-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.dashboard') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/manage/upload" class="nav-link">
            <upload-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.uploadImages') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/manage/images" class="nav-link">
            <camera-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.myImages') }}</span>
          </router-link>
        </li>
         <li class="nav-item">
          <router-link to="/manage/albums" class="nav-link">
            <aim-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.myAlbums') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/manage/profile" class="nav-link">
            <user-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.profileSettings') }}</span>
          </router-link>
        </li>
        <li class="nav-item" v-if="isAdmin">
          <router-link to="/admin/users" class="nav-link">
            <setting-outlined class="nav-icon" />
            <span class="nav-text">{{ t('layout.adminSystem') }}</span>
          </router-link>
        </li>
      </ul>
    </template>
    <template #footer>
      <div class="user-info">
        <span class="user-name">{{ currentUser?.username }}</span>
        <span class="user-role">{{ currentUser?.role.name === 'admin' ? t('layout.admin') : t('layout.user') }}</span>
      </div>
      <a-button type="text" @click="handleLogout" class="logout-button">
        <template #icon>
          <user-outlined />
        </template>
        {{ t('layout.logout') }}
      </a-button>
    </template>
    <router-view />
  </BaseLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { message } from 'ant-design-vue';
import {
  DashboardOutlined,
  AimOutlined,
  UploadOutlined,
  CameraOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons-vue';
import { useAuthStore } from '../stores/auth';
import { authApi } from '../api/services';
import BaseLayout from '../components/BaseLayout.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 路由和状态管理
const router = useRouter();
const authStore = useAuthStore();

// 当前用户
const currentUser = computed(() => authStore.currentUser);
const isAdmin = computed(() => authStore.isAdmin);

// 退出登录
const handleLogout = async () => {
  try {
    // 调用 logout API
    await authApi.logout({ refresh_token: localStorage.getItem('refresh_token') || '' });
    
    // 清除本地状态
    authStore.logout();
    message.success('退出登录成功');
    router.push('/login');
  } catch (error) {
    console.error('Logout error:', error);
    // 即使 API 调用失败，也清除本地状态
    authStore.logout();
    router.push('/login');
  }
};


</script>

<style scoped>
/* 导航菜单样式 */
.nav-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.nav-item {
  margin-bottom: var(--spacing-xs);
}

.nav-link {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-xl);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all var(--transition-fast);
  border-radius: 0 var(--border-radius-md) var(--border-radius-md) 0;
  margin-right: var(--spacing-md);
  cursor: pointer;
}

.nav-icon {
  font-size: 18px;
  width: 24px;
  text-align: center;
}

.nav-link:hover {
  color: var(--primary-color);
  background-color: var(--primary-light);
}

.nav-link.router-link-active {
  color: var(--primary-dark);
  background-color: var(--primary-light);
  font-weight: 500;
}

.nav-text {
  font-size: var(--font-size-base);
}

/* 用户信息样式 */
.user-info {
  display: flex;
  justify-content: center;  
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.user-name {
  font-weight: 500;
  color: var(--text-primary);
}

.user-role {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  background-color: var(--bg-light);
  padding: 2px var(--spacing-sm);
  border-radius: var(--border-radius-sm);
}

/* 退出按钮样式 */
.logout-button {
  color: var(--text-secondary);
  width: 100%;
  justify-content: flex-start;
  gap: var(--spacing-md);
}

.logout-button:hover {
  color: var(--error-color);
  background-color: var(--error-light);
}

/* 响应式设计 */
@media (max-width: 768px) {
  /* 保持PC端样式，添加滚动支持 */
  .nav-menu {
    max-height: 70vh;
    overflow-y: auto;
  }
  
  /* 自定义滚动条 */
  .nav-menu::-webkit-scrollbar {
    width: 4px;
  }
  
  .nav-menu::-webkit-scrollbar-track {
    background: var(--bg-light);
    border-radius: 2px;
  }
  
  .nav-menu::-webkit-scrollbar-thumb {
    background: var(--border-color);
    border-radius: 2px;
  }
  
  .nav-menu::-webkit-scrollbar-thumb:hover {
    background: var(--text-secondary);
  }
}
</style>