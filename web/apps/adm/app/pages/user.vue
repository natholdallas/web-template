<script setup lang="ts">
import { Apis, UserIn } from '~/lib/sdk'

definePageMeta({
  name: 'user',
  middleware: 'auth',
})

const queries = ref({ page: 1, size: 20, column: 'id', desc: true })

const { open } = useDialog()
const { mi, mo, sc, su, reset } = useCrud({ ...inst(UserIn), id: 0 })
const { loading, data, send } = useRequest(
  () =>
    Apis.User.list({
      params: queries.value,
    }),
  {
    initialData: { total: 0, page: 0, content: [] },
  },
).onSuccess(reset)
const { loading: creating, send: create } = useRequest(
  () =>
    Apis.User.create({
      data: { username: mi.value.username, password: mi.value.password },
    }),
  {
    immediate: false,
  },
).onSuccess(send)
const { loading: updating, send: update } = useRequest(
  () =>
    Apis.User.update({
      pathParams: { id: mo.value.id },
      data: { username: mo.value.username, password: mo.value.password },
    }),
  {
    immediate: false,
  },
).onSuccess(send)
const { loading: removing, send: remove } = useRequest(
  () =>
    Apis.User.remove({
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
    Apis.User.resetUserPassword({
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
  <ComCtl :loading="resetting" class="flex flex-col p-4">
    <div class="flex size-full flex-1 flex-col gap-4">
      <div class="flex items-center justify-between gap-4">
        <div class="flex items-center gap-4">
          <div class="flex size-12 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <VIcon icon="mdi-account-group" size="28" />
          </div>
          <div>
            <h1 class="text-2xl font-bold leading-tight">{{ $t('urls.user') }}</h1>
            <p class="text-sm opacity-70">{{ $t('user.desc') }}</p>
          </div>
        </div>

        <VChip prepend-icon="mdi-account-multiple" variant="tonal" class="hidden sm:inline-flex" rounded> {{ $t('total') }}: {{ data.total }} </VChip>
      </div>

      <VCard border rounded="xl" class="flex-1 min-h-0 overflow-hidden">
        <VDataTableServer
          v-model:items-per-page="queries.size"
          v-model:page="queries.page"
          class="h-full"
          :headers="[
            { title: $t('model.id'), key: 'id', sortable: true },
            { title: $t('user.username'), key: 'username', sortable: true },
            { key: 'data-table-expand' },
          ]"
          :items="data.content"
          :items-length="data.total"
          :loading="loading || removing"
          @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
        >
          <template #top>
            <TopTableBar v-model="sc" :text="$t('user')" />
          </template>

          <template #item.data-table-expand="{ internalItem, item, isExpanded, toggleExpand }">
            <div class="flex items-center justify-end gap-1">
              <VxActionBtn
                icon="mdi-pencil"
                :tooltip="$t('update')"
                @click="
                  () => {
                    mo = cpm({ ...item, password: '' })
                    su = true
                  }
                "
              />
              <VxActionBtn
                icon="mdi-delete"
                :tooltip="$t('remove')"
                @click="
                  () => {
                    mo = cpm({ ...item, password: '' })
                    open({ confirm: remove })
                  }
                "
              />
              <VxActionBtn icon="mdi-key-refresh" :tooltip="$t('reset.password')" @click="resetPassword(item.id)" />
              <VxExpandBtn :item="internalItem" @expanded="isExpanded" @toggle="toggleExpand" />
            </div>
          </template>

          <template #expanded-row="{ columns, item }">
            <RecordInfoTable :colspan="columns.length" :info="item" />
          </template>
        </VDataTableServer>
      </VCard>
    </div>

    <template #modals>
      <VxModal v-model="sc" :title="$t('create')">
        <FormUser v-model="mi" :loading="creating" @submit="create" />
      </VxModal>

      <VxModal v-model="su" :title="$t('update')">
        <FormUser v-model="mo" :loading="updating" @submit="update" />
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
