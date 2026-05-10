<script setup lang="ts">
definePageMeta({
  name: 'home',
  middleware: 'auth',
})
</script>

<template>
  <ComCtl class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-2 p-4" scroll>
    hello
  </ComCtl>
</template>
