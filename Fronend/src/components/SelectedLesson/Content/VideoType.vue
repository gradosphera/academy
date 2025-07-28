<template>
  <div class="video-lesson">
    <div class="video-lesson_video-container">
      <div v-show="lessonContentData?.filename?.length || lessonContentData?.metadata" class="rectangle-wrapper">
        <div v-if="videoStatus === 'pending_compressing' || videoStatus === 'pending_move_to_mux'"
             class="compressing-error">
          <div class="error-svg">
            <SVGClose/>
          </div>
          <p v-html="t('general.lesson.compressing_error')"></p>
        </div>
        <div class="mux-rectangle-video">
          <mux-player
              :playback-id="videoSource.rectangle_playback_id"
              :playback-token="videoSource.playback_token"
              :src="videoSource.rectangle_src"
              :poster="getFile(lessonContentData?.cover)"
              stream-type="on-demand"
              :accent-color="getAccentColorForMuxPlayer?.first"
              @playing="submitLessonProgress"
          ></mux-player>
        </div>
        <!--        <RectangularVideo :is-compressing="videoStatus === 'pending_compressing' || videoStatus === 'pending_move_to_mux'" @submit="submitLessonProgress" :src="getFile(lessonContentData?.filename)" :poster="getFile(lessonContentData?.cover)" />-->
      </div>

      <div v-show="lessonContentData?.circle_filename?.length || lessonContentData?.circle_metadata"
           class="rounded-wrapper">
        <div v-if="videoStatus === 'pending_compressing' || videoStatus === 'pending_move_to_mux'"
             class="compressing-error">
          <div class="error-svg">
            <SVGClose/>
          </div>
          <p v-html="t('general.lesson.compressing_error')"></p>
        </div>
        <div class="mux-rounded-video">
          <mux-player
              :playback-id="videoSource.circle_playback_id"
              :playback-token="videoSource.playback_token"
              :src="videoSource.circle_src"
              stream-type="on-demand"
              :accent-color="getAccentColorForMuxPlayer?.first"
              @playing="submitLessonProgress"
          ></mux-player>
        </div>
        <!--        <RoundVideo :is-compressing="videoStatus === 'pending_compressing' || videoStatus === 'pending_move_to_mux'" @submit="submitLessonProgress" :src="getFile(lessonContentData?.circle_filename)" :poster="getFile(lessonContentData?.cover)" />-->
      </div>
      <div v-if="!videoIsStarted && lessonContentData?.youtube?.length" class="video-lesson_image-container"
           @click="startVideo">
        <img class="video-lesson_image" :src="getFile(lessonContentData?.cover)" alt=""/>
        <div class="video-lesson_play-icon">
          <img src="../../../assets/svg/play-icon.svg" alt=""/>
        </div>
      </div>

      <iframe
          v-if="videoIsStarted && lessonContentData?.youtube?.length"
          class="video-lesson_video"
          :src="checkYouTubeUrl(lessonContentData?.youtube)"
          title="YouTube video player"
          frameborder="0"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
          allowfullscreen
      >
      </iframe>
    </div>
    <div class="video-lesson_title-wrapper">
      <div class="video-lesson_title">
        <span v-for="(line, index) in parseMessage(lessonContentData?.title)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(lessonContentData?.title).length - 1"/>
        </span>
      </div>
      <div class="video-lesson_add-to-wishlist-container">
        <button type="button" @click="favouritesStore.toggleFavorite(selectedLesson)"
                class="video-lesson_add-to-wishlist">
          <SVGFavouritesAdded v-if="favouritesStore.isFavorite(selectedLesson?.id)"/>
          <SVGFavouritesEmpty v-else/>
        </button>
      </div>
    </div>
    <div class="video-lesson_description description">
      <div ref="descriptionVideoRef" class="description_text" :class="{ expanded: isExpanded }" :style="{height: descriptionHeight}">
        <span v-for="(line, index) in parseMessage(lessonContentData?.description)" :key="index">
          {{ line }}
          <br v-if="index < parseMessage(lessonContentData?.description).length - 1"/>
        </span>
      </div>
      <button v-if="showReadMore" class="description_read-more-btn" type="button" @click="toggleReadMore">
        {{ isExpanded ? t('general.buttons.lessText') : t('general.buttons.moreText') }}
      </button>
    </div>
  </div>
