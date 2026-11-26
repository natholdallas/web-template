<script setup lang="ts">
import { api, AuthIn } from '@sdk/client/usr'
import type { TelegramAuth } from '@sdk/client/usr'

definePageMeta({
  name: 'entrance',
  layout: 'minimal',
})

const auth = useAuthStore()
const app = useAppStore()
app.$load()
const isTauri = useIsTauri()
const form = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(() => api.auth.signIn(form.value), {
  immediate: false,
}).onSuccess(auth.$signIn)

const google = computed(() => app.cfg.oauth?.google)
const telegram = computed(() => app.cfg.oauth?.telegram)

function signInGoogle() {
  const url = app.googleAuthURL('login')
  if (url) window.location.href = url
}

const { send: signInTelegram } = useRequest((d: TelegramAuth) => api.auth.signInTelegram(d), {
  immediate: false,
}).onSuccess(auth.$signIn)

function onTgAuth(d: TelegramAuth | null) {
  if (d) signInTelegram(d)
}
</script>

<template>
  <ComCtl class="flex items-center justify-center p-4">
    <UiCard class="w-full max-w-112.5">
      <UiCardHeader>
        <UiCardTitle>{{ $t('sign.in') }}</UiCardTitle>
        <UiCardDescription>{{ $t('sign.in.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormLogin v-model="form" :loading="signing" :text="$t('sign.in')" @submit="signIn" />

        <template v-if="!isTauri && (google?.enabled || telegram?.enabled)">
          <div class="my-5 flex items-center gap-3">
            <UiSeparator class="flex-1" />
            <span class="text-xs text-muted-foreground">{{ $t('oauth.divider') }}</span>
            <UiSeparator class="flex-1" />
          </div>

          <div class="flex flex-col gap-2">
            <UiButton v-if="google?.enabled" variant="outline" type="button" @click="signInGoogle">
              {{ $t('sign.in.google') }}
            </UiButton>
            <div v-if="telegram?.enabled" class="flex min-h-10 items-center justify-center">
              <OauthTelegramLogin
                v-if="telegram.botUsername"
                :bot-username="telegram.botUsername"
                @auth="onTgAuth"
              />
            </div>
          </div>
        </template>
      </UiCardContent>
    </UiCard>
  </ComCtl>
</template>