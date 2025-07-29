<script setup>
import Toast from "./components/Toast.vue";

import {useNavTabsStore} from "./store/tabsStore.js";

const tabsStore = useNavTabsStore();

(function () {
  const tg = window.Telegram.WebApp;
  tg.ready();
  tg.expand();

  tg.setHeaderColor("#1B232C");
  tg.setBackgroundColor("#1B232C");

  tg.disableVerticalSwipes();

  Telegram.WebApp.onEvent('backButtonClicked', () => {
    tabsStore.setPreviousTab();
  })
})()
</script>

<template>
  <main>
    <RouterView />
    <Toast/>
  </main>
</template>

<style lang="scss">
@use './assets/styles/_main.scss' as *;
</style>
