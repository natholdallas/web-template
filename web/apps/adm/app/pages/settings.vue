<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'
import { Apis, ProfileIn, ResetPasswordIn } from '~/lib/sdk'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

const conf = useConf()

const profile = ref(inst(ProfileIn, { username: '' }))
const password = ref(inst(ResetPasswordIn))

const { loading: updatingProfile, send: sendProfile } = useRequest(
  () =>
    Apis.Profile.update({
      data: { username: profile.value.username || '' },
    }),
  { immediate: false },
)
const { loading: updatingPassword, send: sendPassword } = useRequest(
  () =>
    Apis.Profile.resetPassword({
      data: password.value,
    }),
  {
    immediate: false,
  },
)

const themeOptions = [
  { title: 'settings.theme.dark', value: 'dark', icon: 'mdi-moon-waning-crescent' },
  { title: 'settings.theme.light', value: 'light', icon: 'mdi-white-balance-sunny' },
]
</script>

<template>
  <ComCtl class="p-4" scroll>
    <div class="mx-auto max-w-2xl flex flex-col gap-4">
      <VCard border>
        <VCardTitle>{{ $t('settings.profile') }}</VCardTitle>
        <VCardSubtitle>{{ $t('settings.profile.desc') }}</VCardSubtitle>

        <VCardText>
          <FormProfile v-model="profile" :loading="updatingProfile" @submit="sendProfile" />
        </VCardText>
      </VCard>

      <VCard border>
        <VCardTitle>{{ $t('settings.password') }}</VCardTitle>
        <VCardSubtitle>{{ $t('settings.password.desc') }}</VCardSubtitle>

        <VCardText>
          <FormPasswd v-model="password" :loading="updatingPassword" @submit="sendPassword" />
        </VCardText>
      </VCard>

      <VCard border>
        <VCardTitle>{{ $t('settings.theme') }}</VCardTitle>

        <VCardText>
          <VBtnToggle v-model="conf.theme" color="blue" divided mandatory variant="outlined">
            <VBtn v-for="opt in themeOptions" :key="opt.value" :value="opt.value">
              <VIcon class="mr-1" :icon="opt.icon" />
              {{ $t(opt.title) }}
            </VBtn>
          </VBtnToggle>
        </VCardText>
      </VCard>

      <VCard border>
        <VCardTitle>{{ $t('settings.locale') }}</VCardTitle>

        <VCardText>
          <VBtnToggle color="blue" divided mandatory :model-value="$i18n.locale" variant="outlined" @update:model-value="$i18n.setLocale($event)">
            <VBtn v-for="l in Locale" :key="l" :value="l">
              {{ $t(locales[l].k) }}
            </VBtn>
          </VBtnToggle>
        </VCardText>
      </VCard>
    </div>
  </ComCtl>
</template>
