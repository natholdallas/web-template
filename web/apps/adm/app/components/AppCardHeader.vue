<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    icon: string
    title: string
    description?: string
    tone?: 'primary' | 'info' | 'success' | 'warning'
  }>(),
  { tone: 'primary', description: '' },
)

const tones = {
  primary: 'bg-primary/10 text-primary',
  info: 'bg-info/10 text-info',
  success: 'bg-success/10 text-success',
  warning: 'bg-warning/10 text-warning',
} as const

const toneClass = computed(() => tones[props.tone])
</script>

<template>
  <div class="flex items-start gap-3 px-5 pt-5">
    <div class="grid size-9 shrink-0 place-items-center rounded-lg" :class="toneClass">
      <VIcon :icon="icon" size="20" />
    </div>
    <div class="min-w-0">
      <p class="text-body-large font-weight-bold text-truncate">{{ title }}</p>
      <p v-if="description" class="text-body-small text-medium-emphasis">{{ description }}</p>
    </div>
  </div>
</template>
