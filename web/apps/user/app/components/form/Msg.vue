<script setup lang="ts">
import type { Msg } from '~/lib/sdk'

defineProps<{
  loading: boolean
  text?: string
}>()

defineEmits<{
  (e: 'submit'): void
}>()

const model = defineModel<Msg>({ required: true })
</script>

<template>
  <NForm @submit="$emit('submit')" :loading="loading" :text="text">
    <UiFormField v-slot="{ componentField }" name="event">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.event') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.event" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="session_id">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.session.id') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.session_id" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="message_key">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.message.key') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.message_key" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="source_name">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.source.name') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.source_name" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="group_name">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.group.name') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.group_name" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
    <UiFormField v-slot="{ componentField }" name="content">
      <UiFormItem>
        <UiFormLabel>{{ $t('msg.content') }}</UiFormLabel>
        <UiFormControl>
          <UiInput v-bind="componentField" v-model="model.content" />
        </UiFormControl>
        <UiFormMessage />
      </UiFormItem>
    </UiFormField>
  </NForm>
</template>
