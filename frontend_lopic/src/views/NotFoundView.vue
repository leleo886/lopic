<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <div class="not-found-icon">
        <frown-outlined />
      </div>
      <h1 class="not-found-title">404</h1>
      <h2 class="not-found-subtitle">{{ t('notFound.subtitle') }}</h2>
      <p class="not-found-description">
        {{ t('notFound.description') }}
      </p>
      <div class="not-found-actions">
        <a-button
          type="primary"
          @click="goHome"
          class="home-button"
        >
          {{ t('notFound.homeButton') }}
        </a-button>
        <a-button
          @click="goBack"
          class="back-button"
        >
          {{ t('notFound.backButton') }}
        </a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { FrownOutlined } from '@ant-design/icons-vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 路由
const router = useRouter();

// 返回首页
const goHome = () => {
  router.push('/manage/dashboard');
};

// 返回上一页
const goBack = () => {
  router.back();
};
</script>

<style scoped>
.not-found-container {
  min-height: 100vh;
  background-color: var(--bg-light);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
}

.not-found-content {
  text-align: center;
  max-width: 500px;
  width: 100%;
  background-color: var(--bg-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-md);
  padding: var(--spacing-xxl);
  animation: fadeIn 0.5s ease;
}

.not-found-icon {
  width: 120px;
  height: 120px;
  border-radius: var(--border-radius-full);
  background-color: var(--bg-dark);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64px;
  color: var(--text-tertiary);
  margin: 0 auto var(--spacing-lg);
}

.not-found-title {
  font-size: 64px;
  font-weight: 700;
  color: var(--primary-color);
  margin: 0 0 var(--spacing-sm);
  line-height: 1;
}

.not-found-subtitle {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 var(--spacing-md);
}

.not-found-description {
  font-size: var(--font-size-base);
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-xl);
  line-height: 1.5;
}

.not-found-actions {
  display: flex;
  gap: var(--spacing-md);
  justify-content: center;
}

.home-button {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
}

.home-button:hover {
  background-color: var(--primary-dark);
  border-color: var(--primary-dark);
}

.back-button {
  background-color: var(--bg-dark);
  border-color: var(--bg-dark);
}

.back-button:hover {
  background-color: var(--border-color);
  border-color: var(--border-color);
}

/* 动画效果 */
@keyframes fadeIn {
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
@media (max-width: 768px) {
  .not-found-content {
    padding: var(--spacing-xl);
  }
  
  .not-found-title {
    font-size: 48px;
  }
  
  .not-found-icon {
    width: 96px;
    height: 96px;
    font-size: 48px;
  }
  
  .not-found-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .home-button,
  .back-button {
    width: 100%;
    max-width: 200px;
  }
}
</style>
