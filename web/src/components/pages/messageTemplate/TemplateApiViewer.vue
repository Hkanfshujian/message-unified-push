<script lang="ts">
import { ref, defineComponent, watch, toRef, computed } from 'vue'
import hljs from 'highlight.js/lib/core'
import bash from 'highlight.js/lib/languages/bash'
import javascript from 'highlight.js/lib/languages/javascript'
import python from 'highlight.js/lib/languages/python'
import php from 'highlight.js/lib/languages/php'
import go from 'highlight.js/lib/languages/go'
import java from 'highlight.js/lib/languages/java'
import rust from 'highlight.js/lib/languages/rust'
import { TemplateApiStrGenerate } from '@/util/viewApi'

hljs.registerLanguage('bash', bash)
hljs.registerLanguage('javascript', javascript)
hljs.registerLanguage('python', python)
hljs.registerLanguage('php', php)
hljs.registerLanguage('go', go)
hljs.registerLanguage('java', java)
hljs.registerLanguage('rust', rust)
import { useInstanceData } from '@/composables/useInstanceData'
import { useApiCodeViewer } from '@/composables/useApiCodeViewer'
import AppDetailDrawer from '@/components/ui/AppDetailDrawer.vue'
import { zhCN } from '@/locales/zh-CN'

export default defineComponent({
  name: 'TemplateApiViewer',
  components: { AppDetailDrawer },
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
    const messages = zhCN.templateApiViewer
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

    const highlightLanguageMap: Record<string, string> = {
      curl: 'bash',
      javascript: 'javascript',
      python: 'python',
      php: 'php',
      golang: 'go',
      java: 'java',
      rust: 'rust'
    }

    const getHighlightedCode = (language: string) => {
      const code = generateApiCode(language)
      return hljs.highlight(code, { language: highlightLanguageMap[language] || 'plaintext' }).value
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
      getHighlightedCode,
      copyToClipboard,
      messages
    }
  }
})
</script>

<template>
  <AppDetailDrawer :model-value="open" :title="messages.title" size="min(900px, 96vw)" @update:model-value="handleUpdateOpen">

      <div class="template-api-content">
        <section class="template-api-overview" aria-labelledby="template-api-overview-title">
          <div class="template-api-identity">
            <h3 id="template-api-overview-title">{{ templateData?.name || messages.unnamedTemplate }}</h3>
            <code>{{ templateData?.id }}</code>
          </div>

          <div class="template-api-endpoint" :aria-label="messages.requestAddress">
            <span class="template-api-method">POST</span>
            <code>/api/v2/message/send</code>
          </div>

          <dl class="template-api-contract">
            <div>
              <dt>{{ messages.requiredParameters }}</dt>
              <dd><code>token</code><code>title</code><code>placeholders</code></dd>
            </div>
            <div>
              <dt>{{ messages.optionalParameters }}</dt>
              <dd><code>recipients</code><code>wait_result</code></dd>
            </div>
            <div>
              <dt>{{ messages.sendChannels }}</dt>
              <dd v-if="enabledChannelNames.length" class="template-api-channel-list">
                <span v-for="name in enabledChannelNames" :key="name">{{ name }}</span>
              </dd>
              <dd v-else class="template-api-empty">{{ messages.noEnabledChannels }}</dd>
            </div>
          </dl>

          <p class="template-api-security-note">
            <strong>{{ messages.authRequirement }}</strong>
            {{ messages.authDescription }}<code>token</code>{{ messages.authDescriptionSuffix }}
          </p>
        </section>

        <section v-if="hasDynamicRecipientInstance" class="template-api-recipient" aria-labelledby="template-api-recipient-title">
          <div>
            <span class="template-api-eyebrow">{{ messages.dynamicRecipients }}</span>
            <h3 id="template-api-recipient-title">{{ messages.recipientTitle }}</h3>
            <p>{{ messages.recipientDescriptionPrefix }}<code>recipients</code>{{ messages.recipientDescriptionSuffix }}</p>
          </div>
          <el-tag type="warning" effect="plain">{{ messages.enabled }}</el-tag>
        </section>

        <section class="template-api-workbench" aria-labelledby="template-api-code-title">
          <header class="template-api-section-header">
            <h3 id="template-api-code-title">{{ messages.codeExamples }}</h3>
            <el-segmented v-model="codeStyle" :options="[{ label: messages.script, value: 'script' }, { label: messages.functionWrapper, value: 'function' }]" />
          </header>

          <el-tabs v-model="activeTab" class="template-api-tabs">
            <el-tab-pane v-for="lang in codeLanguages" :key="lang.value" :name="lang.value" :label="lang.label">
              <div class="template-api-code">
                <div class="template-api-code-toolbar">
                  <span>{{ lang.label }}</span>
                  <el-button size="small" plain @click="copyToClipboard(generateApiCode(lang.value))">{{ messages.copyCode }}</el-button>
                </div>
                <pre><code class="hljs" v-html="getHighlightedCode(lang.value)"></code></pre>
              </div>
            </el-tab-pane>
          </el-tabs>
        </section>

        <section class="template-api-guide" aria-labelledby="template-api-guide-title">
          <header class="template-api-section-header">
            <h3 id="template-api-guide-title">{{ messages.guide }}</h3>
          </header>
          <dl class="template-api-guide-list">
            <div><dt><code>token</code></dt><dd>{{ messages.tokenGuide }}</dd></div>
            <div><dt><code>recipients</code></dt><dd>{{ messages.recipientsGuide }}</dd></div>
            <div><dt><code>placeholders</code></dt><dd>{{ messages.placeholdersGuidePrefix }}<code>{"key": "value"}</code>。</dd></div>
            <div><dt><code>wait_result</code></dt><dd>{{ messages.waitResultGuidePrefix }}<code>true</code>。</dd></div>
          </dl>
          <p class="template-api-guide-footnote">{{ messages.guideFootnote }}</p>
        </section>
      </div>
  </AppDetailDrawer>
