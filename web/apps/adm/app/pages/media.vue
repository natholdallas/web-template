<script setup lang="ts">
import { useDayjs } from '#dayjs'
import { Media, MediaPage, api } from '@sdk/client/adm'

definePageMeta({
  name: 'media',
})

const SITE_TZ = 'Asia/Shanghai'
const day = useDayjs()

const app = useAppStore()
app.$load()
const { openExternal } = useOpener()

const { smAndDown } = useVDisplay()
const { t } = useI18n()
const queries = ref({ page: 1, size: 20, column: 'created_at', desc: true })

const { loading, data, send } = useWatcher(
  () => api.media.list(queries.value),
  // A shallow copy is enough: every mutation replaces a field, and abortLast
  // cancels the in-flight request so results cannot land out of order.
  [() => ({ ...queries.value })],
  {
    immediate: true,
    debounce: 200,
    abortLast: true,
    initialData: inst(MediaPage),
  },
)

const { loading: removing, send: remove } = useRequest((id: number) => api.media.remove(id), {
  immediate: false,
}).onSuccess(() => {
  removeOpen.value = false
  // Deleting the last row of the last page would otherwise show an empty page.
  const remaining = Math.max(0, (data.value?.total ?? 1) - 1)
  queries.value.page = Math.max(1, Math.min(queries.value.page, Math.ceil(remaining / queries.value.size) || 1))
  send()
})

const { loading: uploading, send: upload } = useRequest((body: { file: File; kind: string }) => api.media.upload(body), {
  immediate: false,
}).onSuccess(send)

const fileInput = ref<HTMLInputElement>()
const removeOpen = ref(false)
const target = ref<Media>(inst(Media))

// Mirrors the backend allowlist (internal/srv/internal/media.go). `accept` only
// filters the picker — the server still sniffs the content and re-validates.
const ACCEPT = [
  '.jpg',
  '.jpeg',
  '.png',
  '.webp',
  '.gif',
  '.mp4',
  '.webm',
  '.mov',
  '.m4v',
  '.pdf',
  '.zip',
  '.gz',
  '.7z',
  '.rar',
  '.doc',
  '.docx',
  '.xls',
  '.xlsx',
  '.ppt',
  '.pptx',
  '.txt',
  '.md',
  '.csv',
].join(',')

function onPick(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  el.value = ''
  if (!file) return
  // Reject oversize files before spending bandwidth on an upload the server
  // will refuse (app.body-limit / resources.media-max-size).
  const max = app.cfg.mediaMaxSize
  if (max > 0 && file.size > max * 1024 * 1024) {
    useSnackBar().error(t('media.too.large', { v: max }))
    return
  }
  const kind = file.type.startsWith('video/') ? 'video' : file.type.startsWith('image/') ? 'image' : 'file'
  upload({ file, kind })
}

function formatSize(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${units[i]}`
}

function formatTime(v?: string): string {
  if (!v) return '-'
  return day(v).tz(SITE_TZ).format('YYYY-MM-DD HH:mm:ss')
}

const headers = computed(() => {
  const hidden = smAndDown.value ? ['id', 'created_at'] : []
  return [
    { title: t('model.id'), key: 'id', sortable: true },
    { title: t('media.preview'), key: 'url', sortable: false },
    { title: t('media.name'), key: 'name', sortable: false },
    { title: t('media.kind'), key: 'kind', sortable: true },
    { title: t('media.size'), key: 'size', sortable: true },
    { title: t('model.createdat'), key: 'created_at', sortable: true },
    { key: 'actions', sortable: false, align: 'end' as const },
  ].filter((h) => !hidden.includes(h.key))
})

function askRemove(item: Media) {
  target.value = item
  removeOpen.value = true
}
</script>

<template>
  <ComCtl :loading="[loading, uploading, removing]" class="flex flex-col gap-4 p-4 lg:p-6">
    <AppPageHeader icon="mdi-image-multiple" :title="$t('urls.media')" :description="$t('media.desc')" :count="data.total">
      <template #actions>
        <VBtn color="primary" prepend-icon="mdi-upload" class="text-none" @click="fileInput?.click()">
          {{ $t('upload') }}
        </VBtn>
        <input ref="fileInput" type="file" class="hidden" :accept="ACCEPT" @change="onPick">
      </template>
    </AppPageHeader>

    <VCard border rounded="lg" class="flex-1 min-h-0 overflow-hidden">
      <VDataTableServer
        v-model:items-per-page="queries.size"
        v-model:page="queries.page"
        :headers="headers"
        :items="data.content"
        :items-length="data.total"
        :items-per-page-options="[10, 20, 50, 100]"
        :loading="loading || removing"
        no-data-text="no.data"
        class="adm-table"
        density="comfortable"
        hover
        @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
      >
        <template #[`item.id`]="{ item }"
          ><span class="tabular-nums">{{ item.id }}</span></template
        >
        <template #[`item.url`]="{ item }">
          <VAvatar size="40" rounded="lg" color="surface-variant">
            <VxImg v-if="item.kind === 'image'" :src="app.mediaUrl(item.url)" :alt="item.name" cover />
            <VIcon v-else icon="mdi-file" />
          </VAvatar>
        </template>
        <template #[`item.name`]="{ item }">
          <span class="text-body-medium d-inline-block text-truncate" style="max-width: 16rem">{{ item.name }}</span>
        </template>
        <template #[`item.kind`]="{ item }">
          <VChip size="small" variant="tonal" :color="item.kind === 'image' ? 'primary' : item.kind === 'video' ? 'info' : undefined">
            {{ $t('media.kind.' + item.kind) }}
          </VChip>
        </template>
        <template #[`item.size`]="{ item }">
          <span class="text-body-medium tabular-nums">{{ formatSize(item.size) }}</span>
        </template>
        <template #[`item.created_at`]="{ item }">
          <span class="text-body-medium text-medium-emphasis">{{ formatTime(item.created_at) }}</span>
        </template>
        <template #[`item.actions`]="{ item }">
          <div class="flex items-center justify-end gap-1">
            <AppIconBtn icon="mdi-open-in-new" :tooltip="$t('open')" @click="openExternal(app.mediaUrl(item.url))" />
            <AppIconBtn icon="mdi-delete" tone="error" :tooltip="$t('remove')" @click="askRemove(item)" />
          </div>
        </template>
      </VDataTableServer>
    </VCard>

    <AppModal
      v-model="removeOpen"
      :title="$t('dialog')"
      :description="$t('dialog.desc')"
      icon="mdi-delete"
      tone="error"
      :confirm-text="$t('remove')"
      :loading="removing"
      @cancel="removeOpen = false"
      @confirm="remove(target.id)"
    />
  </ComCtl>
</template>
