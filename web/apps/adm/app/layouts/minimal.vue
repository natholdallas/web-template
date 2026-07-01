<script setup lang="ts">
const conf = useConf()
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VAppBar name="app-bar">
      <VAppBarTitle :text="$t('app.name')" />
      <template #append>
        <div class="flex gap-1">
          <LangSwitcher />
          <VBtn
            @click="conf.theme = conf.theme === 'dark' ? 'light' : 'dark'"
            :text="$t('switch.theme')"
            icon="mdi-theme-light-dark"
            variant="flat"
          />
        </div>
      </template>
    </VAppBar>

    <VMain class="size-full" name="main">
      <slot></slot>
    </VMain>

    <ProviderDialog />
    <ProviderSnackbar />
  </VApp>
</template>
