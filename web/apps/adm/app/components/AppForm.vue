<script setup lang="ts">
defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
  error?: string
  submitDisabled?: boolean
}>()

defineEmits<{
  (e: 'submit'): void
}>()
</script>

<template>
  <VForm fast-fail class="flex flex-col gap-4" @submit.prevent="$emit('submit')">
    <VAlert v-if="error" type="error" variant="tonal">{{ error }}</VAlert>

    <slot />

    <VBtn
      type="submit"
      color="primary"
      size="large"
      block
      class="mt-1 text-none font-weight-semibold"
      :loading="loading"
      :disabled="readonly || submitDisabled"
    >
      {{ submitText || $t('submit') }}
    </VBtn>
  </VForm>
</template>
