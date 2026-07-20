import { ref } from 'vue'
import { notifyError, notifySuccess } from '@/util/uiFeedback'

/**
 * API 代码查看器公共逻辑 Composable
 */
export function useApiCodeViewer() {
  // 当前选中的标签
  const activeTab = ref('curl')

  // 代码语言选项
  const codeLanguages = [
    { value: 'curl', label: 'cURL', icon: '🌐' },
    { value: 'javascript', label: 'JS', icon: '🟨' },
    { value: 'python', label: 'Python', icon: '🐍' },
    { value: 'php', label: 'PHP', icon: '🐘' },
    { value: 'golang', label: 'Go', icon: '🐹' },
    { value: 'java', label: 'Java', icon: '☕' },
    { value: 'rust', label: 'Rust', icon: '🦀' }
  ]

  const copyWithSelection = (text: string) => {
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.setAttribute('readonly', '')
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    const copied = document.execCommand('copy')
    document.body.removeChild(textarea)
    return copied
  }

  // 复制代码到剪贴板
  const copyToClipboard = async (text: string) => {
    try {
      if (navigator.clipboard?.writeText) {
        try {
          await navigator.clipboard.writeText(text)
        } catch {
          if (!copyWithSelection(text)) throw new Error('copy failed')
        }
      } else if (!copyWithSelection(text)) {
        throw new Error('copy failed')
      }
      notifySuccess('复制成功')
    } catch {
      notifyError('复制失败，请手动选择代码复制')
    }
  }

  return {
    activeTab,
    codeLanguages,
    copyToClipboard
  }
}
