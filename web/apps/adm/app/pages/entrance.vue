<script setup lang="ts">
import { Apis, AuthIn } from '~/lib/sdk'

definePageMeta({
  layout: 'minimal',
  name: 'entrance',
  middleware: 'entrance',
})

const auth = useAuth()
const model = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(
  () =>
    Apis.Auth.signIn({
      data: model.value,
    }),
  {
    immediate: false,
  },
).onSuccess(auth.$signIn)
</script>

<template>
  <ComCtl class="flex flex-col justify-center items-center size-full px-2">
    <VCard border class="w-full sm:w-120">
      <VCardTitle>{{ $t('sign.in') }}</VCardTitle>
      <VCardSubtitle> {{ $t('sign.in.desc') }} </VCardSubtitle>

      <VCardText class="flex flex-col gap-2">
        <FormLogin v-model="model" :loading="signing" :submit-text="$t('sign.in')" @submit="signIn" />
      </VCardText>
    </VCard>
  </ComCtl>
</template>
