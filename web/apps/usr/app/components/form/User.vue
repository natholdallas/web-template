<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { User } from '~/lib/sdk'

defineProps<{
  text?: string
  loading: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<User>({ required: true })
const schema = toTypedSchema(
  z.object({
    username: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
  }),
)
</script>

<template>
  <UixForm @submit="$emit('submit')" :loading="loading" :text="text" :schema="schema" :model="model">
    <UixFieldText v-model="model.username" name="username" :label="$t('user.username')" />
  </UixForm>
</template>
