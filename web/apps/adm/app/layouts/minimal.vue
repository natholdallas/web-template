<script setup lang="ts">
const conf = useConfStore()
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VMain class="auth-bg relative size-full overflow-auto">
      <div class="auth-grid absolute inset-0 pointer-events-none" />
      <div
        class="absolute right-4 top-4 z-10 flex items-center gap-1 rounded-lg border border-on-surface/10 bg-surface/70 p-1 backdrop-blur"
      >
        <LangSwitcher />
        <AppThemeBtn />
      </div>
      <div class="relative flex min-h-full items-center justify-center p-4">
        <slot />
      </div>
    </VMain>
  </VApp>
</template>