</template>

<style scoped>
.template-api-content {
  display: grid;
  gap: 16px;
  max-width: 1120px;
  margin: 0 auto;
}

.template-api-overview,
.template-api-workbench,
.template-api-guide {
  overflow: hidden;
  border: 1px solid var(--app-overlay-border);
  border-radius: 10px;
  background: var(--app-overlay-surface);
}

.template-api-identity {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 14px 20px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.template-api-eyebrow {
  display: block;
  margin-bottom: 5px;
  color: var(--admin-text-muted);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.template-api-section-header,
.template-api-recipient {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.template-api-identity h3,
.template-api-section-header h3,
.template-api-recipient h3 {
  margin: 0;
  color: var(--foreground);
  font-size: 16px;
  font-weight: 700;
  line-height: 1.4;
}

.template-api-identity h3 {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-api-identity > code {
  flex: none;
  margin-left: auto;
  color: var(--admin-text-muted);
  font-size: 12px;
}

.template-api-endpoint {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  background: color-mix(in srgb, var(--app-overlay-surface) 82%, var(--admin-surface-muted));
  border-bottom: 1px solid var(--app-overlay-border);
}

.template-api-method {
  flex: none;
  padding: 4px 9px;
  border: 1px solid color-mix(in srgb, var(--el-color-success) 28%, transparent);
  border-radius: 5px;
  background: color-mix(in srgb, var(--el-color-success) 10%, transparent);
  color: var(--el-color-success);
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.04em;
}

.template-api-endpoint > code {
  overflow: hidden;
  color: var(--foreground);
  font-size: 13px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.template-api-contract {
  display: grid;
  grid-template-columns: 1fr 1fr 1.15fr;
  margin: 0;
}

.template-api-contract > div {
  min-width: 0;
  padding: 16px 20px;
}

.template-api-contract > div + div {
  border-left: 1px solid var(--app-overlay-border);
}

.template-api-contract dt,
.template-api-guide-list dt {
  margin-bottom: 9px;
  color: var(--admin-text-muted);
  font-size: 12px;
  font-weight: 600;
}

.template-api-contract dd {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin: 0;
}

.template-api-contract dd code,
.template-api-channel-list span {
  padding: 3px 7px;
  border-radius: 5px;
  background: var(--admin-surface-muted);
  color: var(--foreground);
  font-size: 11px;
}

.template-api-empty {
  color: var(--admin-text-muted);
  font-size: 12px;
}

.template-api-security-note {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin: 0;
  padding: 11px 20px;
  border-top: 1px solid color-mix(in srgb, var(--el-color-warning) 22%, var(--app-overlay-border));
  background: color-mix(in srgb, var(--el-color-warning) 7%, var(--app-overlay-surface));
  color: var(--admin-text-muted);
  font-size: 12px;
  line-height: 1.6;
}

.template-api-security-note strong {
  flex: none;
  color: color-mix(in srgb, var(--el-color-warning) 78%, var(--foreground));
}

.template-api-recipient {
  padding: 15px 18px;
  border-left: 3px solid var(--el-color-warning);
  border-radius: 8px;
  background: color-mix(in srgb, var(--el-color-warning) 6%, var(--app-overlay-surface));
}

.template-api-recipient p {
  margin: 5px 0 0;
  color: var(--admin-text-muted);
  font-size: 12px;
}

.template-api-section-header {
  box-sizing: border-box;
  min-height: 52px;
  padding: 8px 14px 8px 18px;
  border-bottom: 1px solid var(--app-overlay-border);
}

.template-api-tabs {
  padding: 0 18px 18px;
}

.template-api-tabs :deep(.el-tabs__header) {
  margin: 0 0 14px;
}

.template-api-tabs :deep(.el-tabs__item) {
  height: 44px;
  padding: 0 16px;
  font-size: 12px;
}

.template-api-code {
  display: flex;
  flex-direction: column;
  height: 320px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--foreground) 20%, transparent);
  border-radius: 8px;
  background: #172033;
}

.template-api-code-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 44px;
  padding: 6px 8px 6px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.62);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.template-api-code pre {
  flex: 1;
  min-height: 0;
  max-width: 100%;
  margin: 0;
  padding: 18px;
  overflow: scroll;
  scrollbar-color: rgba(148, 163, 184, 0.72) rgba(15, 23, 42, 0.72);
  scrollbar-width: auto;
  color: #e5edf8;
  font-size: 12px;
  line-height: 1.65;
  white-space: pre;
}

.template-api-code pre::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.template-api-code pre::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.72);
}

