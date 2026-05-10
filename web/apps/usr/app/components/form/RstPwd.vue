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
  <NForm @submit="$emit('submit')" :loading="loading" :text="text" :schema="schema">
    <UiFormField v-slot="{ componentField }" name="old">
      <UiFormItem>
        <UiFormLabel>{{ $t('settings.old.password') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.old" type="password" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="new">
      <UiFormItem>
        <UiFormLabel>{{ $t('settings.new.password') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.new" type="password" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="confirm">
      <UiFormItem>
        <UiFormLabel>{{ $t('settings.confirm.password') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.confirm" type="password" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
  </NForm>
</template>
