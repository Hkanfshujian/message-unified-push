import { createApp } from 'vue'
import './index.css'
import 'uno.css'
import 'element-plus/dist/index.css'
import './styles/index.scss'
import App from './App.vue'
import pinia from './store';
//@ts-ignore
import router from './router';
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { permissionDirective } from '@/directives/permission'

// 初始化明暗模式：优先本地存储，其次系统偏好
(() => {
  try {
    const storageKey = 'theme';
    const stored = localStorage.getItem(storageKey);
    const systemPrefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
    const theme = stored || (systemPrefersDark ? 'dark' : 'light');
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  } catch (_) {}
})();

const app = createApp(App);
app.use(router);
app.use(pinia);
app.use(ElementPlus, { locale: zhCn });
app.directive('permission', permissionDirective)
app.mount('#app');
