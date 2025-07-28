<template>
  <div v-if="isPrevLessonAvailable.is_btn_visible || isNextLessonAvailable.is_btn_visible" class="lesson-content__btns_inner">
    <button v-if="isPrevLessonAvailable.is_btn_visible" @click="handleClick('back')">
      <SVGArrowLeft />
      {{t('general.buttons.previous')}}
    </button>
    <div v-if="isPrevLessonAvailable.is_btn_visible && isNextLessonAvailable.is_btn_visible" class="divider"></div>
    <button v-if="isNextLessonAvailable.is_btn_visible" @click="handleClick('next')">
      {{ t("general.buttons.next") }}
      <SVGArrowLeft class="db-arrow"/>
    </button>
  </div>
</template>
<script setup>
import SVGArrowLeft from "../svg/SVGArrowLeft.vue";
import {computed, ref} from "vue";
import {useProductsStore} from "../../store/productsStore.js";
import {storeToRefs} from "pinia";
import {useI18n} from "vue-i18n";
import {getLesson} from "../../api/api.js";
import {isDatePassed, submitAnalyticsData} from "../../helpers/index.js";

const emits = defineEmits(["openModal", "openPayWall"]);

const {t} = useI18n();
const productStore = useProductsStore();
const isPayWallOpen = ref(false);
const {selectedLesson, selectedLessonTab, lessonTabs, isLessonReviewed, isLessonCompleted} = storeToRefs(productStore);
const firstLesson = productStore.selectedProduct?.lessons?.[0];
const lastLesson = computed(() => {
  const lastModule = productStore.selectedProductModules[productStore.selectedProductModules?.length - 1];

  return lastModule.items[lastModule.items.length - 1] || null;
});
const prevLesson = computed(() => {
  if (selectedLesson.value && selectedLesson.value.id === firstLesson.id) return null;

  const indexOfCurrentLesson = productStore.selectedProduct?.lessons?.findIndex(lesson => lesson.id === selectedLesson.value?.id);

  if (indexOfCurrentLesson !== -1) {
    return productStore.selectedProduct?.lessons?.[indexOfCurrentLesson - 1];
  }

  return null;
})
const nextLesson = computed(() => {
  if (selectedLesson.value && selectedLesson.value.id === lastLesson.value.id) return null;

  const indexOfCurrentLesson = productStore.selectedProduct?.lessons?.findIndex(lesson => lesson.id === selectedLesson.value?.id);

  if (indexOfCurrentLesson !== -1) {
    return productStore.selectedProduct?.lessons?.[indexOfCurrentLesson + 1];
  }

  return null;
})

const getPrevBtnText = computed(() => {
  const tab = productStore.selectedLessonTab;

  return 'Back'
});

const getNextBtnText = computed(() => {
  const tab = productStore.selectedLessonTab;

  return 'Next'
});

const isNextLessonAvailable = computed(() => {
  const tabs = lessonTabs.value;
  const currentIndex = tabs.indexOf(selectedLessonTab.value);
  const result = {is_btn_visible: false, info: ''};

  if (!nextLesson.value) {
    result.is_btn_visible = !(currentIndex === tabs.length - 1);
    return result;
  }

  if (currentIndex !== tabs.length - 1) {
    result.is_btn_visible = true;
    return result;
  }

  if (productStore.selectedProduct?.product_levels?.length) {
    function searchNextLessonInLevel() {
      for (const level of productStore.selectedProduct.product_levels) {
        const isLessonFound = level?.product_level_lessons?.find(lesson => lesson.lesson_id === nextLesson.value.id);

        if (isLessonFound) {
          result.is_btn_visible = false;
          return result;
        }
      }

      result.is_btn_visible = true;
      return result;
    }

    const isNextLessonFree = searchNextLessonInLevel();

    if (isNextLessonFree.is_btn_visible) {
      if (!nextLesson.value?.previous_lesson_id?.startsWith('00000000')) {
        result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === nextLesson.value?.previous_lesson_id);
        return result;
      }

      return true;
    } else {
      const isLessonPaid = productStore.paidLessons?.find(el => el.LessonID === nextLesson.value?.id);

      if (!isLessonPaid) {
        result.is_btn_visible = true;
        isPayWallOpen.value = true;
        return result;
      }

      if (!nextLesson.value?.previous_lesson_id?.startsWith('00000000')) {
        result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === nextLesson.value?.previous_lesson_id);
        return result;
      }

      result.is_btn_visible = true;
      return result;
    }

  } else if (!nextLesson.value?.previous_lesson_id?.startsWith('00000000')) {
    result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === nextLesson.value?.previous_lesson_id);
    return result;
  } else if (nextLesson.value?.release_date) {
    if (productStore.selectedProduct?.product_levels?.length) {
      if (nextLesson.value?.product_level_id) {
        const isPaid = productStore.paidLessons?.find(el => el.LessonID === nextLesson.value?.id);

        if (!isPaid) {
          result.is_btn_visible = false;
          return result;
        } else {
          result.is_btn_visible = isDatePassed(nextLesson.value?.release_date);
          return result;
        }
      }
    }

    result.is_btn_visible = isDatePassed(nextLesson.value?.release_date);
    return result;
  } else {
    result.is_btn_visible = true;
    return result;
  }
})

