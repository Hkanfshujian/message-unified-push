import type { ComputedRef, InjectionKey } from 'vue'

export interface DialogContext {
  open: ComputedRef<boolean>
  setOpen: (value: boolean) => void
}

export const dialogContextKey: InjectionKey<DialogContext> = Symbol('dialogContext')
