<!-- SysConf has no business fields yet; this page is intentionally disabled until the model gains fields. -->
<script setup lang="ts">
definePageMeta({
  name: 'sysconf',
})
</script>

<template>
  <div class="flex flex-col gap-4 p-4 lg:p-6">
    <AppPageHeader icon="mdi-database-cog" :title="$t('urls.sysconf')" />

    <VCard border rounded="lg" class="flex flex-1 flex-col items-center justify-center gap-3 p-10 text-center">
      <div class="grid size-12 place-items-center rounded-xl bg-surface-variant text-medium-emphasis">
        <VIcon icon="mdi-database-cog" size="26" />
      </div>
      <p class="text-title-small font-weight-bold">{{ $t('sysconf.disabled') }}</p>
      <p class="max-w-md text-body-small text-medium-emphasis">{{ $t('sysconf.disabled.desc') }}</p>
      <VChip color="grey" variant="tonal" size="small">{{ $t('disabled') }}</VChip>
    </VCard>
  </div>
</template>
