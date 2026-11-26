<script setup lang="ts">
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { AuthIn } from '@sdk/client/usr'

defineProps<{
  loading: boolean
  text?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<AuthIn>({ default: () => inst(AuthIn) })
// Computed so a locale switch re-runs the resolvers; messages are callbacks
// for the same reason (a plain string freezes the text at setup time).
const schema = computed(() =>
  toTypedSchema(
    z.object({
      username: z
        .string({ message: $t('va.required') })
        .min(4, { message: $t('va.min', { v: 4 }) })
        .max(50, { message: $t('va.max', { v: 50 }) }),
      password: z
        .string({ message: $t('va.required') })
        .min(4, { message: $t('va.min', { v: 4 }) })
        .max(72, { message: $t('va.max', { v: 72 }) }),
    }),
  ),
)
</script>

<template>
  <UixForm :model="model" :loading="loading" :text="text" :schema="schema" @submit="$emit('submit')">
    <UixFieldText v-model="model.username" :label="$t('user.username')" :placeholder="$t('user.username.placeholder')" name="username" type="text" />
    <UixFieldText
      v-model="model.password"
      :label="$t('user.password')"
      :placeholder="$t('user.password.placeholder')"
      name="password"
      type="password"
      maxlength="72"
    />
  </UixForm>
</template>
