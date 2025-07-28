<template>
  <div class="podcast-lesson">
    <div class="podcast-lesson_content">
      <div class="podcast-lesson_image-container">
        <img class="podcast-lesson_image" :src="getFile(lessonContentData?.cover)" alt="" />
      </div>
      <div>
        <div class="podcast-lesson_title-wrapper">
          <div class="podcast-lesson_title">
            <span v-for="(line, index) in parseMessage(lessonContentData?.title)" :key="index">
              {{ line }}
              <br v-if="index < parseMessage(lessonContentData?.title).length - 1"/>
            </span>
          </div>
          <div class="podcast-lesson_add-to-wishlist-container">
            <button type="button" @click="favouritesStore.toggleFavorite(selectedLesson)"
                    class="podcast-lesson_add-to-wishlist">
              <SVGFavouritesAdded v-if="favouritesStore.isFavorite(selectedLesson?.id)"/>
              <SVGFavouritesEmpty v-else/>
            </button>
          </div>
        </div>
        <div ref="description" class="podcast-lesson_description" :class="{ expanded: isExpanded }" :style="{height: descriptionHeight}">
            <span v-for="(line, index) in parseMessage(lessonContentData?.description)" :key="index">
              {{ line }}
              <br v-if="index < parseMessage(lessonContentData?.description).length - 1"/>
            </span>
        </div>
        <button v-if="showReadMore" class="podcast-lesson_description-read-more-btn" type="button" @click="toggleReadMore">
          {{ isExpanded ? t('general.buttons.lessText') : t('general.buttons.moreText') }}
        </button>
        <AudioPlayer class="podcast-lesson_audio-player" @submit="submitLessonProgress" :file="getFile(lessonContentData?.filename)" :key="lessonContentData?.filename" />
      </div>
    </div>
  </div>
</template>
<script setup>
import { storeToRefs } from "pinia"
import AudioPlayer from "../AudioPlayer.vue"
import {nextTick, ref, watch} from "vue"
import {useProductsStore} from "../../../store/productsStore.js";
import {getFile, getProductData, parseMessage, submitAnalyticsData} from "../../../helpers/index.js";
import {useI18n} from "vue-i18n";
import {submitLesson} from "../../../api/api.js";
import SVGFavouritesAdded from "../../svg/SVGFavouritesAdded.vue";
import SVGFavouritesEmpty from "../../svg/SVGFavouritesEmpty.vue";
import {useFavoriteLessons} from "../../../store/favouritesStore.js";

const favouritesStore = useFavoriteLessons();
const {t} = useI18n();
const productStore = useProductsStore();
const { lessonContentData, selectedLesson } = storeToRefs(productStore);
const description = ref(null)
const isExpanded = ref(false)
const showReadMore = ref(false)
const descriptionHeight = ref('auto');
const startDescriptionHeight = ref(0);

function toggleReadMore() {
  isExpanded.value = !isExpanded.value

  if (isExpanded.value) {
    descriptionHeight.value = description.value.scrollHeight + 'px';
  } else {
    descriptionHeight.value = startDescriptionHeight.value + 'px';
  }
}

const submitLessonProgress = async() => {
  submitAnalyticsData('lesson_media_play', {
    lesson_title: selectedLesson.value?.title,
    course_title: productStore.selectedProduct?.title || '',
    media_type: 'audio',
  })

  if (!productStore.isLessonCompleted) {
    const data = new FormData();

    await submitLesson(productStore.selectedLesson?.id, data);
    await getProductData(productStore.selectedProduct?.id);
  }
}

watch(
    () => lessonContentData.value?.description,
    async (desc) => {
      if (!desc) return;
      await nextTick();
      setTimeout(() => {
        if (description.value) {
          const el = description.value;
          showReadMore.value = el.scrollHeight > el.clientHeight;
          descriptionHeight.value = el.clientHeight + 'px';
          startDescriptionHeight.value = el.clientHeight;
        }
      }, 200);
    },
    { immediate: true }
);
</script>
<style scoped lang="scss">
.podcast-lesson {
  color: var(--White, var(--system-white, #fff));

  &_bg-image {
    background: var(--Miscellaneous-Bar-border, rgba(0, 0, 0, 0.7));
    height: 100%;
    position: absolute;
    left: 0;
    top: 0;
    opacity: 0.3;
    filter: blur(22px);
    z-index: 0;

    img {
      height: 100%;
      object-fit: cover;
      object-position: 50% 50%;
    }
  }

  &_content {
    position: relative;
    z-index: 2;
  }

  &_image-container {
    position: relative;
    margin-bottom: 30px;
  }

  &_image {
    border-radius: 16px;
    max-width: 100%;
  }

  &_title {
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
    cursor: pointer;
    display: flex;
    justify-content: center;
  }

  &_description {
    margin-bottom: 12px;
    white-space: pre-wrap;
    display: -webkit-box;
    -webkit-line-clamp: 5;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: all .3s ease-in-out;

    & span {
      color: #fff;
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

  &_description-read-more-btn {
    margin-bottom: 24px;
    background-color: unset;
    color: #fff;
    font-size: 16px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &_audio-player {
    margin-bottom: 24px;
  }

  &_next-lesson-btn {
    cursor: pointer;
    width: 100%;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    border-radius: 12px;
    background: var(--12-white, rgba(255, 255, 255, 0.12));

    span {
      color: var(--White, #fff);
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
      /* 20.8px */
    }
  }
}

.v-enter-active,
.v-leave-active {
  transition: all 1s ease;
}

.v-enter-from,
.v-leave-to {
  opacity: 0;
}
</style>
