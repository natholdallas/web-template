<script setup lang="ts">
import { Link2, Lock, Palette, UserRound } from '@lucide/vue'
import { locales } from '~/lib/locale'
import { Apis, Profile, ResetPasswordIn } from '~/lib/sdk'
import type { TelegramAuth } from '~/lib/sdk'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

const app = useAppStore()
app.$load()

const {
  loading: findingUser,
  data: user,
  send: findUser,
} = useRequest(Apis.Profile.find, {
  initialData: inst(Profile),
}).onSuccess(({ data }) => {
  usrForm.value = cpm(data)
})
const usrForm = ref(inst(Profile))
const { loading: savingUsr, send: saveUsr } = useRequest(
  () =>
    Apis.Profile.update({
      data: { username: usrForm.value.username, avatar: usrForm.value.avatar || '' },
    }),
  {
    immediate: false,
  },
).onSuccess(() => {
  rst(usrForm.value, Profile)
  findUser()
})
const pwdForm = ref(inst(ResetPasswordIn))
const { loading: savingPwd, send: savePwd } = useRequest(
  () =>
    Apis.Profile.resetPassword({
      data: pwdForm.value,
    }),
  {
    immediate: false,
  },
).onSuccess(() => {
  rst(pwdForm.value, ResetPasswordIn)
})

// ---- third-party login binding ----
const google = computed(() => app.cfg.oauth?.google)
const telegram = computed(() => app.cfg.oauth?.telegram)

function bindGoogle() {
  const url = app.googleAuthURL('bind')
  if (url) window.location.href = url
}

const {
  loading: bindingTg,
  send: bindTelegram,
} = useRequest((d: TelegramAuth) => Apis.Profile.bindTelegram({ data: d }), {
  immediate: false,
}).onSuccess(() => {
  findUser()
})

function onTgAuth(d: TelegramAuth | null) {
  if (d) bindTelegram(d)
}

const {
  loading: unbinding,
  send: unbind,
} = useRequest((platform: string) => Apis.Profile.unbindPlatform({ pathParams: { platform } }), {
  immediate: false,
}).onSuccess(() => {
  findUser()
})

const boundGoogle = computed(() => !!user.value.googleEmail)
const boundTelegram = computed(() => !!user.value.telegramName)
</script>

