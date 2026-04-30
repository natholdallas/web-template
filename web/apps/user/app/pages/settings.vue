<script setup lang="ts">
import { locales } from '~/lib/locale'
import { FindUser, ResetPasswordIn, RstPwd, UpdateUser, User } from '~/lib/sdk'

definePageMeta({
  name: 'settings',
  middleware: 'auth',
})

const { loading, data, send } = useRequest(FindUser, {
  initialData: inst(User),
}).onSuccess(({ data }) => {
  usrForm.value = cpm(data)
})
const usrForm = ref(inst(User))
const { loading: savingUsr, send: saveUsr } = useRequest(() => UpdateUser(usrForm.value), {
  immediate: false,
}).onSuccess(() => {
  rst(usrForm.value, User)
  send()
})
const pwdForm = ref(inst(ResetPasswordIn))
const { loading: savingPwd, send: savePwd } = useRequest(() => RstPwd(pwdForm.value), {
  immediate: false,
}).onSuccess(() => {
  rst(pwdForm.value, ResetPasswordIn)
})
</script>

<template>
  <ComCtl :loading="loading" class="space-y-6 p-4" scroll>
    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.user.info') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.user.info.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent class="space-y-4">
        <div class="flex items-center justify-between py-2 border-b">
          <span class="text-sm text-muted-foreground">{{ $t('model.id') }}</span>
          <span class="font-medium">{{ data.id }}</span>
        </div>
        <div class="flex items-center justify-between py-2 border-b">
          <span class="text-sm text-muted-foreground">{{ $t('user.username') }}</span>
          <span class="font-medium">{{ data.username }}</span>
        </div>
      </UiCardContent>
    </UiCard>

    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.change.password') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.change.password.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormRstPwd v-model="pwdForm" @submit="savePwd" :loading="savingPwd" />
      </UiCardContent>
    </UiCard>

    <UiCard>
      <UiCardHeader>
        <UiCardTitle>{{ $t('settings.user.info') }}</UiCardTitle>
        <UiCardDescription>{{ $t('settings.user.info.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormUser v-model="usrForm" @submit="saveUsr" :loading="savingUsr" />
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
          <ThemeSwitcher switch />
        </div>
        <div class="flex items-center justify-between gap-4">
          <div class="space-y-0.5">
            <span class="font-medium">{{ $t('locale') }}</span>
            <p class="text-sm text-muted-foreground">{{ $t('locale.desc') }}</p>
          </div>
          <LangSwitcher
            @update="$i18n.setLocale"
            :options="Object.values(locales).map(({ k, v }) => ({ label: $t(k), value: v }))"
            :value="$t(locales[$i18n.locale].k)"
          />
        </div>
      </UiCardContent>
    </UiCard>
  </ComCtl>
</template>
