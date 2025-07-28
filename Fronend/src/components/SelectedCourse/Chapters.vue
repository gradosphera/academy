<template>
  <div ref="chaptersContainerRef" class="selected-course_chapters">
    <div class="chapter" v-for="chapter in productStore.selectedProductModules" :key="chapter.name">
      <p :class="['chapter_title', {'available' : chapter.id === 1 || isLessonAvailable(chapter.id - 1)}]">{{chapter.name}}</p>
      <div class="chapter_lessons">
        <SkeletonLesson v-if="isMediaLoading" v-for="(lesson, i) in chapter.items" :key="lesson.id" />
        <div v-show="!isMediaLoading" class="lesson_wrapper" v-for="(lesson, i) in chapter.items" :key="lesson.id">
          <div @click="openLesson(lesson)" class="lesson">
            <div v-if="lesson?.release_date && !isDatePassed(lesson?.release_date)" class="lesson__schedule">
              <div class="lesson__schedule_text">{{t('general.main.lesson_scheduled')}}</div>
              <div class="lesson__schedule_date">
                <span v-if="timeLeftByLessonId[lesson.id]?.days !== '0'">{{`${timeLeftByLessonId[lesson.id]?.days} ${t('general.main.lesson_schedule_left.days')}`}}</span>
                <span>{{`${timeLeftByLessonId[lesson.id]?.hours} ${t('general.main.lesson_schedule_left.hours')}`}}</span>
                <span v-if="timeLeftByLessonId[lesson.id]?.days === '0'">{{`${timeLeftByLessonId[lesson.id]?.minutes} ${t('general.main.lesson_schedule_left.minutes')}`}}</span>
              </div>
            </div>
            <div class="lesson__content">
              <div class="lesson_image-container">
                <img v-if="lesson.materials?.[1]?.filename?.length" class="lesson_image" :src="getFile(lesson.materials?.[1]?.filename)" alt=""/>
                <div class="icon">
                  <component v-if="lesson?.content_type?.length" :is="getLessonTypeIcon(lesson)"/>
                </div>
              </div>
              <div class="lesson_title_wrapper" :class="{center: !isVideoCompressing(lesson)}">
                <div v-if="isVideoCompressing(lesson)" class="lesson__video_status">
                  {{t('general.main.lesson_video_compressing')}}
                </div>
                <div :class="['lesson_title', {'available': lesson.lessonId === 1 || isLessonAvailable(lesson.lessonId - 1)}]">
                <span v-for="(line, index) in parseMessage(lesson.materials?.[0]?.title)" :key="index">
                  {{ line }}
                  <br v-if="index < parseMessage(lesson.materials?.[0]?.title).length - 1"/>
                </span>
                </div>
              </div>
              <div class="lesson_right">
                <div class="lesson_arrow-icon">
                  <SVGBlockLesson v-if="!isLessonAvailable(lesson)"/>
                  <SVGCompletedLesson v-else-if="isLessonCompleted(lesson.id)"/>
                  <SVGArrowShortRight v-else/>
                </div>
              </div>
            </div>
          </div>
          <div v-if="!isDividerHidden(lesson)" class="lesson_divider">{{t('general.course.dividerText')}}</div>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { useNavTabsStore } from "../../store/tabsStore.js"
import SVGArrowShortRight from "../svg/SVGArrowShortRight.vue";
import SVGCompletedLesson from "../svg/SVGCompletedLesson.vue";
import {useI18n} from "vue-i18n";
import {useProductsStore} from "../../store/productsStore.js";
import {getFile, isDatePassed, parseMessage, submitAnalyticsData} from "../../helpers/index.js";
import SVGAudio from "../svg/SVGAudio.vue";
import SVGPlayVideo from "../svg/SVGPlayVideo.vue";
import SVGLessonTextType from "../svg/SVGLessonTextType.vue";
import {getLesson} from "../../api/api.js";
import SVGBlockLesson from "../svg/SVGBlockLesson.vue";
import SkeletonLesson from "../Skeletons/SkeletonLesson.vue";
import {onBeforeUnmount, onMounted, reactive, ref} from "vue";
import {useMediaLoader} from "../../composable/useMediaLoader.js";
import {useToastStore} from "../../store/toastStore.js";

const props = defineProps({
  progress: Array,
})

const emits = defineEmits(['closeModal', 'openModal']);
const {t} = useI18n();
const toastStore = useToastStore();
const chaptersContainerRef = ref(null);
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();
const navTabsStore = useNavTabsStore()
const productStore = useProductsStore();
const timeLeftByLessonId = reactive({})
let timerInterval = null


const openLesson = async (lesson) => {
  if (!isLessonAvailable(lesson, 'open')) return;

  navTabsStore.updateActiveTab('selected_lesson')

  try {

    const resp = await getLesson(lesson.id);

    if (resp.data) {
      const lesson = resp.data.lesson;
      productStore.setSelectedLesson(resp.data.lesson);

      submitAnalyticsData('lesson_open', {
        lesson_title: lesson?.title,
        course_title: productStore.selectedProduct?.title || '',
      })
    }
  } catch (e) {
    navTabsStore.setPreviousTab();
    toastStore.error({text: t('general.main.failed_open_lesson')});
  }
}

