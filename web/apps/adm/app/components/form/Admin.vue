<script setup lang="ts">
import type { Admin } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<Admin>({ required: true })
</script>

<template>
  <VxForm @submit="$emit('submit')" :loading="loading" :submit-text="submitText" :readonly="readonly">
    <VTextField v-model="model.username" :label="$t('admin.username')" />
    <VxPasswdField v-model="model.password" :label="$t('admin.password')" />
  </VxForm>
</template>
