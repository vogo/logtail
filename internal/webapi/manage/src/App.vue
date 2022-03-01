<script setup lang="ts">
import { computed } from '@vue/reactivity';
import { useRouter } from 'vue-router';
const router = useRouter()
const initActive = computed(() => router.currentRoute.value.path)
</script>

<template>
  <div :class="$style.app">
    <ElMenu :default-active="initActive" mode="horizontal" router>
      <ElMenuItem index="/">Home</ElMenuItem>
      <ElMenuItem index="/transfer">Transfer</ElMenuItem>
      <ElMenuItem index="/router">Router</ElMenuItem>
      <ElMenuItem index="/server">Server</ElMenuItem>
    </ElMenu>
    <RouterView />
  </div>
</template>

<style>
body {
  margin: 0;
}
</style>

<style module>
.app {
  min-height: 100vh;
}
</style>
