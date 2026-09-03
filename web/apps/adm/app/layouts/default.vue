<script setup lang="ts">
const { mdAndUp } = useVDisplay()
const { open } = useDialog()

const auth = useAuthStore()
const conf = useConfStore()
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
            <VListItem prepend-icon="mdi-theme-light-dark" :title="$t('switch.theme')" @click="conf.$switchTheme" />
            <VListItem prepend-icon="mdi-logout" :title="$t('sign.out')" @click="open({ confirm: auth.$signOut })" />
          </VList>
        </VMenu>
      </template>
    </VAppBar>

    <VNavigationDrawer v-model="drawer" :expand-on-hover="mdAndUp" mobile-breakpoint="md" :rail="mdAndUp">
      <VList>
        <VListItem prepend-icon="mdi-home" :title="$t('urls.index')" to="/" />
        <VListItem prepend-icon="mdi-account-supervisor" :title="$t('urls.admin')" to="/admin" />
        <VListItem prepend-icon="mdi-account" :title="$t('urls.user')" to="/user" />
        <VListItem prepend-icon="mdi-cog" :title="$t('urls.settings')" to="/settings" />
      </VList>
    </VNavigationDrawer>

    <VMain class="size-full" name="main">
      <slot />
    </VMain>
  </VApp>
</template>
