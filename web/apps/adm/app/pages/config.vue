<script setup lang="ts">
import { api, ConfStatus } from '@sdk/client/adm'
import type { ConfCheck } from '@sdk/client/adm'

definePageMeta({
  name: 'config',
})

const { loading, data, send } = useRequest(() => api.config.getConfigStatus(), {
  initialData: inst(ConfStatus),
})

// Consecutive grouping keeps the server's display order: the backend already
// emits the probes in the order the page should show them.
const groups = computed(() => {
  const out: { name: string; checks: ConfCheck[] }[] = []
  for (const check of data.value.checks) {
    const last = out[out.length - 1]
    if (last && last.name === check.group) last.checks.push(check)
    else out.push({ name: check.group, checks: [check] })
  }
  return out
})

// Only an enabled feature that is missing values is an actual problem.
const incomplete = computed(() => data.value.checks.filter((c) => c.enabled && !c.configured).length)

function statusIcon(check: ConfCheck): string {
  if (check.configured) return 'mdi-check-circle'
  return check.enabled ? 'mdi-alert-circle' : 'mdi-minus-circle'
}

function statusColor(check: ConfCheck): string {
  if (check.configured) return 'success'
  return check.enabled ? 'warning' : 'grey'
}
</script>

<template>
  <ComCtl :loading="loading" class="flex flex-col gap-4 p-4 lg:p-6" scroll>
    <AppPageHeader icon="mdi-clipboard-check-outline" :title="$t('urls.config')" :description="$t('config.desc')">
      <template #actions>
        <VChip color="grey" variant="tonal" size="small">{{ $t('config.readonly') }}</VChip>
        <VBtn variant="text" prepend-icon="mdi-refresh" class="text-none" :loading="loading" @click="send">
          {{ $t('config.refresh') }}
        </VBtn>
      </template>
    </AppPageHeader>

    <VAlert v-if="incomplete === 0" type="success" variant="tonal" density="comfortable" class="flex-none">
      {{ $t('config.all.ok') }}
    </VAlert>
    <VAlert v-else type="warning" variant="tonal" density="comfortable" class="flex-none">
      {{ $t('config.missing.count', { n: incomplete }) }}
    </VAlert>

    <VCard v-for="group in groups" :key="group.name" border rounded="lg">
      <AppCardHeader icon="mdi-folder-cog" :title="$t('config.group.' + group.name)" />
      <VCardText class="flex flex-col divide-y p-5">
        <div v-for="check in group.checks" :key="check.key" class="flex flex-col gap-2 py-3 first:pt-0 last:pb-0">
          <div class="flex flex-wrap items-center gap-2">
            <VIcon :icon="statusIcon(check)" :color="statusColor(check)" size="18" />
            <span class="text-body-medium font-weight-medium">{{ $t('config.check.' + check.key) }}</span>
            <code class="text-body-small text-medium-emphasis">{{ check.key }}</code>
            <VSpacer />
            <VChip size="small" variant="tonal" :color="check.enabled ? 'primary' : 'grey'">
              {{ check.enabled ? $t('enabled') : $t('disabled') }}
            </VChip>
            <VChip size="small" variant="tonal" :color="check.configured ? 'success' : 'warning'">
              {{ check.configured ? $t('config.configured') : $t('config.not.configured') }}
            </VChip>
          </div>
          <div v-if="check.missing.length" class="flex flex-wrap items-center gap-2 pl-7">
            <span class="text-body-small text-medium-emphasis">{{ $t('config.missing') }}</span>
            <code v-for="key in check.missing" :key="key" class="text-body-small text-warning">{{ key }}</code>
          </div>
        </div>
      </VCardText>
    </VCard>
  </ComCtl>
</template>
