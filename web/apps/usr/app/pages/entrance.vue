<script setup lang="ts">
import { Apis, AuthIn } from '~/lib/sdk'

definePageMeta({
  name: 'entrance',
  middleware: 'entrance',
  layout: 'minimal',
})

const auth = useAuth()
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
      </UiCardContent>
    </UiCard>
  </ComCtl>
</template>
