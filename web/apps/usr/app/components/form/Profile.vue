<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import type { Profile } from '~/lib/sdk'

defineProps<{
  text?: string
  loading: boolean
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<Profile>({ required: true })
const schema = toTypedSchema(
  z.object({
    username: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
  }),
)
</script>

<template>
  <UixForm :loading="loading" :text="text" :schema="schema" :model="model" @submit="$emit('submit')">
    <UixFieldText v-model="model.username" name="username" :label="$t('user.username')" />
  </UixForm>
</template>
