<script setup lang="ts">
import { Close } from '@element-plus/icons-vue'
import { defineProps, PropType, defineEmits } from 'vue'

interface Matcher {
  _contain?: string,
  _not_contain?: string,
  contains: string[],
  not_contains: string[]
}

defineProps({
  matcher: {
    type: Object as PropType<Matcher>,
    required: true
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

defineEmits({
  remove: null,

  removeContain: null,
  addContain: null,
  updateUnderscoreContain: null,
  
  removeNotContain: null,
  addNotContain: null,
  updateUnderscoreNotContain: null,
})

</script>
<template>
  <ElCard
    style="margin-bottom: 10px;"
    :body-style="{ padding: '10px' }"
    shadow="never">
    <ElRow v-if="!disabled" justify="end">
      <ElButton 
        style="padding: 0; margin-bottom: 10px; min-height: auto; color: var(--el-color-danger);"
        type="text"
        circle
        :icon="Close"
        @click="$emit('remove')" />
    </ElRow>
    <ElForm :model="matcher" size="mini" label-width="120px">
      <ElFormItem
        prop="contains"
        label="contains"
        style="margin-bottom: 10px;">
        <ElTag
          style="margin: 0 5px 5px 0;"
          v-for="(tag, index) in matcher.contains"
          :key="tag"
          :closable="!disabled"
          :disable-transitions="false"
          @close="$emit('removeContain', index)"
        >
          {{ tag }}
        </ElTag>
        <ElInput 
          v-if="!disabled"
          style="width: 150px;" 
          :model-value="matcher._contain"
          @update:model-value="$emit('updateUnderscoreContain', $event)"
          @blur="() => $emit('addContain')"
          @keyup.enter="() => $emit('addContain')">
          <template #append>
            <el-button
              @click="() => $emit('addContain')">
              Add
            </el-button>
          </template>
        </ElInput>
      </ElFormItem>
      <ElFormItem
        style="margin-bottom: 0;"
        prop="not_contains"
        label="not_contains">
        <ElTag
          style="margin: 0 5px 5px 0;"
          v-for="(tag, index) in matcher.not_contains"
          :key="tag"
          :closable="!disabled"
          :disable-transitions="false"
          @close="$emit('removeNotContain', index)"
        >
          {{ tag }}
        </ElTag>
        <ElInput
          v-if="!disabled"
          style="width: 150px;" 
          :model-value="matcher._not_contain"
          @update:model-value="$emit('updateUnderscoreNotContain', $event)"
          @blur="() => $emit('addNotContain')"
          @keyup.enter="() => $emit('addNotContain')">
          <template #append>
            <el-button
              @click="() => $emit('addNotContain')">
              Add
            </el-button>
          </template>
        </ElInput>
      </ElFormItem>
    </ElForm>
  </ElCard>
</template>