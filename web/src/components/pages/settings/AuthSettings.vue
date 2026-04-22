<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { request } from '@/api/api'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { CircleHelp } from 'lucide-vue-next'

const loading = ref(false)
const saving = ref(false)

const state = reactive({
  register_enabled: 'false',
  // Casdoor 独立配置
  casdoor_enabled: 'false',
  casdoor_endpoint: '',
  casdoor_client_id: '',
  casdoor_client_secret: '',
  casdoor_redirect_uri: '',
  casdoor_auth_path: '/login/oauth/authorize',
  casdoor_token_path: '/api/login/oauth/access_token',
  casdoor_userinfo_path: '/api/get-account',
  casdoor_logout_path: '/api/logout',
  casdoor_auto_create_user: 'true',
  casdoor_default_group_code: '',
  casdoor_button_text: '企微登录',
  casdoor_button_icon: '',
  // 本地用户配置
  local_default_group_code: ''
})

// Switch 组件用的布尔值计算属性
const registerEnabledBool = computed({
  get: () => state.register_enabled === 'true',
  set: (val: boolean) => { state.register_enabled = val ? 'true' : 'false' }
})
const casdoorEnabledBool = computed({
  get: () => state.casdoor_enabled === 'true',
  set: (val: boolean) => { state.casdoor_enabled = val ? 'true' : 'false' }
})
const casdoorAutoCreateUserBool = computed({
  get: () => state.casdoor_auto_create_user === 'true',
  set: (val: boolean) => { state.casdoor_auto_create_user = val ? 'true' : 'false' }
})

const userGroups = ref<Array<{ id: number, code: string, name: string }>>([])

const loadUserGroups = async () => {
  try {
    const rsp = await request.get('/rbac/groups')
    const list = rsp?.data?.data?.lists || []
    userGroups.value = list.map((item: any) => ({
      id: item.id,
      code: item.code,
      name: item.name
    }))
  } catch {
    userGroups.value = []
  }
}

