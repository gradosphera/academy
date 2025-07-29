<template>
  <div
      @click="setControlsVisible"
      class="rectangle-video"
  >
    <div class="rectangle-video__loader">
      <div v-if="productStore.selectedVideoType === 'default' && isVideoLoading && !isCompressing" class="loaderMedia">
        <SVGLoader/>
      </div>
    </div>
    <video
        ref="videoRef"
        :src="src"
        :poster="poster"
        playsinline
        preload="metadata"
        @playing="onVideoPlaying"
        @waiting="onVideoWaiting"
        @timeupdate="updateTime"
        @loadedmetadata="updateDuration"
        @canplaythrough="onVideoLoaded"
    ></video>

    <div :class="['rectangle-video__controls', {visible: isControlsVisible}]">
      <div class="rectangle-video__controls_center">
        <button class="move" @click="seek(-10)"><SVGMoveVideoLeft/></button>
        <button :class="['play', {visible: isPlayButtonVisible && !isCompressing}]" @click="togglePlay">
          <SVGPlayVideo v-if="!isPlaying"/>
          <SVGPause v-else/>
        </button>
        <button class="move" @click="seek(10)"><SVGMoveVideoRight/></button>
      </div>

      <div class="rectangle-video__controls_bottom">
        <span>{{ formatTime(currentTime) }} / {{ formatTime(duration) }}</span>

        <input
            type="range"
            min="0"
            :max="duration"
            step="0.1"
            v-model="currentTime"
            @input="onSeek"
        />

        <button @click="toggleSpeed">{{ playbackRate }}x</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, watch} from "vue";
import SVGMoveVideoLeft from "../svg/SVGMoveVideoLeft.vue";
import SVGMoveVideoRight from "../svg/SVGMoveVideoRight.vue";
import SVGPlayVideo from "../svg/SVGPlayVideo.vue";
import SVGPause from "../svg/SVGPause.vue";
import SVGLoader from "../svg/SVGLoader.vue";
import {useProductsStore} from "../../store/productsStore.js";

defineProps({
  src: String,
  poster: String,
  isCompressing: {type: Boolean, default: false},
});

const emits = defineEmits(['submit'])

const productStore = useProductsStore();
const videoRef = ref(null);
const isVideoLoading = ref(true);
const isPlaying = ref(false);
const currentTime = ref(0);
const duration = ref(0);
const playbackRate = ref(1);
const isControlsVisible = ref(false);
const isPlayButtonVisible = ref(false);

watch(() => isVideoLoading.value, (newVal) => {
  if (!newVal) {
    isPlayButtonVisible.value = true;
  }
})
const hideControlsTimeout = ref(null);

const setControlsVisible = () => {
  if (isVideoLoading.value) return;

  isControlsVisible.value = true;
  isPlayButtonVisible.value = true;
}

const onVideoPlaying = () => {
  isVideoLoading.value = false;

  if (!isControlsVisible.value && !isPlayButtonVisible.value) {
    isControlsVisible.value = true;
    isPlayButtonVisible.value = true;
  }
}

const onVideoWaiting = () => {
  isVideoLoading.value = true;
  isControlsVisible.value = false;
  isPlayButtonVisible.value = false;
}

const togglePlay = () => {
  const video = videoRef.value;
  if (!video || isVideoLoading.value) return;
  emits('submit');
  if (video.paused) {
    video.play();
    isPlaying.value = true;
    isControlsVisible.value = true;

    if (hideControlsTimeout.value) {
      clearTimeout(hideControlsTimeout.value);
    }

    hideControlsTimeout.value = setTimeout(() => {
      isControlsVisible.value = false;
      isPlayButtonVisible.value = false;
      hideControlsTimeout.value = null;
    }, 1000);
  } else {
    video.pause();
    isPlaying.value = false;
  }
};

const onVideoLoaded = () => {
  isVideoLoading.value = false;
}

const updateTime = () => {
  currentTime.value = videoRef.value.currentTime;
};

const updateDuration = () => {
  duration.value = videoRef.value.duration;
};

const seek = (seconds) => {
  const video = videoRef.value;
  video.currentTime = Math.min(
      Math.max(0, video.currentTime + seconds),
      duration.value
  );
};

const onSeek = () => {
  videoRef.value.currentTime = currentTime.value;
};

const toggleSpeed = () => {
  const video = videoRef.value;
  playbackRate.value = playbackRate.value === 2 ? 1 : playbackRate.value + 0.25;
  video.playbackRate = playbackRate.value;
};

const formatTime = (time) => {
  const minutes = Math.floor(time / 60)
      .toString()
      .padStart(1, "0");
  const seconds = Math.floor(time % 60)
      .toString()
      .padStart(2, "0");
  return `${minutes}:${seconds}`;
};
</script>

<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.rectangle-video {
  position: relative;
  height: 100%;
  width: 100%;
  display: flex;
  justify-content: center;

  &__loader {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  &__controls {
    height: 100%;
    width: 100%;
    position: absolute;
    bottom: 0;
    left: 0;

    &.visible {
      & .move, .rectangle-video__controls_bottom {
        opacity: 1;
        pointer-events: all;
      }
    }

    button {
      background: none;
      border: none;
      color: white;
      font-size: 1.1rem;
      cursor: pointer;
    }

    &_center {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      display: flex;
      align-items: center;
      gap: 16px;

      & button {
        border-radius: 50%;
        display: flex;
        justify-content: center;
        align-items: center;
        background: rgba(0, 0, 0, 0.25);
        backdrop-filter: blur(5px);
      }

      & .move {
        min-width: 32px;
        min-height: 32px;
        max-height: 32px;
        max-width: 32px;
        opacity: 0;
        pointer-events: none;
      }

      & .play {
        min-width: 64px;
        min-height: 64px;
        max-height: 64px;
        max-width: 64px;
        opacity: 0;

        &.visible {
          opacity: 1;
        }

        & svg {
          width: 24px;
          height: 24px;
        }
      }
    }

    &_bottom {
      width: 100%;
      padding: 0 8px;
      height: 24px;
      position: absolute;
      bottom: 8px;
      left: 0;
      display: flex;
      align-items: center;
      gap: 16px;
      opacity: 0;
      pointer-events: none;

      input[type="range"] {
        flex: 1;
        padding: 10px;
        cursor: pointer;
        -webkit-appearance: none;
        appearance: none;
        width: 100%;
        height: 4px;
        background: transparent;
        border-radius: 2px;
        outline: none;
        transition: background 0.3s ease;

        &::-webkit-slider-thumb {
          -webkit-appearance: none;
          appearance: none;
          width: 12px;
          height: 12px;
          background: #fff;
          border-radius: 50%;
          cursor: pointer;
          position: relative;
          z-index: 2;
          margin-top: -4px;

          &::after {
            top: 0;
            right: 0;
            display: flex;
            content: '';
            width: 4px;
            height: 32px;
            border-radius: 100px;
            background-color: #FFF;
          }
        }

        &::-webkit-slider-runnable-track {
          height: 4px;
          border-radius: 3px;
          background: rgba(127, 127, 127, 0.5);
        }
      }

      & span {
        color: #FFF;
        text-align: center;
        font-size: 11px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }

      & button {
        color: #FFF;
        text-align: center;
        font-size: 11px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }
    }
  }
}

video {
  width: 100%;
}
</style>
