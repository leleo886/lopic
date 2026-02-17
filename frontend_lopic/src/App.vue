<template>
  <a-config-provider
    :locale="locale"
    :theme="{
      token: {
        colorPrimary: '#73CDDC',
        colorInfo: '#73CDDC',
      },
    }"
  >
    <div class="app">
      <router-view />
    </div>
  </a-config-provider>
</template>

<script setup lang="ts">
import { ref, provide, onMounted, onUnmounted } from 'vue';
import zhCN from 'ant-design-vue/es/locale/zh_CN';
import en_US from 'ant-design-vue/es/locale/en_US';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import 'dayjs/locale/en';
import i18n from './i18n';

// 语言状态
const currentLanguage = ref('zh');
const locale = ref(zhCN);

// 全局移动端状态
const globalIsMobile = ref(false);

const checkMobile = () => {
  globalIsMobile.value = window.innerWidth < 768;
};

// 切换语言
const changeLanguage = (lang: string) => {
  currentLanguage.value = lang;
  locale.value = lang === 'zh' ? zhCN : en_US;
  dayjs.locale(lang === 'zh' ? 'zh-cn' : 'en');
  
  // 更新i18n语言
  i18n.global.locale.value = lang as 'zh' | 'en';
  
  // 保存到本地存储
  localStorage.setItem('language', lang);
};

// 初始化语言
const initLanguage = () => {
  const savedLang = localStorage.getItem('language');
  if (savedLang) {
    changeLanguage(savedLang);
  }
};

// 提供语言相关方法给子组件
provide('currentLanguage', currentLanguage);
provide('changeLanguage', changeLanguage);
provide('globalIsMobile', globalIsMobile);

// 初始化
initLanguage();

onMounted(() => {
  checkMobile();
  window.addEventListener('resize', checkMobile);
});

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile);
});
</script>

<style scoped>
.app {
  min-height: 100vh;
  background-color: var(--bg-light);
}

/* 路由过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-normal);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
