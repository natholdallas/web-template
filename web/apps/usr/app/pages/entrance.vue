<script setup lang="ts">
import { Apis, AuthIn } from '~/lib/sdk'
import type { TelegramAuth } from '~/lib/sdk'

definePageMeta({
  name: 'entrance',
  middleware: 'entrance',
  layout: 'minimal',
})

const auth = useAuthStore()
const app = useAppStore()
app.$load()
const form = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(
  () =>
    Apis.Auth.signIn({
      data: form.value,
    }),
  {
    immediate: false,
  },
).onSuccess(auth.$signIn)

const google = computed(() => app.cfg.oauth?.google)
const telegram = computed(() => app.cfg.oauth?.telegram)

function signInGoogle() {
  const url = app.googleAuthURL('login')
  if (url) window.location.href = url
}

const { send: signInTelegram } = useRequest(
  (d: TelegramAuth) =>
    Apis.Auth.signInTelegram({
      data: d,
    }),
  {
    immediate: false,
  },
).onSuccess(auth.$signIn)

function onTgAuth(d: TelegramAuth | null) {
  if (d) signInTelegram(d)
}
</script>

<template>
  <ComCtl class="flex justify-center items-center">
    <UiCard class="w-112.5">
      <UiCardHeader>
        <UiCardTitle>{{ $t('sign.in') }}</UiCardTitle>
        <UiCardDescription>{{ $t('sign.in.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormLogin v-model="form" :loading="signing" @submit="signIn" />

        <template v-if="google?.enabled || telegram?.enabled">
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
              <TelegramLogin
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