<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue"
import { notifyError, notifySuccess } from '@/util/uiFeedback'

// @ts-ignore
import { CONSTANT } from '@/constant';
// @ts-ignore
import { authApi } from '@/api/auth'
import { usersApi } from '@/api/users'
import { usePageState } from '@/store/page_sate';
import { useRbacStore, useSessionStore } from '@/store';
import { useMessageCenterStore } from '@/store/message-center'
import { useRouter, useRoute } from 'vue-router';
import { sanitizeInternalRedirect } from '@/router/guard-utils'
// @ts-ignore
import { LocalStieConfigUtils } from '@/util/localSiteConfig';
// @ts-ignore
import config from '../../config.js';


const siteConfigData = JSON.parse(CONSTANT.DEFALUT_SITE_CONFIG);
let logo = ref(siteConfigData.logo?.trimStart?.().startsWith('<') ? 'data:image/svg+xml;base64,' + btoa(siteConfigData.logo) : siteConfigData.logo);
let slogan = ref(siteConfigData.slogan);
let loginTitle = ref(siteConfigData.login_title || '消 息 统 一 推 送 中 台');
// 设置默认网站标题
if (siteConfigData.title) {
  document.title = siteConfigData.title;
}
let pageState = usePageState();
let rbacStore = useRbacStore();
let sessionStore = useSessionStore();
let messageCenterStore = useMessageCenterStore();
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
const dashboardTilt = ref<HTMLElement | null>(null);
let tiltFrame = 0;

const updateDashboardTilt = (event: PointerEvent) => {
  if (event.pointerType === 'touch' || !dashboardTilt.value) return;
  const rect = dashboardTilt.value.getBoundingClientRect();
  const x = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
  const y = Math.max(0, Math.min(1, (event.clientY - rect.top) / rect.height));
  const rotateY = 8 + (x - 0.5) * 8;
  const rotateX = 3 - (y - 0.5) * 6;

  cancelAnimationFrame(tiltFrame);
  tiltFrame = requestAnimationFrame(() => {
    dashboardTilt.value?.style.setProperty('--tilt-x', `${rotateX.toFixed(2)}deg`);
    dashboardTilt.value?.style.setProperty('--tilt-y', `${rotateY.toFixed(2)}deg`);
  });
};

const resetDashboardTilt = () => {
  cancelAnimationFrame(tiltFrame);
  dashboardTilt.value?.style.setProperty('--tilt-x', '3deg');
  dashboardTilt.value?.style.setProperty('--tilt-y', '8deg');
};
let registerOpen = ref(false);
let registerLoading = ref(false)
const loginNotice = ref('')
const registerForm = ref({
  username: '',
  password: '',
  confirmPassword: ''
});

const redirectAfterLogin = () => sanitizeInternalRedirect(route.query.redirect, '/')

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

const handleForgotPassword = () => {
  notifyError('请联系系统管理员重置登录密码')
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
    link.type = logoUrl.toLowerCase().includes('.svg') ? 'image/svg+xml' : 'image/png'
  }
}

onBeforeUnmount(() => {
  cancelAnimationFrame(tiltFrame)
})

