<script setup lang="ts">
import { locales } from '~/lib/locale'

const route = useRoute()
</script>

<template>
  <div class="flex fixed size-full">
    <div class="flex flex-col size-full">
      <header class="h-15 border-b sticky top-0 flex items-center justify-between px-4">
        <div class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground capitalize">{{ route.name }}</span>
        </div>

        <div class="flex items-center gap-1">
          <LangSwitcher
            @update="$i18n.setLocale"
            :options="Object.values(locales).map(({ k, v }) => ({ label: $t(k), value: v }))"
            :value="$t(locales[$i18n.locale].k)"
          />
          <ThemeSwitcher />
        </div>
      </header>

      <main class="flex-1 min-h-0">
        <slot />
      </main>
    </div>
  </div>
</template>
