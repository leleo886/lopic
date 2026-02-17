<template>
  <div class="base-layout" :class="layoutType">
    <!-- 移动端顶部导航栏 -->
    <header class="mobile-header" v-if="isMobile">
      <div class="mobile-header-left">
        <a-button type="text" @click="mobileMenuOpen = true" class="menu-button">
          <template #icon>
            <menu-outlined />
          </template>
        </a-button>
        <h1 class="mobile-title">{{ projectName }}</h1>
      </div>
      <div class="mobile-header-right">
        <a-button type="text" size="small" @click="showAboutModal = true">
          {{ t('layout.about') }}
        </a-button>
        <a-radio-group v-model:value="selectedLanguage" @change="handleLanguageChange" size="small">
          <a-radio-button value="zh">中</a-radio-button>
          <a-radio-button value="en">En</a-radio-button>
        </a-radio-group>
      </div>
    </header>

    <!-- 移动端抽屉导航 -->
    <a-drawer
      v-model:open="mobileMenuOpen"
      placement="left"
      :width="280"
      class="mobile-drawer"
      @close="mobileMenuOpen = false"
    >
      <template #title>
        <div class="drawer-header">
          <h2 class="drawer-title">{{ projectName }}</h2>
        </div>
      </template>
      <nav class="drawer-nav">
        <slot name="sidebar"></slot>
      </nav>
      <div class="drawer-footer">
        <div v-if="showHeaderRight" class="drawer-header-right">
          <slot name="header-right"></slot>
        </div>
        <div v-if="showFooter" class="drawer-footer-slot">
          <slot name="footer"></slot>
        </div>
      </div>
    </a-drawer>

    <!-- 桌面端侧边导航栏 -->
    <aside class="layout-sidebar" v-if="!isMobile">
      <div class="sidebar-header">
        <h1 class="sidebar-title">{{ projectName }}</h1>
      </div>
      <nav class="sidebar-nav">
        <slot name="sidebar"></slot>
      </nav>
      <div v-if="showFooter" class="sidebar-footer">
        <slot name="footer"></slot>
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="layout-content" :class="{ 'mobile-content': isMobile }">
      <!-- 桌面端顶部导航栏 -->
      <header class="layout-header" v-if="!isMobile">
        <div class="header-left">
          <h2 class="header-title">{{ pageTitle }}</h2>
        </div>
        <div v-if="showHeaderRight" class="header-right">
          <slot name="header-right"></slot>
        </div>
        <div class="header-right">
          <div class="header-actions">
            <a-button type="text" size="small" @click="showAboutModal = true">
              {{ t('layout.about') }}
            </a-button>
            <div class="language-selector">
              <a-radio-group v-model:value="selectedLanguage" @change="handleLanguageChange" size="small">
                <a-radio-button value="zh">中文</a-radio-button>
                <a-radio-button value="en">English</a-radio-button>
              </a-radio-group>
            </div>
          </div>
        </div>
      </header>

      <!-- 关于模态框 -->
      <a-modal
        v-model:open="showAboutModal"
        :title="t('layout.aboutTitle')"
        @cancel="showAboutModal = false"
        @ok="showAboutModal = false"
      >
        <div class="about-content">
          <div class="about-logo">
            <h1>{{ projectName }}</h1>
            <a-button type="link" class="github-button" href="https://github.com/leleo886/lopic" target="_blank">
              <GithubOutlined />
            </a-button>
          </div>
          <div class="about-info">
            <p>{{ t('layout.aboutDescription') }}</p>
            <p>{{ t('layout.version') }}: v1.0.0</p>
            <p>{{ t('layout.author') }}: Leleo</p>
            <p>{{ t('layout.license') }}: MIT License</p>
          </div>
        </div>
      </a-modal>

      <!-- 内容区域 -->
      <main class="layout-main">
        <slot></slot>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject, provide, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { MenuOutlined, GithubOutlined } from '@ant-design/icons-vue';

const { t } = useI18n();

const props = defineProps({
  layoutType: {
    type: String,
    default: 'user'
  },
  pageTitle: {
    type: String,
    default: '用户面板'
  },
  showFooter: {
    type: Boolean,
    default: true
  },
  showHeaderRight: {
    type: Boolean,
    default: false
  }
});

