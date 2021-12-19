<script setup lang="ts">
import { computed, ref } from "@vue/reactivity"
import axios from "axios"
import { useTransfers } from "../../service/transfer"
import { Delete, Close } from '@element-plus/icons-vue'
import { useRouters } from "../../service/router"
import RouterMatcher from "./RouterMatcher.vue"

const {
  loading: transferLoading,
  transfers,
} = useTransfers()

const transferOptions = computed(() => {
  if (!Array.isArray(transfers.value)) return []
  return transfers.value.map(({ name }) => ({ label: name, value: name }))
})

const { 
  loading: listLoading,
  routers: list,
  getRouters: initData
} = useRouters()

const data = ref({} as any)
const createDialogVisible = ref(false)
function showCreateDialog() {
  data.value = {
    matchers: [{}]
  }
  createDialogVisible.value = true
}
const saveLoading = ref(false)
function createRouter() {
  saveLoading.value = true
  axios.post('/manage/router/add', data.value)
    .then(() => createDialogVisible.value = false)
    .then(() => initData())
    .finally(() => saveLoading.value = false)
}
function addMatcherContain(matcher: any) {
  if (!matcher._contain) return
  if (!Array.isArray(matcher.contains)) matcher.contains = []
  matcher.contains.push(matcher._contain)
  matcher._contain = ''
}
function addMatcherNotContain(matcher: any) {
  if (!matcher._not_contain) return
  if (!Array.isArray(matcher.not_contains)) matcher.not_contains = []
  matcher.not_contains.push(matcher._not_contain)
  matcher._not_contain = ''
}
function deleteRouter({ name = '' } = {}) {
  if (!name) return
  axios.post('/manage/router/delete', { name })
    .then(() => initData())
}
</script>

<template>
  <div style="padding: 0 20px">
    <ElRow style="padding: 12px 0">
      <ElButton type="primary" @click="showCreateDialog">Create</ElButton>
    </ElRow>
    <ElTable :data="list" height="500">
      <ElTableColumn prop="name" label="name" width="200" />
      <ElTableColumn prop="matchers" label="matchers">
        <template #default="{ row }">
          <RouterMatcher v-for="(matcher, index) in row.matchers"
            :matcher="matcher"
            :disabled="true" />
        </template>
      </ElTableColumn>
      <ElTableColumn prop="transfers" label="transfers">
        <template #default="{ row }">
          <ElTag
            style="margin: 0 5px 5px 0;"
            v-for="(tag, index) in row.transfers"
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
            @confirm="() => deleteRouter(row)">
            <template #reference>
              <ElButton type="danger" size="small">Delete</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>
    <ElDialog 
      title="Create Router"
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
          prop="matchers"
          label="matchers">
          <RouterMatcher v-for="(matcher, index) in data.matchers"
            :matcher="matcher"
            @remove="data.matchers.splice(index, 1)"
            @add-contain="addMatcherContain(matcher)"
            @remove-contain="matcher.contains.splice($event, 1)"
            @update-underscore-contain="matcher._contain = $event"
            @add-not-contain="addMatcherNotContain(matcher)"
            @remove-not-contain="matcher.not_contains.splice($event, 1)"
            @update-underscore-not-contain="matcher._not_contain = $event" />
          <ElButton 
            size="small"
            type="primary" 
            plain
            @click="() => {
              if (!Array.isArray(data.matchers)) return data.matchers = [{}]
              data.matchers.push({})
            }">
            Add Matcher
          </ElButton>
        </ElFormItem>
        <ElFormItem
          prop="transfers"
          label="transfers">
          <ElSelect v-model="data.transfers" multiple>
            <el-option
              v-for="item in transferOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
            </el-option>
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <ElRow justify="end">
        <ElButton 
          :loading="saveLoading" 
          type="primary" 
          @click="createRouter">
          Save
        </ElButton>
      </ElRow>
    </ElDialog>
  </div>
</template>