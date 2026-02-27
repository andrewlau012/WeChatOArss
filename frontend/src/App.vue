
<template>
  <n-config-provider :theme-overrides="themeOverrides">
    <n-global-style />
    <n-message-provider>
      <router-view />
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { NConfigProvider, NGlobalStyle, NMessageProvider } from 'naive-ui'

const themeOverrides = {
  common: {
    primaryColor: '#07c160',
    primaryColorHover: '#09d66d',
  }
}
</script>

<style>
body, html, #app {
  height: 100%;
  margin: 0;
  overflow: hidden;
}
</style>
