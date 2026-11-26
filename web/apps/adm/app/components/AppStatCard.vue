<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    icon: string
    value: number
    label: string
    tone?: 'primary' | 'info' | 'success' | 'warning'
  }>(),
  { tone: 'primary' },
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
  <VCard border rounded="lg" class="relative overflow-hidden">
    <div class="pointer-events-none absolute -right-8 -top-10 size-32 rounded-full bg-primary/5 blur-2xl" />
    <VCardText class="relative flex items-center gap-4 p-5">
      <div class="grid size-12 shrink-0 place-items-center rounded-xl" :class="toneClass">
        <VIcon :icon="icon" size="26" />
      </div>
      <div class="min-w-0">
        <p class="text-headline-small font-weight-bold tabular-nums">{{ value.toLocaleString() }}</p>
        <p class="text-body-small text-medium-emphasis">{{ label }}</p>
      </div>
    </VCardText>
  </VCard>
</template>
