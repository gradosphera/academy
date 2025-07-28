<template>
  <div ref="productHeadContainerRef" class="selected-course_head head">
    <div class="head_image">
      <AnimatedSkeleton v-if="isMediaLoading" height="100px" />
      <img v-show="!isMediaLoading" v-if="productStore.selectedProduct?.cover?.length" :src="getFile(productStore.selectedProduct?.cover)" alt="" />
      <div class="head_progress-bar-container">
        <div class="head_progress-bar" :style="{background: miniAppStore.accentedColor}" />
        <div class="head_progress" :style="{ width: progress() + '%', background: miniAppStore.accentedColor }" />
      </div>
    </div>
    <div class="head_texts">
      <div class="head_title">
        <span v-for="(line, index) in parseMessage(productStore.selectedProduct?.title)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(productStore.selectedProduct?.title).length - 1"/>
        </span>
      </div>
      <div class="head_description">
        <span v-for="(line, index) in parseMessage(productStore.selectedProduct?.description)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(productStore.selectedProduct?.description).length - 1"/>
        </span>
      </div>
    </div>
  </div>
</template>
<script setup>
import {useProductsStore} from "../../store/productsStore.js";
import {getFile, parseMessage} from "../../helpers/index.js";
import {storeToRefs} from "pinia";
import {useMiniAppStore} from "../../store/miniAppStore.js";
import {useMediaLoader} from "../../composable/useMediaLoader.js";
import {onMounted, ref} from "vue";
import AnimatedSkeleton from "../UI/AnimatedSkeleton.vue";

defineProps({
  locale: Object,
})
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();
const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();
const productHeadContainerRef = ref(null);
const {productProgress, selectedProduct} = storeToRefs(productStore);
const progress = () => {
  const step = 100 / selectedProduct.value?.lessons?.length || 0;

  return step * productProgress.value?.length;
}

onMounted(async() => {
  await waitForMediaLoad(productHeadContainerRef.value)
})
</script>
<style scoped lang="scss">
.selected-course {
  .head {
    &_image {
      position: relative;
      height: 100px;
      overflow: hidden;
      background: rgba(255, 255, 255, 0.12);
      img {
        height: 100%;
        width: 100%;
        object-fit: cover;
      }
    }
    &_progress-bar-container {
      position: absolute;
      left: 0;
      bottom: 0;
      height: 10px;
      width: 100%;
    }

    &_progress-bar {
      position: absolute;
      left: 0;
      top: 0;
      height: 100%;
      width: 100%;
      opacity: 0.5;
      background: linear-gradient(180deg, #7581EC 0%, #464D88 67.11%);
    }

    &_progress {
      height: 100%;
      border-radius: 0 99999px 99999px 0;
      background: linear-gradient(180deg, #7581EC 0%, #464D88 67.11%);
      transition: width 0.4s ease;
    }

    &_texts {
      padding: 31px 24px 38px;
      display: flex;
      flex-direction: column;
      gap: 16px;
    }

    &_title {
      & span {
        color: #fff;
        font-size: 22px;
        font-style: normal;
        font-weight: 700;
        line-height: 21px;
      }
    }

    &_description {
      & span {
        opacity: 0.8;
        color: #fff;
        font-size: 16px;
        font-style: normal;
        font-weight: 400;
        line-height: 130%;
        letter-spacing: 0.16px;
      }
    }
  }
}
</style>