<template>
  <ComCtl :loading="findingUser" class="p-4 lg:p-6" scroll>
    <div class="mx-auto max-w-3xl flex flex-col gap-4">
      <UiCard>
        <UiCardContent class="flex items-center gap-4 p-6 sm:gap-5">
          <MediaUpload v-model="usrForm.avatar" :kind="'image'" class="shrink-0" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-lg font-semibold">{{ user.username || $t('settings.no.username') }}</p>
            <p class="mt-1 text-sm text-muted-foreground">ID: {{ user.id }}</p>
          </div>
        </UiCardContent>
      </UiCard>

      <UiCard class="flex flex-col">
        <UiCardHeader class="space-y-1">
          <UiCardTitle class="flex items-center gap-2 text-base">
            <span class="flex size-8 items-center justify-center rounded-lg bg-muted text-muted-foreground">
              <UserRound class="size-4" />
            </span>
            {{ $t('settings.edit.profile') }}
          </UiCardTitle>
          <UiCardDescription>{{ $t('settings.edit.profile.desc') }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent class="flex-1">
          <FormProfile v-model="usrForm" :loading="savingUsr" class="flex h-full flex-col" @submit="saveUsr" />
        </UiCardContent>
      </UiCard>

      <UiCard class="flex flex-col">
        <UiCardHeader class="space-y-1">
          <UiCardTitle class="flex items-center gap-2 text-base">
            <span class="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
              <Link2 class="size-4" />
            </span>
            {{ $t('settings.oauth') }}
          </UiCardTitle>
          <UiCardDescription>{{ $t('settings.oauth.desc') }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent class="flex flex-col divide-y">
          <!-- Google -->
          <div v-if="google?.enabled" class="flex flex-wrap items-center justify-between gap-3 py-3">
            <div class="space-y-0.5">
              <span class="text-sm font-medium">{{ $t('oauth.google') }}</span>
              <p class="text-sm text-muted-foreground">
                {{ boundGoogle ? (user.googleEmail || $t('oauth.bound')) : $t('oauth.bind.desc') }}
              </p>
            </div>
            <UiButton
              v-if="!boundGoogle"
              variant="outline"
              size="sm"
              :disabled="bindingTg || unbinding"
              @click="bindGoogle"
            >
              {{ $t('oauth.bind') }}
            </UiButton>
            <UiButton
              v-else
              variant="ghost"
              size="sm"
              :disabled="bindingTg || unbinding"
              @click="unbind('google')"
            >
              {{ $t('oauth.unbind') }}
            </UiButton>
          </div>

          <!-- Telegram -->
          <div v-if="telegram?.enabled" class="flex flex-wrap items-center justify-between gap-3 py-3">
            <div class="space-y-0.5">
              <span class="text-sm font-medium">{{ $t('oauth.telegram') }}</span>
              <p class="text-sm text-muted-foreground">
                {{ boundTelegram ? user.telegramName : $t('oauth.bind.desc') }}
              </p>
            </div>
            <TelegramLogin
              v-if="!boundTelegram && telegram.botUsername"
              :bot-username="telegram.botUsername"
              @auth="onTgAuth"
            />
            <UiButton
              v-else-if="boundTelegram"
              variant="ghost"
              size="sm"
              :disabled="bindingTg || unbinding"
              @click="unbind('telegram')"
            >
              {{ $t('oauth.unbind') }}
            </UiButton>
          </div>

          <p v-if="!google?.enabled && !telegram?.enabled" class="py-3 text-sm text-muted-foreground">
            {{ $t('oauth.disabled') }}
          </p>
        </UiCardContent>
      </UiCard>

      <UiCard class="flex flex-col">
        <UiCardHeader class="space-y-1">
          <UiCardTitle class="flex items-center gap-2 text-base">
            <span class="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
              <Lock class="size-4" />
            </span>
            {{ $t('settings.change.password') }}
          </UiCardTitle>
          <UiCardDescription>{{ $t('settings.change.password.desc') }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent class="flex-1">
          <FormRstPwd v-model="pwdForm" :loading="savingPwd" class="flex h-full flex-col" @submit="savePwd" />
        </UiCardContent>
      </UiCard>

      <UiCard>
        <UiCardHeader class="space-y-1">
          <UiCardTitle class="flex items-center gap-2 text-base">
            <span class="flex size-8 items-center justify-center rounded-md bg-muted text-muted-foreground">
              <Palette class="size-4" />
            </span>
            {{ $t('settings.appearance') }}
          </UiCardTitle>
          <UiCardDescription>{{ $t('settings.appearance.desc') }}</UiCardDescription>
        </UiCardHeader>
        <UiCardContent class="flex flex-col gap-5">
          <div class="space-y-2.5">
            <div class="space-y-0.5">
              <span class="text-sm font-medium">{{ $t('settings.theme') }}</span>
              <p class="text-sm text-muted-foreground">{{ $t('settings.theme.desc') }}</p>
            </div>
            <ThemeSelect />
          </div>
          <UiSeparator />
          <div class="flex flex-wrap items-center justify-between gap-4">
            <div class="space-y-0.5">
              <span class="text-sm font-medium">{{ $t('locale') }}</span>
              <p class="text-sm text-muted-foreground">{{ $t('locale.desc') }}</p>
            </div>
            <SwitchLang
              :options="Object.values(locales).map(({ k, v }) => ({ label: $t(k), key: k, value: v }))"
              :value="$t(locales[$i18n.locale].k)"
              @update="$i18n.setLocale"
            />
          </div>
        </UiCardContent>
      </UiCard>
    </div>
  </ComCtl>
</template>