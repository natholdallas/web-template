<script setup lang="ts">
defineProps<{ compact?: boolean }>()
</script>

<template>
  <div class="flex items-center gap-3">
    <div class="grid size-10 place-items-center rounded-xl bg-gradient-to-br from-primary to-secondary text-white">
      <VIcon icon="mdi-shield-account" size="22" />
    </div>
    <div v-if="!compact" class="min-w-0">
      <p class="text-body-medium font-weight-bold text-truncate">{{ $t('app.name') }}</p>
      <p class="text-body-small text-medium-emphasis text-truncate">{{ $t('app.console') }}</p>
    </div>
  </div>
</template>