const getLessonTypeIcon = (lesson) => {
  switch (lesson.content_type) {
    case 'text':
      return SVGLessonTextType;
    case 'audio':
      return SVGAudio;
    case 'video':
      return SVGPlayVideo;
  }
}

const isVideoCompressing = (elem) => {
  if (elem) {
    const videoMaterial = elem.materials?.find(material => material.content_type === 'video' || material.content_type === 'circle_video');

    if (videoMaterial) {
      return videoMaterial.status === 'pending_move_to_mux';
    }
  }

  return false;
}

const isLessonAvailable = (lesson, action='') => {
  if (lesson.lessonType === 'stream') return true;

  if (productStore.selectedProduct?.product_levels?.length) {
    if (productStore.productInvite) {
      return productStore.productInvite.product_id === productStore.selectedProduct?.id;
    }

    if (lesson.product_level_id) {
      const item = productStore.paidLessons?.find(el => el.LessonID === lesson.id);

      if (!item) {
        if (action === 'open') {
          emits('openModal', lesson.product_level_id)
        }
        return false;
      } else {
        if (!lesson?.previous_lesson_id?.startsWith('00000000')) {
          return productStore.productProgress.find(el => el.lesson_id === lesson.previous_lesson_id);
        }

        return true;
      }
    } else {
      return true;
    }

  } else if (!lesson?.previous_lesson_id?.startsWith('00000000')) {
    return productStore.productProgress.find(el => el.lesson_id === lesson.previous_lesson_id);
  } else if (lesson?.release_date) {
    if (productStore.selectedProduct?.product_levels?.length) {
      if (lesson.product_level_id) {
        const isPaid = productStore.paidLessons?.find(el => el.LessonID === lesson.id);

        if (!isPaid) {
          if (action === 'open') {
            emits('openModal', lesson.product_level_id)
          }
          return false;
        } else {
          return isDatePassed(lesson?.release_date);
        }
      }
    }

    return isDatePassed(lesson?.release_date);
  } else {
    return true;
  }
}

const isLessonCompleted = (id) => {
  return productStore.productProgress.find(el => el.lesson_id === id);
}

const isDividerHidden = (lesson) => {
  const nextLesson = productStore.selectedProduct?.lessons?.find(elem => elem.index === lesson?.index + 1);

  if (!nextLesson) return true;
  const isNextLessonOpensInProgress = !nextLesson.previous_lesson_id?.startsWith('00000000');

  if (!isNextLessonOpensInProgress) return true;

  function checkLessonPayment(lessonId) {
    return productStore.paidLessons?.find(el => el.LessonID === lessonId);
  }

  const isCurrentLessonPaid = checkLessonPayment(lesson.id);
  const isNextLessonPaid = checkLessonPayment(nextLesson.id);
  const isCurrentLessonCompleted = isLessonCompleted(lesson.id);
  const isCurrentLessonAvailable = isLessonAvailable(lesson);
  const isNextLessonAvailable = isLessonAvailable(nextLesson);
  let isPreviousLessonCompleted;

  if (lesson.index !== 0) {
    const prevLesson = productStore.selectedProduct?.lessons?.find(elem => elem.index === lesson?.index - 1);

    if (prevLesson) {
      isPreviousLessonCompleted = isLessonCompleted(prevLesson.id);
    }
  }

  if (!isCurrentLessonAvailable) return true;

  if (lesson.product_level_id) {
    if (!isCurrentLessonPaid) return true;

    if (lesson.index !== 0) return !isPreviousLessonCompleted;

    return isCurrentLessonCompleted;
  }

  if (nextLesson.product_level_id) {
    if (!isNextLessonPaid) return true;

    return isCurrentLessonCompleted;
  }

  if (isCurrentLessonCompleted && isNextLessonAvailable) return true;

  return !isNextLessonOpensInProgress;
}

function calculateTimeLeft(releaseDate) {
  const now = new Date()
  const target = new Date(releaseDate)
  const diff = target - now

  if (diff <= 0) return { days: '0', hours: '0', minutes: '0' }

  const totalMinutes = Math.floor(diff / 1000 / 60)
  const d = Math.floor(totalMinutes / 60 / 24)
  const h = Math.floor((totalMinutes / 60) % 24)
  const m = totalMinutes % 60

  return {
    days: String(d),
    hours: String(h),
    minutes: String(m),
  }
}

function updateAllLessonsCountdown(lessons) {
  lessons.forEach((lesson) => {
    if (lesson.release_date) {
      timeLeftByLessonId[lesson.id] = calculateTimeLeft(lesson.release_date)
    }
  })
}

