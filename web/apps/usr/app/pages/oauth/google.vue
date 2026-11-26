<script setup lang="ts">
import { Loader2 } from '@lucide/vue'
import { api, GoogleAuthIn } from '@sdk/client/usr'

definePageMeta({
  name: 'oauth-google',
  layout: 'minimal',
})

const auth = useAuthStore()
const route = useRoute()
const loading = ref(true)
const msg = ref('')
let redirect: number | undefined

onMounted(async () => {
  const code = String(route.query.code ?? '')
  const [mode, nonce] = String(route.query.state ?? '').split(':')
  const expected = mode === 'login' || mode === 'bind' ? sessionStorage.getItem('oauth:' + mode) : null
  // Reject callbacks whose nonce does not match the one this tab generated.
  if (!code || !nonce || expected !== nonce) {
    msg.value = 'oauth.failed'
    loading.value = false
    return
  }
  sessionStorage.removeItem('oauth:' + mode)
  try {
    // Returning from Google reloads the app, so the in-memory access token is
    // gone; restore it from the refresh cookie before choosing the branch.
    await auth.$restore()
    if (mode === 'bind' && auth.accessToken) {
      await api.profile.bindGoogle(inst(GoogleAuthIn, { code }))
      await useUserStore().$load(true)
      msg.value = 'oauth.bound'
      redirect = window.setTimeout(() => navigateTo({ name: 'settings' }), 1000)
    } else {
      const data = await api.auth.signInGoogle(inst(GoogleAuthIn, { code }))
      auth.$signIn({ data })
    }
  } catch {
    msg.value = 'oauth.failed'
  }
  loading.value = false
})

onUnmounted(() => {
  clearTimeout(redirect)
})
</script>

<template>
  <ComCtl class="flex items-center justify-center p-4">
    <UiCard class="w-full max-w-112.5 gap-4 p-6 text-center">
      <UiButton class="pointer-events-none" :class="{ hidden: !loading }" variant="ghost" disabled>
        <Loader2 class="size-4 animate-spin" />
      </UiButton>
      <template v-if="!loading">
        <p class="text-muted-foreground">{{ msg ? $t(msg) : '' }}</p>
        <UiButton v-if="msg === 'oauth.failed'" variant="outline" @click="navigateTo({ name: 'entrance' })">
          {{ $t('sign.in') }}
        </UiButton>
      </template>
    </UiCard>
  </ComCtl>
</template>
