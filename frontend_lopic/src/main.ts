import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import './style.css'
import i18n from './i18n'

// 创建应用实例
const app = createApp(App)

// 注册插件
app.use(createPinia())
app.use(router)
app.use(Antd)
app.use(i18n)

// 挂载应用
app.mount('#app')