onMounted(async() => {
  await waitForMediaLoad(chaptersContainerRef.value);

  timerInterval = setInterval(() => {
    productStore.selectedProductModules.forEach((chapter) => {
      updateAllLessonsCountdown(chapter.items)
    })
  }, 60000) // update every minute

  productStore.selectedProductModules.forEach((chapter) => {
    updateAllLessonsCountdown(chapter.items)
  })
})

onBeforeUnmount(() => {
  clearInterval(timerInterval)
})
</script>
<style scoped lang="scss">
@import "../../assets/styles/main";

.selected-course {
  &_chapters {
    padding: 0 24px;

    .chapter {
      margin-bottom: 24px;

      &_title {
        margin-bottom: 16px;
        font-size: 15px;
        font-style: normal;
        font-weight: 600;
        line-height: 120%;
        color: $white;

        &.available {
          color: $white;
        }
      }

      &_lessons {
        display: flex;
        flex-direction: column;
        gap: 16px;

        .lesson {
          position: relative;
          width: 100%;
          cursor: pointer;
          border-radius: 12px;
          background: rgba(255, 255, 255, 0.07);
          backdrop-filter: blur(12px);
          padding: 12px;
          display: flex;
          align-items: center;
          gap: 12px;
          flex-direction: column;

          &:has(.lesson__video_status) {
            align-items: flex-start;

            & .lesson_image-container {
              opacity: .5;
            }
          }

          &__schedule {
            width: 100%;
            display: flex;
            align-items: center;
            justify-content: space-between;

            &_text {
              color: rgba(255, 255, 255, 0.50);
              font-size: 11px;
              font-style: normal;
              font-weight: 500;
              line-height: normal;
            }

            &_date {
              border-radius: 10px;
              background: rgba(255, 255, 255, 0.12);
              padding: 4px 8px;
              display: flex;
              align-items: center;
              gap: 4px;

              color: #FFF;
              font-size: 11px;
              font-style: normal;
              font-weight: 500;
              line-height: normal;
            }
          }

          &__content {
            position: relative;
            width: 100%;
            display: flex;
            align-items: stretch;
            gap: 8px;
          }

          &_wrapper {
            display: flex;
            flex-direction: column;
            gap: 24px;
          }

          &_divider {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 8px;
            color: rgba(255, 255, 255, 0.50);
            font-size: 13px;
            font-style: normal;
            font-weight: 400;
            line-height: normal;
            text-wrap: nowrap;

            &::before, &::after {
              content: '';
              width: 100%;
              height: 1px;
              display: inline-block;
              background: rgba(255, 255, 255, 0.12);
            }
          }

          &_image-container {
            position: relative;
            border-radius: 12px;
            overflow: hidden;
            width: 110px;
            max-width: 110px;
            height: 71px;
            display: flex;
            align-items: center;
            justify-content: center;
            min-width: 110px;

            & .icon {
              position: absolute;
              top: 50%;
              left: 50%;
              transform: translate(-50%, -50%);
              min-width: 32px;
              min-height: 32px;
              max-width: 32px;
              max-height: 32px;
              border-radius: 50%;
              background: rgba(182, 182, 182, 0.15);
              backdrop-filter: blur(4px);
              display: flex;
              align-items: center;
              justify-content: center;
            }

            & img {
              width: 100%;
              height: 100%;
              object-fit: cover;
            }
          }

          &_play-icon {
            display: flex;
            justify-content: center;
            align-items: center;
            width: 32px;
            height: 32px;
            position: absolute;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            border-radius: 199998px;
            background: rgba(182, 182, 182, 0.15);
            backdrop-filter: blur(4px);
          }

          &_title {
            margin-right: 22px;

            & span {
              width: 100%;
              font-size: 13px;
              font-style: normal;
              font-weight: 600;
              line-height: 120%;

              color: rgba(255, 255, 255, 0.50);
            }

            &.available {
              color: $white;
            }

            &_wrapper {
              display: flex;
              flex-direction: column;
              gap: 5px;

              &.center {
                justify-content: center;
              }
            }
          }

          &__video_status {
            width: fit-content;
            padding: 2px 4px;
            border-radius: 4px;
            background: rgba(255, 255, 255, 0.07);
            color: #878787;
            font-size: 13px;
            font-style: normal;
            font-weight: 400;
            line-height: normal;
            display: flex;
            align-items: center;
          }

          &_right {
            position: absolute;
            z-index: 2;
            right: 0;
            top: 50%;
            transform: translateY(-50%);
            margin-left: auto;
            display: flex;
            align-items: center;
            gap: 5px;

            & span {
              max-width: 52px;
              margin-left: auto;
              overflow: hidden;
              color: #FFF;
              text-overflow: ellipsis;
              white-space: nowrap;
              font-size: 14px;
              font-style: normal;
              font-weight: 500;
              line-height: 125%;
              opacity: 0.5;
            }
          }

          &_arrow-icon {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 24px;
            height: 24px;

            svg {
              min-width: 24px;
            }
          }
        }
      }
    }
  }
}
</style>