const isPrevLessonAvailable = computed(() => {
  const result = {is_btn_visible: false, info: ''};
  if (!selectedLesson.value) return result.is_btn_visible = false;

  const tabs = lessonTabs.value;
  const currentIndex = tabs.indexOf(selectedLessonTab.value);

  if (!prevLesson.value) {
    result.is_btn_visible = !(currentIndex === 0)
    return result;

  }

  if (productStore.selectedProduct?.product_levels?.length) {
    function searchPreviousLessonInLevel() {
      for (const level of productStore.selectedProduct.product_levels) {
        const isLessonFound = level?.product_level_lessons?.find(lesson => lesson.lesson_id === prevLesson.value.id);

        if (isLessonFound) {
          result.is_btn_visible = false
          return result;
        }
      }

      result.is_btn_visible = true;
      return result;
    }

    const isPreviousLessonFree = searchPreviousLessonInLevel();

    if (isPreviousLessonFree) {
      if (!prevLesson.value?.previous_lesson_id?.startsWith('00000000')) {
        result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === prevLesson.value?.previous_lesson_id);
        return result;
      }

      result.is_btn_visible = true;
      return result;
    } else {
      const isLessonPaid = productStore.paidLessons?.find(el => el.LessonID === prevLesson.value?.id);

      if (!isLessonPaid) {
        result.is_btn_visible = true;
        isPayWallOpen.value = true;
        return result;
      }

      if (!nextLesson.value?.previous_lesson_id?.startsWith('00000000')) {
        result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === prevLesson.value?.previous_lesson_id);
        return result;
      }

      result.is_btn_visible = true;
      return result;
    }
  } else if (!prevLesson.value?.previous_lesson_id?.startsWith('00000000')) {
    result.is_btn_visible = productStore.productProgress.find(el => el.lesson_id === prevLesson.value?.previous_lesson_id);
    return result;
  } else if (prevLesson.value?.release_date) {
    if (productStore.selectedProduct?.product_levels?.length) {
      if (prevLesson.value?.product_level_id) {
        const isPaid = productStore.paidLessons?.find(el => el.LessonID === prevLesson.value?.id);

        if (!isPaid) {
          result.is_btn_visible = false;
          return result;
        } else {
          result.is_btn_visible = isDatePassed(prevLesson.value?.release_date);
          return result;
        }
      }
    }

    result.is_btn_visible = isDatePassed(prevLesson.value?.release_date);
    return result;
  } else {
    result.is_btn_visible = true;
    return result;
  }
})

const handleClick = async (direction) => {
  const tabs = lessonTabs.value;
  const currentIndex = tabs.indexOf(selectedLessonTab.value);
  if (currentIndex === -1) return;

  const isFirst = currentIndex === 0;
  const isLast = currentIndex === tabs.length - 1;

  const goToLesson = async (el) => {
    if (!el) return;
    const resp = await getLesson(el.id);
    if (resp.data) {
      isLessonCompleted.value = false;
      isLessonReviewed.value = false;
      productStore.setSelectedLesson(resp.data.lesson);
      productStore.setLessonTab('content');
    }
  };

  if (isLast && direction === 'next') {
    if (!isNextLessonAvailable.value.is_btn_visible) return;

    if (isPayWallOpen.value) {
      emits('openPayWall')

      submitAnalyticsData('paywall_shown', {
        source: 'next_lesson_button',
        course_title: productStore.selectedProduct?.title || '',
      })

      return;
    }

    if (isLessonCompleted.value && !isLessonReviewed.value) {
      emits('openModal');
      return;
    }

    await goToLesson(nextLesson.value);
    submitAnalyticsData('next_lesson_click', {
      current_lesson_title: productStore.selectedLesson?.title || '',
      course_title: productStore.selectedProduct?.title || '',
      next_lesson_title: nextLesson.value?.title,
    })
  } else if (isFirst && direction === 'prev') {
    await goToLesson(prevLesson.value);
  } else if ((isLast && direction === 'prev') || (isFirst && direction === 'next')) {
    const newIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1;
    if (tabs[newIndex]) productStore.setLessonTab(tabs[newIndex]);
  } else {
    const newIndex = direction === 'next' ? currentIndex + 1 : currentIndex - 1;
    if (tabs[newIndex]) {
      productStore.setLessonTab(tabs[newIndex]);
    } else {
      const lessonRef = direction === 'next' ? nextLesson.value : prevLesson.value;
      await goToLesson(lessonRef);
    }
  }
};
</script>
<style scoped lang="scss">
.lesson-content {
  &__btns {
    &_inner {
      height: 100%;
      display: flex;
      align-items: center;
      overflow: hidden;
      border-radius: 12px;
      background: #2B2B2B;

      & button {
        height: 100%;
        width: 100%;
        background: transparent;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;

        color: #FFF;
        font-size: 16px;
        font-style: normal;
        font-weight: 500;
        line-height: 130%;

        & .db-arrow {
          transform: rotate(180deg);
        }
      }

      & .divider {
        height: 60%;
        width: 1px;
        background: rgba(255, 255, 255, 0.12);
      }
    }
  }
}
</style>