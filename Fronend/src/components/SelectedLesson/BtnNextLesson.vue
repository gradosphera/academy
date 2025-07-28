<template>
  <div class="btns">
    <div class="btns_inner">
      <button v-if="isPrevButtonVisible" type="button" @click="handlePrevClick" class="btns_prev-lesson-btn">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="25"
            viewBox="0 0 24 25"
            fill="none">
          <path
              d="M18.2714 16.7773C18.0525 16.7773 17.8636 16.7128 17.6497 16.5837L12.5521 13.5942C12.2587 13.4204 12.0747 13.2168 12.01 12.9735V15.8388C12.01 16.4446 11.642 16.7773 11.2044 16.7773C10.9856 16.7773 10.8016 16.7128 10.5827 16.5837L5.48516 13.5942C5.11216 13.3757 4.92318 13.1076 4.92318 12.7749C4.92318 12.4372 5.10719 12.1839 5.48516 11.9605L10.5827 8.97101C10.8016 8.8419 10.9905 8.77734 11.2094 8.77734C11.647 8.77734 12.01 9.11006 12.01 9.71589V12.5713C12.0747 12.3329 12.2537 12.1343 12.5521 11.9605L17.6497 8.97101C17.8636 8.8419 18.0525 8.77734 18.2714 8.77734C18.709 8.77734 19.077 9.11006 19.077 9.71589V15.8388C19.077 16.4446 18.709 16.7773 18.2714 16.7773Z"
              fill="#0A0A0A"
          />
        </svg>
        <span>{{t('general.buttons.previous')}}</span>
      </button>
      <div class="btns_divider"
           v-if="isPrevButtonVisible && isNextBtnVisible"/>
      <button
          v-if="isNextBtnVisible"
          type="button"
          @click="handleNextClick"
          class="btns_next-lesson-btn"
      >
        <span>{{ nextBtnText }}</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="25" viewBox="0 0 24 25" fill="none">
          <path
              d="M5.72864 16.7773C5.94746 16.7773 6.13644 16.7128 6.35029 16.5837L11.4479 13.5942C11.7413 13.4204 11.9253 13.2168 11.99 12.9735V15.8388C11.99 16.4446 12.358 16.7773 12.7956 16.7773C13.0144 16.7773 13.1984 16.7128 13.4173 16.5837L18.5148 13.5942C18.8878 13.3757 19.0768 13.1076 19.0768 12.7749C19.0768 12.4372 18.8928 12.1839 18.5148 11.9605L13.4173 8.97101C13.1984 8.8419 13.0095 8.77734 12.7906 8.77734C12.353 8.77734 11.99 9.11006 11.99 9.71589V12.5713C11.9253 12.3329 11.7463 12.1343 11.4479 11.9605L6.35029 8.97101C6.13644 8.8419 5.94746 8.77734 5.72864 8.77734C5.29099 8.77734 4.92297 9.11006 4.92297 9.71589V15.8388C4.92297 16.4446 5.29099 16.7773 5.72864 16.7773Z"
              fill="#0A0A0A"
          />
        </svg>
      </button>
    </div>
  </div>
</template>
<script setup>
import {useSelectedLesson} from "../../store/selectedLesson.js"
import {useSelectedCourse} from "../../store/selectedCourse.js"
import {computed} from "vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();
const selectedCourseStore = useSelectedCourse();
const selectedLessonStore = useSelectedLesson();
const tabs = selectedLessonStore.selectedLesson?.existedTabs || [];
const currentTabIndex = tabs.findIndex(item => item === selectedLessonStore.lessonActiveTab);

const handleNextClick = () => {
  if (tabs.length > 1) {
    if (currentTabIndex !== tabs.length - 1) {
      selectedLessonStore.updateLessonActiveTab(tabs[currentTabIndex + 1]);

      return;
    }
  }

  selectedLessonStore.toNextLesson();
  selectedLessonStore.updateLessonActiveTab(1);
}

const isNextBtnVisible = computed(() => {
  let isStreamCompleted = true;
  const isLastLesson = selectedLessonStore.selectedLesson.lessonId !== selectedCourseStore.selectedCourse.lessonsAmount;

  if (selectedLessonStore.selectedLesson?.lessonType === 'stream') {
    const progress = selectedCourseStore.courseProgress?.["1"] || [];

    isStreamCompleted = progress.find(lesson => lesson.lesson_id === selectedLessonStore.selectedLesson.lessonId);
  }

  return isStreamCompleted && isLastLesson;
})

const nextBtnText = computed(() => {
  if (tabs.length === 1) return t("general.buttons.next");

  if (currentTabIndex === tabs.length - 1) return t("general.buttons.nextLesson");

  return t("general.buttons.next");
})

const handlePrevClick = () => {
  if (selectedLessonStore.lessonActiveTab === 1) {
    selectedLessonStore.toPreviousLesson();
    selectedLessonStore.updateLessonActiveTab(1);

    return;
  }

  selectedLessonStore.updateLessonActiveTab(1);
}

const isPrevButtonVisible = computed(() => {
  const isCurrentlyFirstLesson = selectedLessonStore.selectedLesson?.lessonId === 1;
  const lastLesson = selectedLessonStore.allLessons?.[selectedLessonStore.allLessons?.length - 1];
  const isCurrentlyLastLesson = selectedLessonStore.selectedLesson?.lessonId === lastLesson?.lessonId;

  if (selectedLessonStore.lessonActiveTab === 1 && isCurrentlyFirstLesson) return false;

  if (currentTabIndex === tabs.length - 1 && isCurrentlyFirstLesson) return false;

  if (isCurrentlyLastLesson || tabs.length === 1) return true;

  return currentTabIndex !== tabs.length - 1;
})

</script>
<style scoped lang="scss">
.btns {
  cursor: pointer;
  position: relative;
  z-index: 10;
  height: 48px;
  padding: 0 24px;

  &_inner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 100%;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.12);
    backdrop-filter: blur(6px);
  }

  &_prev-lesson-btn {
    height: 100%;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    background: unset;

    span {
      color: #FFFFFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }

    & svg {
      ::v-deep(path) {
        fill: #FFF;
      }
    }
  }

  &_divider {
    height: 24px;
    width: 1px;
    background-color: rgba(255, 255, 255, 0.50);
  }

  &_next-lesson-btn {
    height: 100%;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    background: unset;

    span {
      color: #FFFFFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }

    & svg {
      ::v-deep(path) {
        fill: #FFF;
      }
    }
  }
}
</style>
