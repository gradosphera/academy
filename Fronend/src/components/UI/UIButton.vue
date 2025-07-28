<template>
  <button
      :disabled="isDisabled"
      :class="['ui-button', {'border': isBorder}]"
      :style="{background: bg, width: width, color, height: height}"
      @click="emits('cta')"
  >
    <component :is="svg"/>
    <slot></slot>
  </button>
</template>
<script setup>
defineProps({
  isDisabled: {type: Boolean, required: false, default: false},
  bg: {type: String, required: false, default: 'transparent'},
  isBorder: {type: Boolean, required: false, default: false},
  width: {type: String, required: false, default: '100%'},
  svg: {type: Object, required: false, default: null},
  color: {type: String, required: false, default: '#fff'},
  height: {type: String, required: false, default: '45px'},
})

const emits = defineEmits(['cta']);
</script>
<style scoped lang="scss">
@import "../../assets/styles/main.scss";

.ui-button {
  height: 45px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 12px;
  color: $white;
  font-size: 16px;
  font-style: normal;
  font-weight: 500;
  line-height: 130%;
  gap: 8px;

  &:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  &.border {
    border: 1px solid rgba(255, 255, 255, 0.17);
  }
}
</style>