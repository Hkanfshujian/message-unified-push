import type { Directive } from 'vue'
import { hasAnyPermissionFromStorage, hasPermissionFromStorage } from '@/util/rbacAuthz'

type PermissionBindingValue = string | string[]

const updateElementVisibility = (el: HTMLElement, value: PermissionBindingValue) => {
  let allowed = true
  if (typeof value === 'string') {
    allowed = hasPermissionFromStorage(value)
  } else if (Array.isArray(value)) {
    allowed = hasAnyPermissionFromStorage(value)
  }
  el.style.display = allowed ? '' : 'none'
}

export const permissionDirective: Directive<HTMLElement, PermissionBindingValue> = {
  mounted(el, binding) {
    updateElementVisibility(el, binding.value)
  },
  updated(el, binding) {
    updateElementVisibility(el, binding.value)
  }
}

