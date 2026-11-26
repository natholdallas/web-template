<script setup lang="ts">
import type { ProfileIn } from '@sdk/client/adm'

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
  <AppForm :loading="loading" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField
      v-model="model.username"
      autocomplete="username"
      :label="$t('settings.username')"
      variant="outlined"
      density="comfortable"
      hide-details="auto"
    />
  </AppForm>
</template>
