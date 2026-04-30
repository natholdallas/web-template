<script setup lang="ts">
import { BaseQueries, CreateUser, ListUser, Page, RemoveUser, UpdateUser, User } from '~/sdk'

definePageMeta({
  name: 'index',
  middleware: 'auth',
})

const queries = ref(inst(BaseQueries, { column: 'id', desc: true }))

const { open } = useDialog()
const { mi, mo, sc, su, reset } = useCrud(inst(User))
const { loading, data, send } = useRequest(() => ListUser(queries.value), {
  initialData: inst(Page),
}).onSuccess(reset)
const { loading: creating, send: create } = useRequest(() => CreateUser(mi.value), {
  immediate: false,
}).onSuccess(send)
const { loading: updating, send: update } = useRequest(() => UpdateUser(mo.value), {
  immediate: false,
}).onSuccess(send)
const { loading: removing, send: remove } = useRequest(() => RemoveUser(mo.value.id), {
  immediate: false,
}).onSuccess(send)

watch(queries, send, { deep: true })
</script>

<template>
  <ComCtl>
    <VDataTableServer
      v-model:items-per-page="queries.size"
      v-model:page="queries.page"
      @update:options="({ sortBy }) => vtables.sort(queries, sortBy)"
      :items-length="data.total"
      :loading="loading || removing"
      :items="data.content"
      :headers="[
        { title: $t('model.id'), key: 'id' },
        { title: $t('user.username'), key: 'username' },
        { title: $t('user.password'), key: 'password' },
        { key: 'data-table-expand' },
      ]"
      class="h-full"
    >
      <template #top>
        <TopTableBar v-model="sc" :text="$t('urls.index')" />
      </template>
      <template #item.data-table-expand="{ internalItem, item, isExpanded, toggleExpand }">
        <div class="flex gap-2 items-center">
          <NActionBtn
            @click="
              () => {
                mo = cpm(item)
                su = true
              }
            "
            icon="mdi-pencil"
          />
          <NActionBtn
            @click="
              () => {
                mo = cpm(item)
                open({
                  confirm: remove,
                })
              }
            "
            icon="mdi-delete"
          />
          <NExpandBtn @expanded="isExpanded" @toggle="toggleExpand" :item="internalItem" />
        </div>
      </template>
      <template #expanded-row="{ columns, item }">
        <RecordInfoTable :colspan="columns.length" :info="item" />
      </template>
    </VDataTableServer>
    <template #modals>
      <NModal v-model="sc" :title="$t('create')">
        <FormUser v-model="mi" @submit="create" :loading="creating" />
      </NModal>
      <NModal v-model="su" :title="$t('update')">
        <FormUser v-model="mo" @submit="update" :loading="updating" />
      </NModal>
    </template>
  </ComCtl>
</template>
