import { onBeforeUnmount, ref, shallowRef } from 'vue'

export const useRichEditor = (initialContent = '') => {
  const editor = shallowRef<unknown>(null)
  const content = ref(initialContent)
  const dirty = ref(false)
  const errors = ref<string[]>([])

  const setEditor = (instance: unknown) => {
    editor.value = instance
  }

  const setContent = (value: string) => {
    content.value = value
    dirty.value = true
  }

  const validateRequired = (message = '内容不能为空') => {
    errors.value = content.value.trim() ? [] : [message]
    return errors.value.length === 0
  }

  const reset = (value = '') => {
    content.value = value
    dirty.value = false
    errors.value = []
  }

  const destroy = () => {
    const instance = editor.value as { destroy?: () => void } | null
    instance?.destroy?.()
    editor.value = null
  }

  onBeforeUnmount(destroy)

  return {
    editor,
    content,
    dirty,
    errors,
    setEditor,
    setContent,
    validateRequired,
    reset,
    destroy
  }
}
