<script setup lang="ts">
import type { AuthIn } from '@sdk/client/adm'

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
  <AppForm :loading="loading" :readonly="readonly" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField
      v-model="model.username"
      autocomplete="username"
      :counter="100"
      :label="$t('user.username')"
      variant="outlined"
      density="comfortable"
      hide-details="auto"
    />
    <AppPasswordField v-model="model.password" autocomplete="current-password" :label="$t('user.password')" />
  </AppForm>
</template>
