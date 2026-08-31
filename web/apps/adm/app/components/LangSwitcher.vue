<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'

const showLang = shallowRef(false)
</script>

<template>
  <div>
    <VBtn icon="mdi-web" variant="flat" @click="showLang = true" />

    <VxModal v-model="showLang" :title="$t('switch.locale')">
      <VBtn
        v-for="i in Locale"
        :key="i"
        block
        flat
        :text="$t(locales[i].k)"
        @click="
          () => {
            $i18n.setLocale(i)
            showLang = false
          }
        "
      />
    </VxModal>
  </div>
</template>
