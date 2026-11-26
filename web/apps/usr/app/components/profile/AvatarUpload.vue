<script setup lang="ts">
import { Camera, Loader2, RefreshCw, RotateCcw, RotateCw, UserRound, ZoomIn, ZoomOut } from '@lucide/vue'
import { toast } from 'vue-sonner'
import { api } from '@sdk/client/usr'

// Loaded on demand: the cropper plus its stylesheet is ~110 KB that the
// settings page should not pay for until the user actually opens the dialog.
const Cropper = defineAsyncComponent(() => import('vue-advanced-cropper').then((m) => m.Cropper))

// The dialog is a sibling root, so the parent's class must land on the row.
defineOptions({ inheritAttrs: false })

// Avatar picker: the user selects a local image, crops it in a modal, and the
// cropped result is uploaded at the source resolution. The backend stores the
// bytes verbatim, so the aspect ratio and output size are decided here.
const model = defineModel<string>({ default: '' })

const { t } = useI18n()
const app = useAppStore()
app.$load()

const ACCEPT = ['image/jpeg', 'image/png', 'image/webp']
const MAX_INPUT_SIZE = 10 * 1024 * 1024
const MAX_INPUT_DIMENSION = 4096

const PREVIEW_CANVAS = { width: 160, height: 160 }
const MAX_PANEL_WIDTH = 320
const MAX_PANEL_HEIGHT = 420
const MAX_OUTPUT = 512

// Corner handles only: the stencil is locked to 1:1, so edge handles add nothing.
const STENCIL_PROPS = {
  aspectRatio: 1,
  handlers: { eastNorth: true, westNorth: true, eastSouth: true, westSouth: true },
}

const MIN_ZOOM = 1
const MAX_ZOOM = 3
const ZOOM_STEP = 0.2

type CropperRef = {
  getResult: () => { canvas: HTMLCanvasElement | null }
  zoom: (factor: number) => void
  rotate: (angle: number) => void
  reset: () => void
  refresh: () => void
}

const src = ref('')
const cropping = ref(false)
const fileInput = ref<HTMLInputElement>()
const cropper = ref<CropperRef>()

// The panel follows the source aspect so the cropper can show the whole image
// instead of a square band of it.
const sourceSize = ref({ width: 0, height: 0 })
// `true` renders at the stencil's native size; the small canvas is the live preview.
const canvasSize = ref<{ width: number; height: number } | true>({ ...PREVIEW_CANVAS })

const panelStyle = computed(() => {
  const { width, height } = sourceSize.value
  if (!width || !height) return { width: `${MAX_PANEL_WIDTH}px`, height: `${MAX_PANEL_HEIGHT}px` }
  const scale = Math.min(MAX_PANEL_WIDTH / width, MAX_PANEL_HEIGHT / height)
  return { width: `${Math.round(width * scale)}px`, height: `${Math.round(height * scale)}px` }
})

const zoom = ref(1)
const previewUrl = ref('')
// The cropper's zoom() takes a multiplier, so the applied value is tracked apart
// from the slider position.
let lastZoom = 1

const { loading: uploading, send: upload } = useRequest(
  (file: File) => api.profile.uploadAvatar({ file }),
  { immediate: false },
).onSuccess(({ data }) => {
  model.value = data.avatar
  cropping.value = false
})

const { loading: removing, send: remove } = useRequest(() => api.profile.deleteAvatar(), {
  immediate: false,
}).onSuccess(({ data }) => {
  model.value = data.avatar
})

function clearSrc() {
  if (src.value) {
    URL.revokeObjectURL(src.value)
    src.value = ''
  }
}

function setZoom(next: number) {
  const clamped = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, next))
  if (Math.abs(clamped - lastZoom) < 0.001) return
  cropper.value?.zoom(clamped / lastZoom)
  lastZoom = clamped
  zoom.value = clamped
}

