<script setup lang="ts">
import type { ResetPasswordIn } from '@sdk/client/adm'

defineProps<{
  loading: boolean
  submitText?: string
}>()
defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<ResetPasswordIn>({ required: true })

const confirm = ref('')

// The dialog is reused across opens; drop the confirmation once the parent
// resets the model after a successful submit.
watch(
  () => model.value.new,
  (v) => {
    if (!v) confirm.value = ''
  },
)

// Bounds mirror the server contract (4-72 characters) so the API's 400 is not
// the first feedback.
const tooShort = computed(() => model.value.new.length > 0 && model.value.new.length < 4)
const tooLong = computed(() => model.value.new.length > 72 || model.value.old.length > 72)
const mismatch = computed(() => !tooShort.value && confirm.value.length > 0 && model.value.new !== confirm.value)
const valid = computed(
  () =>
    model.value.old.length >= 4 &&
    model.value.new.length >= 4 &&
    !tooLong.value &&
    model.value.new === confirm.value,
)
</script>

<template>
  <AppForm :loading="loading" :submit-text="submitText" :submit-disabled="!valid" @submit="$emit('submit')">
    <AppPasswordField v-model="model.old" autocomplete="current-password" :label="$t('settings.password.old')" />
    <AppPasswordField v-model="model.new" autocomplete="new-password" :label="$t('settings.password.new')" />
    <AppPasswordField v-model="confirm" autocomplete="new-password" :label="$t('reset.password.confirm')" />
    <p v-if="tooShort" class="text-body-small text-error">{{ $t('reset.password.min') }}</p>
    <p v-else-if="tooLong" class="text-body-small text-error">{{ $t('reset.password.max') }}</p>
    <p v-else-if="mismatch" class="text-body-small text-error">{{ $t('reset.password.mismatch') }}</p>
  </AppForm>
</template>