const loadConfig = async () => {
  loading.value = true
  try {
    const rsp = await request.get('/system/auth-config')
    const data = rsp?.data?.data || {}
    state.register_enabled = data.register_enabled || 'false'
    // Casdoor 配置
    state.casdoor_enabled = data.casdoor_enabled || 'false'
    state.casdoor_endpoint = data.casdoor_endpoint || ''
    state.casdoor_client_id = data.casdoor_client_id || ''
    state.casdoor_client_secret = data.casdoor_client_secret || ''
    state.casdoor_redirect_uri = data.casdoor_redirect_uri || ''
    state.casdoor_auth_path = data.casdoor_auth_path || '/login/oauth/authorize'
    state.casdoor_token_path = data.casdoor_token_path || '/api/login/oauth/access_token'
    state.casdoor_userinfo_path = data.casdoor_userinfo_path || '/api/get-account'
    state.casdoor_logout_path = data.casdoor_logout_path || '/api/logout'
    state.casdoor_auto_create_user = data.casdoor_auto_create_user || 'true'
    state.casdoor_default_group_code = data.casdoor_default_group_code || ''
    state.casdoor_button_text = data.casdoor_button_text || '企微登录'
    state.casdoor_button_icon = data.casdoor_button_icon || ''
    // 本地用户
    state.local_default_group_code = data.local_default_group_code || ''
  } finally {
    loading.value = false
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    const rsp = await request.post('/system/auth-config', {
      data: {
        register_enabled: state.register_enabled,
        // Casdoor 配置
        casdoor_enabled: state.casdoor_enabled,
        casdoor_endpoint: state.casdoor_endpoint,
        casdoor_client_id: state.casdoor_client_id,
        casdoor_client_secret: state.casdoor_client_secret,
        casdoor_redirect_uri: state.casdoor_redirect_uri,
        casdoor_auth_path: state.casdoor_auth_path,
        casdoor_token_path: state.casdoor_token_path,
        casdoor_userinfo_path: state.casdoor_userinfo_path,
        casdoor_logout_path: state.casdoor_logout_path,
        casdoor_auto_create_user: state.casdoor_auto_create_user,
        casdoor_default_group_code: state.casdoor_default_group_code,
        casdoor_button_text: state.casdoor_button_text,
        casdoor_button_icon: state.casdoor_button_icon,
        // 本地用户
        local_default_group_code: state.local_default_group_code
      }
    })
    if (rsp?.data?.code === 200) {
      toast.success('认证配置保存成功')
      return
    }
    toast.error(rsp?.data?.msg || '认证配置保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadUserGroups()
  await loadConfig()
})
</script>

<template>
  <div class="space-y-5">
    <!-- 基础开关 -->
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">基础开关</div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Switch v-model="registerEnabledBool" />
            <span class="text-sm font-medium" :class="registerEnabledBool ? 'text-foreground' : 'text-muted-foreground'">
              允许新用户注册
            </span>
          </div>
        </div>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <Switch v-model="casdoorEnabledBool" />
            <span class="text-sm font-medium" :class="casdoorEnabledBool ? 'text-foreground' : 'text-muted-foreground'">
              启用 Casdoor 登录
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Casdoor 配置 -->
    <div v-if="casdoorEnabledBool" class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">Casdoor 配置</div>
      <TooltipProvider>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-1.5">
              <span class="text-sm text-slate-700 dark:text-slate-200">服务地址</span>
              <Tooltip>
                <TooltipTrigger>
                  <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>Casdoor 服务地址，如：https://sso.example.com</p>
                </TooltipContent>
              </Tooltip>
            </div>
            <Input v-model="state.casdoor_endpoint" maxlength="255" autocomplete="off" />
          </div>
          <div class="space-y-2">
            <div class="flex items-center gap-1.5">
              <span class="text-sm text-slate-700 dark:text-slate-200">回调地址</span>
              <Tooltip>
                <TooltipTrigger>
                  <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>登录成功后的回调地址，如：https://your-domain.com/auth/casdoor/callback</p>
                </TooltipContent>
              </Tooltip>
            </div>
            <Input v-model="state.casdoor_redirect_uri" maxlength="255" autocomplete="off" />
          </div>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <div class="flex items-center gap-1.5">
              <span class="text-sm text-slate-700 dark:text-slate-200">客户端ID</span>
              <Tooltip>
                <TooltipTrigger>
                  <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>在 Casdoor 注册的应用客户端 ID</p>
                </TooltipContent>
              </Tooltip>
            </div>
            <Input v-model="state.casdoor_client_id" maxlength="255" autocomplete="off" />
          </div>
          <div class="space-y-2">
            <div class="flex items-center gap-1.5">
              <span class="text-sm text-slate-700 dark:text-slate-200">客户端密钥</span>
              <Tooltip>
                <TooltipTrigger>
                  <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>在 Casdoor 注册的应用客户端密钥</p>
                </TooltipContent>
              </Tooltip>
            </div>
            <Input v-model="state.casdoor_client_secret" type="password" maxlength="255" autocomplete="off" />
          </div>
        </div>

        <!-- OAuth2 端点配置 -->
        <div class="pt-4 mt-2 border-t border-slate-200 dark:border-slate-700">
          <div class="text-sm font-medium text-slate-700 dark:text-slate-200 mb-2">OAuth2 端点配置</div>
          <div class="text-xs text-slate-500 dark:text-slate-400 mb-3">Casdoor 默认端点已预设，一般无需修改</div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <div class="flex items-center gap-1.5">
                <span class="text-sm text-slate-700 dark:text-slate-200">授权路径</span>
                <Tooltip>
                  <TooltipTrigger>
                    <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>OAuth2 授权端点路径，默认：/login/oauth/authorize</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input v-model="state.casdoor_auth_path" maxlength="255" autocomplete="off" />
            </div>
            <div class="space-y-2">
              <div class="flex items-center gap-1.5">
                <span class="text-sm text-slate-700 dark:text-slate-200">令牌路径</span>
                <Tooltip>
                  <TooltipTrigger>
                    <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>OAuth2 令牌端点路径，默认：/api/login/oauth/access_token</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input v-model="state.casdoor_token_path" maxlength="255" autocomplete="off" />
            </div>
            <div class="space-y-2">
              <div class="flex items-center gap-1.5">
                <span class="text-sm text-slate-700 dark:text-slate-200">用户信息路径</span>
                <Tooltip>
                  <TooltipTrigger>
                    <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>用户信息端点路径，默认：/api/get-account</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input v-model="state.casdoor_userinfo_path" maxlength="255" autocomplete="off" />
            </div>
            <div class="space-y-2">
              <div class="flex items-center gap-1.5">
                <span class="text-sm text-slate-700 dark:text-slate-200">登出路径</span>
                <Tooltip>
                  <TooltipTrigger>
                    <CircleHelp class="w-4 h-4 text-slate-400 cursor-help" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>统一登出端点路径，默认：/api/logout</p>
                  </TooltipContent>
                </Tooltip>
              </div>
              <Input v-model="state.casdoor_logout_path" maxlength="255" autocomplete="off" />
            </div>
          </div>
        </div>
      </TooltipProvider>
    </div>

    <!-- 登录策略 -->
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="text-sm font-semibold">登录策略</div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-2">
          <div class="text-sm text-slate-700 dark:text-slate-200">本地新用户默认组</div>
          <select
            v-model="state.local_default_group_code"
            class="w-full h-9 px-3 rounded-md border border-input bg-background text-sm"
          >
            <option value="">不自动绑定</option>
            <option v-for="group in userGroups" :key="group.code" :value="group.code">
              {{ group.name }}
            </option>
          </select>
        </div>
        <div v-if="casdoorEnabledBool" class="space-y-2">
          <div class="text-sm text-slate-700 dark:text-slate-200">Casdoor 新用户默认组</div>
          <select
            v-model="state.casdoor_default_group_code"
            class="w-full h-9 px-3 rounded-md border border-input bg-background text-sm"
          >
            <option value="">不自动绑定</option>
            <option v-for="group in userGroups" :key="group.code" :value="group.code">
              {{ group.name }}
            </option>
          </select>
        </div>
      </div>
      <div v-if="casdoorEnabledBool" class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="flex items-center gap-3">
          <Switch v-model="casdoorAutoCreateUserBool" />
          <span class="text-sm font-medium" :class="casdoorAutoCreateUserBool ? 'text-foreground' : 'text-muted-foreground'">
            Casdoor 自动创建本地用户
          </span>
        </div>
        <div class="space-y-2">
          <div class="text-sm text-slate-700 dark:text-slate-200">登录按钮文案</div>
          <Input v-model="state.casdoor_button_text" placeholder="企微登录" maxlength="50" />
        </div>
      </div>
      <div v-if="casdoorEnabledBool" class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-2">
          <div class="text-sm text-slate-700 dark:text-slate-200 flex items-center gap-1">
            登录按钮图标
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger as-child>
                  <CircleHelp class="w-4 h-4 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p>填写图标 URL，如企微、钉钉等 Logo 地址</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <Input v-model="state.casdoor_button_icon" placeholder="https://example.com/icon.png" />
        </div>
        <div v-if="state.casdoor_button_icon" class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground">预览：</span>
          <img :src="state.casdoor_button_icon" alt="button-icon" class="w-6 h-6 object-contain" @error="(e: Event) => (e.target as HTMLImageElement).style.display = 'none'" />
        </div>
      </div>
    </div>

    <div class="flex justify-end pt-2 border-t border-slate-200/80 dark:border-slate-700/80">
      <Button :disabled="loading || saving" @click="saveConfig">{{ saving ? '保存中...' : '保存设置' }}</Button>
    </div>
  </div>
</template>
