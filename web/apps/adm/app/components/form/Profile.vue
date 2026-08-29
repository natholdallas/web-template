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
  <VxForm @submit="$emit('submit')" :loading="loading" :submit-text="submitText">
    <VTextField v-model="model.username" :label="$t('settings.username')" />
  </VxForm>
</template>
