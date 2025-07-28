<template>
  <div class="stream-lesson">
    <div @click="startVideo" class="stream-lesson_image-container">
      <div v-if="isLessThan24Hours" class="info_date timer">
        <SVGTimer/>
        <span>{{ `${hours}:${minutes}:${seconds}` }}</span>
      </div>

      <div v-if="selectedLesson.videoLink.length && !videoIsStarted" class="stream-lesson_play-icon">
        <img src="../../../assets/svg/play-icon.svg" alt=""/>
      </div>

      <img v-if="!videoIsStarted" class="stream-lesson_image" :src="selectedLesson.previewImage" alt=""/>

      <iframe
          v-else
          class="stream-lesson_video"
          :src="`${selectedLesson.videoLink}&autoplay=1`"
          title="YouTube video player"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
          allowfullscreen
      >
      </iframe>
    </div>

    <div :class="['stream-lesson_info info']">
      <div class="info_is-stream">
        <SVGStreamIcon/>
        <span>Live Stream</span>
      </div>
      <div class="info_badge info_date">
        <SVGStreamCalendarIcon/>
        <span>{{ language?.formatedDate }}</span>
      </div>
    </div>
    <div class="stream-lesson_title-wrapper">
      <p class="stream-lesson_title">{{ language?.title }}</p>
      <div class="stream-lesson_add-to-wishlist-container">
        <button type="button" @click="favouritesStore.toggleFavorite(selectedLesson.lessonId)"
                class="stream-lesson_add-to-wishlist">
          <SVGFavouritesAdded v-if="favouritesStore.isFavorite(selectedLesson.lessonId)"/>
          <SVGFavouritesEmpty v-else/>
        </button>
      </div>
    </div>
    <button v-if="eventInProgress" @click="openLink" class="stream-lesson_join-btn">{{t('general.buttons.joinStreamNow')}}</button>
    <div class="stream-lesson_description description">
      <p ref="streamDescription" class="description_text" :class="{ expanded: isExpanded }"
         v-html="language?.description"></p>
      <button v-if="showReadMore" class="description_read-more-btn" type="button" @click="toggleReadMore">
        {{ isExpanded ? t('general.buttons.lessText') : t('general.buttons.moreText') }}
      </button>
    </div>
    <BtnNextLesson v-if="isButtonVisible" class="stream-lesson_controls"/>
  </div>
</template>
<script setup>
import {computed, onMounted, onUnmounted, ref, watch} from "vue"
import {storeToRefs} from "pinia"
import {useSelectedLesson} from "../../../store/selectedLesson.js"
import {useFavoriteLessons} from "../../../store/favouritesStore.js"
import BtnNextLesson from "../BtnNextLesson.vue"
import SVGTimer from "../../svg/SVGTimer.vue";
import SVGFavouritesEmpty from "../../svg/SVGFavouritesEmpty.vue";
import SVGFavouritesAdded from "../../svg/SVGFavouritesAdded.vue";
import SVGStreamIcon from "../../svg/SVGStreamIcon.vue";
import SVGStreamCalendarIcon from "../../svg/SVGStreamCalendarIcon.vue";
import {useSelectedCourse} from "../../../store/selectedCourse.js";
import {useI18n} from "vue-i18n";

defineProps({
  language: Object,
})

const {t} = useI18n();
const selectedLessonStore = useSelectedLesson();
const {selectedLesson} = storeToRefs(selectedLessonStore);
const courseStore = useSelectedCourse();
const now = ref(new Date());
const eventInProgress = computed(() => now.value >= selectedLesson.value.start && now.value <= selectedLesson.value.end);
const favouritesStore = useFavoriteLessons();
const isExpanded = ref(false);
const showReadMore = ref(false);
const streamDescription = ref(null);
const videoIsStarted = ref(false)

function toggleReadMore() {
  isExpanded.value = !isExpanded.value
}

function startVideo() {
  if (!selectedLesson.value.videoLink.length) return;

  videoIsStarted.value = true
}

function checkTextHeight() {
  const element = streamDescription.value;
  if (element) {
    const lineHeight = parseFloat(getComputedStyle(element).lineHeight);
    const maxHeight = lineHeight * 5;
    showReadMore.value = element.scrollHeight > maxHeight;
  }
}

const isButtonVisible = computed(() => {
  const {lessonType, lessonId} = selectedLesson.value;

  if (lessonType === 'stream') {
    const progress = courseStore.courseProgress["1"] || [];

    return progress.find(val => val.lesson_id === lessonId);
  }

  return false;
})

