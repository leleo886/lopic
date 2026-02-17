<template>
  <div class="profile-container">
    <!-- 主要内容 -->
    <main class="profile-main card">
      <div class="profile-content">
        <!-- 用户信息表单 -->
        <a-form
          :model="userForm"
          :rules="userRules"
          ref="userFormRef"
          @finish="handleUpdateUser"
          class="profile-form"
        >
          <p>{{ t('profile.email') }} {{ currentUser?.email }}</p>
          <a-form-item name="username" :label="t('profile.username')">
            <a-input v-model:value="userForm.username" :placeholder="t('profile.enterUsername')" />
          </a-form-item>
          <a-form-item name="password" :label="t('profile.password')" v-if="showPassword">
            <a-input-password v-model:value="userForm.password" :placeholder="t('profile.enterNewPassword')" />
          </a-form-item>

          <a-form-item>
            <a-button type="text" @click="togglePassword">
              {{ showPassword ? t('profile.cancelPassword') : t('profile.changePassword') }}
            </a-button>
          </a-form-item>

          <a-form-item>
            <a-button type="primary" html-type="submit" :loading="loading">
              {{ t('profile.saveChanges') }}
            </a-button>
          </a-form-item>
        </a-form>

        <!-- 用户统计信息 -->
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-label">{{ t('profile.userRole') }}</div>
            <div class="stat-value">
              {{ currentUser?.role.name }} : {{ currentUser?.role.description }}
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-label">{{ t('profile.registerTime') }}</div>
            <div class="stat-value">{{ formatDateTime(currentUser?.created_at || '') }}</div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { message } from 'ant-design-vue';
import { userApi,authApi } from '../api/services';
import { useAuthStore } from '../stores/auth';
import type { UserRequest, User } from '../types/api';
import { formatDateTime } from '../utils/index';
import { useRouter } from 'vue-router';
import { getErrorMessage } from '../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const authStore = useAuthStore();
const router = useRouter();

// 加载状态
const loading = ref(false);
// 显示密码输入框
const showPassword = ref(false);
// 用户表单
const userFormRef = ref();
const userForm = reactive<UserRequest>({
  username: '',
  password: '',
});
// 当前用户信息
const currentUser = ref<User | null>(null);

// 表单验证规则
const userRules = {
  username: [
    { required: true, message: () => t('profile.validation.usernameRequired'), trigger: 'blur' },
    { min: 3, max: 20, message: () => t('profile.validation.usernameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: () => t('profile.validation.emailRequired'), trigger: 'blur' },
    { type: 'email', message: () => t('profile.validation.emailFormat'), trigger: 'blur' },
  ],
  password: [
    { min: 6, message: () => t('profile.validation.passwordLength'), trigger: 'blur' },
  ],
};

// 切换密码输入框显示
const togglePassword = () => {
  showPassword.value = !showPassword.value;
  if (!showPassword.value) {
    userForm.password = '';
  }
};

// 获取当前用户信息
const fetchCurrentUser = async () => {
  try {
    loading.value = true;
    const response = await userApi.getCurrentUser();
    currentUser.value = response.data;
    userForm.username = response.data.username;
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('profile.fetchUserFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};


// 更新用户信息
const handleUpdateUser = async () => {
  if (!userFormRef.value) return;

  try {
    await userFormRef.value.validate();
    loading.value = true;

    // 准备更新数据
    const updateData: UserRequest = {
      username: userForm.username
    };

    // 如果填写了密码，则添加到更新数据中
    if (userForm.password) {
      updateData.password = userForm.password;
    }

    // 调用更新API
    await userApi.updateCurrentUser(updateData);
    message.success(t('profile.updateSuccess'));

     try {
      await authApi.logout({ refresh_token: localStorage.getItem('refresh_token') || '' });
      authStore.logout();
      router.push('/login');
    } catch (error) {
      console.error('Logout error:', error);
      authStore.logout();
      router.push('/login');
    }
    
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('profile.updateFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};

// 初始化
onMounted(() => {
  fetchCurrentUser();
});
</script>

<style scoped>
.profile-container {
  width: 100%;
  max-width: 800px;
  margin: 0 auto;
}

.profile-main {
  padding: var(--spacing-xl);
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.profile-form {
  width: 100%;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.stat-item {
  background-color: var(--bg-light);
  padding: var(--spacing-md);
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-xs);
}

.stat-value {
  font-size: var(--font-size-base);
  font-weight: 500;
  color: var(--text-primary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .profile-container {
    max-width: 100%;
    padding: 0;
  }
  
  .profile-main {
    padding: var(--spacing-md);
  }
  
  .profile-content {
    gap: var(--spacing-md);
  }
  
  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-sm);
  }
  
  .stat-item {
    padding: var(--spacing-sm);
  }
}
</style>