.template-api-code pre::-webkit-scrollbar-thumb {
  border: 2px solid rgba(15, 23, 42, 0.72);
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.72);
}

.template-api-code pre::-webkit-scrollbar-thumb:hover {
  background: rgba(203, 213, 225, 0.88);
}

.template-api-code pre::-webkit-scrollbar-corner {
  background: rgba(15, 23, 42, 0.72);
}

.template-api-code :deep(.hljs) {
  color: #d8dee9;
  background: transparent;
}

.template-api-code :deep(.hljs-comment),
.template-api-code :deep(.hljs-quote) {
  color: #718096;
  font-style: italic;
}

.template-api-code :deep(.hljs-keyword),
.template-api-code :deep(.hljs-selector-tag),
.template-api-code :deep(.hljs-literal),
.template-api-code :deep(.hljs-section),
.template-api-code :deep(.hljs-link) {
  color: #c792ea;
}

.template-api-code :deep(.hljs-string),
.template-api-code :deep(.hljs-title),
.template-api-code :deep(.hljs-name),
.template-api-code :deep(.hljs-type),
.template-api-code :deep(.hljs-attribute),
.template-api-code :deep(.hljs-symbol),
.template-api-code :deep(.hljs-bullet),
.template-api-code :deep(.hljs-addition),
.template-api-code :deep(.hljs-variable),
.template-api-code :deep(.hljs-template-tag),
.template-api-code :deep(.hljs-template-variable) {
  color: #addb67;
}

.template-api-code :deep(.hljs-number),
.template-api-code :deep(.hljs-meta),
.template-api-code :deep(.hljs-built_in),
.template-api-code :deep(.hljs-builtin-name),
.template-api-code :deep(.hljs-params) {
  color: #f78c6c;
}

.template-api-code :deep(.hljs-function .hljs-title),
.template-api-code :deep(.hljs-class .hljs-title),
.template-api-code :deep(.hljs-title.function_) {
  color: #82aaff;
}

.template-api-code :deep(.hljs-property),
.template-api-code :deep(.hljs-attr) {
  color: #89ddff;
}

.template-api-guide-list {
  display: block;
  margin: 0;
  padding: 6px 0;
}

.template-api-guide-list > div {
  display: grid;
  grid-template-columns: 148px minmax(0, 1fr);
  min-height: 40px;
  margin: 0 10px;
  border-radius: 7px;
}

.template-api-guide-list > div + div {
  border-top: 1px solid var(--app-overlay-border);
  border-radius: 0;
}

.template-api-guide-list dt,
.template-api-guide-list dd {
  display: flex;
  align-items: center;
  margin: 0;
}

.template-api-guide-list dt {
  padding: 6px 14px;
}

.template-api-guide-list dd {
  padding: 6px 16px;
  color: color-mix(in srgb, var(--foreground) 78%, var(--admin-text-muted));
  font-size: 12px;
  line-height: 1.5;
}

.template-api-guide-list dt code {
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 2px 7px;
  border: 1px solid color-mix(in srgb, var(--brand-500) 18%, var(--app-overlay-border));
  border-radius: 5px;
  background: var(--app-overlay-surface);
  color: color-mix(in srgb, var(--brand-700) 76%, var(--foreground));
}

.template-api-guide-list code,
.template-api-security-note code,
.template-api-recipient code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  color: var(--foreground);
  font-size: 11px;
  font-weight: 600;
}

.template-api-guide-footnote {
  position: relative;
  margin: 0;
  padding: 12px 18px 12px 34px;
  border-top: 1px solid var(--app-overlay-border);
  background: color-mix(in srgb, var(--brand-50) 38%, var(--app-overlay-surface));
  color: var(--admin-text-muted);
  font-size: 12px;
  line-height: 1.65;
}

.template-api-guide-footnote::before {
  position: absolute;
  top: 18px;
  left: 18px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--brand-500);
  content: '';
}

pre,
code {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

@media (max-width: 760px) {
  .template-api-content {
    gap: 12px;
  }

  .template-api-recipient,
  .template-api-security-note {
    align-items: flex-start;
    flex-direction: column;
  }

  .template-api-section-header {
    min-height: 48px;
    padding-block: 7px;
  }

  .template-api-contract {
    grid-template-columns: 1fr;
  }

  .template-api-contract > div + div {
    border-left: 0;
  }

  .template-api-contract > div + div {
    border-top: 1px solid var(--app-overlay-border);
  }

  .template-api-guide-list > div {
    grid-template-columns: 1fr;
    gap: 5px;
  }

  .template-api-section-header :deep(.el-segmented) {
    width: 100%;
  }

  .template-api-tabs {
    padding-inline: 12px;
  }

  .template-api-tabs :deep(.el-tabs__nav-wrap) {
    overflow-x: auto;
  }
}
</style>
