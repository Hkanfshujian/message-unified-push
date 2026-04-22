import { createRouter, createWebHistory, createWebHashHistory } from 'vue-router'
// import LoginInex from '../components/Login.vue'
import { CONSTANT } from '../constant'
import axios from 'axios'
import config from '../../config.js'
import { clearAuthzDataStorage, hasAnyPermissionFromStorage, readAuthzDataFromStorage, writeAuthzDataToStorage } from '@/util/rbacAuthz'
import { getFirstAccessibleRoutePath, LEGACY_SETTINGS_PATH, PROFILE_SETTINGS_PATH, SYSTEM_SETTINGS_PATH, resolveLegacySettingsRedirect } from './guard-utils'
import { profileSettingsChildren, systemSettingsChildren } from './settings-route-config'

const router = createRouter({
  // 使用 HTML5 History 模式，确保 URL 变化反映在浏览器地址栏中
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component:() => import('../components/Login.vue')
    },
    {
      // 兼容历史链接：/sendlogs -> /logs/task
      path: '/sendlogs',
      redirect: to => ({ path: '/logs/task', query: to.query })
    },
    {
      path: '/',
      name: 'index',
      component: () => import('../components/Index.vue'),
      children: [
        {
          // 默认子路由，显示 Dashboard
          path: '',
          name: 'dashboard',
          component: () => import('../components/pages/dashboard/Dashboard.vue'),
          meta: { requiredPermissions: ['dashboard:view'] }
        },
        {
          path: 'logs',
          name: 'logs',
          redirect: '/logs/task',
          meta: { requiredPermissions: ['message:sendlogs:view'] }
        },
        {
          path: 'logs/task',
          name: 'logs-task',
          component: () => import('../components/pages/sendLogs/SendLogs.vue'),
          meta: { requiredPermissions: ['message:sendlogs:view'] }
        },
        {
          path: 'logs/login',
          name: 'logs-login',
          component: () => import('../components/pages/settings/LoginLogs.vue'),
          meta: { requiredPermissions: ['system:loginlogs:view'] }
        },
        {
          path: 'settings',
          name: 'settings-legacy',
          component: () => import('../components/pages/profile/ProfileSettings.vue')
        },
        {
          path: 'system/settings',
          name: 'system-settings',
          component: () => import('../components/pages/systemManagement/SystemSettings.vue'),
          meta: { requiredPermissions: ['system:settings:view'] },
          children: systemSettingsChildren
        },
        {
          path: 'profile/settings',
          name: 'profile-settings',
          component: () => import('../components/pages/profile/ProfileSettings.vue'),
          meta: { requiredPermissions: ['profile:settings:view'] },
          children: profileSettingsChildren
        },
        {
          path: 'sendways',
          name: 'sendways',
          component: () => import('../components/pages/sendWays/SendWays.vue'),
          meta: { requiredPermissions: ['message:sendways:view'] }
        },
        {
          path: 'cronmessages',
          name: 'cronmessages',
          component: () => import('../components/pages/cronMessages/CronMessages.vue'),
          meta: { requiredPermissions: ['message:cron:view'] }
        },
        {
          path: 'templates',
          name: 'templates',
          component: () => import('../components/pages/messageTemplate/MessageTemplate.vue'),
          meta: { requiredPermissions: ['message:template:view'] }
        },
        {
          path: 'system/roles',
          name: 'system-roles',
          component: () => import('../components/pages/systemManagement/RolesManagement.vue'),
          meta: { requiredPermissions: ['system:rbac:role'] }
        },
        {
          path: 'system/groups',
          name: 'system-groups',
          component: () => import('../components/pages/systemManagement/GroupsManagement.vue'),
          meta: { requiredPermissions: ['system:rbac:group'] }
        },
        {
          path: 'system/permissions',
          name: 'system-permissions',
          component: () => import('../components/pages/systemManagement/PermissionsManagement.vue'),
          meta: { requiredPermissions: ['system:rbac:permission'] }
        },
        {
          path: 'system/users',
          name: 'system-users',
          component: () => import('../components/pages/systemManagement/UsersManagement.vue'),
          meta: { requiredPermissions: ['system:rbac:user'] }
        },
        {
          path: 'system/relations',
          name: 'system-relations',
          redirect: '/system/users',
          meta: { requiredPermissions: ['system:rbac:user'] }
        },
        {
          path: 'data/mq-sources',
          name: 'data-mq-sources',
          component: () => import('../components/pages/dataManagement/MQSources.vue'),
          meta: { requiredPermissions: ['data:mq-source:view'] }
        },
        {
          path: 'message/subscriptions',
          name: 'message-subscriptions',
          component: () => import('../components/pages/subscriptions/Subscriptions.vue'),
          meta: { requiredPermissions: ['data:subscription:view'] }
        },
        {
          path: 'logs/consume',
          name: 'logs-consume',
          component: () => import('../components/pages/consumeLogs/ConsumeLogs.vue'),
          meta: { requiredPermissions: ['data:consume-log:view'] }
        }
      ]
    },
    // {
    //   path: '/settings',
    //   name: 'settings',
    //   component: () => import('../views/tabsTools/settings/settings.vue')
    // },   
    {
      path: '/:catchAll(.*)',
      name: '404',
      component: () => import('../components/404.vue')
    },
  ]
})

