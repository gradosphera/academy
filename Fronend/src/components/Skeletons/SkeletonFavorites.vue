<template>
  <div class="favorite-item">
    <div class="favorite-item_image">
      <AnimatedSkeleton height="160px" border-radius="12px" />
    </div>
    <div class="favorite-item_bottom">
      <div class="favorite-item_title">
        <AnimatedSkeleton width="90px" height="22px" border-radius="12px" />
      </div>
    </div>
  </div>
</template>
<script setup>
import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";
</script>
<style scoped lang="scss">
.favorite-item {
  cursor: pointer;
  color: white;
  width: 100%;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  padding: 16px;
  flex-direction: column;
  gap: 12px;

  &_image {
    position: relative;
  }

  &_bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
}
</style>