<script setup>
import Author from "./Author.vue"
import {useMiniAppStore} from "../store/miniAppStore.js";

const miniAppStore = useMiniAppStore();
</script>

<template>
  <div class="bottom-content" :style="{background: miniAppStore.accentedColor}">
    <Author />
  </div>
</template>

<style scoped lang="scss">
@import "../assets/styles/main.scss";
.bottom-content {
  position: fixed;
  z-index: 0;
  top: 0;
  left: 0;
  width: 100dvw;
  max-height: 100dvh;
  height: 100%;
  padding: 16px 24px 45px;
  background: linear-gradient(180deg, #7581EC 0%, #464C88 67.11%);
}
</style>
