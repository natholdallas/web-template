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
  <NForm @submit="$emit('submit')" :loading="loading" :text="text" :schema="schema">
    <UiFormField v-slot="{ componentField }" name="username">
      <UiFormItem>
        <UiFormLabel>{{ $t('user.username') }}</UiFormLabel>
        <UiFormControl>
          <UiInput
            v-bind="componentField"
            v-model="model.username"
            :placeholder="$t('user.username.placeholder')"
            type="text"
            autocomplete="username"
          />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="password">
      <UiFormItem>
        <UiFormLabel>{{ $t('user.password') }}</UiFormLabel>
        <UiFormControl>
          <UiInput
            v-bind="componentField"
            v-model="model.password"
            :placeholder="$t('user.password.placeholder')"
            type="password"
            autocomplete="current-password"
          />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
  </NForm>
</template>
