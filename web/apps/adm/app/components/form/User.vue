<script setup lang="ts">
import type { UserIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  submitText?: string
  readonly?: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<UserIn>({ required: true })
</script>

<template>
  <VxForm :loading="loading" :readonly="readonly" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField v-model="model.username" :label="$t('user.username')" />
    <VxPasswdField v-model="model.password" :label="$t('user.password')" />
  </VxForm>
</template>
