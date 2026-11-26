<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    icon: string
    title: string
    description?: string
    count?: number
    tone?: 'primary' | 'info' | 'success' | 'warning'
  }>(),
  { tone: 'primary', description: '', count: undefined },
)

// Static map: Tailwind cannot see class names built at runtime.
const tones = {
  primary: 'bg-primary/10 text-primary',
  info: 'bg-info/10 text-info',
  success: 'bg-success/10 text-success',
  warning: 'bg-warning/10 text-warning',
} as const

const toneClass = computed(() => tones[props.tone])
</script>

<template>
  <div class="flex flex-wrap items-start justify-between gap-4">
    <div class="flex min-w-0 items-center gap-4">
      <div class="grid size-12 shrink-0 place-items-center rounded-xl" :class="toneClass">
        <VIcon :icon="icon" size="26" />
      </div>
      <div class="min-w-0">
        <div class="flex items-center gap-2">
          <h1 class="text-title-medium font-weight-bold text-truncate">{{ title }}</h1>
          <VChip v-if="count !== undefined" size="small" variant="tonal" color="primary">{{ count }}</VChip>
        </div>
        <p v-if="description" class="text-body-small text-medium-emphasis">{{ description }}</p>
      </div>
    </div>

    <div class="flex items-center gap-2">
      <slot name="actions" />
    </div>
  </div>
</template>
