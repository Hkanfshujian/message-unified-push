import type { Directive } from 'vue'
import { useRbacStore } from '@/store/rbac'

type PermissionBindingValue = string | string[]

const updateElementVisibility = (el: HTMLElement, value: PermissionBindingValue) => {
  let allowed = true
  const rbacStore = useRbacStore()
  if (typeof value === 'string') {
    allowed = rbacStore.hasPermission(value)
  } else if (Array.isArray(value)) {
    allowed = rbacStore.hasAnyPermission(value)
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

