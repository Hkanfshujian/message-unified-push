import { createApp } from 'vue'
import './index.css'
import App from './App.vue'
import pinia from './store';
//@ts-ignore
import router from './router';
import { permissionDirective } from '@/directives/permission'
import { applyTheme, getStoredTheme } from '@/util/theme'

// 初始化主题：优先本地存储，其次系统偏好
(() => {
  try {
    // 1. 初始化主题色（brand颜色）
    const themeColor = getStoredTheme()
    applyTheme(themeColor)
    
    // 2. 初始化明暗模式
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
app.directive('permission', permissionDirective)
app.mount('#app');
