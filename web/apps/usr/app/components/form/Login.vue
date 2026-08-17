<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { AuthIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  text?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<AuthIn>({ default: AuthIn })
const schema = toTypedSchema(
  z.object({
    username: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
    password: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
  }),
)
</script>

<template>
  <UixForm @submit="$emit('submit')" :model="model" :loading="loading" :text="text" :schema="schema">
    <UixFieldText
      v-model="model.username"
      :label="$t('user.username')"
      :placeholder="$t('user.username.placeholder')"
      name="username"
      type="text"
    />
    <UixFieldText
      v-model="model.password"
      :label="$t('user.password')"
      :placeholder="$t('user.password.placeholder')"
      name="password"
      type="password"
    />
  </UixForm>
</template>
