<template>
  <div class="pdf-viewer__container">
    <div class="pdf-viewer__inner">
      <div class="pdf-viewer">
        <VuePDF ref="vuePDFRef" :rotation="rotation" :pdf="pdf" :page="page" fit-parent/>
      </div>
    </div>
    <div class="pdf-viewer__controls">
      <button class="pdf-viewer__page_slide" @click="handlePrevPage" :disabled="page === 1">
        <SVGArrowLeft />
      </button>
      <button class="pdf-viewer__rotate" @click="rotatePDF">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="19" viewBox="0 0 18 19" fill="none">
          <path d="M5.5175 5.28814L0 10.7881L5.52602 16.2797L11.052 10.7881L5.5175 5.28814ZM2.40965 10.7881L5.52602 7.68644L8.63387 10.7881L5.5175 13.8898L2.40965 10.7881ZM15.7521 5.48305C14.2621 3.99153 12.2952 3.24576 10.3368 3.24576V0.5L6.72659 4.09322L10.3368 7.68644V4.94068C11.8609 4.94068 13.3851 5.51695 14.5516 6.67797C16.8761 8.99153 16.8761 12.7542 14.5516 15.0678C13.3851 16.2288 11.8609 16.8051 10.3368 16.8051C9.51088 16.8051 8.68496 16.6271 7.91864 16.2881L6.64995 17.5508C7.7824 18.178 9.0596 18.5 10.3368 18.5C12.2952 18.5 14.2621 17.7542 15.7521 16.2627C18.7493 13.2881 18.7493 8.45763 15.7521 5.48305Z" fill="white"/>
        </svg>
      </button>
      <button class="pdf-viewer__page_slide right" @click="handleNextPage" :disabled="page === pages">
        <SVGArrowLeft />
      </button>
    </div>
  </div>
</template>
<script setup>
import {ref} from 'vue';
import { VuePDF, usePDF } from '@tato30/vue-pdf'
import SVGArrowLeft from "../svg/SVGArrowLeft.vue";

const props = defineProps({
  source: {
    type: String,
    required: true,
  },
});
const vuePDFRef = ref(null);
const page = ref(1);
const rotation = ref(0);
const { pdf, pages } = usePDF(props.source || '');
const rotatePDF = () => {
  if (rotation.value === 0) {
    rotation.value += 90;
  } else if (rotation.value === 90) {
    rotation.value -= 90;
  }

  if (vuePDFRef.value) {
    vuePDFRef.value.reload();
  }
}

const handleNextPage = () => {
  page.value = page.value < pages.value ? page.value + 1 : page.value
}

const handlePrevPage = () => {
  page.value = page.value > 1 ? page.value - 1 : page.value
}
</script>
<style scoped lang="scss">
.pdf-viewer {
  &__container {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  &__controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-top: auto;

    & button {
      background: transparent;
    }
  }

  &__page {
    &_slide {
      & svg {
        width: 24px;
        height: 24px;
      }

      &.right {
        & svg {
          transform: rotate(180deg);
        }
      }

      &:disabled {
        opacity: .5;
      }
    }
  }

  &__rotate {
    background: transparent;
    display: flex;
    align-items: center;
    color: #FFF;
    font-size: 15px;
    font-style: normal;
    font-weight: 500;
    line-height: 130%;
    gap: 8px;

    & svg {
      width: 20px;
      height: 20px;
    }
  }

  &__inner {
    margin-bottom: 24px;
    overflow-x: scroll;
  }
}
</style>
