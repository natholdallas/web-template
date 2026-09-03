<script setup lang="ts">
import { Palette, Lock, UserRound } from '@lucide/vue'
import { locales } from '~/lib/locale'
import { Apis, Profile, ResetPasswordIn } from '~/lib/sdk'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

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
      data: { username: usrForm.value.username },
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
</script>

<template>
  <ComCtl :loading="findingUser" class="p-4 lg:p-6" scroll>
    <div class="mx-auto max-w-3xl flex flex-col gap-4">
      <UiCard>
        <UiCardContent class="flex items-center gap-4 p-6 sm:gap-5">
          <UiAvatar class="size-16 rounded-none">
            <UiAvatarFallback class="bg-secondary text-lg secondary-foreground">
              {{ user.username?.slice(0, 1).toUpperCase() || '?' }}
            </UiAvatarFallback>
          </UiAvatar>
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
