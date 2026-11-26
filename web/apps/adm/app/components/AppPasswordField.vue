<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
    autocomplete?: string
  }>(),
  { autocomplete: 'current-password', label: '' },
)

const model = defineModel<string>({ default: '' })
const show = ref(false)
</script>

<template>
  <VTextField
    v-model="model"
    :type="show ? 'text' : 'password'"
    :label="label"
    :autocomplete="autocomplete"
    :clearable="false"
    maxlength="72"
    :append-inner-icon="show ? 'mdi-eye-off' : 'mdi-eye'"
    variant="outlined"
    density="comfortable"
    hide-details="auto"
    @click:append-inner="show = !show"
  />
</template>
