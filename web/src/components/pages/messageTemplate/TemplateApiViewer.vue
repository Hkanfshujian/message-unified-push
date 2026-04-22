<script lang="ts">
import { ref, defineComponent, watch, toRef, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { TemplateApiStrGenerate } from '@/util/viewApi'
import { useInstanceData } from '@/composables/useInstanceData'
import { useApiCodeViewer } from '@/composables/useApiCodeViewer'

export default defineComponent({
  name: 'TemplateApiViewer',
  components: {
    Button,
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
    Badge
  },
  props: {
    open: {
      type: Boolean,
      default: false
    },
    templateData: {
      type: Object,
      default: null
    }
  },
  emits: ['update:open'],
  setup(props, { emit }) {
    // 处理关闭事件
    const handleUpdateOpen = (value: boolean) => {
      emit('update:open', value)
    }

    // 使用实例数据管理 composable
    const { hasDynamicRecipientInstance, dynamicRecipientWayTypes, enabledChannelNames } = useInstanceData(
      toRef(props, 'templateData'),
      toRef(props, 'open')
    )

    const recipientExampleByWayType: Record<string, string[]> = {
      QyWeiXinApp: ['zhangsan', 'lisi'],
      WXOA: ['oAbCdEfGhOpenId1', 'oAbCdEfGhOpenId2'],
      AliyunSMS: ['13800138000', '13900139000'],
      Email: ['user1@example.com', 'user2@example.com']
    }

    const recipientExample = computed(() => {
      for (const wayType of dynamicRecipientWayTypes.value) {
        if (recipientExampleByWayType[wayType]) {
          return recipientExampleByWayType[wayType]
        }
      }
      return ['target1', 'target2']
    })

    // 使用 API 代码查看器 composable
    const { activeTab, codeLanguages, copyToClipboard } = useApiCodeViewer()

    // 可选参数选项
    const showRecipients = ref(false)
    const codeStyle = ref('script') // 'script' or 'function'
    
    // 监听动态接收实例变化，自动勾选
    watch(hasDynamicRecipientInstance, (newVal) => {
      if (newVal) {
        showRecipients.value = true
      }
    })
    
    // 监听弹窗关闭，重置状态
    watch(() => props.open, (newVal) => {
      if (!newVal) {
        showRecipients.value = false
      }
    })

    // 生成API代码示例
    const generateApiCode = (language: string) => {
      const templateId = props.templateData?.id || 'TEMPLATE_ID'
      const placeholders = props.templateData?.placeholders || '[]'
      const options = {
        recipients: showRecipients.value,
        waitResult: true,
        recipientExample: recipientExample.value
      }

      const isFunction = codeStyle.value === 'function'

      switch (language) {
        case 'curl':
          return TemplateApiStrGenerate.getCurlString(templateId, placeholders, options, isFunction)
        case 'javascript':
          return TemplateApiStrGenerate.getNodeString(templateId, placeholders, options, isFunction)
        case 'python':
          return TemplateApiStrGenerate.getPythonString(templateId, placeholders, options, isFunction)
        case 'php':
          return TemplateApiStrGenerate.getPHPString(templateId, placeholders, options, isFunction)
        case 'golang':
          return TemplateApiStrGenerate.getGolangString(templateId, placeholders, options, isFunction)
        case 'java':
          return TemplateApiStrGenerate.getJavaString(templateId, placeholders, options, isFunction)
        case 'rust':
          return TemplateApiStrGenerate.getRustString(templateId, placeholders, options, isFunction)
        default:
          return '// 请选择一种编程语言查看示例代码'
      }
    }

    return {
      handleUpdateOpen,
      activeTab,
      hasDynamicRecipientInstance,
      dynamicRecipientWayTypes,
      recipientExample,
      enabledChannelNames,
      showRecipients,
      codeLanguages,
      codeStyle,
      generateApiCode,
      copyToClipboard
    }
  }
})
</script>

<template>
  <Dialog :open="open" @update:open="handleUpdateOpen">
    <DialogContent class="w-[min(855px,98vw)] !max-w-[98vw] sm:!max-w-[98vw] max-h-[90vh] overflow-hidden flex flex-col">
      <DialogHeader class="flex-shrink-0 border-b border-border/60 pb-3">
        <DialogTitle class="flex items-center gap-2 text-lg sm:text-xl">
          <span>模板API接口</span>
          <Badge v-if="templateData" variant="outline">{{ templateData.name }}</Badge>
        </DialogTitle>
        <p class="text-xs text-muted-foreground">通过 V2 接口发送模板消息，支持动态接收者与同步结果回传。</p>
      </DialogHeader>

      <div class="space-y-5 flex-1 overflow-y-auto pr-1 sm:pr-2 mt-4">
        <!-- API 信息概览 -->
        <div class="border border-border/60 rounded-xl p-4 space-y-3 bg-muted/20">
          <div class="flex items-center gap-2 flex-wrap">
            <Badge variant="default">POST</Badge>
            <code class="text-sm bg-gray-100 dark:bg-slate-800 px-2 py-1 rounded">/api/v2/message/send</code>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs text-muted-foreground">
            <p><strong class="text-foreground">模板ID:</strong> <code class="bg-gray-100 dark:bg-slate-800 px-1 py-0.5 rounded">{{ templateData?.id }}</code></p>
            <p><strong class="text-foreground">必填参数:</strong> token, title, placeholders</p>
            <p><strong class="text-foreground">可选参数:</strong> recipients, wait_result</p>
            <p class="text-amber-600 dark:text-amber-400"><strong>⚠️ 注意:</strong> 使用加密 token，不支持明文模板ID</p>
          </div>
          
          <!-- 已启用的渠道列表 -->
          <div v-if="enabledChannelNames.length > 0" class="mt-2 pt-3 border-t border-border/60">
            <p class="text-xs font-medium text-muted-foreground mb-2">已启用发送渠道</p>
            <div class="flex flex-wrap gap-2">
              <Badge 
                v-for="(name, index) in enabledChannelNames" 
                :key="index" 
                variant="secondary"
                class="text-xs rounded-md"
              >
                {{ name }}
              </Badge>
            </div>
          </div>
          <div v-else class="mt-2 pt-3 border-t border-border/60">
            <p class="text-xs text-amber-600 dark:text-amber-400">⚠️ 该模板暂无启用的发送渠道</p>
          </div>
        </div>

        <!-- 可选参数 -->
        <div v-if="hasDynamicRecipientInstance" class="border border-border/60 rounded-xl p-4 bg-muted/20">
          <h3 class="font-semibold mb-3">可选参数</h3>
          <div class="flex flex-wrap gap-4">
            <label class="flex items-center gap-2 cursor-not-allowed opacity-75">
              <input 
                type="checkbox" 
                v-model="showRecipients" 
                disabled
                class="rounded cursor-not-allowed"
              >
              <span class="text-sm">动态接收者</span>
              <Badge variant="secondary" class="text-xs">必填</Badge>
            </label>
          </div>
          <div class="space-y-1 text-xs text-muted-foreground mt-3">
            <p>📧 动态接收者：该模板配置了动态接收实例，发送时必须通过 API 指定接收者列表（群发模式）。</p>
            <p class="text-amber-600 dark:text-amber-400">⚠️ 此参数已自动勾选且不可取消，因为模板已配置动态接收实例</p>
          </div>
        </div>

        <!-- 代码示例 -->
        <div class="space-y-4 border border-border/60 rounded-xl p-4 bg-muted/10">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold">代码示例</h3>
            <Tabs v-model="codeStyle" class="w-40">
              <TabsList class="grid w-full grid-cols-2">
                <TabsTrigger value="script">脚本</TabsTrigger>
                <TabsTrigger value="function">函数封装</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <Tabs v-model="activeTab" class="w-full">
            <TabsList class="grid w-full grid-cols-7 gap-1 h-9">
              <TabsTrigger v-for="lang in codeLanguages" :key="lang.value" :value="lang.value"
                class="flex items-center gap-1 px-2 py-1 text-xs">
                <span>{{ lang.icon }}</span>
                <span class="hidden sm:inline">{{ lang.label }}</span>
                <span class="sm:hidden">{{ lang.label.slice(0, 3) }}</span>
              </TabsTrigger>
            </TabsList>

            <TabsContent v-for="lang in codeLanguages" :key="lang.value" :value="lang.value" class="mt-4">
              <div class="relative rounded-xl border border-slate-800 overflow-hidden">
                <Button size="sm" variant="secondary" class="absolute top-2 right-2 z-10"
                  @click="copyToClipboard(generateApiCode(lang.value))">
                  复制代码
                </Button>
                <pre
                  class="bg-gray-950 text-gray-100 p-4 overflow-x-auto text-xs leading-relaxed max-w-full whitespace-pre-wrap break-words"><code class="text-xs font-mono">{{ generateApiCode(lang.value) }}</code></pre>
              </div>
            </TabsContent>
          </Tabs>
        </div>

        <!-- 说明 -->
        <div class="border border-brand-200/70 dark:border-brand-800/60 bg-brand-50/70 dark:bg-brand-900/20 p-4 rounded-xl text-xs space-y-1">
          <p class="font-semibold text-brand-900 dark:text-brand-200">💡 使用说明</p>
          <ul class="text-brand-800 dark:text-brand-300 space-y-1 ml-4 list-disc leading-5">
            <li><strong>token 参数：</strong>需要使用加密后的 token，不能直接使用明文模板ID（安全考虑）</li>
            <li><strong>placeholders 参数：</strong>用于替换模板中的占位符，格式为 <code class="bg-brand-100 dark:bg-brand-900 px-1 rounded">{"key": "value"}</code></li>
            <li><strong>recipients 参数：</strong>动态接收者列表；企业微信应用传企微用户ID、微信公众号传OpenID、短信传手机号、邮件传邮箱（仅模板启用动态接收者实例时需要）</li>
            <li><strong>wait_result 参数：</strong>是否同步等待发送结果，示例 <code class="bg-brand-100 dark:bg-brand-900 px-1 rounded">"wait_result": true</code>（联调排错建议开启）</li>
            <li>如果模板配置了@提醒，会自动应用到发送的消息中</li>
            <li>支持 Text、HTML、Markdown 三种格式，根据实例配置精确发送对应类型</li>
            <li>系统会自动遍历所有启用的实例进行发送</li>
          </ul>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
/* 代码块样式优化 */
pre {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
