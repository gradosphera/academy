<script setup>
import BgBlurSquare from "./BgBlurSquare.vue"
</script>

<template>
  <div class="page-layout">
    <div class="page-layout_container">
      <slot />
      <BgBlurSquare />
    </div>
  </div>
</template>

<style scoped lang="scss">
.page-layout {
  margin-top: 72px;
  height: calc(100vh - 60px);
  position: relative;
  background: linear-gradient(0deg, rgba(10,10,10, 1) 80%, rgba(36,36,36,1) 190%);
  color: var(--White, var(--system-white, #fff));
  border-radius: 24px 24px 0 0;
  padding: 24px 24px 102px;

  &_container {
    height: 100%;
    overflow-y: auto;
  }
}
</style>
