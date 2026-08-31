<script setup lang="ts">
const conf = useConf()
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VAppBar name="app-bar">
      <VAppBarTitle :text="$t(`app.name`).toUpperCase()" />

      <template #append>
        <div class="flex gap-1">
          <LangSwitcher />
          <VBtn icon="mdi-theme-light-dark" :text="$t('switch.theme')" variant="flat" @click="conf.$switchTheme" />
        </div>
      </template>
    </VAppBar>

    <VMain class="size-full" name="main">
      <slot />
    </VMain>
  </VApp>
</template>
