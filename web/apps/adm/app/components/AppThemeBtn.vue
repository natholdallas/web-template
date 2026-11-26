<script setup lang="ts">
import { usePreferredDark } from '@vueuse/core'

const conf = useConfStore()
const preferredDark = usePreferredDark()
const isDark = computed(() => conf.theme === 'dark' || (conf.theme === 'system' && preferredDark.value))
</script>

<template>
  <VTooltip :text="$t('switch.theme')">
    <template #activator="{ props }">
      <VBtn
        v-bind="props"
        :icon="isDark ? 'mdi-white-balance-sunny' : 'mdi-weather-night'"
        variant="text"
        @click="conf.$switchTheme"
      />
    </template>
  </VTooltip>
</template>
