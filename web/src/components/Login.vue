<script setup lang="ts">
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { toast } from "vue-sonner"
import { ref, onMounted } from "vue"
import { Card, CardContent } from '@/components/ui/card'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'

// @ts-ignore
import { CONSTANT } from '@/constant';
// @ts-ignore
import { request } from '@/api/api';
import { usePageState } from '@/store/page_sate';
import { useRbacAuthzStore } from '@/store/rbac_authz';
import { useRouter, useRoute } from 'vue-router';
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig';
// @ts-ignore
import config from '../../config.js';
// @ts-ignore
import { applyTheme } from '@/util/theme';


const siteConfigData = JSON.parse(CONSTANT.DEFALUT_SITE_CONFIG);
let logo = ref('data:image/svg+xml;base64,' + btoa(siteConfigData.logo));
let slogan = ref(siteConfigData.slogan);
let loginTitle = ref(siteConfigData.login_title || '消 息 统 一 推 送 中 台');
// 设置默认网站标题
if (siteConfigData.title) {
  document.title = siteConfigData.title;
}
let pageState = usePageState();
let rbacAuthzState = useRbacAuthzStore();
let router = useRouter();
let route = useRoute();
let account = ref("");
let password = ref("");
let loading = ref(false);
let oidcLoading = ref(false);
let oidcEnabled = ref(false);
let registerEnabled = ref(false);
let oidcLoginButtonText = ref('企微登录');
let oidcLoginButtonIcon = ref('');
let oidcErrorMsg = ref('');
let registerOpen = ref(false);
const registerForm = ref({
  username: '',
  password: '',
  confirmPassword: ''
});

const resolveLogoPath = (logoValue: string) => {
  const raw = (logoValue || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw) || raw.startsWith('data:')) return raw
  const normalized = raw.startsWith('/') ? raw : `/${raw}`
  return normalized
}

// 站点演示模式
const demoSiteSet = () => {
    // @ts-ignore
    if (import.meta.env.VITE_RUN_MODE === 'demo') {
        account.value = 'admin';
        password.value = '123456';
    }
}

let handleSubmit = async function () {
  registerForm.value = {
    username: '',
    password: '',
    confirmPassword: ''
  }
  registerOpen.value = true
}