const currentLanguage = inject('currentLanguage', ref('zh'));
const changeLanguage = inject('changeLanguage', (_: string) => {});

const selectedLanguage = ref(currentLanguage.value);
const showAboutModal = ref(false);
const mobileMenuOpen = ref(false);

const globalIsMobile = inject<ReturnType<typeof ref<boolean>>>('globalIsMobile');
const isMobile = computed(() => globalIsMobile?.value ?? false);

const handleLanguageChange = (e: any) => {
  const lang = e.target.value;
  selectedLanguage.value = lang;
  changeLanguage(lang);
};

const openMobileMenu = () => {
  mobileMenuOpen.value = true;
};

provide('openMobileMenu', openMobileMenu);
provide('closeMobileMenu', () => { mobileMenuOpen.value = false; });
provide('isMobile', isMobile);

import { projectName } from '../api/axios';
</script>

<style scoped>
.base-layout {
  display: flex;
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 移动端顶部导航栏 */
.mobile-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-md);
  z-index: 1000;
}

.mobile-header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.menu-button {
  padding: var(--spacing-xs);
}

.mobile-title {
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: var(--primary-color);
  margin: 0;
}

.mobile-header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.mobile-header-right .ant-radio-group {
  font-size: var(--font-size-xs);
}

/* 抽屉样式 */
.drawer-header {
  display: flex;
  align-items: center;
}

.drawer-title {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--primary-color);
  margin: 0;
}

.drawer-nav {
  flex: 1;
}

.drawer-footer {
  padding: var(--spacing-lg) 0;
  border-top: 1px solid var(--border-color);
  margin-top: var(--spacing-lg);
}

.drawer-header-right {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  padding: 0 var(--spacing-md);
  margin-bottom: var(--spacing-lg);
}

.drawer-footer-slot {
  padding: 0 var(--spacing-md);
}

/* 桌面端侧边导航栏 */
.layout-sidebar {
  width: 200px;
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 1000;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: var(--spacing-xl);
  border-bottom: 1px solid var(--border-color);
}

.sidebar-title {
  font-size: var(--font-size-xl);
  font-weight: 800;
  color: var(--primary-color);
  margin: 0;
}

.sidebar-nav {
  flex: 1;
  padding: var(--spacing-lg) 0;
}

.sidebar-footer {
  padding: var(--spacing-lg);
  border-top: 1px solid var(--border-color);
}

/* 主内容区 */
.layout-content {
  flex: 1;
  margin-left: 200px;
  min-height: 100vh;
}

.layout-content.mobile-content {
  margin-left: 0;
  padding-top: 56px;
}

/* 顶部导航栏 */
.layout-header {
  background-color: var(--bg-color);
  box-shadow: var(--shadow-sm);
  padding: 0 var(--spacing-xl);
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.language-selector {
  margin-left: 0;
}

/* 关于模态框样式 */
.about-content {
  padding: var(--spacing-md) 0;
}

.about-logo {
  text-align: center;
  margin-bottom: var(--spacing-lg);
}

.about-logo h1 {
  color: var(--primary-color);
  margin: 0;
}

.github-button {
  font-size: 24px;
}

.about-info {
  text-align: center;
}

.about-info p {
  margin-bottom: var(--spacing-sm);
  color: var(--text-secondary);
}

.about-info p:last-child {
  margin-bottom: 0;
}

/* 内容区域 */
.layout-main {
  padding: var(--spacing-xl);
  min-height: calc(100vh - 64px);
}

.base-layout.user .header-title {
  color: var(--text-primary);
}

.base-layout.admin .header-title {
  color: var(--text-primary);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .layout-sidebar {
    width: 200px;
  }
  
  .layout-content {
    margin-left: 200px;
  }
  
  .sidebar-header {
    padding: var(--spacing-lg);
  }
}

@media (max-width: 768px) {
  .layout-sidebar {
    display: none;
  }
  
  .layout-content {
    margin-left: 0;
  }
  
  .layout-main {
    padding: var(--spacing-md);
    min-height: calc(100vh - 56px);
  }
  
  .layout-header {
    display: none;
  }
}
</style>
