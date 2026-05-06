<script setup lang="ts">
import { Home, LogOut, Menu, Settings } from 'lucide-vue-next'
import { locales } from '~/lib/locale'

const auth = useAuth()
const route = useRoute()
const isSidebarOpen = ref(true)
const isMobileOpen = ref(false)
const isLogoutDialogOpen = ref(false)

const menuItems = computed(() => [
  { label: $t('urls.home'), icon: Home, to: '/' },
  { label: $t('urls.settings'), icon: Settings, to: '/settings' },
])

function isActive(to: string) {
  return route.path === to
}

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}
</script>

<template>
  <div class="flex fixed size-full">
    <UiSheet v-model:open="isMobileOpen">
      <UiSheetContent side="left" class="w-72">
        <div class="flex flex-col h-full">
          <div class="flex items-center gap-2 px-4 py-4 border-b">
            <span class="font-semibold text-lg">{{ $t('app.name') }}</span>
          </div>
          <UiScrollArea class="flex-1 px-2">
            <nav class="space-y-1 py-2">
              <NuxtLink
                v-for="item in menuItems"
                :key="item.to"
                :to="item.to"
                @click="isMobileOpen = false"
                :class="[
                  'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  isActive(item.to)
                    ? 'bg-primary text-primary-foreground'
                    : 'text-muted-foreground hover:text-foreground hover:bg-secondary',
                ]"
              >
                <component :is="item.icon" class="w-5 h-5" />
                <span>{{ item.label }}</span>
              </NuxtLink>
            </nav>
          </UiScrollArea>
        </div>
      </UiSheetContent>
    </UiSheet>

    <div
      :class="[
        'hidden md:flex flex-col border-r bg-card transition-all duration-300',
        isSidebarOpen ? 'w-64' : 'w-16',
      ]"
    >
      <div class="p-4 flex items-center justify-between border-b min-h-14">
        <span v-if="isSidebarOpen" class="font-semibold text-lg">{{ $t('app.name') }}</span>
        <UiButton variant="ghost" size="icon" @click="toggleSidebar">
          <Menu class="w-4 h-4" />
        </UiButton>
      </div>

      <UiScrollArea class="flex-1 px-2">
        <nav class="space-y-1 py-2">
          <NuxtLink
            v-for="item in menuItems"
            :key="item.to"
            :to="item.to"
            :class="[
              'flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
              isActive(item.to)
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:text-foreground hover:bg-secondary',
            ]"
          >
            <component :is="item.icon" class="w-5 h-5 shrink-0" />
            <span v-if="isSidebarOpen">{{ item.label }}</span>
          </NuxtLink>
        </nav>
      </UiScrollArea>
    </div>

    <div class="flex flex-col size-full">
      <header class="h-15 border-b sticky top-0 flex items-center justify-between px-4">
        <div class="flex items-center gap-2">
          <UiButton variant="ghost" class="md:hidden" @click="isMobileOpen = true">
            <Menu class="w-5 h-5" />
          </UiButton>
          <span class="text-sm text-muted-foreground capitalize">{{ route.name }}</span>
        </div>

        <div class="flex items-center gap-1">
          <LangSwitcher
            @update="$i18n.setLocale"
            :options="Object.values(locales).map(({ k, v }) => ({ label: $t(k), value: v }))"
            :value="$t(locales[$i18n.locale].k)"
          />
          <ThemeSwitcher />
          <UiButton variant="ghost" size="icon" @click="isLogoutDialogOpen = true">
            <LogOut class="w-5 h-5" />
          </UiButton>
        </div>
      </header>

      <main class="flex-1 min-h-0">
        <slot />
      </main>
    </div>

    <UiDialog v-model:open="isLogoutDialogOpen">
      <UiDialogContent>
        <UiDialogHeader>
          <UiDialogTitle>{{ $t('sign.out') }}</UiDialogTitle>
          <UiDialogDescription>{{ $t('sign.out.desc') }}</UiDialogDescription>
        </UiDialogHeader>
        <UiDialogFooter>
          <UiButton variant="outline" @click="isLogoutDialogOpen = false">
            {{ $t('cancel') }}
          </UiButton>
          <UiButton @click="auth.$signOut">
            {{ $t('confirm') }}
          </UiButton>
        </UiDialogFooter>
      </UiDialogContent>
    </UiDialog>
  </div>
</template>