// 更新favicon
const updateFavicon = (logoValue: string) => {
  if (!logoValue) return
  // 查找现有的favicon链接
  let link = document.querySelector("link[rel*='icon']") as HTMLLinkElement
  if (!link) {
    // 如果不存在，创建新的favicon链接
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  // 判断是 SVG 文本还是图片 URL
  if (logoValue.trimStart().startsWith('<')) {
    // SVG 文本（旧数据）
    const svgBlob = new Blob([logoValue], { type: 'image/svg+xml' })
    link.href = URL.createObjectURL(svgBlob)
    link.type = 'image/svg+xml'
  } else {
    // 图片 URL（新数据）
    const logoUrl = resolveLogoPath(logoValue)
    link.href = logoUrl
    link.type = 'image/png'
  }
}

// 登录页面加载时获取最新配置
onMounted(async () => {

  demoSiteSet();

  const casdoorToken = typeof route.query.casdoor_token === 'string' ? route.query.casdoor_token : '';
  const casdoorLogout = typeof route.query.casdoor_logout === 'string' ? route.query.casdoor_logout : '';
  const casdoorError = typeof route.query.casdoor_error === 'string' ? route.query.casdoor_error : '';
  const casdoorErrorMsg = typeof route.query.casdoor_error_msg === 'string' ? route.query.casdoor_error_msg : '';
  
  console.log('[Casdoor] 登录页回调参数:', { casdoorToken: casdoorToken ? '有' : '无', casdoorLogout, casdoorError });
  
  // Casdoor 统一登出回调
  if (casdoorLogout === '1') {
    console.log('[Casdoor] 统一登出回调，清除本地状态');
    pageState.setToken('');
    localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME);
    rbacAuthzState.clearAuthzData();
    window.history.replaceState({}, document.title, window.location.pathname);
    toast.success('已完成统一登出');
    return;
  }
  
  // Casdoor 登录成功回调
  if (casdoorToken) {
    console.log('[Casdoor] 登录成功，存储token');
    pageState.setToken(casdoorToken);
    localStorage.setItem(CONSTANT.STORE_AUTH_SOURCE_NAME, 'casdoor');
    window.history.replaceState({}, document.title, window.location.pathname);
    await rbacAuthzState.fetchCurrentUserPermissions({ silent: true });
    await loadAndApplyUserPreference();
    toast.success('登录成功');
    router.replace('/');
    return;
  }
  
  // Casdoor 登录失败回调
  if (casdoorError === '1') {
    oidcErrorMsg.value = casdoorErrorMsg || 'Casdoor 登录失败';
    toast.error(oidcErrorMsg.value);
    window.history.replaceState({}, document.title, window.location.pathname);
  }
  
  // 登录页面加载时获取最新配置
  try {
    await LocalStieConfigUtils.getLatestLocalConfig();
    // 更新页面状态中的站点配置
    pageState.setSiteConfigData(LocalStieConfigUtils.getLocalConfig());
    // 更新当前页面的logo、slogan和网站标题
      const siteConfig = LocalStieConfigUtils.getLocalConfig();
      // 判断是 SVG 文本还是图片 URL
      if (siteConfig.logo) {
        if (siteConfig.logo.trimStart().startsWith('<')) {
          // SVG 文本（旧数据）
          logo.value = 'data:image/svg+xml;base64,' + btoa(siteConfig.logo);
        } else {
          // 图片 URL（新数据）
          logo.value = resolveLogoPath(siteConfig.logo);
        }
      }
      slogan.value = siteConfig.slogan;
      loginTitle.value = siteConfig.login_title || '消 息 统 一 推 送 中 台';
      // 更新网站标题
      if (siteConfig.title) {
        document.title = siteConfig.title;
      }
      // 更新favicon
      if (siteConfig.logo) {
        updateFavicon(siteConfig.logo);
      }
      // 应用站点主题色（覆盖默认主题色）
      if (siteConfig.theme_color) {
        applyTheme(siteConfig.theme_color)
      }
  } catch (error) {
    console.warn('获取站点配置失败:', error);
  }

  try {
    const rsp = await request.get('/auth/public-config', {
      params: { t: Date.now() }
    });
    const configData = rsp?.data?.data || {};
    oidcEnabled.value = configData.casdoor_enabled === 'true' || configData.casdoor_enabled === '1';
    registerEnabled.value = configData.register_enabled === 'true' || configData.register_enabled === '1';
    oidcLoginButtonText.value = configData.casdoor_button_text || '企微登录';
    oidcLoginButtonIcon.value = configData.casdoor_button_icon || '';
  } catch (error) {
    oidcEnabled.value = false;
    registerEnabled.value = false;
    oidcLoginButtonText.value = '企微登录';
    oidcLoginButtonIcon.value = '';
  }
});

// 加载并应用用户个性设置（主题色、主题模式、侧边栏背景）
const loadAndApplyUserPreference = async () => {
  try {
    const rsp = await request.get('/profile/theme')
    const data = rsp?.data?.data || {}
    
    // 应用主题色
    if (data.theme_color) {
      applyTheme(data.theme_color)
    }
    
    // 应用主题模式
    if (data.theme_mode === 'light' || data.theme_mode === 'dark' || data.theme_mode === 'system') {
      localStorage.setItem('themePreference', data.theme_mode)
      const systemDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
      const effective = data.theme_mode === 'system' ? (systemDark ? 'dark' : 'light') : data.theme_mode
      if (effective === 'dark') {
        document.documentElement.classList.add('dark')
      } else {
        document.documentElement.classList.remove('dark')
      }
    }
    
    // 应用侧边栏背景色
    if (data.sidebar_bg) {
      document.documentElement.style.setProperty('--sidebar-bg', data.sidebar_bg)
    }
  } catch {
    // 加载失败时使用默认设置，不阻断登录流程
  }
}

// 登录
let clickLogin = async () => {
  // 防止重复提交
  if (loading.value) {
    return;
  }
  
  // 验证输入
  if (!account.value.trim() || !password.value.trim()) {
    toast.error('请输入账号和密码');
    return;
  }
  
  loading.value = true;
  
  try {
    const rspe = await request.post('/auth', { username: account.value, passwd: password.value });
    const rsp = rspe.data;
    if (rsp.code != 200) {
        // OIDC 用户特殊提示
        if (rsp.msg === 'OIDC_USER_USE_OIDC_LOGIN') {
          toast.error('该账号为 Casdoor 用户，请使用下方统一登录按钮');
        } else {
          toast.error(rsp.msg);
        }
    } else {
        pageState.setToken(rsp.data.token);
        localStorage.setItem(CONSTANT.STORE_AUTH_SOURCE_NAME, 'local');
        const loaded = await rbacAuthzState.fetchCurrentUserPermissions({ silent: true });
        if (!loaded) {
          toast.error('登录成功，但权限初始化失败，请检查 token 或权限接口');
          return;
        }
        // 登录成功后立即加载并应用用户个性设置
        await loadAndApplyUserPreference()
        toast.success('登录成功');
        router.replace('/');
    }
  } catch (error) {
    toast.error('登录失败，请检查网络连接');
  } finally {
    loading.value = false;
  }
};