// 登录页面加载时获取最新配置
onMounted(async () => {

  demoSiteSet();

  const casdoorToken = typeof route.query.casdoor_token === 'string' ? route.query.casdoor_token : '';
  const casdoorLogout = typeof route.query.casdoor_logout === 'string' ? route.query.casdoor_logout : '';
  const casdoorError = typeof route.query.casdoor_error === 'string' ? route.query.casdoor_error : '';
  const casdoorErrorMsg = typeof route.query.casdoor_error_msg === 'string' ? route.query.casdoor_error_msg : '';
  const loginReason = typeof route.query.reason === 'string' ? route.query.reason : ''
  loginNotice.value = loginReason === 'expired' ? '登录状态已失效，请重新登录。登录后将返回原页面。' : loginReason === 'required' ? '请先登录，登录后将返回原页面。' : ''
  
  // Casdoor 统一登出回调
  if (casdoorLogout === '1') {
    sessionStore.clear();
    rbacStore.clear();
    messageCenterStore.resetState();
    window.history.replaceState({}, document.title, window.location.pathname);
    notifySuccess('已完成统一登出');
    return;
  }
  
  // Casdoor 登录成功回调
  if (casdoorToken) {
    sessionStore.setToken(casdoorToken);
    sessionStore.setAuthSource('casdoor');
    messageCenterStore.resetState();
    window.history.replaceState({}, document.title, window.location.pathname);
    try {
      const loaded = await rbacStore.loadCurrentUserPermissions();
      if (!loaded) throw new Error('权限初始化失败');
      await loadAndApplyUserPreference();
      notifySuccess('登录成功');
      router.replace(redirectAfterLogin());
    } catch {
      sessionStore.clear();
      rbacStore.clear();
      messageCenterStore.resetState();
      notifyError('登录成功，但权限初始化失败，请重新登录');
    }
    return;
  }
  
  // Casdoor 登录失败回调
  if (casdoorError === '1') {
    oidcErrorMsg.value = casdoorErrorMsg || 'Casdoor 登录失败';
    notifyError(oidcErrorMsg.value);
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
  } catch {
    notifyError('获取站点配置失败');
  }

  try {
    const rsp = await authApi.publicConfig();
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

// 加载并应用用户显示模式
const loadAndApplyUserPreference = async () => {
  try {
    const rsp = await usersApi.getTheme()
    const data = rsp?.data?.data || {}
    
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
    notifyError('请输入账号和密码');
    return;
  }
  
  loading.value = true;
  
  try {
    const rspe = await authApi.login(account.value, password.value);
    const rsp = rspe.data;
    if (rsp.code != 200) {
        // OIDC 用户特殊提示
        if (rsp.msg === 'OIDC_USER_USE_OIDC_LOGIN') {
          notifyError('该账号为 Casdoor 用户，请使用下方统一登录按钮');
        } else {
          notifyError(rsp.msg);
        }
    } else {
        sessionStore.setToken(rsp.data.token);
        sessionStore.setAuthSource('local');
        messageCenterStore.resetState();
        const loaded = await rbacStore.loadCurrentUserPermissions();
        if (!loaded) {
          sessionStore.clear();
          rbacStore.clear();
          messageCenterStore.resetState();
          notifyError('登录成功，但权限初始化失败，请重新登录');
          return;
        }
        // 登录成功后立即加载并应用用户个性设置
        await loadAndApplyUserPreference()
        notifySuccess('登录成功');
        router.replace(redirectAfterLogin());
    }
  } catch (error) {
    sessionStore.clear();
    rbacStore.clear();
    messageCenterStore.resetState();
    notifyError('登录失败，请检查网络连接');
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
  if (registerLoading.value) return
  if (!registerEnabled.value) {
    notifyError('当前环境已关闭注册，请联系管理员')
    return
  }
  const username = registerForm.value.username.trim()
  const password = registerForm.value.password.trim()
  const confirmPassword = registerForm.value.confirmPassword.trim()
  if (!username || username.length < 3) {
    notifyError('用户名至少3位')
    return
  }
  if (!password || password.length < 6) {
    notifyError('密码至少6位')
    return
  }
  if (password !== confirmPassword) {
    notifyError('两次密码不一致')
    return
  }
  const rsp = await authApi.register({
    username,
    passwd: password,
    confirm_passwd: confirmPassword
  })
  const result = rsp.data
  if (result.code !== 200) {
    notifyError(result.msg || '注册失败')
    return
  }
  sessionStore.setToken(result.data.token)
  sessionStore.setAuthSource('local')
  messageCenterStore.resetState()
  try {
    const loaded = await rbacStore.loadCurrentUserPermissions()
    if (!loaded) throw new Error('权限初始化失败')
    registerOpen.value = false
    notifySuccess('注册并登录成功')
    router.push(redirectAfterLogin())
  } catch {
    sessionStore.clear()
    rbacStore.clear()
    messageCenterStore.resetState()
    notifyError('注册成功，但权限初始化失败，请重新登录')
  }
}

</script>

<template>
  <main class="login-page">
    <section class="brand-stage" aria-hidden="true">
      <header class="brand-copy">
        <div class="brand-mark">
          <span class="brand-logo-wrap"><img :src="logo" alt="" class="brand-logo" /></span>
          <span class="brand-copy-text">
            <strong class="brand-name">{{ loginTitle }}</strong>
            <small>统一推送 · 实时监控 · 稳定触达</small>
          </span>
        </div>
      </header>

      <div class="dashboard-perspective">
        <div class="dashboard-float">
          <div
            ref="dashboardTilt"
            class="dashboard-tilt"
            @pointermove="updateDashboardTilt"
            @pointerleave="resetDashboardTilt"
            @pointercancel="resetDashboardTilt"
          >
          <div class="dashboard-mock">
            <header class="mock-toolbar">
              <div class="mock-heading">
                <span class="mock-app-icon">M</span>
                <div><strong>消息运营驾驶舱</strong><small>全链路实时监控</small></div>
              </div>
              <div class="mock-toolbar-meta">
                <span class="live-status"><i></i>系统运行正常</span>
                <span>2026-07-19</span>
              </div>
            </header>

            <div class="mock-metrics">
              <article class="mock-metric metric-blue"><span>发送总量</span><strong>128,640</strong><small>较昨日 +12.6%</small></article>
              <article class="mock-metric metric-green"><span>成功率</span><strong>99.82%</strong><small>稳定高于 SLA</small></article>
              <article class="mock-metric metric-cyan"><span>活跃渠道</span><strong>12</strong><small>全部正常接入</small></article>
              <article class="mock-metric metric-amber"><span>待处理</span><strong>3</strong><small>需关注异常任务</small></article>
            </div>

            <div class="mock-analysis">
              <section class="mock-panel trend-panel">
                <div class="mock-panel-head"><div><strong>发送趋势</strong><small>近 7 日推送表现</small></div><span>实时</span></div>
                <div class="trend-chart">
                  <div class="chart-grid"></div>
                  <svg viewBox="0 0 420 130" preserveAspectRatio="none">
                    <defs><linearGradient id="loginChartFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#22b8cf" stop-opacity=".32"/><stop offset="1" stop-color="#22b8cf" stop-opacity="0"/></linearGradient></defs>
                    <path d="M0 105 L55 88 L110 94 L168 56 L224 67 L278 40 L334 51 L382 24 L420 34 L420 130 L0 130 Z" fill="url(#loginChartFill)" />
                    <polyline points="0,105 55,88 110,94 168,56 224,67 278,40 334,51 382,24 420,34" fill="none" stroke="#22b8cf" stroke-width="4" stroke-linecap="round" stroke-linejoin="round" />
                    <polyline points="0,112 55,96 110,102 168,66 224,76 278,49 334,60 382,32 420,42" fill="none" stroke="#2f80ed" stroke-width="2" stroke-dasharray="5 5" />
                  </svg>
                  <div class="chart-labels"><span>周一</span><span>周三</span><span>周五</span><span>今日</span></div>
                </div>
              </section>

              <section class="mock-panel health-panel">
                <div class="mock-panel-head"><div><strong>渠道健康</strong><small>核心通道状态</small></div><span>12 / 12</span></div>
                <div class="health-list">
                  <div><span><i class="channel-icon wx">企</i>企业微信</span><em class="status-success">正常</em></div>
                  <div><span><i class="channel-icon sms">短</i>短信网关</span><em class="status-warning">波动</em></div>
                  <div><span><i class="channel-icon mail">邮</i>邮件服务</span><em class="status-success">正常</em></div>
                  <div><span><i class="channel-icon hook">W</i>Webhook</span><em class="status-failed">异常</em></div>
                </div>
              </section>
            </div>

            <section class="mock-panel mock-logs">
              <div class="mock-panel-head"><div><strong>最新链路日志</strong><small>实时追踪投递结果</small></div><span>查看全部</span></div>
              <div class="log-table">
                <div class="log-row log-head"><span>任务</span><span>渠道</span><span>状态</span><span>时间</span></div>
                <div class="log-row"><span>主机告警通知</span><span>企业微信</span><em class="status-success">成功</em><time>18:42:16</time></div>
                <div class="log-row"><span>订单超时提醒</span><span>短信网关</span><em class="status-warning">重试中</em><time>18:41:03</time></div>
                <div class="log-row"><span>支付结果回调</span><span>Webhook</span><em class="status-failed">失败</em><time>18:39:48</time></div>
                <div class="log-row"><span>日终运营报告</span><span>邮件服务</span><em class="status-pending">待处理</em><time>18:38:20</time></div>
              </div>
            </section>
          </div>
          </div>
        </div>
      </div>
    </section>

    <section class="login-access" aria-labelledby="login-title">
      <div class="login-shell">
        <header class="login-brand">
          <span class="login-logo-wrap"><img :src="logo" :alt="`${loginTitle} 标识`" class="login-logo" /></span>
          <div class="login-brand-text">
            <h1 id="login-title">{{ loginTitle }}</h1>
            <p>{{ slogan }}</p>
          </div>
        </header>

        <section class="login-card" aria-labelledby="welcome-title">
          <div class="welcome-copy">
            <h2 id="welcome-title">欢迎回来</h2>
            <p>登录您的账号以继续</p>
          </div>

          <div v-if="oidcErrorMsg" class="login-oidc-error" role="alert"><strong>统一登录异常</strong><span>{{ oidcErrorMsg }}</span></div>

          <el-form class="login-form" @submit.prevent="clickLogin">
            <div class="form-field">
              <label for="account">账号</label>
              <el-input id="account" v-model="account" type="text" autocomplete="username" placeholder="请输入账号" clearable />
            </div>
            <div class="form-field">
              <label for="password">密码</label>
              <el-input id="password" v-model="password" type="password" autocomplete="current-password" placeholder="请输入密码" show-password />
              <div class="password-action-row"><button type="button" class="forgot-password" @click="handleForgotPassword">忘记密码？</button></div>
            </div>
            <el-button type="primary" native-type="submit" class="login-submit" :loading="loading" :disabled="loading">
              {{ loading ? '登录中...' : '登录' }}
            </el-button>
            <template v-if="oidcEnabled">
              <div class="login-divider"><span>或使用其他方式继续</span></div>
              <el-button type="default" native-type="button" class="oidc-button" :loading="oidcLoading" :disabled="oidcLoading" @click="clickOIDCLogin">
                <span class="oidc-content">
                  <img v-if="oidcLoginButtonIcon && !oidcLoading" :src="oidcLoginButtonIcon" :alt="`${oidcLoginButtonText}图标`" @error="(e: Event) => (e.target as HTMLImageElement).style.display = 'none'" />
                  <span>{{ oidcLoading ? '跳转中...' : oidcLoginButtonText }}</span>
                </span>
              </el-button>
            </template>
          </el-form>
        </section>

        <div v-if="registerEnabled" class="register-entry">
          还没有账号？<button type="button" class="login-register-link" @click="handleSubmit">立即注册</button>
        </div>
        <p class="login-copyright">© 2026 {{ loginTitle }}</p>
      </div>
    </section>

    <el-dialog v-model="registerOpen" title="注册账号" width="560px" class="max-w-[90vw]">
      <el-form label-position="top" @submit.prevent="submitRegister">
        <el-form-item label="用户名">
          <el-input v-model="registerForm.username" autocomplete="username" placeholder="至少 3 位" clearable />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="registerForm.password" type="password" autocomplete="new-password" placeholder="至少 6 位" show-password />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="registerForm.confirmPassword" type="password" autocomplete="new-password" placeholder="再次输入密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button :disabled="registerLoading" @click="registerOpen = false">取消</el-button>
        <el-button type="primary" :loading="registerLoading" :disabled="registerLoading" @click="submitRegister">{{ registerLoading ? '注册中...' : '注册并登录' }}</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<style scoped>
.login-page {
  display: grid;
  grid-template-columns: minmax(0, 4fr) minmax(400px, 3fr);
  min-height: 100svh;
  background: var(--dora-layout-bg);
  color: var(--foreground);
}

.brand-stage {
  position: relative;
  display: grid;
  min-width: 0;
  overflow: hidden;
  place-items: center;
  padding: clamp(28px, 3vw, 52px);
  background: #071b2f;
  color: #f5fbff;
  isolation: isolate;
}

.brand-stage::before,
.brand-stage::after {
  position: absolute;
  z-index: -1;
  border: 1px solid rgba(91, 207, 222, .12);
  border-radius: 50%;
  content: "";
}

.brand-stage::before { width: 460px; height: 460px; top: -250px; right: -120px; }
.brand-stage::after { width: 300px; height: 300px; bottom: -190px; left: -110px; }

.brand-copy {
  position: absolute;
  z-index: 3;
  top: clamp(24px, 3vw, 44px);
  left: clamp(28px, 3.5vw, 60px);
}

.brand-mark { display: flex; align-items: center; gap: 12px; }
.brand-copy-text { display: flex; height: 44px; flex-direction: column; justify-content: center; gap: 2px; }
.brand-copy-text small { color: #8fa8bb; font-size: 10px; font-weight: 500; letter-spacing: .035em; line-height: 1.3; }
.brand-logo-wrap,
.login-logo-wrap {
  display: grid;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 12px;
}
.brand-logo-wrap { width: 44px; height: 44px; border: 1px solid rgba(255,255,255,.16); background: rgba(255,255,255,.08); }
.brand-logo { width: 28px; height: 28px; object-fit: contain; }
.brand-name { font-size: 15px; font-weight: 700; letter-spacing: .04em; }
.brand-copy p { margin: 8px 0 0 56px; color: #8fa8bb; font-size: 11px; letter-spacing: .05em; }

.dashboard-perspective {
  width: min(94%, 800px);
  margin: 3vh auto 0;
  padding: 22px 32px 42px 12px;
  perspective: 1500px;
}

.dashboard-float {
  position: relative;
  animation: dashboard-float 7s ease-in-out infinite;
}

.dashboard-tilt {
  --tilt-x: 3deg;
  --tilt-y: 8deg;
  position: relative;
  transform: rotateX(var(--tilt-x)) rotateY(var(--tilt-y)) rotateZ(-1deg);
  transform-style: preserve-3d;
  transition: transform 160ms cubic-bezier(.2, .75, .25, 1);
  will-change: transform;
}

.dashboard-tilt::before,
.dashboard-tilt::after {
  position: absolute;
  inset: 12px -14px -16px 14px;
  border-radius: 18px;
  background: #0c3146;
  content: "";
  transform: translateZ(-28px);
}
.dashboard-tilt::after { inset: 24px -25px -29px 30px; background: rgba(3, 13, 25, .62); filter: blur(18px); transform: translateZ(-52px); }

.dashboard-mock {
  position: relative;
  z-index: 1;
  overflow: hidden;
  border: 1px solid #d7e2ec;
  border-radius: 18px;
  padding: 15px;
  background: #f4f7fa;
  box-shadow: 0 26px 56px rgba(0, 5, 14, .34), 0 3px 0 rgba(255,255,255,.42) inset;
  color: #17263a;
  transform-style: preserve-3d;
}

.mock-toolbar,
.mock-panel,
.mock-metric { border: 1px solid #dfe7ee; background: #fff; box-shadow: 0 8px 22px rgba(19, 46, 73, .07); }
.mock-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; border-radius: 12px; padding: 11px 13px; }
.mock-heading { display: flex; align-items: center; gap: 9px; }
.mock-heading strong,
.mock-panel-head strong { display: block; font-size: 10px; line-height: 1.35; }
.mock-heading small,
.mock-panel-head small { display: block; margin-top: 2px; color: #8796a8; font-size: 7px; }
.mock-app-icon { display: grid; width: 28px; height: 28px; place-items: center; border-radius: 8px; background: #0d6075; color: white; font-size: 11px; font-weight: 800; }
.mock-toolbar-meta { display: flex; align-items: center; gap: 12px; color: #77879a; font-size: 7px; }
.live-status { display: flex; align-items: center; gap: 5px; color: #17845d; }
.live-status i { width: 5px; height: 5px; border-radius: 50%; background: #20b77a; box-shadow: 0 0 0 3px rgba(32,183,122,.12); }

.mock-metrics { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 9px; margin: 10px 0; }
.mock-metric { min-width: 0; border-radius: 11px; padding: 11px 12px; border-top: 2px solid var(--metric-color); }
.mock-metric span { display: block; color: #6f8092; font-size: 7px; }
.mock-metric strong { display: block; margin: 7px 0 5px; color: #1c2d42; font-size: clamp(14px, 1.5vw, 20px); letter-spacing: -.03em; }
.mock-metric small { display: block; overflow: hidden; color: var(--metric-color); font-size: 6.5px; text-overflow: ellipsis; white-space: nowrap; }
.metric-blue { --metric-color: #2f80ed; } .metric-green { --metric-color: #20a878; } .metric-cyan { --metric-color: #168da2; } .metric-amber { --metric-color: #d98a12; }

.mock-analysis { display: grid; grid-template-columns: minmax(0, 1.5fr) minmax(160px, .8fr); gap: 10px; }
.mock-panel { overflow: hidden; border-radius: 12px; }
.mock-panel-head { display: flex; align-items: center; justify-content: space-between; gap: 10px; border-bottom: 1px solid #e9eef3; padding: 9px 11px; }
.mock-panel-head > span { border-radius: 999px; padding: 3px 7px; background: #edf7f8; color: #13788b; font-size: 6.5px; }
.trend-chart { position: relative; height: 128px; margin: 8px 11px 10px; overflow: hidden; }
.chart-grid { position: absolute; inset: 0 0 16px; background: linear-gradient(to right, #e9eff4 1px, transparent 1px) 0 0 / 20% 100%, linear-gradient(to bottom, #e9eff4 1px, transparent 1px) 0 0 / 100% 33.33%; }
.trend-chart svg { position: absolute; inset: 0 0 15px; width: 100%; height: calc(100% - 15px); }
.chart-labels { position: absolute; right: 0; bottom: 0; left: 0; display: flex; justify-content: space-between; color: #91a0af; font-size: 6px; }
.health-list { padding: 3px 10px 7px; }
.health-list > div { display: flex; min-height: 30px; align-items: center; justify-content: space-between; gap: 7px; border-bottom: 1px solid #edf1f4; font-size: 7px; }
.health-list > div:last-child { border-bottom: 0; }
.health-list span { display: flex; align-items: center; gap: 6px; }
.channel-icon { display: grid; width: 19px; height: 19px; place-items: center; border-radius: 6px; background: #edf5f7; color: #168da2; font-size: 6px; font-style: normal; font-weight: 800; }
.sms { color: #2f80ed; } .mail { color: #7657cb; } .hook { color: #d86660; }
.status-success, .status-warning, .status-failed, .status-pending { display: inline-flex; justify-self: start; border-radius: 999px; padding: 3px 7px; font-size: 6px; font-style: normal; font-weight: 700; white-space: nowrap; }
.status-success { background: #e9f8f1; color: #17845d; } .status-warning { background: #fff5df; color: #b37108; } .status-failed { background: #fff0ef; color: #c8524b; } .status-pending { background: #edf1f5; color: #697a8d; }

.mock-logs { margin-top: 10px; }
.log-table { padding: 3px 10px 7px; }
.log-row { display: grid; grid-template-columns: 1.4fr 1fr .7fr .7fr; gap: 8px; align-items: center; min-height: 24px; border-bottom: 1px solid #edf1f4; color: #6b7c8e; font-size: 6.5px; }
.log-row:last-child { border-bottom: 0; }
.log-row.log-head { color: #9aa7b5; font-weight: 700; }
.log-row time { text-align: right; }

.login-access {
  display: flex;
  min-width: 0;
  min-height: 100svh;
  align-items: center;
  justify-content: center;
  overflow-y: auto;
  padding: 34px clamp(28px, 4vw, 58px);
  background-color: var(--dora-layout-bg);
  background-image: linear-gradient(color-mix(in srgb, var(--dora-border) 38%, transparent) 1px, transparent 1px), linear-gradient(90deg, color-mix(in srgb, var(--dora-border) 38%, transparent) 1px, transparent 1px);
  background-size: 64px 64px;
}
.login-shell { display: grid; width: min(100%, 448px); gap: 22px; }
.login-brand { display: grid; justify-items: center; gap: 10px; margin: 0; text-align: center; }
.login-logo-wrap { width: 58px; height: 58px; border: 1px solid var(--dora-border); background: var(--dora-container-bg); box-shadow: 0 12px 28px color-mix(in srgb, var(--brand-600) 18%, transparent); }
.login-logo { width: 38px; height: 38px; object-fit: contain; }
.login-brand-text { min-width: 0; }
.login-brand-text h1 { margin: 0; overflow: hidden; color: var(--foreground); font-size: 21px; font-weight: 760; line-height: 1.35; letter-spacing: .04em; text-overflow: ellipsis; white-space: nowrap; }
.login-brand-text p { margin: 5px 0 0; overflow: hidden; color: var(--muted-foreground); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.login-card { border: 1px solid color-mix(in srgb, var(--dora-border) 86%, transparent); border-radius: 18px; padding: 30px 32px; background: color-mix(in srgb, var(--dora-container-bg) 97%, transparent); box-shadow: 0 22px 60px rgba(15, 23, 42, .12); backdrop-filter: blur(16px); }
.welcome-copy { text-align: center; }
.welcome-copy h2 { margin: 0; color: var(--foreground); font-size: 24px; font-weight: 760; line-height: 1.3; letter-spacing: -.025em; }
.welcome-copy p { margin: 7px 0 0; color: var(--muted-foreground); font-size: 13px; line-height: 1.6; }
.login-oidc-error { display: grid; gap: 2px; margin-top: 18px; border: 1px solid rgba(218, 76, 76, .3); border-radius: 9px; padding: 10px 12px; background: rgba(218, 76, 76, .08); color: #c84e4e; font-size: 12px; line-height: 1.5; }
.login-form { display: grid; gap: 17px; margin-top: 25px; }
.form-field { display: grid; gap: 8px; }
.form-field label { color: var(--foreground); font-size: 13px; font-weight: 650; }
.password-action-row { display: flex; justify-content: flex-end; margin-top: -2px; }
.forgot-password,
.login-register-link { border: 0; padding: 2px; background: transparent; color: var(--brand-600); font: inherit; font-weight: 650; cursor: pointer; }
.forgot-password { font-size: 12px; }
.login-submit,
.oidc-button { width: 100%; min-height: 44px; margin-left: 0 !important; border-radius: 9px; font-weight: 650; }
.login-submit { margin-top: 1px; }
.login-divider { display: flex; align-items: center; gap: 12px; color: var(--muted-foreground); font-size: 11px; }
.login-divider::before, .login-divider::after { height: 1px; flex: 1; background: var(--dora-border); content: ""; }
.login-divider span { white-space: nowrap; }
.oidc-content { display: inline-flex; align-items: center; gap: 9px; }
.oidc-content img { width: 20px; height: 20px; object-fit: contain; }
.register-entry { color: var(--muted-foreground); font-size: 13px; text-align: center; }
.login-copyright { margin: 2px 0 0; color: color-mix(in srgb, var(--muted-foreground) 75%, transparent); font-size: 11px; text-align: center; }

:deep(.el-input__wrapper) { min-height: 44px; border-radius: 9px; }
.forgot-password:focus-visible,
.login-register-link:focus-visible { border-radius: 4px; outline: 2px solid var(--brand-600); outline-offset: 2px; }

@keyframes dashboard-float {
  0%, 100% { transform: translateY(-9px); }
  50% { transform: translateY(9px); }
}

@media (max-width: 1180px) {
  .brand-stage { padding: 30px; }
  .brand-copy { top: 26px; left: 32px; }
  .dashboard-perspective { width: 98%; padding: 10px 24px 30px 0; }
}

@media (max-width: 899px) {
  .login-page { display: block; min-height: 100svh; }
  .brand-stage { display: none; }
  .login-access { min-height: 100svh; padding: 48px 28px; }
}

@media (max-width: 480px) {
  .login-access { align-items: flex-start; padding: 26px 16px 34px; background-size: 48px 48px; }
  .login-shell { gap: 18px; }
  .login-card { border-radius: 15px; padding: 24px 20px; }
  .welcome-copy h2 { font-size: 23px; }
  .login-form { gap: 17px; margin-top: 23px; }
  .login-submit, .oidc-button, :deep(.el-input__wrapper) { min-height: 46px; }
}

@media (max-height: 680px) and (min-width: 900px) {
  .login-access { align-items: flex-start; padding-top: 20px; padding-bottom: 20px; }
  .login-shell { gap: 14px; }
  .login-brand { grid-template-columns: auto 1fr; justify-items: start; text-align: left; }
  .login-logo-wrap { width: 46px; height: 46px; }
  .login-logo { width: 30px; height: 30px; }
  .login-card { padding: 22px 28px; }
  .login-form { margin-top: 19px; gap: 12px; }
  .dashboard-perspective { margin-top: 6vh; transform: scale(.8); transform-origin: center; }
}

@media (prefers-reduced-motion: reduce) {
  .dashboard-perspective { perspective: none; }
  .dashboard-float { animation: none; transform: none; }
  .dashboard-tilt { transform: rotateX(1deg) rotateY(2deg); transition: none; }
}
</style>
