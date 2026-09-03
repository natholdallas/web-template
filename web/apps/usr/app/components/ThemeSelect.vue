<script setup lang="ts">
import { useColorMode } from '@vueuse/core'
import { Monitor, Moon, Sun } from '@lucide/vue'

const mode = useColorMode()
const modeStore = mode.store

const options: { label: string; value: 'auto' | 'light' | 'dark'; icon: typeof Monitor }[] = [
  { label: 'settings.theme.system', value: 'auto', icon: Monitor },
  { label: 'settings.theme.light', value: 'light', icon: Sun },
  { label: 'settings.theme.dark', value: 'dark', icon: Moon },
]
</script>

<template>
  <div class="flex w-full gap-1 rounded-lg bg-muted p-1" role="radiogroup" aria-label="Theme">
    <UiButton
      v-for="opt in options"
      :key="opt.value"
      type="button"
      size="sm"
      class="flex-1 gap-1.5"
      :variant="modeStore === opt.value ? 'default' : 'ghost'"
      :aria-checked="modeStore === opt.value"
      :role="'radio'"
      @click="modeStore = opt.value"
    >
      <component :is="opt.icon" class="size-4 shrink-0" />
      <span class="text-xs">{{ $t(opt.label) }}</span>
    </UiButton>
  </div>
</template>