</template>
<script setup>
import {storeToRefs} from "pinia"
import {useFavoriteLessons} from "../../../store/favouritesStore.js"
import {computed, nextTick, onMounted, ref, watch} from "vue"
import SVGFavouritesAdded from "../../svg/SVGFavouritesAdded.vue";
import SVGFavouritesEmpty from "../../svg/SVGFavouritesEmpty.vue";
import {useI18n} from "vue-i18n";
import {useProductsStore} from "../../../store/productsStore.js";
import {getFile, getProductData, parseMessage, submitAnalyticsData} from "../../../helpers/index.js";
import {getMaterialIdForMuxVideo, submitLesson} from "../../../api/api.js";
import SVGClose from "../../svg/SVGClose.vue";
import '@mux/mux-player';
import {useMiniAppStore} from "../../../store/miniAppStore.js";

defineProps({
  language: Object,
})

const {t} = useI18n();
const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();
const {lessonContentData, selectedLesson} = storeToRefs(productStore);
const favouritesStore = useFavoriteLessons();
const showReadMore = ref(false)
const descriptionVideoRef = ref(null)
const descriptionHeight = ref('auto');
const startDescriptionHeight = ref(0);
const videoSource = ref({
  rectangle_src: '',
  rectangle_playback_id: '',
  circle_src: '',
  circle_playback_id: '',
  playback_token: '',
})

const isExpanded = ref(false);

function toggleReadMore() {
  isExpanded.value = !isExpanded.value

  if (isExpanded.value) {
    descriptionHeight.value = descriptionVideoRef.value.scrollHeight + 'px';
  } else {
    descriptionHeight.value = startDescriptionHeight.value + 'px';
  }
}

const videoStatus = computed(() => {
  const videoItem = selectedLesson.value?.materials?.find(el => el.content_type === 'video' || el.content_type === 'circle_video');

  if (videoItem) {
    return videoItem.status;
  }

  return ''
})

const videoPosterBackgroundColor = computed(() => {
  if (lessonContentData.value?.cover?.length) {
    return '#212121'
  }
  return 'transparent'
})

const getAccentColorForMuxPlayer = computed(() => {
  const str = miniAppStore.miniAppData?.color_theme?.accent_color || '';

  const arr = str.split('/');

  return {first: arr[0]?.length ? arr[0] : '#0061D2', second: arr[1] || ''};
})

const videoIsStarted = ref(false)

const submitLessonProgress = async () => {
  submitAnalyticsData('lesson_media_play', {
    lesson_title: selectedLesson.value?.title,
    course_title: productStore.selectedProduct?.title || '',
    media_type: 'video',
  })

  if (!productStore.isLessonCompleted) {
    const data = new FormData();

    await submitLesson(productStore.selectedLesson?.id, data);
    await getProductData(productStore.selectedProduct?.id);
    productStore.isLessonCompleted = true;
  }
}

async function startVideo() {
  videoIsStarted.value = true;

  await submitLessonProgress()
}

async function checkTextHeight() {
  await nextTick(() => {
    const element = descriptionVideoRef.value;
    if (element) {
      showReadMore.value = element.scrollHeight > element.clientHeight;
    }
  })
}

const checkYouTubeUrl = (url) => {
  if (url) {
    if (url.includes('watch?v=')) {
      return url.replace('watch?v=', 'embed/') + '?autoplay=1';
    } else {
      return url + '?autoplay=1';
    }
  }
}


const loadVideo = async () => {
  const {filename, metadata, circle_filename, circle_metadata, video_material_id} = lessonContentData.value || {};
  // Case 1: Local preview (user-selected file via input)
  if (filename?.startsWith('blob')) {
    videoSource.value.rectangle_playback_id = '';
    videoSource.value.rectangle_src = filename;
    return;
  }

  if (circle_filename?.startsWith('blob')) {
    videoSource.value.circle_playback_id = '';
    videoSource.value.circle_src = circle_filename;
    return;
  }

  // Case 2: Load from Mux API (with signed URL)
  if (metadata || circle_metadata) {
    const resp = await getMaterialIdForMuxVideo(video_material_id);
    if (resp?.data?.playback_id && resp?.data?.token) {
      videoSource.value.playback_token = resp?.data?.token;

      if (metadata) {
        videoSource.value.rectangle_src = '';
        videoSource.value.rectangle_playback_id = `${resp.data.playback_id}`;
      }

      if (circle_metadata) {
        videoSource.value.circle_src = '';
        videoSource.value.circle_playback_id = `${resp.data.playback_id}`;
      }
    }
  }
};

