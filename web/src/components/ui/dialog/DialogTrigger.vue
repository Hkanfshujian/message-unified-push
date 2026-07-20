<script setup lang="ts">
defineProps<{ asChild?: boolean }>()
</script>

<template>
  <button
    data-slot="dialog-trigger"
    type="button"
  >
    <slot />
  </button>
</template>
