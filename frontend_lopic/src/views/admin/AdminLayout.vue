<template>
  <BaseLayout 
    layoutType="admin" 
    :pageTitle="t('admin.pageTitle')"
    :showFooter="false"
    :showHeaderRight="true"
  >
    <template #sidebar>
      <ul class="nav-menu">
        <li class="nav-item">
          <router-link to="/admin/users" class="nav-link">
            <user-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.users') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/admin/roles" class="nav-link">
            <team-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.roles') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/admin/albums" class="nav-link">
            <aim-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.albums') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/admin/images" class="nav-link">
            <camera-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.images') }}</span>
          </router-link>
        </li>
                <li class="nav-item">
          <router-link to="/admin/storage" class="nav-link">
            <database-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.storage') }}</span>
          </router-link>
        </li>
         <li class="nav-item">
          <router-link to="/admin/backup" class="nav-link">
            <file-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.backup') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/admin/system" class="nav-link">
            <setting-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.system') }}</span>
          </router-link>
        </li>
        <li class="nav-item">
          <router-link to="/manage/dashboard" class="nav-link">
            <rollback-outlined class="nav-icon" />
            <span class="nav-text">{{ t('admin.menu.dashboard') }}</span>
          </router-link>
        </li>
      </ul>
    </template>
    <template #header-right>
      <div class="user-info">
        <span class="user-name">{{ currentUser?.username }}</span>
        <span class="user-role">{{ t('admin.userRole') }}</span>
      </div>
      <a-button type="text" @click="handleLogout" class="logout-button">
        <template #icon>
          <logout-outlined />
        </template>
        {{ t('admin.logout') }}
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
  UserOutlined,
  TeamOutlined,
  AimOutlined,
  CameraOutlined,
  SettingOutlined,
  DatabaseOutlined,
  LogoutOutlined,
  RollbackOutlined,
  FileOutlined
} from '@ant-design/icons-vue';
import { useAuthStore } from '../../stores/auth';
import BaseLayout from '../../components/BaseLayout.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 路由和状态管理
const router = useRouter();
const authStore = useAuthStore();

// 当前用户
const currentUser = computed(() => authStore.currentUser);

// 退出登录
const handleLogout = async () => {
  try {
    // 清除本地状态
    authStore.logout();
    message.success(t('admin.logoutSuccess'));
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
  align-items: center;
  gap: var(--spacing-md);
  margin-right: var(--spacing-lg);
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
}

.logout-button:hover {
  color: var(--error-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  /* 导航菜单滚动支持 */
  .nav-menu {
    max-height: 50vh;
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
  
  /* 移动端用户信息样式 */
  .user-info {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-xs);
    margin-right: 0;
    margin-bottom: var(--spacing-sm);
  }
  
  .user-name {
    font-size: var(--font-size-base);
  }
  
  .user-role {
    font-size: var(--font-size-xs);
  }
  
  /* 移动端退出按钮样式 */
  .logout-button {
    width: 100%;
    justify-content: center;
    margin-top: var(--spacing-sm);
  }
}
</style>
