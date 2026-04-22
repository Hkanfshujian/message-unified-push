export const systemSettingsChildren = [
  {
    path: '',
    name: 'system-settings-default',
    redirect: '/system/settings/site'
  },
  {
    path: 'site',
    name: 'system-settings-site',
    component: () => import('../components/pages/settings/SiteSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'auth',
    name: 'system-settings-auth',
    component: () => import('../components/pages/settings/AuthSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'storage',
    name: 'system-settings-storage',
    component: () => import('../components/pages/settings/StorageSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'clean',
    name: 'system-settings-clean',
    component: () => import('../components/pages/settings/CleanSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'mq-status-policy',
    name: 'system-settings-mq-status-policy',
    component: () => import('../components/pages/settings/MQStatusPolicySettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'token-tool',
    name: 'system-settings-token-tool',
    component: () => import('../components/pages/settings/TokenToolSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  },
  {
    path: 'about',
    name: 'system-settings-about',
    component: () => import('../components/pages/settings/AboutSettings.vue'),
    meta: { requiredPermissions: ['system:settings:view'] }
  }
]

export const profileSettingsChildren = [
  {
    path: '',
    name: 'profile-settings-default',
    redirect: '/profile/settings/password'
  },
  {
    path: 'password',
    name: 'profile-settings-password',
    component: () => import('../components/pages/settings/PasswordSettings.vue'),
    meta: { requiredPermissions: ['profile:settings:edit'] }
  },
  {
    path: 'preference',
    name: 'profile-settings-preference',
    component: () => import('../components/pages/profile/ThemePreferenceSettings.vue'),
    meta: { requiredPermissions: ['profile:settings:view'] }
  }
]
