<template>
  <teleport to="#modals">
    <div v-if="isOpen" class="modal" :class="position">
      <div v-if="position === 'center'" class="modal__content">
        <div class="modal__top" :class="titlePosition">
          <span :style="{color: titleColor}">{{ title }}</span>
          <SVGClose :style="{cursor: 'pointer'}" @click="emits('closeModal')"/>
        </div>
        <slot></slot>
      </div>
      <div v-if="position === 'bottom'" class="modal__bottom">
        <slot name="bottom"></slot>
      </div>
    </div>
  </teleport>
</template>
<script setup>
import SVGClose from "../../components/svg/SVGClose.vue";

const props = defineProps({
  isOpen: Boolean,
  title: String,
  position: {type: String, default: 'center'},
  titlePosition: {type: String, default: 'left'},
  titleColor: {type: String, default: '#FFF'},
});

const emits = defineEmits(['closeModal']);
</script>
<style scoped lang="scss">
@import "../../assets/styles/main.scss";

.modal {
  position: fixed;
  top: 0;
  left: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100vw;
  min-height: 100vh;
  height: 100%;
  background: rgba(0, 0, 0, 0.50);
  backdrop-filter: blur(5px);
  padding: 0 24px;
  z-index: 10000;

  &.bottom {
    align-items: flex-end;
    padding: 0;

    & .modal {
      &__content {
        border-bottom-left-radius: 0;
        border-bottom-right-radius: 0;
      }
    }
  }

  &__content {
    width: 100%;
    padding: 24px 16px;
    border-radius: 16px;
    background: #181818;
    display: flex;
    flex-direction: column;

    color: $white;
    font-size: 17px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
  }

  &__bottom {
    width: 100%;
    background-color: #181818;
  }

  &__top {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    padding-right: 24px;

    & svg {
      position: absolute;
      top: 0;
      right: 0;
    }

    &.left {
      justify-content: flex-start;
    }

    &.center {
      justify-content: center;
    }
  }
}


</style>