const clickOIDCLogin = async () => {
  if (oidcLoading.value) return;
  oidcLoading.value = true;
  try {
    const pathPrefix = config.pathPrefix || '';
    // 使用新的 Casdoor 登录接口
    window.location.href = `${config.apiUrl}${pathPrefix}/auth/casdoor/login`;
  } finally {
    oidcLoading.value = false;
  }
}

const submitRegister = async () => {
  if (!registerEnabled.value) {
    toast.error('当前环境已关闭注册，请联系管理员')
    return
  }
  const username = registerForm.value.username.trim()
  const password = registerForm.value.password.trim()
  const confirmPassword = registerForm.value.confirmPassword.trim()
  if (!username || username.length < 3) {
    toast.error('用户名至少3位')
    return
  }
  if (!password || password.length < 6) {
    toast.error('密码至少6位')
    return
  }
  if (password !== confirmPassword) {
    toast.error('两次密码不一致')
    return
  }
  const rsp = await request.post('/auth/register', {
    username,
    passwd: password,
    confirm_passwd: confirmPassword
  })
  const result = rsp.data
  if (result.code !== 200) {
    toast.error(result.msg || '注册失败')
    return
  }
  pageState.setToken(result.data.token)
  localStorage.setItem(CONSTANT.STORE_AUTH_SOURCE_NAME, 'local')
  await rbacAuthzState.fetchCurrentUserPermissions({ silent: true })
  registerOpen.value = false
  toast.success('注册并登录成功')
  router.push('/')
}

</script>

