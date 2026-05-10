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
  <NForm @submit="$emit('submit')" :loading="loading" :text="text" :schema="schema">
    <UiFormField v-slot="{ componentField }" name="username">
      <UiFormItem>
        <UiFormLabel>{{ $t('user.username') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.username" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
  </NForm>
</template>
