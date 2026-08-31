<script setup lang="ts">
import type { ProfileIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<ProfileIn>({ required: true })
</script>

<template>
  <VxForm :loading="loading" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField v-model="model.username" :label="$t('settings.username')" />
  </VxForm>
</template>