function onZoomInput(event: Event) {
  setZoom(Number((event.target as HTMLInputElement).value))
}

function rotateBy(angle: number) {
  cropper.value?.rotate(angle)
}

function resetCropper() {
  cropper.value?.reset()
  zoom.value = 1
  lastZoom = 1
  previewUrl.value = ''
}

function onChange({ canvas }: { canvas?: HTMLCanvasElement | null }) {
  previewUrl.value = canvas?.toDataURL('image/png') ?? ''
}

// The object URL is only valid while the cropper is open; the cropper needs a
// layout pass before it can measure itself, hence the refresh on the next tick.
watch(cropping, async (open) => {
  if (!open) {
    clearSrc()
    previewUrl.value = ''
    return
  }
  zoom.value = 1
  lastZoom = 1
  previewUrl.value = ''
  // The lazy component needs its stylesheet before it measures itself.
  await import('vue-advanced-cropper/dist/style.css')
  await nextTick()
  cropper.value?.refresh()
})

onUnmounted(clearSrc)

// Read the dimensions before decoding a bitmap: a small file can still expand
// into a multi-gigabyte bitmap, so oversize inputs are rejected first.
function probeSize(url: string): Promise<{ width: number; height: number } | null> {
  return new Promise((resolve) => {
    const img = new Image()
    img.onload = () => resolve({ width: img.naturalWidth, height: img.naturalHeight })
    img.onerror = () => resolve(null)
    img.src = url
  })
}

// Bake the EXIF orientation into a plain bitmap before the cropper sees it.
// Browsers disagree on whether an <img> is auto-oriented (and the cropper has
// its own EXIF handling on top), which is what mirrored a user's avatar. An
// already-upright, EXIF-free source makes every crop path identical.
async function bakeOrientation(
  file: File,
  fallback: { width: number; height: number },
): Promise<{ url: string; width: number; height: number } | null> {
  try {
    const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
    const canvas = document.createElement('canvas')
    canvas.width = bitmap.width
    canvas.height = bitmap.height
    const ctx = canvas.getContext('2d')
    if (!ctx) {
      bitmap.close()
      return null
    }
    ctx.drawImage(bitmap, 0, 0)
    const { width, height } = canvas
    bitmap.close()
    const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, 'image/png'))
    return blob ? { url: URL.createObjectURL(blob), width, height } : null
  } catch {
    // WebViews without createImageBitmap fall back to the raw file; the
    // cropper's own orientation pass still applies there.
    return { url: URL.createObjectURL(file), ...fallback }
  }
}

async function onPick(e: Event) {
  const el = e.target as HTMLInputElement
  const file = el.files?.[0]
  // Reset immediately so picking the same file twice fires change again.
  el.value = ''
  if (!file) return
  if (!ACCEPT.includes(file.type)) {
    toast.error(t('avatar.invalid.type'))
    return
  }
  if (file.size > MAX_INPUT_SIZE) {
    toast.error(t('avatar.invalid.size'))
    return
  }
  const probeUrl = URL.createObjectURL(file)
  const size = await probeSize(probeUrl)
  URL.revokeObjectURL(probeUrl)
  if (!size) {
    toast.error(t('avatar.invalid.type'))
    return
  }
  if (size.width > MAX_INPUT_DIMENSION || size.height > MAX_INPUT_DIMENSION) {
    toast.error(t('avatar.invalid.dimensions'))
    return
  }
  const baked = await bakeOrientation(file, size)
  if (!baked) {
    toast.error(t('avatar.invalid.type'))
    return
  }
  clearSrc()
  previewUrl.value = ''
  sourceSize.value = { width: baked.width, height: baked.height }
  src.value = baked.url
  cropping.value = true
}

