<script setup lang="ts">
import { AuthIn, SignIn } from '~/lib/sdk'

definePageMeta({
  name: 'entrance',
  middleware: 'entrance',
  layout: 'minimal',
})

const auth = useAuth()
const form = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(() => SignIn(form.value), {
  immediate: false,
}).onSuccess(auth.$signIn)
</script>

<template>
  <ComCtl class="flex justify-center items-center">
    <UiCard class="w-[450px]">
      <UiCardHeader>
        <UiCardTitle>{{ $t('sign.in') }}</UiCardTitle>
        <UiCardDescription>{{ $t('sign.in.desc') }}</UiCardDescription>
      </UiCardHeader>
      <UiCardContent>
        <FormLogin v-model="form" @submit="signIn" :loading="signing" />
      </UiCardContent>
    </UiCard>
  </ComCtl>
</template>
