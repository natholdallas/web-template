<script setup lang="ts">
import { Language, languages } from '~/sdk'

const showLang = shallowRef(false)
</script>

<template>
  <div>
    <VBtn @click="showLang = true" icon="mdi-web" variant="flat" />
    <NModal v-model="showLang" :title="$t('switch.locale')">
      <VBtn
        v-for="i in Language"
        @click="
          () => {
            $i18n.setLocale(i)
            showLang = false
          }
        "
        :text="$t(languages[i].k)"
        flat
        block
      />
    </NModal>
  </div>
</template>
