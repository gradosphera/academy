<template>
  <div class="podcast-player">
    <div class="progress-bar">
      <input type="range" min="0" :max="duration" v-model="currentTime" @input="seek" :style="{ background: progressBackground }" />
    </div>
    <div class="time-info">
      <span>{{ formatTimer(currentTime) }}</span>
      <span>-{{ formatTimer(duration - currentTime) }}</span>
    </div>
    <div class="controls">
      <div class="empty"></div>
      <div class="controls_main">
        <button @click="rewind" class="controls_10s-btn">
          <img src="../../assets/svg/player-minus10-sec.svg" alt="" />
        </button>
        <button @click="togglePlayPause" class="controls_play-pause-btn">
          <img v-if="isPlaying" src="../../assets/svg/player-play.svg" alt="" />
          <img v-else src="../../assets/svg/player-pause.svg" alt="" />
        </button>
        <button @click="forward" class="controls_10s-btn">
          <img src="../../assets/svg/player-plus10-sec.svg" alt="" />
        </button>
      </div>
      <button @click="changeSpeed" class="controls_speed-btn">{{ playbackRate }}x</button>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted, computed, onUnmounted } from "vue"
import { formatTimer } from "../../helpers/timer"
import { useFavoriteLessons } from "../../store/favouritesStore.js"

const props = defineProps(["file"])
const emits = defineEmits(['submit'])

const favouritesStore = useFavoriteLessons()

const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const playbackRate = ref(1)

const audio = new Audio(props.file)

const togglePlayPause = () => {
  if (isPlaying.value) {
    audio.pause()
  } else {
    audio.play()
    emits('submit');
  }
  isPlaying.value = !isPlaying.value
}

const rewind = () => {
  currentTime.value = Math.max(0, currentTime.value - 10)
  audio.currentTime = currentTime.value
}

const forward = () => {
  currentTime.value = Math.min(duration.value, currentTime.value + 10)
  audio.currentTime = currentTime.value
}

const changeSpeed = () => {
  playbackRate.value = playbackRate.value === 2 ? 1 : playbackRate.value + 0.5
  audio.playbackRate = playbackRate.value
}

const seek = () => {
  audio.currentTime = currentTime.value
}

const progressBackground = computed(() => {
  const progressPercent = (currentTime.value / duration.value) * 100
  return `linear-gradient(to right, #FFF ${progressPercent}%, #7a7a7a ${progressPercent}%)`
})

const toggleWishlist = () => {
  // Add functionality to toggle wishlist
}

onMounted(() => {
  audio.addEventListener("timeupdate", () => {
    currentTime.value = Math.floor(audio.currentTime)
  })

  audio.addEventListener("loadedmetadata", () => {
    duration.value = Math.floor(audio.duration)
  })

  audio.addEventListener("ended", () => {
    isPlaying.value = false
  })
})

onUnmounted(() => {
  audio.pause()
  isPlaying.value = false
})
</script>
<style scoped lang="scss">
.podcast-player {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  color: #ffffff;

  .progress-bar {
    width: 100%;
    margin-bottom: 12px;

    input[type="range"] {
      width: 100%;
      -webkit-appearance: none;
      appearance: none;
      background: transparent;
      outline: none;
      height: 7px;
      border-radius: 4px;

      &::-webkit-slider-runnable-track {
        height: 4px;
        border-radius: 4px;
      }

      &::-moz-range-track {
        height: 4px;
        border-radius: 4px;
      }

      /* Hide the thumb */
      &::-webkit-slider-thumb {
        -webkit-appearance: none;
        appearance: none;
        width: 0;
        height: 0;
        background: transparent;
        border: none;
      }

      &::-moz-range-thumb {
        width: 0;
        height: 0;
        background: transparent;
        border: none;
      }
    }
  }

  .time-info {
    margin-bottom: 24px;
    display: flex;
    justify-content: space-between;
    width: 100%;
    color: var(--no-variables-label-vibrant-secondary, rgba(127, 127, 127, 0.5));
    font-size: 11px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  .controls {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    & .empty {
      width: 35px;
      height: 35px;
    }

    &_wishlist {
      width: 35px;
      height: 35px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 5713.714px;
      background: rgba(255, 255, 255, 0.04);
    }

    &_main {
      display: flex;
      align-items: center;
      gap: 40px;
    }

    &_10s-btn {
      background: transparent;
      width: 24px;
      height: 24px;
    }

    &_play-pause-btn {
      background: transparent;
      width: 57px;
      height: 57px;
    }

    &_speed-btn {
      width: 35px;
      height: 35px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 5713.714px;
      background: rgba(255, 255, 255, 0.04);
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 21px; /* 131.25% */
      letter-spacing: 0.41px;
      color: var(--70-white, rgba(255, 255, 255, 0.7));
    }
  }
}
</style>
