<script setup lang="ts">
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
  <ComCtl :loading="findingUser" class="space-y-6 p-4" scroll>
    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.user.info') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.user.info.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent class="space-y-4">
        <div class="flex items-center justify-between py-2 border-b">
          <span class="text-sm text-muted-foreground">{{ $t('model.id') }}</span>
          <span class="font-medium">{{ user.id }}</span>
        </div>
        <div class="flex items-center justify-between py-2 border-b">
          <span class="text-sm text-muted-foreground">{{ $t('user.username') }}</span>
          <span class="font-medium">{{ user.username }}</span>
        </div>
      </UiCardContent>
    </UiCard>

    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.change.password') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.change.password.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormRstPwd v-model="pwdForm" :loading="savingPwd" @submit="savePwd" />
      </UiCardContent>
    </UiCard>

    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.user.info') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.user.info.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormProfile v-model="usrForm" :loading="savingUsr" @submit="saveUsr" />
      </UiCardContent>
    </UiCard>

    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.appearance') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.appearance.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent class="flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <div class="space-y-0.5">
            <span class="font-medium">{{ $t('settings.dark.mode') }}</span>
            <p class="text-sm text-muted-foreground">{{ $t('settings.dark.mode.desc') }}</p>
          </div>
          <SwitchTheme switch />
        </div>
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-0.5">
            <span class="font-medium">{{ $t('locale') }}</span>
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
  </ComCtl>
</template>
