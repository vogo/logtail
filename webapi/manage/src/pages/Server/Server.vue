<script setup lang="ts">
import axios from 'axios'
import { omit, pick } from 'lodash-es'
import { computed, ref } from 'vue'
import { useRouters } from '../../service/router'
import { useServers, useServerTypes } from '../../service/server'

const {
  loading: listLoading,
  servers: list,
  getServers: initData
} = useServers()

const data = ref({} as any)
const createDialogVisible = ref(false)
function showCreateDialog() {
  data.value = {}
  createDialogVisible.value = true
}
const { 
  loading: serverTypeLoading,
  serverTypes,
} = useServerTypes()
const serverTypeOptions = computed(() => {
  if (!Array.isArray(serverTypes.value)) return []
  return serverTypes.value.map(type => ({ label: type, value: type }))
})
const {
  loading: routerLoading,
  routers
} = useRouters()
const routerOptions = computed(() => {
  if (!Array.isArray(routers.value)) return []
  return routers.value.map(({ name }) => ({ label: name, value: name }))
})
const saveLoading = ref(false)
function createServer() {
  saveLoading.value = true
  axios.post('/manage/server/add', pick(data.value, ['name', 'routers', data.value.type]))
    .then(() => createDialogVisible.value = false)
    .then(() => initData())
    .finally(() => saveLoading.value = false)
}
function deleteServer({ name = '' } = {}) {
  if (!name) return
  axios.post('/manage/server/delete', { name })
    .then(() => initData())
}
</script>

<template>
  <div style="padding: 0 20px">
    <ElRow style="padding: 12px 0">
      <ElButton type="primary" @click="showCreateDialog">Create</ElButton>
    </ElRow>
    <ElTable :data="list" height="500" :loading="listLoading">
      <ElTableColumn prop="name" label="name" width="200" />
      <ElTableColumn prop="routers" label="routers">
        <template #default="{ row }">
          <ElTag
            style="margin: 0 5px 5px 0;"
            v-for="(tag, index) in row.routers"
            :key="tag"
          >
            {{ tag }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="others" label="others">
        <template #default="{ row }">
          <ElTag
            style="margin: 0 5px 5px 0;"
            v-for="(tag, index) in  Object.keys(row)
              .filter(key => !['name', 'routers'].includes(key))
              .map(key => `${key}: ${row[key]}`)"
            :key="tag"
          >
            {{ tag }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn fixed="right" label="operations" width="120">
        <template #default="{ row }">
          <ElPopconfirm 
            title="Are you sure to delete this?" 
            confirmButtonType="danger"
            @confirm="() => deleteServer(row)">
            <template #reference>
              <ElButton type="danger" size="small">Delete</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>
    <ElDialog 
      title="Create Server"
      v-model="createDialogVisible" 
      :close-on-click-modal="false"
      :close-on-press-escape="false">
      <ElForm :model="data" label-width="120px">
        <ElFormItem
          prop="name"
          label="name">
          <ElInput v-model="data.name"></ElInput>
        </ElFormItem>
        <ElFormItem
          prop="routers"
          label="routers">
          <ElSelect v-model="data.routers" :loading="routerLoading" multiple>
            <ElOption
              v-for="item in routerOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </ElOption>
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          prop="type"
          label="type">
          <ElSelect v-model="data.type" :loading="serverTypeLoading">
            <ElOption
              v-for="item in serverTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </ElOption>
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          v-if="data.type"
          :prop="data.type"
          :label="data.type">
          <ElInput v-model="data[data.type]"></ElInput>
        </ElFormItem>
      </ElForm>
      <ElRow justify="end">
        <ElButton 
          :loading="saveLoading" 
          type="primary" 
          @click="createServer">
          Save
        </ElButton>
      </ElRow>
    </ElDialog>
  </div>
</template>