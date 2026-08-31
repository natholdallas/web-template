<script setup lang="ts">
import type { AdminIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<AdminIn>({ required: true })
</script>

<template>
  <VxForm :loading="loading" :readonly="readonly" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField v-model="model.username" :label="$t('admin.username')" />
    <VxPasswdField v-model="model.password" :label="$t('admin.password')" />
  </VxForm>
</template>
