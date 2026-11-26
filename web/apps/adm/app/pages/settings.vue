<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'
import { ProfileIn, ResetPasswordIn, api } from '@sdk/client/adm'
import type { ThemeMode } from '~/stores/conf'

definePageMeta({
  name: 'settings',
})

const conf = useConfStore()
const { locale: currentLocale } = useI18n()

// One shared profile store drives the settings card, the app bar and the
// sidebar, so an avatar upload is reflected everywhere without a reload.
const store = useProfileStore()
const { profile, loading: finding } = storeToRefs(store)
store.$load()

const editOpen = ref(false)
const profileForm = ref(inst(ProfileIn, { username: '' }))
const passwordOpen = ref(false)
const password = ref(inst(ResetPasswordIn))

// Seed the form when the dialog opens, not on every store change: a background
// profile refresh must not overwrite what the user is typing.
function openEdit() {
  profileForm.value = inst(ProfileIn, { username: profile.value.username || '' })
  editOpen.value = true
}

function openPassword() {
  password.value = inst(ResetPasswordIn)
  passwordOpen.value = true
}

const { loading: updatingProfile, send: sendProfile } = useRequest(
  () => api.profile.update({ username: profileForm.value.username || '' }),
  { immediate: false },
).onSuccess(() => {
  editOpen.value = false
  store.$load(true)
})
const { loading: updatingPassword, send: sendPassword } = useRequest(() => api.profile.resetPassword(password.value), {
  immediate: false,
}).onSuccess(() => {
  passwordOpen.value = false
  password.value = inst(ResetPasswordIn)
})

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
    <div class="mx-auto flex max-w-6xl flex-col gap-6">
      <AppPageHeader icon="mdi-cog" :title="$t('urls.settings')" :description="$t('settings.desc')" />

      <div class="grid gap-6 lg:grid-cols-2">
        <VCard border rounded="lg" class="flex flex-col">
          <AppCardHeader icon="mdi-account" :title="$t('settings.profile')" :description="$t('settings.profile.desc')" />
          <VCardText class="flex flex-1 flex-col justify-between gap-5 p-5">
            <AppAvatarUpload
              v-model="profile.avatar"
              :name="profile.username"
              :label="profile.username || $t('settings.no.username')"
              :sublabel="$t('role.admin')"
            />
            <VBtn variant="tonal" color="primary" prepend-icon="mdi-pencil" class="text-none font-weight-semibold self-start" @click="openEdit">
              {{ $t('update') }}
            </VBtn>
          </VCardText>
        </VCard>

        <VCard border rounded="lg" class="flex flex-col">
          <AppCardHeader icon="mdi-lock" :title="$t('settings.password')" :description="$t('settings.password.desc')" tone="warning" />
          <VCardText class="flex flex-1 flex-col justify-between gap-5 p-5">
            <div class="flex items-center gap-3 rounded-lg border border-on-surface/10 bg-surface-variant px-4 py-3">
              <VIcon icon="mdi-form-textbox-password" size="20" class="text-medium-emphasis" />
              <span class="text-body-medium tracking-widest text-medium-emphasis">••••••••••</span>
            </div>
            <VBtn
              variant="tonal"
              color="warning"
              prepend-icon="mdi-lock-reset"
              class="text-none font-weight-semibold self-start"
              @click="openPassword"
            >
              {{ $t('reset.password') }}
            </VBtn>
          </VCardText>
        </VCard>

        <VCard border rounded="lg" class="flex flex-col lg:col-span-2">
          <AppCardHeader icon="mdi-palette" :title="$t('settings.theme')" :description="$t('settings.theme.desc')" tone="info" />
          <VCardText class="p-5">
            <div class="grid gap-6 sm:grid-cols-2">
              <div>
                <p class="mb-2 text-body-small font-weight-medium text-medium-emphasis">{{ $t('settings.theme') }}</p>
                <div class="grid grid-cols-1 gap-1 rounded-lg border border-on-surface/10 bg-surface-variant p-1 sm:grid-cols-3">
                  <VBtn
                    v-for="opt in themeOptions"
                    :key="opt.value"
                    class="text-none"
                    size="small"
                    :color="conf.theme === opt.value ? 'primary' : undefined"
                    :variant="conf.theme === opt.value ? 'flat' : 'text'"
                    :prepend-icon="opt.icon"
                    @click="conf.theme = opt.value"
                  >
                    {{ $t(opt.title) }}
                  </VBtn>
                </div>
              </div>

              <div>
                <p class="mb-2 text-body-small font-weight-medium text-medium-emphasis">{{ $t('settings.locale') }}</p>
                <div class="grid grid-cols-1 gap-1 rounded-lg border border-on-surface/10 bg-surface-variant p-1 sm:grid-cols-2">
                  <VBtn
                    v-for="opt in localeOptions"
                    :key="opt.value"
                    class="justify-start text-none"
                    size="small"
                    :color="currentLocale === opt.value ? 'primary' : undefined"
                    :variant="currentLocale === opt.value ? 'flat' : 'text'"
                    :append-icon="currentLocale === opt.value ? 'mdi-check' : undefined"
                    @click="$i18n.setLocale(opt.value)"
                  >
                    {{ $t(opt.title) }}
                  </VBtn>
                </div>
              </div>
            </div>
          </VCardText>
        </VCard>
      </div>
    </div>

    <AppModal v-model="editOpen" :title="$t('settings.profile')" :description="$t('settings.profile.desc')" icon="mdi-account-edit" no-footer>
      <FormProfile v-model="profileForm" :loading="updatingProfile" :submit-text="$t('update')" @submit="sendProfile" />
    </AppModal>

    <AppModal v-model="passwordOpen" :title="$t('settings.password')" :description="$t('settings.password.desc')" icon="mdi-lock-reset" no-footer>
      <FormPasswd v-model="password" :loading="updatingPassword" :submit-text="$t('confirm')" @submit="sendPassword" />
    </AppModal>
  </ComCtl>
</template>
