<template>
<div class="selected-lesson__skeleton">
  <div class="selected-lesson__skeleton_nav">
    <AnimatedSkeleton width="50%" height="26px" border-radius="6px"/>
  </div>
  <div class="selected-lesson__skeleton_cover">
    <AnimatedSkeleton width="100%" height="190px" border-radius="12px"/>
  </div>
  <div class="selected-lesson__skeleton_title">
    <AnimatedSkeleton width="50%" height="28px" border-radius="12px"/>
  </div>
  <div class="selected-lesson__skeleton_description">
    <AnimatedSkeleton width="100%" height="28px" border-radius="12px"/>
    <AnimatedSkeleton width="80%" height="28px" border-radius="12px"/>
    <AnimatedSkeleton width="60%" height="28px" border-radius="12px"/>
  </div>
</div>
</template>
<script setup>
import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";
</script>
<style scoped lang="scss">
.selected-lesson {
  &__skeleton {
    &_nav {
      border-radius: 8px;
      background: #2C2C2C;
      padding: 2px;
      margin-bottom: 24px;
    }

    &_cover {
      margin-bottom: 16px;
    }

    &_title {
      margin-bottom: 18px;
    }

    &_description {
      display: flex;
      flex-direction: column;
      gap: 8px;
    }
  }
}
</style>