const openLink = () => {
  window.Telegram.WebApp.openLink(selectedLesson.value.link);
}

// const isStreamCompleted = computed((() => {
//   const progress = courseStore.courseProgress?.["1"] || [];
//
//   return progress.find(lesson => lesson.lesson_id === selectedLesson.value.lessonId);
// }))

const isLessThan24Hours = computed(() => {
  const start = selectedLesson.value.start.getTime();
  const now = Date.now();
  const millisecondsIn24Hours = 24 * 60 * 60 * 1000;
  const val = start - now;

  if (val < 0) return false;

  return val < millisecondsIn24Hours;
});

const hours = ref('00');
const minutes = ref('00');
const seconds = ref('00');

const calculateTimeLeft = () => {
  const now = new Date().getTime();
  const startDate = new Date(selectedLesson.value.start);
  const timeLeft = startDate - now;

  if (timeLeft > 0) {
    const h = Math.floor((timeLeft % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const m = Math.floor((timeLeft % (1000 * 60 * 60)) / (1000 * 60));
    const s = Math.floor((timeLeft % (1000 * 60)) / 1000);

    hours.value = String(h).padStart(2, '0');
    minutes.value = String(m).padStart(2, '0');
    seconds.value = String(s).padStart(2, '0');
  } else {
    hours.value = '00';
    minutes.value = '00';
    seconds.value = '00';
  }
};

watch(() => selectedLesson.description, () => {
  checkTextHeight();
});

onMounted(() => {
  checkTextHeight()
  if (isLessThan24Hours.value) {
    calculateTimeLeft();
    const timer = setInterval(calculateTimeLeft, 1000);
    onUnmounted(() => {
      clearInterval(timer);
    });
  }
})
</script>
<style scoped lang="scss">
.stream-lesson {
  color: #fff;
  padding-bottom: 16px;

  &_image-container {
    margin-bottom: 16px;
    position: relative;
    width: 100%;
    aspect-ratio: 16 / 9;
    cursor: pointer;
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

  .info {
    margin-bottom: 16px;
    display: flex;
    gap: 8px;

    &_is-stream {
      display: flex;
      justify-content: flex-start;
      align-items: center;
      height: 24px;
      padding: 0 8px;
      gap: 8px;
      border-radius: 5px;
      background: #e72e2e;

      span {
        color: var(--system-white, #fff);
        font-size: 11px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }
    }

    &_date {
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 0 8px;
      height: 24px;
      gap: 6px;
      border-radius: 5px;
      background: rgba(255, 255, 255, 0.07);

      &.timer {
        position: absolute;
        top: 11px;
        right: 9px;
        background: #1B1A1D;

        & span {
          width: 44px;
        }
      }

      span {
        color: var(--system-white, #fff);
        font-size: 11px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }
    }
  }

  &_title {
    font-size: 22px;
    font-style: normal;
    font-weight: 700;
    line-height: 125%;

    &-wrapper {
      margin-bottom: 16px;

      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  &_video {
    border-radius: 16px;
    width: 100%;
    height: 100%;
  }

  &_play-icon {
    display: flex;
    justify-content: center;
    align-items: center;
    width: 64px;
    height: 64px;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    border-radius: 199998px;
    background: rgba(182, 182, 182, 0.15);
    backdrop-filter: blur(4px);

    img {
      width: 24px;
      height: 24px;
    }
  }

  &_join-btn {
    margin-bottom: 24px;
    display: flex;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 44px;
    border-radius: 12px;
    background: linear-gradient(180deg, #464D88 5%, #727EE6 105%);
    color: #FFF;
    font-size: 16px;
    font-style: normal;
    font-weight: 500;
    line-height: 130%;
  }

  .description {
    margin-bottom: 70px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    padding: 16px;

    &_subtitle {
      margin-bottom: 12px;
      opacity: 0.3;
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
    }

    &_text {
      white-space: pre-wrap;
      margin-bottom: 12px;
      color: #fff;
      display: -webkit-box;
      -webkit-line-clamp: 5;
      -webkit-box-orient: vertical;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      opacity: 0.8;

      &.expanded {
        -webkit-line-clamp: unset;
        display: block;
      }

      ::v-deep(ul) {
        padding-left: 20px;
        list-style: disc;
      }
    }

    &_read-more-btn {
      background-color: unset;
      color: var(--White, var(--system-white, #fff));
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
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
      line-height: 130%; /* 20.8px */
    }
  }
}
</style>
