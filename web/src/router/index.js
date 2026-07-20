import { createRouter, createWebHashHistory } from 'vue-router'
import { CONSTANT } from '../constant'
import axios from 'axios'
import config from '../../config.js'
import { clearAuthzDataStorage, hasAnyPermissionFromStorage, readAuthzDataFromStorage, writeAuthzDataToStorage } from '@/util/rbacAuthz'
import { getFirstAccessibleRoutePath, isNotFoundRoute, LEGACY_SETTINGS_PATH, PROFILE_SETTINGS_PATH, resolveLegacySettingsRedirect, normalizeRequiredPermissions, resolveAuthRedirect, resolvePermissionDeniedRoute } from './guard-utils'
import { profileSettingsChildren, systemSettingsChildren } from './settings-route-config'

const lazyPage = (loader) => loader

const LoginPage = lazyPage(() => import('../components/Login.vue'))
const IndexPage = lazyPage(() => import('../components/Index.vue'))
const NotFoundPage = lazyPage(() => import('../components/404.vue'))
const DashboardPage = lazyPage(() => import('../components/pages/dashboard/Dashboard.vue'))
const SendLogsPage = lazyPage(() => import('../components/pages/sendLogs/SendLogs.vue'))
const LoginLogsPage = lazyPage(() => import('../components/pages/settings/LoginLogs.vue'))
const SystemSettingsPage = lazyPage(() => import('../components/pages/systemManagement/SystemSettings.vue'))
const ProfileSettingsPage = lazyPage(() => import('../components/pages/profile/ProfileSettings.vue'))
const SendWaysPage = lazyPage(() => import('../components/pages/sendWays/SendWays.vue'))
const CronMessagesPage = lazyPage(() => import('../components/pages/cronMessages/CronMessages.vue'))
const MessageTemplatePage = lazyPage(() => import('../components/pages/messageTemplate/MessageTemplate.vue'))
const RolesManagementPage = lazyPage(() => import('../components/pages/systemManagement/RolesManagement.vue'))
const GroupsManagementPage = lazyPage(() => import('../components/pages/systemManagement/GroupsManagement.vue'))
const PermissionsManagementPage = lazyPage(() => import('../components/pages/systemManagement/PermissionsManagement.vue'))
const UsersManagementPage = lazyPage(() => import('../components/pages/systemManagement/UsersManagement.vue'))
const SystemMessagesManagementPage = lazyPage(() => import('../components/pages/systemManagement/SystemMessagesManagement.vue'))
const MQSourcesPage = lazyPage(() => import('../components/pages/dataManagement/MQSources.vue'))
const SubscriptionsPage = lazyPage(() => import('../components/pages/subscriptions/Subscriptions.vue'))
const ConsumeLogsPage = lazyPage(() => import('../components/pages/consumeLogs/ConsumeLogs.vue'))

const router = createRouter({
  // 使用 HTML5 History 模式，确保 URL 变化反映在浏览器地址栏中
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginPage
    },
    {
      // 兼容历史链接：/sendlogs -> /logs/task
      path: '/sendlogs',
      redirect: to => ({ path: '/logs/task', query: to.query })
    },
    {
      path: '/',
      name: 'index',
      component: IndexPage,
      children: [
        {
          // 默认子路由，显示 Dashboard
          path: '',
          name: 'dashboard',
          component: DashboardPage,
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
          component: SendLogsPage,
          meta: { requiredPermissions: ['message:sendlogs:view'] }
        },
        {
          path: 'logs/login',
          name: 'logs-login',
          component: LoginLogsPage,
          meta: { requiredPermissions: ['system:loginlogs:view'] }
        },
        {
          path: 'settings',
          name: 'settings-legacy',
          component: ProfileSettingsPage
        },
        {
          path: 'system/settings/messages',
          redirect: to => ({ path: '/system/messages', query: to.query })
        },
        {
          path: 'system/settings',
          name: 'system-settings',
          component: SystemSettingsPage,
          meta: { requiredPermissions: ['system:settings:view'] },
          children: systemSettingsChildren
        },
        {
          path: 'profile/settings',
          name: 'profile-settings',
          component: ProfileSettingsPage,
          meta: { requiredPermissions: ['profile:settings:view'] },
          children: profileSettingsChildren
        },
        {
          path: 'sendways',
          name: 'sendways',
          component: SendWaysPage,
          meta: { requiredPermissions: ['message:sendways:view'] }
        },
        {
          path: 'cronmessages',
          name: 'cronmessages',
          component: CronMessagesPage,
          meta: { requiredPermissions: ['message:cron:view'] }
        },
        {
          path: 'templates',
          name: 'templates',
          component: MessageTemplatePage,
          meta: { requiredPermissions: ['message:template:view'] }
        },
        {
          path: 'system/roles',
          name: 'system-roles',
          component: RolesManagementPage,
          meta: { requiredPermissions: ['system:rbac:role'] }
        },
        {
          path: 'system/groups',
          name: 'system-groups',
          component: GroupsManagementPage,
          meta: { requiredPermissions: ['system:rbac:group'] }
        },
        {
          path: 'system/permissions',
          name: 'system-permissions',
          component: PermissionsManagementPage,
          meta: { requiredPermissions: ['system:rbac:permission'] }
        },
        {
          path: 'system/users',
          name: 'system-users',
          component: UsersManagementPage,
          meta: { requiredPermissions: ['system:rbac:user'] }
        },
        {
          path: 'system/messages',
          name: 'system-messages',
          component: SystemMessagesManagementPage,
          meta: { requiredPermissions: ['message:system:view'] }
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
          component: MQSourcesPage,
          meta: { requiredPermissions: ['data:mq-source:view'] }
        },
        {
          path: 'message/subscriptions',
          name: 'message-subscriptions',
          component: SubscriptionsPage,
          meta: { requiredPermissions: ['data:subscription:view'] }
        },
        {
          path: 'logs/consume',
          name: 'logs-consume',
          component: ConsumeLogsPage,
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
      component: NotFoundPage
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

const clearAuthSession = () => {
  localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
  localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME)
  localStorage.removeItem(CONSTANT.STORE_OPEN_TABS_NAME || '__message_nest_open_tabs_v1')
  clearAuthzDataStorage()
}

// 登录失效重定向到登录页面
router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME);
  const isAuthenticated = Boolean(token && token.trim() !== '');

  // 404页面不需要登录验证
  if (isNotFoundRoute(to.name)) {
    next();
    return;
  }
  
  const authRedirect = resolveAuthRedirect(isAuthenticated, to.path);
  if (authRedirect) {
    next({ path: authRedirect, query: { reason: 'required', redirect: to.fullPath } });
  } else {
    if (to.path === LEGACY_SETTINGS_PATH) {
      const loaded = await ensureAuthzLoaded(token);
      if (!loaded) {
        next(PROFILE_SETTINGS_PATH);
        return;
      }
      next(resolveLegacySettingsRedirect(hasAnyPermissionFromStorage(['system:settings:view'])));
      return;
    }
    const requiredPermissions = normalizeRequiredPermissions(to.meta?.requiredPermissions);
    if (requiredPermissions.length > 0) {
      const loaded = await ensureAuthzLoaded(token);
      if (!loaded) {
        clearAuthSession();
        next({ path: '/login', query: { reason: 'expired', redirect: to.fullPath } });
        return;
      }
      if (!hasAnyPermissionFromStorage(requiredPermissions)) {
        next(resolvePermissionDeniedRoute(to.path, to.name, getFirstAccessibleRoute));
        return;
      }
    }
    next();
  }
});

export default router
