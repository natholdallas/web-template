<script setup lang="ts">
import type { ResetPasswordIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<ResetPasswordIn>({ required: true })
</script>

<template>
  <VxForm @submit="$emit('submit')" :loading="loading" :submit-text="submitText">
    <VxPasswdField v-model="model.old" :label="$t('settings.password.old')" />
    <VxPasswdField v-model="model.new" :label="$t('settings.password.new')" />
  </VxForm>
</template>