// Avatars never need more than MAX_OUTPUT px: bigger crops are scaled down by
// one factor on both sides, so the square (or any future aspect) is preserved.
function limitOutput(source: HTMLCanvasElement): HTMLCanvasElement {
  const longest = Math.max(source.width, source.height)
  if (longest <= MAX_OUTPUT) return source
  const scale = MAX_OUTPUT / longest
  const canvas = document.createElement('canvas')
  canvas.width = Math.max(1, Math.round(source.width * scale))
  canvas.height = Math.max(1, Math.round(source.height * scale))
  const ctx = canvas.getContext('2d')
  if (!ctx) return source
  ctx.imageSmoothingQuality = 'high'
  ctx.drawImage(source, 0, 0, canvas.width, canvas.height)
  return canvas
}

async function confirm() {
  // Swap to the native canvas only for the final render, so the live preview
  // stays cheap while the uploaded file keeps the crop's full resolution.
  canvasSize.value = true
  await nextTick()
  const canvas = cropper.value?.getResult().canvas
  canvasSize.value = { ...PREVIEW_CANVAS }
  if (!canvas) return
  // Crops at or below 512 px keep their native resolution; bigger ones are
  // downscaled by limitOutput. WebP keeps 512² small (and supports
  // transparency); browsers without WebP encoding fall back to PNG, so the
  // filename follows the produced type.
  limitOutput(canvas).toBlob(
    (blob) => {
      if (!blob) return
      const ext = blob.type === 'image/webp' ? 'webp' : blob.type === 'image/jpeg' ? 'jpg' : 'png'
      upload(new File([blob], `avatar.${ext}`, { type: blob.type }))
    },
    'image/webp',
    0.92,
  )
}
</script>

