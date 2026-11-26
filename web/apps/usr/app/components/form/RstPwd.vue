<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { ResetPasswordForm } from '~/lib/forms'

defineProps<{
  loading: boolean
  text?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<ResetPasswordForm>({ default: () => inst(ResetPasswordForm) })
const schema = computed(() =>
  toTypedSchema(
    z
      .object({
        old: z
          .string({ message: $t('va.required') })
          .min(4, { message: $t('va.min', { v: 4 }) })
          .max(72, { message: $t('va.max', { v: 72 }) }),
        new: z
          .string({ message: $t('va.required') })
          .min(4, { message: $t('va.min', { v: 4 }) })
          .max(72, { message: $t('va.max', { v: 72 }) }),
        confirm: z.string({ message: $t('va.required') }),
      })
      .refine((data) => data.new === data.confirm, {
        message: $t('va.eq.passwd', { v: 4 }),
        path: ['confirm'],
      }),
  ),
)
</script>

<template>
  <UixForm :loading="loading" :text="text" :schema="schema" :model="model" @submit="$emit('submit')">
    <UixFieldText v-model="model.old" :label="$t('settings.old.password')" name="old" type="password" maxlength="72" />
    <UixFieldText v-model="model.new" :label="$t('settings.new.password')" name="new" type="password" maxlength="72" />
    <UixFieldText v-model="model.confirm" :label="$t('settings.confirm.password')" name="confirm" type="password" />
  </UixForm>
</template>
