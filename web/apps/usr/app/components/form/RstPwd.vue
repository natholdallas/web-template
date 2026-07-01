<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { ResetPasswordIn } from '~/lib/sdk'

defineProps<{
  loading: boolean
  text?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<ResetPasswordIn>({ default: ResetPasswordIn })
const schema = toTypedSchema(
  z
    .object({
      old: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
      new: z.string({ message: $t('va.required') }).min(4, { message: $t('va.min', { v: 4 }) }),
      confirm: z.string({ message: $t('va.required') }),
    })
    .refine((data) => data.new === data.confirm, {
      message: $t('va.eq.passwd', { v: 6 }),
      path: ['confirm'],
    }),
)
</script>

<template>
  <UixForm @submit="$emit('submit')" :loading="loading" :text="text" :schema="schema" :model="model">
    <UixFieldText v-model="model.old" :label="$t('settings.old.password')" name="old" type="password" />
    <UixFieldText v-model="model.new" :label="$t('settings.new.password')" name="new" type="password" />
    <UixFieldText v-model="model.confirm" :label="$t('settings.confirm.password')" name="confirm" type="password" />
  </UixForm>
</template>
