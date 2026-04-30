<script setup lang="ts">
import type { AuthIn } from '~/sdk'

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
  <NForm
    @submit="$emit('submit')"
    :loading="loading"
    :submit-text="submitText"
    :readonly="readonly"
  >
    <VTextField
      v-model="model.username"
      :label="$t('user.username')"
      :counter="100"
      autocomplete="username"
    />
    <NPasswdField
      v-model="model.password"
      :label="$t('user.password')"
      autocomplete="current-password"
    />
  </NForm>
</template>
