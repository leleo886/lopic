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
        <p class="login-subtitle">{{ t('resetPassword.subtitle') }}</p>
      </div>
      <div class="login-body">
        <a-form
          :model="resetForm"
          :rules="resetRules"
          ref="resetFormRef"
          @finish="handleResetPassword"
        >
          <a-form-item name="email" :label="t('resetPassword.email')">
            <a-input
              v-model:value="resetForm.email"
              :placeholder="t('resetPassword.enterEmail')"
              size="large"
              prefix-icon="mail"
            />
          </a-form-item>
          <a-form-item name="code" :label="t('resetPassword.code')">
            <a-row :gutter="12">
              <a-col :span="14">
                <a-input
                  v-model:value="resetForm.code"
                  :placeholder="t('resetPassword.enterCode')"
                  size="large"
                  prefix-icon="verification"
                />
              </a-col>
              <a-col :span="10">
                <a-button
                  type="default"
                  size="large"
                  :loading="sendingCode"
                  :disabled="countdown > 0"
                  @click="handleSendCode"
                  class="send-code-button"
                >
                  {{ countdown > 0 ? `${countdown}s` : t('resetPassword.sendCode') }}
                </a-button>
              </a-col>
            </a-row>
          </a-form-item>
          <a-form-item name="new_password" :label="t('resetPassword.newPassword')">
            <a-input-password
              v-model:value="resetForm.new_password"
              :placeholder="t('resetPassword.enterNewPassword')"
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
              {{ t('resetPassword.resetPassword') }}
            </a-button>
          </a-form-item>
        </a-form>
        <div class="login-footer">
          <span>{{ t('resetPassword.rememberPassword') }}</span>
          <a @click="router.push('/login')" class="login-link">{{ t('resetPassword.login') }}</a>
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
const resetFormRef = ref();

// 加载状态
const loading = ref(false);
const sendingCode = ref(false);

// 倒计时
const countdown = ref(0);
let countdownTimer: number | null = null;

// 重置表单
const resetForm = reactive({
  email: '',
  code: '',
  new_password: '',
});

// 表单验证规则
const resetRules = {
  email: [
    { required: true, message: () => t('resetPassword.validation.emailRequired'), trigger: 'blur' },
    { type: 'email', message: () => t('resetPassword.validation.emailFormat'), trigger: 'blur' },
  ],
  code: [
    { required: true, message: () => t('resetPassword.validation.codeRequired'), trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: () => t('resetPassword.validation.passwordRequired'), trigger: 'blur' },
    { min: 6, message: () => t('resetPassword.validation.passwordLength'), trigger: 'blur' },
  ],
};

// 处理发送验证码
const handleSendCode = async () => {
  if (!resetForm.email) {
    message.error(t('resetPassword.enterEmailFirst'));
    return;
  }
  
  try {
    sendingCode.value = true;
    
    // 调用发送验证码API，添加locale字段
    await authApi.requestPasswordReset({ 
      email: resetForm.email,
      locale: selectedLanguage.value
    });
    
    message.success(t('resetPassword.codeSent'));
    
    // 开始倒计时
    startCountdown();
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('resetPassword.sendCodeFailed');
    message.error(errorMessage);
  } finally {
    sendingCode.value = false;
  }
};

// 开始倒计时
const startCountdown = () => {
  countdown.value = 60;
  
  if (countdownTimer) {
    clearInterval(countdownTimer);
  }
  
  countdownTimer = window.setInterval(() => {
    countdown.value--;
    if (countdown.value <= 0) {
      if (countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
      }
    }
  }, 1000);
};

// 处理密码重置
const handleResetPassword = async () => {
  if (!resetFormRef.value) return;
  
  try {
    await resetFormRef.value.validate();
    loading.value = true;
    
    // 调用密码重置API
    await authApi.resetPassword(resetForm);
    
    message.success(t('resetPassword.resetSuccess'));
    
    // 跳转到登录页面
    router.push('/login');
  } catch (error: any) {
    const errorCode = error.response?.data?.code;
    const errorMessage = errorCode ? getErrorMessage(errorCode) : t('resetPassword.resetFailed');
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

.send-code-button {
  width: 100%;
  height: 40px;
  font-size: var(--font-size-sm);
}

/* 调整表单元素间距 */
.login-body .ant-form-item {
  margin-bottom: var(--spacing-lg);
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