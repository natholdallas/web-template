<script setup lang="ts">
const { mdAndUp } = useVDisplay()
const route = useRoute()
const auth = useAuthStore()
const conf = useConfStore()
const { t } = useI18n()

const drawer = ref(mdAndUp.value)
const rail = ref(false)

watch(mdAndUp, (v) => {
  drawer.value = v
  if (v) rail.value = false
})
watch(
  () => route.fullPath,
  () => {
    if (!mdAndUp.value) drawer.value = false
  },
)

function toggleNav() {
  if (mdAndUp.value) rail.value = !rail.value
  else drawer.value = !drawer.value
}

// Shared with the settings page, so an avatar change shows up here immediately.
const store = useProfileStore()
const { profile } = storeToRefs(store)
store.$load()
const app = useAppStore()
app.$load()

const displayName = computed(() => profile.value.username || `#${auth.id}`)
const initial = computed(() => (profile.value.username || String(auth.id || '?')).slice(0, 1).toUpperCase())

const title = computed(() => t(`urls.${String(route.name)}`))
const subtitle = computed(() => {
  switch (route.name) {
    case 'index':
      return t('dashboard.desc')
    case 'admin':
      return t('admin.desc')
    case 'user':
      return t('user.desc')
    case 'media':
      return t('media.desc')
    case 'rate':
      return t('rate.desc')
    case 'sysconf':
      return t('sysconf.disabled.desc')
    case 'config':
      return t('config.desc')
    case 'settings':
      return t('settings.desc')
    default:
      return ''
  }
})

const signOutOpen = ref(false)

type NavItem = {
  name: 'index' | 'admin' | 'user' | 'media' | 'rate' | 'config' | 'settings'
  icon: string
  title: string
}

const navGroups: { label: string; items: NavItem[] }[] = [
  { label: 'nav.overview', items: [{ name: 'index', icon: 'mdi-view-dashboard', title: 'urls.index' }] },
  {
    label: 'nav.management',
    items: [
      { name: 'admin', icon: 'mdi-shield-account', title: 'urls.admin' },
      { name: 'user', icon: 'mdi-account-group', title: 'urls.user' },
      { name: 'media', icon: 'mdi-image-multiple', title: 'urls.media' },
      { name: 'rate', icon: 'mdi-currency-usd', title: 'urls.rate' },
    ],
  },
  {
    label: 'nav.system',
    items: [
      { name: 'config', icon: 'mdi-clipboard-check-outline', title: 'urls.config' },
      { name: 'settings', icon: 'mdi-cog', title: 'urls.settings' },
    ],
  },
]
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VAppBar flat height="64" color="surface" class="border-b border-on-surface/10">
      <template #prepend>
        <VAppBarNavIcon @click="toggleNav" />
      </template>

      <div class="min-w-0">
        <div class="text-body-medium font-weight-medium text-truncate">{{ title }}</div>
        <div class="text-body-small text-medium-emphasis d-none d-sm-block text-truncate">{{ subtitle }}</div>
      </div>

      <VSpacer />

      <template #append>
        <LangSwitcher />
        <AppThemeBtn />

        <VMenu location="bottom end">
          <template #activator="{ props }">
            <VBtn v-bind="props" icon variant="text" :aria-label="displayName">
              <VAvatar size="32" color="primary" variant="tonal" class="overflow-hidden">
                <VImg v-if="profile.avatar" :src="app.mediaUrl(profile.avatar)" cover alt="" />
                <span v-else class="text-body-medium font-weight-bold">{{ initial }}</span>
              </VAvatar>
            </VBtn>
          </template>

          <VList density="comfortable" min-width="210">
            <VListItem :title="displayName" :subtitle="$t('role.admin')" prepend-icon="mdi-account-circle" />
            <VDivider />
            <VListItem prepend-icon="mdi-cog" :title="$t('urls.settings')" :to="{ name: 'settings' }" />
            <VListItem prepend-icon="mdi-logout" :title="$t('sign.out')" @click="signOutOpen = true" />
          </VList>
        </VMenu>
      </template>
    </VAppBar>

    <VNavigationDrawer
      v-model="drawer"
      :rail="rail && mdAndUp"
      :permanent="mdAndUp"
      :temporary="!mdAndUp"
      width="264"
      rail-width="76"
      color="surface"
      class="border-e border-on-surface/10"
    >
      <div class="flex h-full flex-col">
        <div class="flex shrink-0 items-center px-4 py-4" :class="rail ? 'justify-center' : ''">
          <AppBrand :compact="rail && mdAndUp" />
        </div>
        <VDivider />

        <div class="min-h-0 flex-1 overflow-y-auto">
          <template v-for="group in navGroups" :key="group.label">
            <div v-if="!(rail && mdAndUp)" class="nav-group-label px-4 pt-5 pb-1">{{ $t(group.label) }}</div>
            <VList nav density="comfortable" class="px-2 py-1">
              <VListItem
                v-for="item in group.items"
                :key="item.name"
                class="nav-item rounded-lg mb-0.5"
                :to="{ name: item.name }"
                :prepend-icon="item.icon"
                :title="$t(item.title)"
              />
            </VList>
          </template>
        </div>

        <div class="shrink-0 border-t border-on-surface/10 p-3">
          <div class="flex items-center gap-3" :class="rail && mdAndUp ? 'justify-center' : ''">
            <VAvatar size="36" color="primary" variant="tonal" class="overflow-hidden">
              <VImg v-if="profile.avatar" :src="app.mediaUrl(profile.avatar)" cover alt="" />
              <span v-else class="text-body-medium font-weight-bold">{{ initial }}</span>
            </VAvatar>
            <div v-if="!(rail && mdAndUp)" class="min-w-0 flex-1">
              <p class="text-body-medium font-weight-medium text-truncate">{{ displayName }}</p>
              <p class="text-body-small text-medium-emphasis">{{ $t('role.admin') }}</p>
            </div>
            <VBtn
              v-if="!(rail && mdAndUp)"
              icon="mdi-logout"
              variant="text"
              size="small"
              :title="$t('sign.out')"
              @click="signOutOpen = true"
            />
          </div>
        </div>
      </div>
    </VNavigationDrawer>

    <VMain class="size-full" name="main"><slot /></VMain>

    <AppModal
      v-model="signOutOpen"
      :title="$t('sign.out')"
      :description="$t('sign.out.desc')"
      icon="mdi-logout"
      tone="error"
      :confirm-text="$t('sign.out')"
      @confirm="auth.$signOut"
    />
  </VApp>
</template>
