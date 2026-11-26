<script setup lang="ts">
import { useDayjs } from '#dayjs'

const props = defineProps<{
  info?: { id?: number; created_at?: string; updated_at?: string }
}>()

const day = useDayjs()
const SITE_TZ = 'Asia/Shanghai'

const rows = computed(() => [
  { label: 'model.id', value: props.info?.id === undefined ? '-' : String(props.info.id) },
  { label: 'model.createdat', value: format(props.info?.created_at) },
  { label: 'model.updatedat', value: format(props.info?.updated_at) },
])

function format(v?: string): string {
  if (!v) return '-'
  return day(v).tz(SITE_TZ).format('YYYY-MM-DD HH:mm:ss')
}
</script>

<template>
  <div class="flex flex-col divide-y divide-on-surface/10">
    <div v-for="row in rows" :key="row.label" class="flex items-center justify-between gap-4 py-2.5">
      <span class="text-body-medium text-medium-emphasis">{{ $t(row.label) }}</span>
      <span class="text-body-medium font-weight-medium text-right">{{ row.value }}</span>
    </div>
  </div>
</template>
