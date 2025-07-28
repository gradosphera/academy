<template>
  <Transition name="toast">
    <div class="toaster__wrapper">
      <TransitionGroup name="toast" tag="ul">
        <li
            v-for="toast in toastStore.toasts"
            :class="['toaster__inner', toast.status]"
            :key="toast.text"
        >
          <component class="toaster__inner-icon" :is="toastIconMap[toast.status]"/>

          <span class="toaster__inner-text">
              {{ toast.text }}
            </span>
        </li>
      </TransitionGroup>
    </div>
  </Transition>
</template>
<script setup>
import SVGSuccessIcon from "./svg/SVGSuccessIcon.vue";
import SVGWarningWhite from "./svg/SVGWarningWhite.vue";
import {useToastStore} from "../store/toastStore.js";

const toastStore = useToastStore();

const toastIconMap = {
  error: SVGWarningWhite,
  success: SVGSuccessIcon,
};
</script>

<style scoped lang="scss">
.toast-enter-from,
.toast-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

.toast-enter-active,
.toast-leave-active {
  transition: 0.25s ease all;
}

.toaster {
  &__wrapper {
    position: fixed;
    bottom: 100px;
    right: 0;
    width: 100%;
    z-index: 15000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 0 24px;
  }

  &__inner {
    height: 48px;
    color: #000;
    display: flex;
    align-items: center;
    gap: 1rem;
    border-radius: 12px;
    border: 1px solid transparent;
    background: #2B2B2B;
    padding: 12px;

    &.success {
      background: #2CC069;
      color: #FFF;
    }

    &.error {
      background: #EB5151;
      color: #FFF;
    }

    &-icon {
      width: 1.8rem;
      aspect-ratio: 1/1;
    }

    &-text {
      font-size: 15px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }
  }
}
</style>