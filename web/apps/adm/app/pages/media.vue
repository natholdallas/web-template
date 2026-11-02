<script setup lang="ts">
import type { Media } from '~/lib/sdk'
import { Apis } from '~/lib/sdk'

definePageMeta({
  name: 'media',
  middleware: 'auth',
})

const queries = ref({ page: 1, size: 20, column: 'created_at', desc: true })
const { open } = useDialog()
const { loading, data, send } = useRequest(
  () =>
    Apis.Media.list({
      params: queries.value,
    }),
  {
    initialData: { total: 0, page: 0, content: [] },
  },
)

const { loading: removing, send: remove } = useRequest(
  (id: number) =>
    Apis.Media.remove({
      pathParams: { id },
    }),
  {
    immediate: false,
  },
).onSuccess(send)

const { loading: uploading, send: upload } = useRequest(
  (body: { file: File; kind: string }) =>
    Apis.Media.upload({
      data: body,
    }),
  {
    immediate: false,
  },
).onSuccess(send)

const fileInput = ref<HTMLInputElement>()

function onPick(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  if (file) {
    const kind = file.type.startsWith('video/') ? 'video' : file.type.startsWith('image/') ? 'image' : 'file'
    upload({ file, kind })
  }
  el.value = ''
}

function openUrl(url: string) {
  window.open(url, '_blank', 'noopener')
}

function formatSize(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${units[i]}`
}

watch(queries, send, { deep: true })
</script>

<template>
  <ComCtl :loading="uploading" class="flex flex-col p-4">
    <div class="flex size-full flex-1 flex-col gap-4">
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <div class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <VIcon icon="mdi-image-multiple" size="28" />
          </div>
          <div>
            <h1 class="text-2xl font-bold leading-tight">{{ $t('urls.media') }}</h1>
            <p class="text-sm opacity-70">{{ $t('media.desc') }}</p>
          </div>
        </div>

        <div class="flex items-center gap-3">
          <VChip prepend-icon="mdi-image-multiple" variant="tonal" class="hidden sm:inline-flex" rounded>
            {{ $t('total') }}: {{ data.total }}
          </VChip>
          <VBtn prepend-icon="mdi-upload" color="primary" @click="fileInput?.click()">
            {{ $t('upload') }}
          </VBtn>
          <input ref="fileInput" type="file" class="hidden" @change="onPick">
        </div>
      </div>

      <VCard border rounded="xl" class="flex-1 min-h-0 overflow-hidden">
        <VDataTableServer
          v-model:items-per-page="queries.size"
          v-model:page="queries.page"
          class="h-full"
          :headers="[
            { title: $t('model.id'), key: 'id', sortable: true },
            { title: $t('media.preview'), key: 'url', sortable: false },
            { title: $t('media.name'), key: 'name', sortable: false },
            { title: $t('media.kind'), key: 'kind', sortable: true },
            { title: $t('media.size'), key: 'size', sortable: true },
            { title: $t('model.createdat'), key: 'created_at', sortable: true },
            { key: 'actions' },
          ]"
          :items="data.content"
          :items-length="data.total"
          :loading="loading || removing"
          @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
        >
          <template #item.url="{ item }">
            <div class="flex items-center">
              <VAvatar size="40" rounded="lg" :color="item.kind === 'image' ? 'primary' : 'surface-variant'">
                <VImg v-if="item.kind === 'image'" :src="item.url" :alt="item.name" cover />
                <VIcon v-else icon="mdi-file" />
              </VAvatar>
            </div>
          </template>
          <template #item.size="{ item }">
            <span class="text-sm">{{ formatSize((item as Media).size) }}</span>
          </template>
          <template #item.kind="{ item }">
            <VChip :color="item.kind === 'image' ? 'primary' : item.kind === 'video' ? 'info' : undefined" size="small" variant="tonal">
              {{ $t(`media.kind.${item.kind}`) }}
            </VChip>
          </template>
          <template #item.actions="{ item }">
            <div class="flex items-center justify-end gap-1">
              <VxActionBtn icon="mdi-open-in-new" :tooltip="$t('open')" @click="openUrl(item.url)" />
              <VxActionBtn
                icon="mdi-delete"
                :tooltip="$t('remove')"
                @click="open({ confirm: () => remove(item.id) })"
              />
            </div>
          </template>
        </VDataTableServer>
      </VCard>
    </div>
  </ComCtl>
</template>