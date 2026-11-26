<script setup lang="ts">
import { Stats, api } from '@sdk/client/adm'

definePageMeta({
  name: 'index',
})

const { loading, data, send } = useRequest(api.admin.getStats, {
  initialData: inst(Stats),
})

useInterval(send, 30_000)

const quickLinks = [
  { to: '/user', icon: 'mdi-account-group', title: 'dashboard.quick.users', desc: 'dashboard.quick.users.desc' },
  { to: '/media', icon: 'mdi-image-multiple', title: 'dashboard.quick.media', desc: 'dashboard.quick.media.desc' },
]
</script>

<template>
  <ComCtl :loading="loading" class="p-4 lg:p-6" scroll>
    <div class="mx-auto flex max-w-5xl flex-col gap-6">
      <AppPageHeader icon="mdi-view-dashboard" :title="$t('dashboard.title')" :description="$t('dashboard.desc')">
        <template #actions>
          <VChip color="success" variant="tonal" prepend-icon="mdi-circle">{{ $t('dashboard.live') }}</VChip>
          <AppIconBtn icon="mdi-refresh" :tooltip="$t('dashboard.refresh')" @click="send" />
        </template>
      </AppPageHeader>

      <div class="grid gap-4 sm:grid-cols-2">
        <AppStatCard
          icon="mdi-shield-account"
          :value="data.admins"
          :label="$t('dashboard.total.admins')"
          tone="primary"
        />
        <AppStatCard icon="mdi-account-group" :value="data.users" :label="$t('dashboard.total.users')" tone="info" />
      </div>

      <div class="flex flex-col gap-3">
        <h2 class="text-body-large font-weight-bold">{{ $t('dashboard.quick') }}</h2>
        <div class="grid gap-4 sm:grid-cols-2">
          <NuxtLink v-for="link in quickLinks" :key="link.to" :to="link.to" class="block">
            <VCard border rounded="lg" class="h-full transition hover:border-primary/40">
              <VCardText class="flex items-center gap-4 p-5">
                <div class="grid size-10 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary">
                  <VIcon :icon="link.icon" size="22" />
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-body-medium font-weight-bold text-truncate">{{ $t(link.title) }}</p>
                  <p class="text-body-small text-medium-emphasis">{{ $t(link.desc) }}</p>
                </div>
                <VIcon icon="mdi-chevron-right" size="20" class="text-medium-emphasis" />
              </VCardText>
            </VCard>
          </NuxtLink>
        </div>
      </div>
    </div>
  </ComCtl>
</template>
