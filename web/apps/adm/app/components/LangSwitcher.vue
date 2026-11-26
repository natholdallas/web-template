<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'

const { locale } = useI18n()
</script>

<template>
  <VMenu location="bottom end">
    <template #activator="{ props }">
      <VBtn v-bind="props" icon="mdi-translate" variant="text" :title="$t('switch.locale')" />
    </template>

    <VList density="comfortable" min-width="200">
      <VListItem
        v-for="l in Object.values(Locale)"
        :key="l"
        :title="$t(locales[l].k)"
        :active="locale === l"
        :append-icon="locale === l ? 'mdi-check' : undefined"
        @click="$i18n.setLocale(l)"
      />
    </VList>
  </VMenu>
</template>
