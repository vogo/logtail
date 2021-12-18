<script setup lang="ts">
import { useRouter } from 'vue-router';
const router = useRouter()
console.log(router)
</script>

<template>
  <div :class="$style.app">
    <ElMenu mode="horizontal">
      <ElMenuItem index="1">Home</ElMenuItem>
      <ElMenuItem index="2">Transfer</ElMenuItem>
      <ElMenuItem index="3">Router</ElMenuItem>
      <ElMenuItem index="4">Server</ElMenuItem>
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
