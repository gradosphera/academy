<template>
  <div class="skeleton">
    &nbsp;
  </div>
</template>
<script setup>
defineProps({
  height: {type: String},
  width: {type: String},
  borderRadius: {type: String}
})
</script>
<style scoped lang="scss">
.skeleton {
  height: v-bind(height);
  width: v-bind(width);
  border-radius: v-bind(borderRadius);
  background-image: linear-gradient(
          to right,
          rgba(214, 215, 216, 0.2) 0%,
          rgba(226, 227, 228, 0.3) 10%,
          rgba(214, 215, 216, 0.2) 20%,
          rgba(214, 215, 216, 0.2) 100%
  );
  background-size: 200% 100%;
  animation: bgAnimation 2s infinite linear;
}

@keyframes bgAnimation {
  0% {
    background-position: 50% 0;
  }

  100% {
    background-position: -150% 0;
  }
}
</style>