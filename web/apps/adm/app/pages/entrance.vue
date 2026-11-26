<script setup lang="ts">
import { api, AuthIn } from '@sdk/client/adm'

definePageMeta({
  layout: 'minimal',
  name: 'entrance',
})

const auth = useAuthStore()
const model = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(() => api.auth.signIn(model.value), {
  immediate: false,
}).onSuccess(auth.$signIn)
</script>

<template>
  <ComCtl class="flex size-full items-center justify-center p-4" :loading="signing">
    <VCard border rounded="lg" class="w-full max-w-[26rem] overflow-hidden" elevation="12">
      <div class="h-1 bg-gradient-to-r from-primary to-secondary" />
      <VCardText class="flex flex-col gap-6 p-6 sm:p-8">
        <div class="flex justify-center"><AppBrand /></div>

        <div class="space-y-1 text-center">
          <h1 class="text-title-large font-weight-bold">{{ $t('sign.in') }}</h1>
          <p class="text-body-medium text-medium-emphasis">{{ $t('sign.in.desc') }}</p>
        </div>

        <FormLogin v-model="model" :loading="signing" :submit-text="$t('sign.in')" @submit="signIn" />
      </VCardText>
    </VCard>
  </ComCtl>
</template>
