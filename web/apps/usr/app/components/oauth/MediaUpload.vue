<script setup lang="ts">
import { Camera, Loader2, UserRound } from '@lucide/vue'
import type { Media } from '~/lib/sdk'
import Apis from '~/lib/sdk'

const props = withDefaults(
  defineProps<{
    accept?: string
    kind?: string
    disabled?: boolean
  }>(),
  { accept: 'image/*', kind: 'image', disabled: false },
)

const model = defineModel<string>()

const { loading, send } = useRequest(
  (f: File) => Apis.Media.upload({ data: { file: f, kind: props.kind! } }),
  { immediate: false },
).onSuccess(({ data }) => {
  model.value = (data as Media).url || ''
})

function onPick(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  if (file) send(file)
  el.value = ''
}
</script>

<template>
  <label
    class="relative grid size-16 cursor-pointer place-items-center overflow-hidden rounded-full bg-muted text-muted-foreground transition hover:opacity-80"
    :class="{ 'pointer-events-none opacity-50': disabled || loading }"
  >
    <UiAvatar class="size-full rounded-full">
      <UiAvatarImage v-if="model" :src="model" alt="" class="size-full object-cover" />
      <UiAvatarFallback v-else class="bg-transparent">
        <UserRound class="size-6" />
      </UiAvatarFallback>
    </UiAvatar>
    <span class="absolute bottom-1 right-1 grid size-5 place-items-center rounded-full bg-background shadow-sm">
      <Camera class="size-3" />
    </span>
    <Loader2 v-if="loading" class="absolute animate-spin" />
    <input type="file" class="sr-only" :accept="accept" :disabled="disabled" @change="onPick">
  </label>
</template>