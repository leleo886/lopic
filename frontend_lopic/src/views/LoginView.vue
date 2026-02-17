<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <div class="language-selector">
          <a-radio-group v-model:value="selectedLanguage" @change="handleLanguageChange">
            <a-radio-button value="zh">中文</a-radio-button>
            <a-radio-button value="en">English</a-radio-button>
          </a-radio-group>
        </div>
        <h1 class="login-title">{{ t('login.title') }}</h1>
        <p class="login-subtitle">{{ t('login.subtitle') }}</p>
      </div>
      <div class="login-body">
        <a-form
          :model="loginForm"
          :rules="loginRules"
          ref="loginFormRef"
          @finish="handleLogin"
        >
          <a-form-item name="username" :label="t('login.username')">
            <a-input
              v-model:value="loginForm.username"
              :placeholder="t('login.username')"
              size="large"
              prefix-icon="user"
            />
          </a-form-item>
          <a-form-item name="password" :label="t('login.password')">
            <a-input-password
              v-model:value="loginForm.password"
              :placeholder="t('login.password')"
              size="large"
              prefix-icon="lock"
              visibilityToggle
            />
          </a-form-item>
          <a-form-item>
            <div class="login-links">
              <a @click="router.push('/reset-password')" class="password-reset-link">{{ t('login.forgotPassword') }}</a>
            </div>
          </a-form-item>
          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              class="login-button"
              size="large"
              :loading="loading"
            >
              {{ t('login.login') }}
            </a-button>
          </a-form-item>
        </a-form>
        <div class="login-footer">
          <span>{{ t('login.noAccount') }}</span>
          <a @click="router.push('/register')" class="login-link">{{ t('login.register') }}</a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, inject } from 'vue';
import { useRouter } from 'vue-router';
import { message } from 'ant-design-vue';
import { authApi } from '../api/services';
import { useAuthStore } from '../stores/auth';
import { getErrorMessage } from '../types/errorMessages';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 注入语言相关方法
const currentLanguage = inject('currentLanguage', ref('zh'));
const changeLanguage = inject('changeLanguage', (_lang: string) => {});

// 语言选择状态
const selectedLanguage = ref(currentLanguage.value);

// 处理语言切换
const handleLanguageChange = (e: any) => {
  const lang = e.target.value;
  selectedLanguage.value = lang;
  changeLanguage(lang);
};

// 路由和状态管理
const router = useRouter();
const authStore = useAuthStore();

// 表单引用
const loginFormRef = ref();

// 加载状态
const loading = ref(false);

// 登录表单
const loginForm = reactive({
  username: '',
  password: '',
});

// 表单验证规则
const loginRules = {
  username: [
    { required: true, message: () => t('login.username'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: () => t('login.password'), trigger: 'blur' },
  ],
};

// 处理登录
const handleLogin = async () => {
  if (!loginFormRef.value) return;
  
  try {
    await loginFormRef.value.validate();
    loading.value = true;
    
    // 调用登录API
    const response = await authApi.login(loginForm);
    
    // 从API返回获取用户信息和令牌
    const user = response.data.user;
    const accessToken = response.data.token_response.access_token;
    const refreshToken = response.data.token_response.refresh_token;
    
    // 存储刷新令牌到本地存储
    localStorage.setItem('refresh_token', refreshToken);
    
    // 设置认证状态
    authStore.setAuth(user, accessToken);
    
    message.success(t('login.loginSuccess'));
    
    // 跳转到仪表盘
    router.push('/manage/dashboard');
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('login.loginFailed');
    message.error(errorMessage);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ccf9ff 0%, #7ed9e8 100%);
  padding: var(--spacing-md);
}

.login-card {
  width: 100%;
  max-width: 400px;
  background-color: var(--bg-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  animation: slideUp 0.5s ease;
}

.login-header {
  padding: var(--spacing-xl);
  background-color: var(--bg-light);
  text-align: center;
  border-bottom: 1px solid var(--border-color);
  position: relative;
}

.language-selector {
  position: absolute;
  top: var(--spacing-md);
  right: var(--spacing-md);
  z-index: 10;
}

.language-selector .ant-radio-button-wrapper {
  font-size: var(--font-size-xs);
  padding: 2px 8px;
  height: 28px;
}

.language-selector .ant-radio-button-wrapper:first-child {
  border-radius: 14px 0 0 14px;
}

.language-selector .ant-radio-button-wrapper:last-child {
  border-radius: 0 14px 14px 0;
}

.login-title {
  font-size: 28px;
  font-weight: 600;
  color: var(--primary-color);
  margin-bottom: var(--spacing-xs);
}

.login-subtitle {
  font-size: 15px;
  color: var(--text-secondary);
  margin: 0;
}

.login-body {
  padding: var(--spacing-xl);
}

.login-button {
  width: 100%;
  background-color: var(--primary-color);
  border-color: var(--primary-color);
  height: 48px;
  font-size: var(--font-size-base);
  font-weight: 500;
}

.login-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

/* 动画效果 */
@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.login-links {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--spacing-sm);
}

.password-reset-link {
  color: var(--primary-color);
  font-size: var(--font-size-sm);
  cursor: pointer;
}

.password-reset-link:hover {
  text-decoration: underline;
}

.login-footer {
  margin-top: var(--spacing-lg);
  text-align: center;
  font-size: var(--font-size-sm);
  color: var(--text-secondary);
}

.login-link {
  margin-left: var(--spacing-xs);
  color: var(--primary-color);
  cursor: pointer;
}

.login-link:hover {
  text-decoration: underline;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-card {
    margin: 0;
    border-radius: var(--border-radius-md);
  }
  
  .login-header,
  .login-body {
    padding: var(--spacing-lg);
  }
}
</style>
