<script setup lang="ts">
import { AdminIn, Apis } from '~/lib/sdk'

definePageMeta({
  name: 'admin',
  middleware: 'auth',
})

const queries = ref({ page: 1, size: 20, column: 'id', desc: true })

const { open } = useDialog()
const { mi, mo, sc, su, reset } = useCrud({ ...inst(AdminIn), id: 0 })
const { loading, data, send } = useRequest(
  () =>
    Apis.Admin.list({
      params: queries.value,
    }),
  {
    initialData: { total: 0, page: 0, content: [] },
  },
).onSuccess(reset)
const { loading: creating, send: create } = useRequest(
  () =>
    Apis.Admin.create({
      data: { username: mi.value.username, password: mi.value.password },
    }),
  {
    immediate: false,
  },
).onSuccess(send)
const { loading: updating, send: update } = useRequest(
  () =>
    Apis.Admin.update({
      pathParams: { id: mo.value.id },
      data: { username: mo.value.username, password: mo.value.password },
    }),
  {
    immediate: false,
  },
).onSuccess(send)
const { loading: removing, send: remove } = useRequest(
  () =>
    Apis.Admin.remove({
      pathParams: { id: mo.value.id },
    }),
  {
    immediate: false,
  },
).onSuccess(send)

const resetPwd = ref('')
const resetPwdOpen = ref(false)
const { loading: resetting, send: resetPassword } = useRequest(
  (id: number) =>
    Apis.Admin.resetAdminPassword({
      pathParams: { id },
    }),
  {
    immediate: false,
  },
).onSuccess(({ data }) => {
  resetPwd.value = data.password
  resetPwdOpen.value = true
})

watch(queries, send, { deep: true })
</script>

<template>
  <ComCtl :loading="resetting">
    <VDataTableServer
      v-model:items-per-page="queries.size"
      v-model:page="queries.page"
      class="h-full"
      :headers="[{ title: $t('model.id'), key: 'id' }, { title: $t('user.username'), key: 'username' }, { key: 'data-table-expand' }]"
      :items="data.content"
      :items-length="data.total"
      :loading="loading || removing"
      @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
    >
      <template #top>
        <TopTableBar v-model="sc" />
      </template>

      <template #item.data-table-expand="{ internalItem, item, isExpanded, toggleExpand }">
        <div class="flex gap-2 items-center">
          <VxActionBtn
            icon="mdi-pencil"
            @click="
              () => {
                mo = cpm({ ...item, password: '' })
                su = true
              }
            "
          />

          <VxActionBtn
            icon="mdi-delete"
            @click="
              () => {
                mo = cpm({ ...item, password: '' })
                open({ confirm: remove })
              }
            "
          />

          <VxActionBtn icon="mdi-key-refresh" @click="resetPassword(item.id)" />
          <VxExpandBtn :item="internalItem" @expanded="isExpanded" @toggle="toggleExpand" />
        </div>
      </template>

      <template #expanded-row="{ columns, item }">
        <RecordInfoTable :colspan="columns.length" :info="item" />
      </template>
    </VDataTableServer>

    <template #modals>
      <VxModal v-model="sc" :title="$t('create')">
        <FormAdmin v-model="mi" :loading="creating" @submit="create" />
      </VxModal>

      <VxModal v-model="su" :title="$t('update')">
        <FormAdmin v-model="mo" :loading="updating" @submit="update" />
      </VxModal>

      <VxModal v-model="resetPwdOpen" :title="$t('reset.password')">
        <div class="flex flex-col items-center gap-2 p-2">
          <p>{{ $t('reset.password.desc') }}</p>
          <VxCopyable :text="resetPwd" />
        </div>
      </VxModal>
    </template>
  </ComCtl>
</template>
