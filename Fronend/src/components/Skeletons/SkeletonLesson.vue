<template>
  <div class="lesson_wrapper">
    <div class="lesson">
      <div class="lesson_image-container">
        <AnimatedSkeleton width="120px" height="71px" border-radius="12px"/>
      </div>
      <div :class="['lesson_title', ]">
        <AnimatedSkeleton width="120px" height="15px" border-radius="6px"/>
        <AnimatedSkeleton width="80px" height="15px" border-radius="6px"/>
      </div>
    </div>
  </div>
</template>
<script setup>
import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";
</script>
<style scoped lang="scss">
.lesson {
  width: 100%;
  cursor: pointer;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.07);
  backdrop-filter: blur(12px);
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 8px;

  &_wrapper {
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  &_title {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  &_image-container {
    position: relative;
    border-radius: 12px;
    overflow: hidden;
    width: 125px;
    height: 71px;
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 120px;
    max-width: 120px;
  }
}
</style>