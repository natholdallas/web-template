<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'
import { Apis, Admin, ProfileIn, ResetPasswordIn } from '~/lib/sdk'
import type { ThemeMode } from '~/composables/conf'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

const conf = useConf()
const { locale: currentLocale } = useI18n()

const {
  loading: finding,
  data: admin,
  send: findAdmin,
} = useRequest(Apis.Profile.find, {
  initialData: inst(Admin),
}).onSuccess(({ data }) => {
  profile.value.username = data.username
})
const profile = ref(inst(ProfileIn, { username: '' }))
const password = ref(inst(ResetPasswordIn))

const { loading: updatingProfile, send: sendProfile } = useRequest(
  () =>
    Apis.Profile.update({
      data: { username: profile.value.username || '' },
    }),
  { immediate: false },
).onSuccess(() => {
  findAdmin()
})
const { loading: updatingPassword, send: sendPassword } = useRequest(
  () =>
    Apis.Profile.resetPassword({
      data: password.value,
    }),
  {
    immediate: false,
  },
)

const themeOptions: { title: string; value: ThemeMode; icon: string }[] = [
  { title: 'settings.theme.system', value: 'system', icon: 'mdi-monitor' },
  { title: 'settings.theme.light', value: 'light', icon: 'mdi-white-balance-sunny' },
  { title: 'settings.theme.dark', value: 'dark', icon: 'mdi-moon-waning-crescent' },
]

const localeOptions = Object.values(Locale).map((l) => ({
  title: locales[l].k,
  value: l,
}))
</script>

<template>
  <ComCtl :loading="finding" class="p-4 lg:p-6" scroll>
    <div class="mx-auto max-w-3xl flex flex-col gap-4">
      <VCard border rounded="xl" class="overflow-hidden">
        <div class="flex items-center gap-4 border-b border-on-surface/10 px-6 py-5">
          <VAvatar :color="admin.username ? 'primary' : 'surface-variant'" size="52" rounded="lg">
            <span class="text-lg font-bold">{{ admin.username?.slice(0, 1).toUpperCase() || '?' }}</span>
          </VAvatar>
          <div class="min-w-0">
            <p class="truncate text-lg font-semibold">{{ admin.username || $t('settings.no.username') }}</p>
            <p class="text-sm opacity-70">#{{ admin.id }}</p>
          </div>
        </div>
        <VCardText>
          <FormProfile v-model="profile" :loading="updatingProfile" @submit="sendProfile" />
        </VCardText>
      </VCard>

      <VCard border rounded="xl" class="overflow-hidden">
        <div class="flex items-center gap-3 px-6 pt-5">
          <div class="flex size-10 items-center justify-center rounded-xl bg-warning/10 text-warning">
            <VIcon icon="mdi-lock-outline" size="22" />
          </div>
          <div>
            <VCardTitle class="px-0 text-base">{{ $t('settings.password') }}</VCardTitle>
            <VCardSubtitle class="px-0">{{ $t('settings.password.desc') }}</VCardSubtitle>
          </div>
        </div>
        <VCardText>
          <FormPasswd v-model="password" :loading="updatingPassword" @submit="sendPassword" />
        </VCardText>
      </VCard>

      <div class="grid gap-4 md:grid-cols-2">
        <VCard border rounded="xl" class="overflow-hidden">
          <div class="flex items-center gap-3 px-6 pt-5">
            <div class="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <VIcon icon="mdi-theme-light-dark" size="22" />
            </div>
            <div>
              <VCardTitle class="px-0 text-base">{{ $t('settings.theme') }}</VCardTitle>
              <VCardSubtitle class="px-0">{{ $t('settings.theme.desc') }}</VCardSubtitle>
            </div>
          </div>
          <VCardText>
            <div class="flex w-full gap-1 rounded-md p-1">
              <VBtn
                v-for="opt in themeOptions"
                :key="opt.value"
                :color="conf.theme === opt.value ? 'primary' : undefined"
                :variant="conf.theme === opt.value ? 'flat' : 'text'"
                class="flex-1"
                @click="conf.theme = opt.value"
              >
                <VIcon class="mr-1" :icon="opt.icon" size="18" />
                <span class="text-xs">{{ $t(opt.title) }}</span>
              </VBtn>
            </div>
          </VCardText>
        </VCard>

        <VCard border rounded="xl" class="overflow-hidden">
          <div class="flex items-center gap-3 px-6 pt-5">
            <div class="flex size-10 items-center justify-center rounded-xl bg-info/10 text-info">
              <VIcon icon="mdi-translate" size="22" />
            </div>
            <div>
              <VCardTitle class="px-0 text-base">{{ $t('settings.locale') }}</VCardTitle>
              <VCardSubtitle class="px-0">{{ $t('settings.locale.desc') }}</VCardSubtitle>
            </div>
          </div>
          <VCardText>
            <div class="flex w-full gap-1 rounded-md p-1">
              <VBtn
                v-for="opt in localeOptions"
                :key="opt.value"
                :color="currentLocale === opt.value ? 'primary' : undefined"
                :variant="currentLocale === opt.value ? 'flat' : 'text'"
                size="small"
                class="flex-1"
                @click="$i18n.setLocale(opt.value)"
              >
                <span class="text-xs">{{ $t(opt.title) }}</span>
              </VBtn>
            </div>
          </VCardText>
        </VCard>
      </div>
    </div>
  </ComCtl>
</template>
