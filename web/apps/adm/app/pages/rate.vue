<script setup lang="ts">
import { RatesPage, api } from '@sdk/client/adm'

definePageMeta({
  name: 'rate',
})

const { t } = useI18n()
const queries = ref({ page: 1, size: 20, column: 'base_code', desc: false })

const { loading, data, send } = useWatcher(
  () => api.rate.list(queries.value),
  // A shallow copy is enough: every mutation replaces a field, and abortLast
  // cancels the in-flight request so results cannot land out of order.
  [() => ({ ...queries.value })],
  {
    immediate: true,
    debounce: 200,
    abortLast: true,
    initialData: inst(RatesPage),
  },
)
// The sync rebuilds the whole table server-side, so reload afterwards.
const { loading: syncing, send: sync } = useRequest(() => api.rate.sync(), {
  immediate: false,
}).onSuccess(send)

const headers = computed(() => [
  { title: t('rate.base'), key: 'base_code', sortable: true },
  { title: t('rate.code'), key: 'code', sortable: true },
  { title: t('rate.value'), key: 'rate', sortable: true },
])
</script>

<template>
  <ComCtl :loading="[loading, syncing]" class="flex flex-col gap-4 p-4 lg:p-6">
    <AppPageHeader icon="mdi-currency-usd" :title="$t('urls.rate')" :description="$t('rate.desc')" :count="data.total">
      <template #actions>
        <VBtn color="primary" prepend-icon="mdi-sync" class="text-none" :loading="syncing" :disabled="!data.enabled" @click="sync">
          {{ $t('rate.sync') }}
        </VBtn>
      </template>
    </AppPageHeader>

    <VAlert v-if="!data.enabled" type="warning" variant="tonal" density="comfortable" class="flex-none">
      {{ $t('rate.disabled') }}
    </VAlert>

    <VCard border rounded="lg" class="flex-1 min-h-0 overflow-hidden">
      <VDataTableServer
        v-model:items-per-page="queries.size"
        v-model:page="queries.page"
        :headers="headers"
        :items="data.content"
        :items-length="data.total"
        :items-per-page-options="[10, 20, 50, 100]"
        :loading="loading"
        no-data-text="no.data"
        class="adm-table"
        density="comfortable"
        hover
        @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
      >
        <template #[`item.base_code`]="{ item }">
          <VChip size="small" variant="tonal" color="primary">{{ item.base_code }}</VChip>
        </template>
        <template #[`item.code`]="{ item }">
          <span class="text-body-medium">{{ item.code }}</span>
        </template>
        <template #[`item.rate`]="{ item }">
          <span class="text-body-medium tabular-nums">{{ item.rate }}</span>
        </template>
      </VDataTableServer>
    </VCard>
  </ComCtl>
</template>
