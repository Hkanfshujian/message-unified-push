<script setup lang="ts">
import { computed, ref } from 'vue'
import AppActionButton from '@/components/ui/AppActionButton.vue'
import { useRbacStore } from '@/store'
import { allocateRowActions, type AppRowAction, type AppRowActionPermission } from './rowActions'

const props = withDefaults(defineProps<{
  actions?: AppRowAction[]
}>(), {
  actions: () => []
})

const rbacStore = useRbacStore()
const pendingKeys = ref(new Set<string>())
const hasPermission = (permission: AppRowActionPermission) => typeof permission === 'string'
  ? rbacStore.hasPermission(permission)
  : rbacStore.hasAnyPermission(permission)
const allocation = computed(() => allocateRowActions(props.actions, hasPermission))

const runAction = async (action: AppRowAction) => {
  if (action.disabled || action.loading || pendingKeys.value.has(action.key)) return
  const nextPending = new Set(pendingKeys.value)
  nextPending.add(action.key)
  pendingKeys.value = nextPending
  try {
    await action.onClick()
  } finally {
    const clearedPending = new Set(pendingKeys.value)
    clearedPending.delete(action.key)
    pendingKeys.value = clearedPending
  }
}

const isPending = (action: AppRowAction) => action.loading || pendingKeys.value.has(action.key)
const actionType = (action: AppRowAction) => action.danger ? 'danger' : action.type
</script>

<template>
  <div class="page-actions app-row-actions">
    <template v-if="actions.length">
      <AppActionButton
        v-for="action in allocation.direct"
        :key="action.key"
        :type="actionType(action)"
        :disabled="action.disabled || isPending(action)"
        :loading="isPending(action)"
        @click="runAction(action)"
      >
        {{ action.label }}
      </AppActionButton>
      <el-dropdown v-if="allocation.more.length" trigger="click" placement="bottom-end" popper-class="app-row-actions-popper">
        <el-button text class="app-row-actions-more" aria-label="更多操作">
          <span class="app-row-actions-more-dots">•••</span>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-for="action in allocation.more"
              :key="action.key"
              :class="{ 'app-row-action-danger': action.danger || action.type === 'danger' }"
              :disabled="action.disabled || isPending(action)"
              @click="runAction(action)"
            >
              {{ isPending(action) ? `${action.label}…` : action.label }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>
    <template v-else>
      <slot />
      <el-dropdown v-if="$slots.more" trigger="click" placement="bottom-end" popper-class="app-row-actions-popper">
        <el-button text class="app-row-actions-more" aria-label="更多操作">
          <span class="app-row-actions-more-dots">•••</span>
        </el-button>
        <template #dropdown>
          <el-dropdown-menu>
            <slot name="more" />
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </template>
  </div>
</template>