<template>
  <div class="flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
    <div class="w-full max-w-sm md:max-w-3xl">
      <div class="flex flex-col gap-6">
        <Card class="overflow-hidden">
          <CardContent class="grid p-0 md:grid-cols-2">
            <form class="p-6 md:p-8">
              <div class="flex flex-col gap-6">
                <!-- 小屏时显示logo和标题 -->
                <div class="flex flex-col items-center text-center md:hidden">
                  <div class="w-20 h-20 mb-4 flex justify-center items-center bg-card dark:bg-muted border border-border rounded-full shadow-sm">
                    <img :src="logo" alt="logo" class="w-12 h-12 object-contain" />
                  </div>
                  <h1 class="text-2xl font-bold text-foreground">
                    {{ loginTitle }}
                  </h1>
                  <p class="text-sm text-muted-foreground mt-1">
                    {{ slogan }}
                  </p>
                </div>
                
                <!-- 大屏时显示标题 -->
                <div class="hidden md:flex flex-col items-center text-center">
                  <h1 class="text-2xl font-bold">
                    {{ loginTitle }}
                  </h1>
                  <p class="text-balance text-muted-foreground">
                    {{ slogan }}
                  </p>
                </div>
                <div class="grid gap-2">
                  <Label for="account">账号</Label>
                  <Input id="account" type="text" placeholder="" v-model="account" required />
                </div>
                <div class="grid gap-2">
                  <div class="flex items-center">
                    <Label for="password">密码</Label>
                    <!-- <a
                  href="#"
                  class="ml-auto text-sm underline-offset-2 hover:underline"
                >
                  Forgot your password?
                </a> -->
                  </div>
                  <Input id="password" type="password" v-model="password" required />
                </div>
                <Button type="submit" class="w-full active:scale-95 active:bg-gray-100 transition duration-150"
                  @click.prevent="clickLogin" :disabled="loading">
                  <svg v-if="loading" class="mr-2 h-4 w-4 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  {{ loading ? '登录中...' : '登录' }}
                </Button>
                <Button v-if="oidcEnabled" type="button" variant="outline" class="w-full" :disabled="oidcLoading" @click="clickOIDCLogin">
                  <span class="inline-flex items-center gap-2">
                    <img v-if="oidcLoginButtonIcon && !oidcLoading" :src="oidcLoginButtonIcon" alt="icon" class="w-5 h-5 object-contain" @error="(e: Event) => (e.target as HTMLImageElement).style.display = 'none'" />
                    <span>{{ oidcLoading ? '跳转中...' : oidcLoginButtonText }}</span>
                  </span>
                </Button>
                <div v-if="oidcErrorMsg" class="text-xs text-red-500 bg-red-50 p-2 rounded border border-red-100">
                  <div>统一登录异常：{{ oidcErrorMsg }}</div>
                </div>
                <!-- <div class="relative text-center text-sm after:absolute after:inset-0 after:top-1/2 after:z-0 after:flex after:items-center after:border-t after:border-border">
              <span class="relative z-10 bg-background px-2 text-muted-foreground">
                Or continue with
              </span>
            </div>
            <div class="grid grid-cols-3 gap-4">
              <Button variant="outline" class="w-full">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path
                    d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701"
                    fill="currentColor"
                  />
                </svg>
                <span class="sr-only">Login with Apple</span>
              </Button>
              <Button variant="outline" class="w-full">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path
                    d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
                    fill="currentColor"
                  />
                </svg>
                <span class="sr-only">Login with Google</span>
              </Button>
              <Button variant="outline" class="w-full">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path
                    d="M6.915 4.03c-1.968 0-3.683 1.28-4.871 3.113C.704 9.208 0 11.883 0 14.449c0 .706.07 1.369.21 1.973a6.624 6.624 0 0 0 .265.86 5.297 5.297 0 0 0 .371.761c.696 1.159 1.818 1.927 3.593 1.927 1.497 0 2.633-.671 3.965-2.444.76-1.012 1.144-1.626 2.663-4.32l.756-1.339.186-.325c.061.1.121.196.183.3l2.152 3.595c.724 1.21 1.665 2.556 2.47 3.314 1.046.987 1.992 1.22 3.06 1.22 1.075 0 1.876-.355 2.455-.843a3.743 3.743 0 0 0 .81-.973c.542-.939.861-2.127.861-3.745 0-2.72-.681-5.357-2.084-7.45-1.282-1.912-2.957-2.93-4.716-2.93-1.047 0-2.088.467-3.053 1.308-.652.57-1.257 1.29-1.82 2.05-.69-.875-1.335-1.547-1.958-2.056-1.182-.966-2.315-1.303-3.454-1.303zm10.16 2.053c1.147 0 2.188.758 2.992 1.999 1.132 1.748 1.647 4.195 1.647 6.4 0 1.548-.368 2.9-1.839 2.9-.58 0-1.027-.23-1.664-1.004-.496-.601-1.343-1.878-2.832-4.358l-.617-1.028a44.908 44.908 0 0 0-1.255-1.98c.07-.109.141-.224.211-.327 1.12-1.667 2.118-2.602 3.358-2.602zm-10.201.553c1.265 0 2.058.791 2.675 1.446.307.327.737.871 1.234 1.579l-1.02 1.566c-.757 1.163-1.882 3.017-2.837 4.338-1.191 1.649-1.81 1.817-2.486 1.817-.524 0-1.038-.237-1.383-.794-.263-.426-.464-1.13-.464-2.046 0-2.221.63-4.535 1.66-6.088.454-.687.964-1.226 1.533-1.533a2.264 2.264 0 0 1 1.088-.285z"
                    fill="currentColor"
                  />
                </svg>
                <span class="sr-only">Login with Meta</span>
              </Button>
            </div> -->
                <div v-if="registerEnabled" class="text-center text-sm">
                  还没有账号？
                  <a href="#" class="underline underline-offset-4" @click="handleSubmit">
                    注册
                  </a>
                </div>
              </div>
            </form>
            <div class="relative hidden md:block bg-card">
              <div class="w-full h-full flex justify-center items-center bg-card">
                <img :src="logo" alt="logo" class="max-w-full max-h-full" />
              </div>
            </div>
          </CardContent>
        </Card>
        <!-- <div class="text-balance text-center text-xs text-muted-foreground [&_a]:underline [&_a]:underline-offset-4 hover:[&_a]:text-primary">
      By clicking continue, you agree to our <a href="#">Terms of Service</a>
      and <a href="#">Privacy Policy</a>.
    </div> -->
      </div>
    </div>
    <Dialog v-model:open="registerOpen">
      <DialogContent class="w-[560px] max-w-[90vw]">
        <DialogHeader>
          <DialogTitle>注册账号</DialogTitle>
        </DialogHeader>
        <div class="space-y-3">
          <Input v-model="registerForm.username" placeholder="用户名（至少3位）" />
          <Input v-model="registerForm.password" type="password" placeholder="密码（至少6位）" />
          <Input v-model="registerForm.confirmPassword" type="password" placeholder="确认密码" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="registerOpen = false">取消</Button>
          <Button @click="submitRegister">注册并登录</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
