<script setup lang="ts">
import type { TelegramAuth } from '~/lib/sdk'

const props = defineProps<{
  botUsername: string
}>()

const emit = defineEmits<{
  (e: 'auth', user: TelegramAuth | null): void
}>()

const widget = ref<HTMLElement>()
// unique callback name so re-mounted widgets never clash
const cbName = `onTelegramAuth${Date.now().toString(36)}`
const reg = window as unknown as Record<string, unknown>

function handleUser(user: unknown) {
  const u = user as TelegramAuth | null | undefined
  if (u && u.id && u.hash) {
    emit('auth', u)
  } else {
    emit('auth', null)
  }
}

onMounted(() => {
  reg[cbName] = handleUser
  const script = document.createElement('script')
  script.async = true
  script.src = 'https://telegram.org/js/telegram-widget.js?22'
  script.setAttribute('data-telegram-login', props.botUsername)
  script.setAttribute('data-size', 'large')
  script.setAttribute('data-request-access', 'write')
  script.onerror = () => emit('auth', null)
  widget.value?.appendChild(script)
})

onUnmounted(() => {
  reg[cbName] = undefined
})
</script>

<template>
  <div ref="widget" class="inline-flex min-h-10 items-center" :aria-label="$t('sign.in.telegram')" />
</template>