const getPathPrefix = () => config.pathPrefix || ''

const fetchCurrentUserPermissions = async (token) => {
  const baseURL = config.apiUrl + getPathPrefix()
  const response = await axios.get(`${baseURL}/api/v1/rbac/me/permissions`, {
    headers: { 'm-token': token }
  })
  if (response?.status === 200 && response?.data?.code === 200 && response?.data?.data) {
    writeAuthzDataToStorage(response.data.data)
    return true
  }
  return false
}

const ensureAuthzLoaded = async (token) => {
  if (!token) return false
  const localAuthz = readAuthzDataFromStorage()
  if (Array.isArray(localAuthz.permissions) && localAuthz.permissions.length > 0) {
    return true
  }
  try {
    return await fetchCurrentUserPermissions(token)
  } catch {
    return false
  }
}

const getFirstAccessibleRoute = () => getFirstAccessibleRoutePath(hasAnyPermissionFromStorage)

// 登录失效重定向到登录页面
router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
  const isAuthenticated = Boolean(token && token.trim() !== '');

  // 404页面不需要登录验证
  if (to.name === '404') {
    next();
    return;
  }
  
  // 如果没有token且不是访问登录页，跳转到登录页
  if (!isAuthenticated && to.path !== '/login') {
    next('/login');
  } 
  // 如果有token且访问登录页，跳转到首页
  else if (isAuthenticated && to.path === '/login') {
    next('/');
  } 
  // 其他情况正常访问
  else {
    if (to.path === LEGACY_SETTINGS_PATH) {
      const loaded = await ensureAuthzLoaded(token);
      if (!loaded) {
        next(PROFILE_SETTINGS_PATH);
        return;
      }
      next(resolveLegacySettingsRedirect(hasAnyPermissionFromStorage(['system:settings:view'])));
      return;
    }
    const requiredPermissions = to.meta?.requiredPermissions;
    if (Array.isArray(requiredPermissions) && requiredPermissions.length > 0) {
      const loaded = await ensureAuthzLoaded(token);
      if (!loaded) {
        localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME);
        localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME);
        localStorage.removeItem(CONSTANT.STORE_OPEN_TABS_NAME || '__message_nest_open_tabs_v1');
        clearAuthzDataStorage();
        next('/login');
        return;
      }
      if (!hasAnyPermissionFromStorage(requiredPermissions)) {
        if (to.path === '/' || to.name === 'dashboard') {
          const fallbackRoute = getFirstAccessibleRoute();
          if (fallbackRoute && fallbackRoute !== to.path) {
            next(fallbackRoute);
            return;
          }
        }
        next('/404');
        return;
      }
    }
    next();
  }
});

export default router
