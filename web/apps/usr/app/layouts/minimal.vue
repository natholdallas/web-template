<script setup lang="ts">
import { locales } from '~/lib/locale'
</script>

<template>
  <div class="flex fixed size-full overflow-hidden">
    <div class="flex flex-col size-full overflow-hidden">
      <header class="h-15 border-b sticky top-0 flex items-center justify-between px-4 shrink-0">
        <div class="flex items-center gap-2">
          <span class="text-sm text-muted-foreground capitalize">{{ $route.name }}</span>
        </div>

        <div class="flex items-center gap-1">
          <SwitchLang
            @update="$i18n.setLocale"
            :options="Object.values(locales).map(({ k, v }) => ({ label: $t(k), key: k, value: v }))"
            :value="$t(locales[$i18n.locale].k)"
          />
          <SwitchTheme />
        </div>
      </header>

      <main class="flex-1 min-h-0 overflow-hidden">
        <slot />
      </main>
    </div>
    <div class="contents">
      <slot name="clean" />
    </div>
  </div>
</template>
