<script setup lang="ts">
defineProps<{ asChild?: boolean }>()
</script>

<template>
  <button
    data-slot="dialog-close"
    type="button"
  >
    <slot />
  </button>
</template>
