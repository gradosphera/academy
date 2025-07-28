<template>
  <div class="text-lesson">
    <div class="text-lesson_image-container">
      <img v-if="lessonData?.cover?.length" class="text-lesson_image" :src="getFile(lessonData?.cover)" alt=""/>
    </div>
    <div class="text-lesson_title-wrapper">
      <div class="text-lesson_title">
        <span v-for="(line, index) in parseMessage(lessonData?.title)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(lessonData?.title).length - 1"/>
        </span>
      </div>
      <div class="text-lesson_add-to-wishlist-container">
        <button type="button" @click="favouritesStore.toggleFavorite(productStore.selectedLesson)"
                class="text-lesson_add-to-wishlist">
          <SVGFavouritesEmpty v-if="!favouritesStore.isFavorite(productStore.selectedLesson?.id)"/>
          <SVGFavouritesAdded v-else/>
        </button>
      </div>
    </div>
    <div ref="textDescriptionRef" class="text-lesson_text-content text-content description">
      <div ref="descriptionText" class="description_text" :class="{ expanded: isExpanded }" :style="{height: descriptionHeight}">
        <span v-for="(line, index) in parseMessage(lessonData?.description)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(lessonData?.description).length - 1"/>
        </span>
      </div>
      <button v-if="showReadMore" class="description_read-more-btn" type="button" @click="toggleReadMore">
        {{ isExpanded ? t('general.buttons.lessText') : t('general.buttons.moreText') }}
      </button>
    </div>
  </div>
</template>
<script setup>
import {useFavoriteLessons} from "../../../store/favouritesStore.js"
import SVGFavouritesEmpty from "../../svg/SVGFavouritesEmpty.vue";
import SVGFavouritesAdded from "../../svg/SVGFavouritesAdded.vue";
import {computed, nextTick, onMounted, ref, watch} from "vue";
import {useProductsStore} from "../../../store/productsStore.js";
import {getFile, getProductData, parseMessage} from "../../../helpers/index.js";
import {useI18n} from "vue-i18n";
import {submitLesson} from "../../../api/api.js";

const {t} = useI18n();
const productStore = useProductsStore();
const lessonData = computed(() => {
  return productStore.lessonContentData;
})
const showReadMore = ref(false)
const descriptionText = ref(null)
const favouritesStore = useFavoriteLessons();
const isExpanded = ref(false);
const descriptionHeight = ref('auto');
const startDescriptionHeight = ref(0);

function toggleReadMore() {
  isExpanded.value = !isExpanded.value
  if (isExpanded.value) {
    descriptionHeight.value = descriptionText.value.scrollHeight + 'px';
  } else {
    descriptionHeight.value = startDescriptionHeight.value + 'px';
  }
}

watch(
    () => lessonData.value?.description,
    async (desc) => {
      if (!desc) return;
      await nextTick();
      setTimeout(() => {
        if (descriptionText.value) {
          const el = descriptionText.value;
          showReadMore.value = el.scrollHeight > el.clientHeight;
          descriptionHeight.value = el.clientHeight + 'px';
          startDescriptionHeight.value = el.clientHeight;
        }
      }, 200);
    },
    { immediate: true }
);

onMounted(async () => {
  if (!productStore.isLessonCompleted) {
    const data = new FormData();

    await submitLesson(productStore.selectedLesson?.id, data);

    await getProductData(productStore.selectedProduct?.id);
  }
})
</script>
<style scoped lang="scss">
.text-lesson {
  padding-bottom: 60px;
  color: var(--White, var(--system-white, #fff));

  &_image-container {
    position: relative;
    margin-bottom: 16px;
  }

  &_image {
    border-radius: 16px;
    max-width: 100%;
  }

  &_controls {
    width: 100%;
    position: fixed;
    bottom: 104px;
    left: 50%;
    transform: translateX(-50%);
  }

  .text-content {
    margin-bottom: 24px;
    padding: 16px;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);

    & .description_text {
      display: -webkit-box;
      -webkit-line-clamp: 5;
      -webkit-box-orient: vertical;
      overflow: hidden;
      text-overflow: ellipsis;
      transition: all .3s ease-in-out;


      & span {
        font-size: 16px;
        font-style: normal;
        font-weight: 400;
        line-height: 130%;
        opacity: 0.8;
      }

      &.expanded {
        -webkit-line-clamp: unset;
        display: block;
      }
    }
  }

  &_title {
    margin-bottom: 16px;
    & span {
      font-size: 20px;
      font-style: normal;
      font-weight: 700;
      line-height: 125%;
    }

    &-wrapper {
      margin-bottom: 16px;

      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  &_add-to-wishlist {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.04);

    & svg {
      width: 60%;
    }
  }

  &_add-to-wishlist-container {
    margin-bottom: 24px;
    cursor: pointer;
    display: flex;
    justify-content: center;
  }
}

.description {
  &_read-more-btn {
    background-color: unset;
    color: var(--White, var(--system-white, #fff));
    font-size: 16px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }
}
</style>
