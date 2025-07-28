<template>
  <div class="onboarding-skeleton">
    <div class="onboarding-skeleton__image">
      <AnimatedSkeleton height="100%" width="100%" border-radius="12px" />
    </div>
  </div>
</template>
<script setup>

import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";
</script>

<style scoped lang="scss">
.onboarding-skeleton {
  width: 100%;
  height: 100%;
  display: flex;

  &__image {
    width: 100%;
    height: 100%;
    display: flex;
  }
}
</style>