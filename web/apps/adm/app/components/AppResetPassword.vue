<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    loading: boolean
    result?: string
  }>(),
  { result: '' },
)

const emit = defineEmits<{
  confirm: [password: string]
}>()

const open = defineModel<boolean>({ default: false })
const { t } = useI18n()

const mode = ref<'random' | 'manual'>('random')
const password = ref('')
const confirmPassword = ref('')

// Reopening must never leak the previous target's input.
watch(open, (v) => {
  if (!v) return
  mode.value = 'random'
  password.value = ''
  confirmPassword.value = ''
})

const error = computed(() => {
  if (mode.value !== 'manual' || password.value.length >= 4) {
    return password.value !== confirmPassword.value && confirmPassword.value ? t('reset.password.mismatch') : ''
  }
  return t('reset.password.min')
})

const valid = computed(() => mode.value === 'random' || (password.value.length >= 4 && password.value === confirmPassword.value))

function submit() {
  if (!valid.value) return
  emit('confirm', mode.value === 'manual' ? password.value : '')
}
</script>

<template>
  <AppModal v-model="open" :title="$t('reset.password')" icon="mdi-lock-reset" no-footer>
    <template v-if="props.result">
      <VAlert type="success" variant="tonal" density="comfortable" class="mb-4">
        {{ $t('reset.password.result') }}
      </VAlert>
      <VTextField
        :model-value="props.result"
        readonly
        :clearable="false"
        variant="outlined"
        density="comfortable"
        append-inner-icon="mdi-content-copy"
        @click:append-inner="vcopy(props.result, t('success'))"
      />
      <div class="mt-4 flex justify-end">
        <VBtn variant="text" class="text-none" @click="open = false">{{ $t('reset.password.close') }}</VBtn>
      </div>
    </template>

    <template v-else>
      <VAlert type="warning" variant="tonal" density="comfortable" class="mb-4">
        {{ $t('reset.password.warning') }}
      </VAlert>

      <p class="mb-1 text-body-small font-weight-medium text-medium-emphasis">{{ $t('reset.password.mode') }}</p>
      <VRadioGroup v-model="mode" hide-details class="mb-3">
        <VRadio :label="$t('reset.password.random')" value="random" />
        <VRadio :label="$t('reset.password.manual')" value="manual" />
      </VRadioGroup>

      <div v-if="mode === 'manual'" class="flex flex-col gap-3">
        <AppPasswordField v-model="password" autocomplete="new-password" :label="$t('reset.password.new')" />
        <AppPasswordField
          v-model="confirmPassword"
          autocomplete="new-password"
          :label="$t('reset.password.confirm')"
        />
        <p v-if="error" class="text-body-small text-error">{{ error }}</p>
      </div>

      <div class="mt-4 flex justify-end gap-2">
        <VBtn variant="text" class="text-none" @click="open = false">{{ $t('cancel') }}</VBtn>
        <VBtn color="error" class="text-none" :loading="loading" :disabled="!valid" @click="submit">
          {{ $t('confirm') }}
        </VBtn>
      </div>
    </template>
  </AppModal>
</template>