<template>
  <div v-bind="$attrs" class="flex items-center gap-4">
    <button
      type="button"
      class="relative grid size-16 shrink-0 cursor-pointer place-items-center overflow-hidden rounded-full bg-muted text-muted-foreground transition hover:opacity-80 disabled:pointer-events-none disabled:opacity-50"
      :disabled="uploading || removing"
      @click="fileInput?.click()"
    >
      <UiAvatar class="size-full rounded-full">
        <UiAvatarImage v-if="model" :src="app.mediaUrl(model)" alt="" class="size-full object-cover" />
        <UiAvatarFallback v-else class="bg-transparent">
          <UserRound class="size-6" />
        </UiAvatarFallback>
      </UiAvatar>
      <span class="absolute bottom-1 right-1 grid size-5 place-items-center rounded-full bg-background shadow-sm">
        <Camera class="size-3" />
      </span>
      <Loader2 v-if="uploading" class="absolute size-5 animate-spin" />
    </button>
    <input
      ref="fileInput"
      type="file"
      class="sr-only"
      accept="image/jpeg,image/png,image/webp"
      :disabled="uploading || removing"
      @change="onPick"
    >
    <div class="min-w-0 space-y-1">
      <p class="text-sm font-medium">{{ $t('avatar') }}</p>
      <UiButton
        v-if="model"
        variant="link"
        size="xs"
        class="h-auto p-0 text-muted-foreground"
        :disabled="uploading || removing"
        @click="remove()"
      >
        {{ $t('avatar.remove') }}
      </UiButton>
    </div>
  </div>

  <UiDialog v-model:open="cropping">
    <UiDialogContent :style="{ maxWidth: 'min(42rem, calc(100vw - 2rem))' }">
      <UiDialogHeader>
        <UiDialogTitle>{{ $t('avatar.crop.title') }}</UiDialogTitle>
        <UiDialogDescription>{{ $t('avatar.crop.desc') }}</UiDialogDescription>
      </UiDialogHeader>

      <div class="flex min-w-0 flex-col gap-5 sm:flex-row">
        <div class="avatar-cropper mx-auto min-w-0 shrink-0 bg-muted sm:mx-0" :style="panelStyle">
          <Cropper
            ref="cropper"
            :src="src"
            :stencil-props="STENCIL_PROPS"
            image-restriction="fit-area"
            :canvas="canvasSize"
            :debounce="false"
            :resize-image="{ wheel: false, touch: true }"
            class="block size-full"
            @change="onChange"
          />
        </div>

        <div class="flex min-w-0 flex-1 flex-col gap-4">
          <div class="flex items-center gap-3">
            <img
              v-if="previewUrl"
              :src="previewUrl"
              alt=""
              class="size-16 rounded-full bg-muted object-cover ring-1 ring-border"
            >
            <div v-else class="grid size-16 place-items-center rounded-full bg-muted text-muted-foreground">
              <UserRound class="size-6" />
            </div>
            <div class="min-w-0">
              <p class="text-sm font-medium">{{ $t('avatar.crop.preview') }}</p>
              <p class="text-xs text-muted-foreground">{{ $t('avatar.crop.hint') }}</p>
            </div>
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">{{ $t('avatar.crop.zoom') }}</span>
              <span class="text-xs tabular-nums text-muted-foreground">{{ Math.round(zoom * 100) }}%</span>
            </div>
            <input
              type="range"
              min="1"
              max="3"
              step="0.01"
              :value="zoom"
              class="w-full accent-primary"
              :aria-label="$t('avatar.crop.zoom')"
              @input="onZoomInput"
            >
          </div>

          <div class="grid grid-cols-4 gap-2">
            <UiButton
              variant="outline"
              size="icon-sm"
              :title="$t('avatar.crop.zoom.out')"
              :aria-label="$t('avatar.crop.zoom.out')"
              @click="setZoom(zoom - ZOOM_STEP)"
            >
              <ZoomOut class="size-4" />
            </UiButton>
            <UiButton
              variant="outline"
              size="icon-sm"
              :title="$t('avatar.crop.zoom.in')"
              :aria-label="$t('avatar.crop.zoom.in')"
              @click="setZoom(zoom + ZOOM_STEP)"
            >
              <ZoomIn class="size-4" />
            </UiButton>
            <UiButton
              variant="outline"
              size="icon-sm"
              :title="$t('avatar.crop.rotate.left')"
              :aria-label="$t('avatar.crop.rotate.left')"
              @click="rotateBy(-90)"
            >
              <RotateCcw class="size-4" />
            </UiButton>
            <UiButton
              variant="outline"
              size="icon-sm"
              :title="$t('avatar.crop.rotate.right')"
              :aria-label="$t('avatar.crop.rotate.right')"
              @click="rotateBy(90)"
            >
              <RotateCw class="size-4" />
            </UiButton>
          </div>

          <UiButton variant="ghost" size="sm" class="self-start" @click="resetCropper">
            <RefreshCw class="size-4" />{{ $t('avatar.crop.reset') }}
          </UiButton>
        </div>
      </div>

      <UiDialogFooter>
        <UiButton variant="outline" :disabled="uploading" @click="cropping = false">
          {{ $t('cancel') }}
        </UiButton>
        <UiButton :disabled="uploading" @click="confirm">
          <Loader2 v-if="uploading" class="animate-spin" />
          {{ $t('confirm') }}
        </UiButton>
      </UiDialogFooter>
    </UiDialogContent>
  </UiDialog>
</template>

<style scoped>
.avatar-cropper :deep(.vue-advanced-cropper__background) {
  background: var(--muted);
}
/* The library's own scrim outside the stencil (black @ 0.5) reads as a neutral
   dim; a theme colour there tints the whole ring. */
.avatar-cropper :deep(.vue-line-wrapper) {
  display: none;
}
/* White dot with a theme ring so the corner stays visible on any image. */
.avatar-cropper :deep(.vue-simple-handler) {
  width: 12px;
  height: 12px;
  border: 2px solid var(--primary);
  border-radius: 9999px;
  background: var(--background);
}
.avatar-cropper :deep(.vue-advanced-cropper__stencil-wrapper) {
  outline: 1px solid var(--border);
}
</style>
