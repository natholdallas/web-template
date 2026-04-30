<script setup lang="ts">
import { AuthIn, SignIn } from '~/sdk'

definePageMeta({
  name: 'entrance',
  middleware: 'entrance',
})

const auth = useAuth()
const model = ref(inst(AuthIn))

const { loading: signing, send: signIn } = useRequest(() => SignIn(model.value), {
  immediate: false,
}).onSuccess(auth.$signIn)
</script>

<template>
  <ComCtl class="flex flex-col fixed justify-center items-center size-full px-2">
    <VCard class="w-full sm:w-120" border>
      <VCardTitle> {{ $t('sign.in') }} </VCardTitle>
      <VCardSubtitle> {{ $t('sign.in.desc') }} </VCardSubtitle>
      <VCardText class="flex flex-col gap-2">
        <FormLogin
          v-model="model"
          @submit="signIn"
          :loading="signing"
          :submit-text="$t('sign.in')"
        />
      </VCardText>
    </VCard>
  </ComCtl>
</template>
