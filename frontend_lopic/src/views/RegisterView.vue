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
        <p class="login-subtitle">{{ t('register.subtitle') }}</p>
      </div>
      <div class="login-body">
        <a-form
          :model="registerForm"
          :rules="registerRules"
          ref="registerFormRef"
          @finish="handleRegister"
        >
          <a-form-item name="username" :label="t('register.username')">
            <a-input
              v-model:value="registerForm.username"
              :placeholder="t('register.enterUsername')"
              size="large"
              prefix-icon="user"
            />
          </a-form-item>
          <a-form-item name="email" :label="t('register.email')">
            <a-input
              v-model:value="registerForm.email"
              :placeholder="t('register.enterEmail')"
              size="large"
              prefix-icon="mail"
            />
          </a-form-item>
          <a-form-item name="password" :label="t('register.password')">
            <a-input-password
              v-model:value="registerForm.password"
              :placeholder="t('register.enterPassword')"
              size="large"
              prefix-icon="lock"
              visibilityToggle
            />
          </a-form-item>
          <a-form-item>
            <a-button
              type="primary"
              html-type="submit"
              class="login-button"
              size="large"
              :loading="loading"
            >
              {{ t('register.register') }}
            </a-button>
          </a-form-item>
        </a-form>
        <div class="login-footer">
          <span>{{ t('register.hasAccount') }}</span>
          <a @click="router.push('/login')" class="login-link">{{ t('register.login') }}</a>
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

// 路由
const router = useRouter();

// 表单引用
const registerFormRef = ref();

// 加载状态
const loading = ref(false);

// 注册表单
const registerForm = reactive({
  username: '',
  email: '',
  password: '',
});

// 表单验证规则
const registerRules = {
  username: [
    { required: true, message: () => t('register.validation.usernameRequired'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: () => t('register.validation.emailRequired'), trigger: 'blur' },
    { type: 'email', message: () => t('register.validation.emailFormat'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: () => t('register.validation.passwordRequired'), trigger: 'blur' },
    { min: 6, message: () => t('register.validation.passwordLength'), trigger: 'blur' },
  ],
};

// 处理注册
const handleRegister = async () => {
  if (!registerFormRef.value) return;
  
  try {
    await registerFormRef.value.validate();
    loading.value = true;
    
    // 调用注册API，添加locale字段
    const response = await authApi.register({
      ...registerForm,
      locale: selectedLanguage.value
    });

    console.log(response.data);

    if (response.data.message === 'Register_EmailSent') {
      message.success(t('register.emailSent'));
    } else if (response.data.message === 'Register_NoMail') {
      message.success(t('register.registerSuccess'));
    } else {
      message.success(t('register.registerSuccess'));
    }
    
    // 跳转到登录页面
    router.push('/login');
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('register.registerFailed');
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
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary-color) 100%);
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