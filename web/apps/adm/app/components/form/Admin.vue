<script setup lang="ts">
import type { AdminIn } from '@sdk/client/adm'

withDefaults(
  defineProps<{
    loading: boolean
    submitText?: string
    readonly?: boolean
    showPassword?: boolean
  }>(),
  { submitText: '', readonly: false, showPassword: true },
)
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<AdminIn>({ required: true })
</script>

<template>
  <AppForm :loading="loading" :readonly="readonly" :submit-text="submitText" @submit="$emit('submit')">
    <VTextField
      v-model="model.username"
      autocomplete="off"
      :label="$t('admin.username')"
      variant="outlined"
      density="comfortable"
      hide-details="auto"
    />
    <!-- Updates never change the password; it is reset through its own action. -->
    <AppPasswordField
      v-if="showPassword"
      v-model="model.password"
      autocomplete="new-password"
      :label="$t('admin.password')"
    />
  </AppForm>
</template>
