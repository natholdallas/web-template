<script setup lang="ts">
import { KeyRound, Link2, Lock, Palette, Pencil, ShieldCheck, UserRound } from '@lucide/vue'
import { locales } from '~/lib/locale'
import { ResetPasswordForm } from '~/lib/forms'
import { api, ProfileIn } from '@sdk/client/usr'
import type { TelegramAuth } from '@sdk/client/usr'

definePageMeta({
  name: 'settings',
})

const app = useAppStore()
const user = useUserStore()
app.$load()
user.$load()
const isTauri = useIsTauri()
const { profile, loading: finding } = storeToRefs(user)
const { locale } = useI18n()

type Section = 'profile' | 'privacy' | 'appearance'
const section = ref<Section>('profile')
const tabs = [
  { value: 'profile' as const, label: 'settings.tab.profile', icon: UserRound },
  { value: 'privacy' as const, label: 'settings.tab.privacy', icon: ShieldCheck },
  { value: 'appearance' as const, label: 'settings.tab.appearance', icon: Palette },
]

// Username dialog. Seed on open, never on store change, so a background
// profile refresh cannot overwrite typing.
const editOpen = ref(false)
const usrForm = ref(inst(ProfileIn, { username: '' }))

function openEdit() {
  usrForm.value = inst(ProfileIn, { username: profile.value.username || '' })
  editOpen.value = true
}

const { loading: savingUsr, send: saveUsr } = useRequest(() => api.profile.update({ username: usrForm.value.username }), {
  immediate: false,
}).onSuccess(() => {
  editOpen.value = false
  user.$load(true)
})

const pwdOpen = ref(false)
const pwdForm = ref(inst(ResetPasswordForm))

function openPwd() {
  pwdForm.value = inst(ResetPasswordForm)
  pwdOpen.value = true
}

// `confirm` is a client-only form field: send only the two server-known keys.
const { loading: savingPwd, send: savePwd } = useRequest(
  () => api.profile.resetPassword({ old: pwdForm.value.old, new: pwdForm.value.new }),
  { immediate: false },
).onSuccess(
  () => {
    pwdOpen.value = false
    rst(pwdForm.value, ResetPasswordForm)
  },
)

const joined = computed(() => {
  const d = new Date(profile.value.created_at)
  return Number.isNaN(d.getTime()) ? '—' : d.toLocaleDateString(locale.value, { year: 'numeric', month: 'long', day: 'numeric' })
})

// ---- third-party login binding ----
const google = computed(() => app.cfg.oauth?.google)
const telegram = computed(() => app.cfg.oauth?.telegram)
const boundGoogle = computed(() => !!profile.value.googleEmail)
const boundTelegram = computed(() => !!profile.value.telegramName)

function bindGoogle() {
  const url = app.googleAuthURL('bind')
  if (url) window.location.href = url
}

const { loading: bindingTg, send: bindTelegram } = useRequest((d: TelegramAuth) => api.profile.bindTelegram(d), {
  immediate: false,
}).onSuccess(() => {
  user.$load(true)
})

function onTgAuth(d: TelegramAuth | null) {
  if (d) bindTelegram(d)
}

const { loading: unbinding, send: unbind } = useRequest((platform: string) => api.profile.unbindPlatform(platform), {
  immediate: false,
}).onSuccess(() => {
  user.$load(true)
})
</script>

