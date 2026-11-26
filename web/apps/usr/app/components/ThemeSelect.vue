<script setup lang="ts">
import { useColorMode } from '@vueuse/core'
import { Check, Monitor, Moon, Sun } from '@lucide/vue'

const mode = useColorMode()
const modeStore = mode.store

const options: { label: string; value: 'auto' | 'light' | 'dark'; icon: typeof Monitor }[] = [
  { label: 'settings.theme.system', value: 'auto', icon: Monitor },
  { label: 'settings.theme.light', value: 'light', icon: Sun },
  { label: 'settings.theme.dark', value: 'dark', icon: Moon },
]

type Scheme = 'light' | 'dark'

// Preview palette per scheme; `auto` renders both halves, so its entry stays empty.
const preview: Record<'auto' | Scheme, { shell: string; side: string; bar: string }> = {
  auto: { shell: '', side: '', bar: '' },
  light: { shell: 'bg-zinc-100', side: 'bg-zinc-200', bar: 'bg-zinc-300' },
  dark: { shell: 'bg-zinc-900', side: 'bg-zinc-800', bar: 'bg-zinc-700' },
}
const halves: Scheme[] = ['light', 'dark']
const bars = ['w-3/4', 'w-1/2', 'w-2/3']
</script>

<template>
  <div class="grid gap-3 sm:grid-cols-3" role="radiogroup" :aria-label="$t('settings.theme')">
    <button
      v-for="opt in options"
      :key="opt.value"
      type="button"
      role="radio"
      :aria-checked="modeStore === opt.value"
      class="group flex flex-col overflow-hidden border bg-card p-2 text-left transition-colors hover:border-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      :class="modeStore === opt.value ? 'border-primary ring-1 ring-primary' : 'border-border'"
      @click="modeStore = opt.value"
    >
      <span class="relative block h-24 overflow-hidden border">
        <span v-if="opt.value === 'auto'" class="flex h-full">
          <span v-for="half in halves" :key="half" class="flex h-full w-1/2" :class="preview[half].shell">
            <span class="h-full w-3 shrink-0" :class="preview[half].side" />
            <span class="flex flex-1 flex-col gap-1 p-2">
              <span v-for="bar in bars" :key="bar" class="h-1.5" :class="[bar, preview[half].bar]" />
            </span>
          </span>
        </span>
        <span v-else class="flex h-full" :class="preview[opt.value].shell">
          <span class="h-full w-3 shrink-0" :class="preview[opt.value].side" />
          <span class="flex flex-1 flex-col gap-1 p-2">
            <span v-for="bar in bars" :key="bar" class="h-1.5" :class="[bar, preview[opt.value].bar]" />
          </span>
        </span>
      </span>

      <span class="mt-2 flex items-center justify-between px-0.5">
        <span class="flex items-center gap-1.5">
          <component :is="opt.icon" class="size-3.5" />
          <span class="text-sm font-medium">{{ $t(opt.label) }}</span>
        </span>
        <Check v-if="modeStore === opt.value" class="size-4 text-primary" />
      </span>
    </button>
  </div>
</template>
