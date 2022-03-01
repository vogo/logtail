<script setup lang="ts">
import { ref } from '@vue/reactivity'
import { onMounted } from '@vue/runtime-core'
import axios from 'axios'
import { useTransfers } from '../../service/transfer'

const { 
  loading: listLoading,
  transfers: list,
  getTransfers: initData
} = useTransfers()

const types = ref([] as any[])
const typeLoading = ref(false)
function initTypes() {
  typeLoading.value = true
  axios.get('/manage/transfer/types')
    .then(({ data }) => {
      if (!Array.isArray(data)) return
      types.value = data.map(item => ({ label: item, value: item }))
    })
    .finally(() => typeLoading.value = false)
}
onMounted(() => initTypes())

const data = ref({} as any)
const createDialogVisible = ref(false)
function showCreateDialog() {
  data.value = {}
  createDialogVisible.value = true
}
const saveLoading = ref(false)
function createTransfer() {
  saveLoading.value = true
  axios.post('/manage/transfer/add', data.value)
    .then(() => createDialogVisible.value = false)
    .then(() => initData())
    .finally(() => saveLoading.value = false)
}
function deleteTransfer({ name = '' } = {}) {
  if (!name) return
  axios.post('/manage/transfer/delete', { name })
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
      <ElTableColumn prop="type" label="type" />
      <ElTableColumn prop="dir" label="dir" />
      <ElTableColumn prop="url" label="url" />
      <ElTableColumn fixed="right" label="operations" width="120">
        <template #default="{ row }">
          <ElPopconfirm 
            title="Are you sure to delete this?" 
            confirmButtonType="danger"
            @confirm="() => deleteTransfer(row)">
            <template #reference>
              <ElButton type="danger" size="small">Delete</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>
    <ElDialog 
      title="Create Transfer"
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
          prop="type"
          label="type">
          <ElSelect v-model="data.type" :loading="typeLoading">
            <ElOption
              v-for="item in types"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </ElOption>
          </ElSelect>
        </ElFormItem>
        <ElFormItem
          prop="dir"
          label="dir">
          <ElInput v-model="data.dir"></ElInput>
        </ElFormItem>
        <ElFormItem
          prop="url"
          label="url">
          <ElInput v-model="data.url"></ElInput>
        </ElFormItem>
      </ElForm>
      <ElRow justify="end">
        <ElButton 
          :loading="saveLoading" 
          type="primary" 
          @click="createTransfer">
          Save
        </ElButton>
      </ElRow>
    </ElDialog>
  </div>
</template>