watch(
    () => lessonContentData.value?.description,
    async (desc) => {
      if (!desc) return;
      await nextTick();
      setTimeout(() => {
        if (descriptionVideoRef.value) {
          const el = descriptionVideoRef.value;
          showReadMore.value = el.scrollHeight > el.clientHeight;
          descriptionHeight.value = el.clientHeight + 'px';
          startDescriptionHeight.value = el.clientHeight;
        }
      }, 200);
    },
    { immediate: true }
);

watch(() => [
  lessonContentData.value?.filename,
  lessonContentData.value?.metadata,
  lessonContentData.value?.circle_metadata,
  lessonContentData.value?.circle_filename
], () => {
  loadVideo();
});

onMounted(async () => {
  await loadVideo()
})
</script>

<style scoped lang="scss">
.video-lesson {
  padding-bottom: 16px;
  color: var(--White, var(--system-white, #fff));

  &_video-container {
    width: 100%;
    cursor: pointer;
    position: relative;
    margin-bottom: 16px;
    overflow: hidden;
    border-radius: 12px;

    & video {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }

    & .compressing-error {
      position: absolute;
      top: 0;
      left: 0;
      z-index: 10;
      width: 100%;
      height: 100%;
      border-radius: 12px;
      background: linear-gradient(0deg, rgba(13, 13, 15, 0.90) 0%, rgba(13, 13, 15, 0.90) 100%);
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      gap: 24px;

      & .error-svg {
        display: flex;
        align-items: center;
        justify-content: center;
        min-width: 50px;
        min-height: 50px;
        max-width: 50px;
        max-height: 50px;
        border-radius: 50%;
        background: #EB4F5726;

        & svg {
          width: 100%;
          height: 100%;

          ::v-deep(path) {
            stroke: #EB4F57;
          }
        }
      }

      & p {
        color: #EB4F57;
        text-align: center;
        font-size: 14px;
        font-style: normal;
        font-weight: 400;
        line-height: 18px;
        letter-spacing: -0.14px;
      }
    }

    & .rectangle-wrapper {
      position: relative;
      width: 100%;
      height: 192px;
    }

    & .rounded-wrapper {
      display: flex;
      justify-content: center;

      & .compressing-error {
        width: 193px;
        height: 193px;
        border-radius: 50%;
        left: 50%;
        transform: translateX(-50%);

        & p {
          font-size: 12px;
        }
      }
    }
  }

  &_image-container {
    position: relative;
    width: 100%;
    height: 100%;
    cursor: pointer;
    margin-bottom: 16px;
  }

  &_image {
    object-fit: cover;
    border-radius: 16px;
    width: 100%;
    max-height: 100%;
  }

  &_controls {
    width: 100%;
    position: fixed;
    bottom: 104px;
    left: 50%;
    transform: translateX(-50%);
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

  .description {
    margin-bottom: 60px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.04);
    padding: 16px;

    &_subtitle {
      margin-bottom: 12px;
      opacity: 0.3;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_text {
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

      ::v-deep(a) {
        color: #7581EC;
        cursor: pointer;
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
      line-height: 130%;
      /* 20.8px */
    }
  }
}

.mux-rectangle-video {
  ::v-deep(mux-player) {
    aspect-ratio: 16/9;
    height: 192px;
  }
}

::v-deep(mux-player) {
  &::part(poster) {
    background-color: v-bind(videoPosterBackgroundColor);
  }
}

.mux-rounded-video {
  width: 192px;
  height: 192px;
  position: relative;
  overflow: hidden;
  border-radius: 50%;
  display: flex;

  ::v-deep(mux-player) {
    aspect-ratio: 1/1;
    width: 192px;
    height: 192px;
    --bottom-controls: none;
    --media-object-fit: cover;
  }

  ::v-deep(media-poster-image) {
    height: 192px !important;
    object-fit: cover !important;
  }
}
</style>
