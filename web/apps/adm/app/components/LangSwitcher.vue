<script setup lang="ts">
import { Locale, locales } from '~/lib/locale'

const showLang = shallowRef(false)
</script>

<template>
  <div>
    <VBtn @click="showLang = true" icon="mdi-web" variant="flat" />
    <VxModal v-model="showLang" :title="$t('switch.locale')">
      <VBtn
        v-for="i in Locale"
        @click="
          () => {
            $i18n.setLocale(i)
            showLang = false
          }
        "
        :text="$t(locales[i].k)"
        flat
        block
      />
    </VxModal>
  </div>
</template>
