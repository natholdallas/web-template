<script setup lang="ts">
const { mdAndUp } = useVDisplay()
const { open } = useDialog()

const auth = useAuth()
const conf = useConf()
const drawer = ref(mdAndUp.value)
</script>

<template>
  <VApp class="fixed size-full overflow-hidden" :theme="conf.theme">
    <VAppBar name="app-bar">
      <template #prepend>
        <VAppBarNavIcon @click="drawer = !drawer" />
      </template>
      <VAppBarTitle :text="$t(`urls.${$route.name}`)" />
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
            <VListItem @click="open({ confirm: auth.$signOut })" :title="$t('sign.out')" prepend-icon="mdi-logout" />
          </VList>
        </VMenu>
      </template>
    </VAppBar>

    <VNavigationDrawer v-model="drawer" :expand-on-hover="mdAndUp" :rail="mdAndUp" mobile-breakpoint="md">
      <VList>
        <VListItem :title="$t('urls.index')" :to="{ name: 'index' }" prepend-icon="mdi-home" />
        <VListItem :title="$t('urls.admin')" :to="{ name: 'admin' }" prepend-icon="mdi-account-supervisor" />
        <VListItem :title="$t('urls.user')" :to="{ name: 'user' }" prepend-icon="mdi-account" />
        <VListItem :title="$t('urls.settings')" :to="{ name: 'settings' }" prepend-icon="mdi-cog" />
      </VList>
    </VNavigationDrawer>

    <VMain class="size-full" name="main">
      <slot></slot>
    </VMain>
  </VApp>
</template>