<template>
  <ComCtl :loading="finding" class="p-4 lg:p-6" scroll>
    <div class="mx-auto flex max-w-4xl flex-col gap-6">
      <div class="space-y-1">
        <h1 class="text-2xl font-semibold tracking-tight">{{ $t('settings') }}</h1>
        <p class="text-sm text-muted-foreground">{{ $t('settings.desc') }}</p>
      </div>

      <UiCard class="relative overflow-hidden">
        <div aria-hidden="true" class="pointer-events-none absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-primary/5" />
        <UiCardContent class="relative flex flex-col gap-4 p-6 sm:flex-row sm:items-center sm:gap-6">
          <ProfileAvatarUpload v-model="profile.avatar" class="shrink-0" />
          <div class="min-w-0 flex-1">
            <p class="truncate text-xl font-semibold tracking-tight">
              {{ profile.username || $t('settings.no.username') }}
            </p>
            <div class="mt-1 flex flex-wrap items-center gap-2">
              <UiBadge variant="secondary" class="font-mono">ID {{ profile.id || '—' }}</UiBadge>
            </div>
          </div>
          <UiButton variant="outline" class="self-start sm:self-center" @click="openEdit">
            <Pencil class="size-4" />{{ $t('settings.edit.profile') }}
          </UiButton>
        </UiCardContent>
      </UiCard>

      <UiTabs v-model="section" class="lg:grid lg:grid-cols-[220px_minmax(0,1fr)] lg:items-start lg:gap-6">
        <UiTabsList class="h-auto w-full justify-start overflow-x-auto p-1.5 lg:sticky lg:top-0 lg:w-full lg:flex-col lg:items-stretch">
          <UiTabsTrigger v-for="tab in tabs" :key="tab.value" :value="tab.value" class="h-auto justify-start gap-2 px-3 py-2 lg:flex-none">
            <component :is="tab.icon" class="size-4" />{{ $t(tab.label) }}
          </UiTabsTrigger>
        </UiTabsList>

        <UiTabsContent value="profile" class="flex flex-col gap-4">
          <UiCard>
            <UiCardHeader class="space-y-1">
              <UiCardTitle class="flex items-center gap-2 text-base">
                <span class="flex size-8 items-center justify-center bg-muted text-muted-foreground">
                  <UserRound class="size-4" />
                </span>
                {{ $t('settings.account') }}
              </UiCardTitle>
              <UiCardDescription>{{ $t('settings.account.desc') }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent class="divide-y">
              <div class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                <div class="space-y-0.5">
                  <span class="text-sm text-muted-foreground">{{ $t('user.username') }}</span>
                  <p class="text-sm font-medium">{{ profile.username || $t('settings.no.username') }}</p>
                </div>
                <UiButton variant="outline" size="sm" @click="openEdit"> <Pencil class="size-3.5" />{{ $t('update') }} </UiButton>
              </div>
              <div class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                <div class="space-y-0.5">
                  <span class="text-sm text-muted-foreground">{{ $t('settings.user.id') }}</span>
                  <p class="font-mono text-sm">#{{ profile.id || '—' }}</p>
                </div>
              </div>
              <div class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                <div class="space-y-0.5">
                  <span class="text-sm text-muted-foreground">{{ $t('settings.joined') }}</span>
                  <p class="text-sm font-medium">{{ joined }}</p>
                </div>
              </div>
            </UiCardContent>
          </UiCard>
        </UiTabsContent>

        <UiTabsContent value="privacy" class="flex flex-col gap-4">
          <UiCard>
            <UiCardHeader class="space-y-1">
              <UiCardTitle class="flex items-center gap-2 text-base">
                <span class="flex size-8 items-center justify-center bg-muted text-muted-foreground">
                  <Lock class="size-4" />
                </span>
                {{ $t('settings.change.password') }}
              </UiCardTitle>
              <UiCardDescription>{{ $t('settings.change.password.desc') }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent>
              <div class="flex flex-wrap items-center justify-between gap-4">
                <div class="flex items-center gap-3">
                  <span class="grid size-10 place-items-center bg-muted text-muted-foreground">
                    <KeyRound class="size-4" />
                  </span>
                  <div class="space-y-0.5">
                    <p class="text-sm font-medium">{{ $t('settings.password') }}</p>
                    <p class="font-mono tracking-[0.3em] text-muted-foreground">••••••••••</p>
                  </div>
                </div>
                <UiButton variant="outline" size="sm" @click="openPwd"> <Lock class="size-3.5" />{{ $t('settings.change.password') }} </UiButton>
              </div>
            </UiCardContent>
          </UiCard>

          <UiCard v-if="!isTauri">
            <UiCardHeader class="space-y-1">
              <UiCardTitle class="flex items-center gap-2 text-base">
                <span class="flex size-8 items-center justify-center bg-muted text-muted-foreground">
                  <Link2 class="size-4" />
                </span>
                {{ $t('settings.oauth') }}
              </UiCardTitle>
              <UiCardDescription>{{ $t('settings.oauth.desc') }}</UiCardDescription>
            </UiCardHeader>
            <UiCardContent class="divide-y">
              <div v-if="google?.enabled" class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium">{{ $t('oauth.google') }}</span>
                    <UiBadge :variant="boundGoogle ? 'secondary' : 'outline'">
                      {{ boundGoogle ? $t('oauth.bound') : $t('oauth.bind.desc') }}
                    </UiBadge>
                  </div>
                  <p class="text-sm text-muted-foreground">
                    {{ boundGoogle ? profile.googleEmail || $t('settings.oauth.hint') : $t('settings.oauth.hint') }}
                  </p>
                </div>
                <UiButton v-if="!boundGoogle" variant="outline" size="sm" :disabled="bindingTg || unbinding" @click="bindGoogle">
                  {{ $t('oauth.bind') }}
                </UiButton>
                <UiButton v-else variant="ghost" size="sm" :disabled="bindingTg || unbinding" @click="unbind('google')">
                  {{ $t('oauth.unbind') }}
                </UiButton>
              </div>

              <div v-if="telegram?.enabled" class="flex flex-wrap items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                <div class="space-y-0.5">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium">{{ $t('oauth.telegram') }}</span>
                    <UiBadge :variant="boundTelegram ? 'secondary' : 'outline'">
                      {{ boundTelegram ? $t('oauth.bound') : $t('oauth.bind.desc') }}
                    </UiBadge>
                  </div>
                  <p class="text-sm text-muted-foreground">
                    {{ boundTelegram ? profile.telegramName || $t('settings.oauth.hint') : $t('settings.oauth.hint') }}
                  </p>
                </div>
                <OauthTelegramLogin v-if="!boundTelegram && telegram.botUsername" :bot-username="telegram.botUsername" @auth="onTgAuth" />
                <UiButton v-else-if="boundTelegram" variant="ghost" size="sm" :disabled="bindingTg || unbinding" @click="unbind('telegram')">
                  {{ $t('oauth.unbind') }}
                </UiButton>
              </div>

              <p v-if="!google?.enabled && !telegram?.enabled" class="py-3 text-sm text-muted-foreground">
                {{ $t('oauth.disabled') }}
              </p>
            </UiCardContent>
          </UiCard>
        </UiTabsContent>

        <UiTabsContent value="appearance" class="flex flex-col gap-4">
          <UiCard>
            <UiCardHeader class="space-y-1">
              <UiCardTitle class="flex items-center gap-2 text-base">
                <span class="flex size-8 items-center justify-center bg-muted text-muted-foreground">
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
        </UiTabsContent>
      </UiTabs>

      <UixModal v-model:open="editOpen" :title="$t('settings.edit.profile')" :description="$t('settings.edit.profile.desc')" no-footer>
        <FormProfile v-model="usrForm" :loading="savingUsr" :text="$t('update')" @submit="saveUsr" />
      </UixModal>

      <UixModal v-model:open="pwdOpen" :title="$t('settings.change.password')" :description="$t('settings.change.password.desc')" no-footer>
        <FormRstPwd v-model="pwdForm" :loading="savingPwd" :text="$t('update')" @submit="savePwd" />
      </UixModal>
    </div>
  </ComCtl>
</template>
