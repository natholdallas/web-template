<script setup lang="ts">
import { ProfileIn, ResetPasswordIn, UpdateProfile, ResetPassword } from '~/lib/sdk'
import { Locale, locales } from '~/lib/locale'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

const conf = useConf()

const profile = ref(inst(ProfileIn, { username: '' }))
const password = ref(inst(ResetPasswordIn))

const { loading: updatingProfile, send: sendProfile } = useRequest(
  () => {
    profile.value.username = profile.value.username || ''
    return UpdateProfile(profile.value)
  },
  { immediate: false },
)
const { loading: updatingPassword, send: sendPassword } = useRequest(() => ResetPassword(password.value), {
  immediate: false,
})

const themeOptions = [
  { title: 'settings.theme.dark', value: 'dark', icon: 'mdi-moon-waning-crescent' },
  { title: 'settings.theme.light', value: 'light', icon: 'mdi-white-balance-sunny' },
]
</script>

<template>
  <ComCtl scroll class="p-4">
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
          <VBtnToggle v-model="conf.theme" mandatory divided variant="outlined" color="primary">
            <VBtn v-for="opt in themeOptions" :key="opt.value" :value="opt.value">
              <VIcon :icon="opt.icon" class="mr-1" />
              {{ $t(opt.title) }}
            </VBtn>
          </VBtnToggle>
        </VCardText>
      </VCard>

      <VCard border>
        <VCardTitle>{{ $t('settings.locale') }}</VCardTitle>
        <VCardText>
          <VBtnToggle
            :model-value="$i18n.locale"
            @update:model-value="$i18n.setLocale($event as string)"
            mandatory
            divided
            variant="outlined"
            color="primary"
          >
            <VBtn v-for="l in Locale" :key="l" :value="l">
              {{ $t(locales[l].k) }}
            </VBtn>
          </VBtnToggle>
        </VCardText>
      </VCard>
    </div>
  </ComCtl>
</template>
