<script setup lang="ts">
const { mdAndUp } = useDisplay()
const { open } = useDialog()

const conf = useConf()
const drawer = ref(mdAndUp.value)
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VAppBar name="app-bar">
      <template #prepend>
        <VAppBarNavIcon @click="drawer = !drawer" />
      </template>
      <VAppBarTitle text="picfans" />
      <template #append>
        <VMenu>
          <template #activator="{ props }">
            <LangSwitcher />
            <VBtn icon="mdi-dots-vertical" variant="text" v-bind="props" />
          </template>
          <VList>
            <VListItem
              @click="conf.theme = conf.theme === 'dark' ? 'light' : 'dark'"
              :title="$t('switch.theme')"
              prepend-icon="mdi-theme-light-dark"
            />
            <VListItem
              @click="open({ confirm: () => {} })"
              :title="$t('sign.out')"
              prepend-icon="mdi-logout"
            />
          </VList>
        </VMenu>
      </template>
    </VAppBar>

    <VNavigationDrawer
      v-model="drawer"
      :expand-on-hover="mdAndUp"
      :rail="mdAndUp"
      mobile-breakpoint="md"
    >
      <VList>
        <VListItem :title="$t('urls.index')" to="/" prepend-icon="mdi-account" />
        <VListItem :title="$t('urls.admin')" to="/admin" prepend-icon="mdi-account-supervisor" />
      </VList>
    </VNavigationDrawer>

    <VMain class="size-full" name="main">
      <slot></slot>
    </VMain>

    <ProviderDialog />
    <ProviderSnackbar />
  </VApp>
</template>
