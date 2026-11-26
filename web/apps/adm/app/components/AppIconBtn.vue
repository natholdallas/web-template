<script setup lang="ts">
withDefaults(
  defineProps<{
    icon: string
    tooltip?: string
    tone?: string
    disabled?: boolean
  }>(),
  { tone: 'medium-emphasis', tooltip: '' },
)

defineEmits<{
  (e: 'click'): void
}>()
</script>

<template>
  <VTooltip v-if="tooltip" :text="tooltip">
    <template #activator="{ props }">
      <VBtn
        v-bind="props"
        :icon="icon"
        :color="tone"
        :disabled="disabled"
        variant="text"
        density="comfortable"
        size="small"
        @click="$emit('click')"
      />
    </template>
  </VTooltip>
  <VBtn
    v-else
    :icon="icon"
    :color="tone"
    :disabled="disabled"
    variant="text"
    density="comfortable"
    size="small"
    @click="$emit('click')"
  />
</template>
