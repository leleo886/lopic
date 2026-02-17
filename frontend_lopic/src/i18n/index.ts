import { createI18n } from 'vue-i18n';
import zh from '../locales/zh';
import en from '../locales/en';

// 获取本地存储的语言设置
const savedLanguage = localStorage.getItem('language') || 'zh';

// 创建i18n实例
const i18n = createI18n({
  locale: savedLanguage,
  fallbackLocale: 'zh',
  messages: {
    zh,
    en
  },
  legacy: false, // 使用Composition API
  globalInjection: true // 全局注入
});

export default i18n;
export { savedLanguage };
