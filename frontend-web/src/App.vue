<script setup lang="ts">
import { RouterView } from 'vue-router'
import { Toaster } from '@/components/ui/sonner'
</script>

<template>
  <div id="loading-bar"></div>
  <Toaster />
  <RouterView />
</template>

<style>
/* Reset basic styles and ensure full height */
html,
body,
#app {
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>
