<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    description?: string
    icon?: string
    maxWidth?: string | number
    noFooter?: boolean
    loading?: boolean
    confirmText?: string
    tone?: 'primary' | 'error'
  }>(),
  { maxWidth: 520, tone: 'primary', title: '', description: '', icon: '', confirmText: '' },
)

const emit = defineEmits<{ cancel: []; confirm: [] }>()

const open = defineModel<boolean>({ default: false })

function cancel() {
  open.value = false
  emit('cancel')
}
</script>

<template>
  <VDialog v-model="open" :max-width="maxWidth" persistent>
    <VCard rounded="lg" class="overflow-hidden">
      <div class="flex items-start gap-3 px-6 pt-6">
        <div
          v-if="icon"
          class="grid size-10 shrink-0 place-items-center rounded-xl"
          :class="tone === 'error' ? 'bg-error/10 text-error' : 'bg-primary/10 text-primary'"
        >
          <VIcon :icon="icon" size="22" />
        </div>
        <div class="min-w-0 flex-1">
          <p v-if="title" class="text-title-medium font-weight-bold text-truncate">{{ title }}</p>
          <p v-if="description" class="text-body-medium text-medium-emphasis">{{ description }}</p>
        </div>
        <VBtn icon="mdi-close" variant="text" size="small" class="ml-auto" @click="open = false" />
      </div>

      <VCardText class="max-h-[70vh] overflow-y-auto p-6">
        <slot />
      </VCardText>

      <VCardActions v-if="!noFooter" class="border-t border-on-surface/10 px-6 py-4">
        <VSpacer />
        <VBtn variant="text" class="text-none" @click="cancel">{{ $t('cancel') }}</VBtn>
        <VBtn
          :color="tone === 'error' ? 'error' : 'primary'"
          :loading="loading"
          class="text-none font-weight-semibold"
          @click="$emit('confirm')"
        >
          {{ confirmText || $t('confirm') }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>
