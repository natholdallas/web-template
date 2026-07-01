<script setup lang="ts">
import type { User } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<User>({ required: true })
</script>

<template>
  <VxForm @submit="$emit('submit')" :loading="loading" :submit-text="submitText" :readonly="readonly">
    <VTextField v-model="model.username" :label="$t('user.username')" />
    <VxPasswdField v-model="model.password" :label="$t('user.password')" />
  </VxForm>
</template>
