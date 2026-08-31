<script setup lang="ts">
import type { AuthIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<AuthIn>({ required: true })
</script>

<template>
  <VxForm :loading="loading" :readonly="readonly" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField v-model="model.username" autocomplete="username" :counter="100" :label="$t('user.username')" />
    <VxPasswdField v-model="model.password" autocomplete="current-password" :label="$t('user.password')" />
  </VxForm>
</template>
