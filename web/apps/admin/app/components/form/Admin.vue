<script setup lang="ts">
import type { Admin } from '~/sdk'

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
  <NForm
    @submit="$emit('submit')"
    :loading="loading"
    :submit-text="submitText"
    :readonly="readonly"
  >
    <VTextField v-model="model.username" :label="$t('user.username')" />
    <NPasswdField v-model="model.password" :label="$t('user.password')" />
  </NForm>
</template>
