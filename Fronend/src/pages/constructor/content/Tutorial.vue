<template>
  <div class="tutorial">
    <div ref="onboardContainerRef" class="tutorial_container">
      <div class="tutorial_swiper-outer">
        <swiper
            @swiper="getSlider"
            @slide-change="onSlideChange"
            :spaceBetween="10"
            class="tutorial_swiper"
        >
          <swiper-slide
              v-for="(card, i) in tutorialData"
              :key="i"
          >
            <div class="tutorial_content">
              <div class="tutorial_top">
                <div v-if="card.filename.length" class="tutorial_image-box">
                  <OnboardingSkeleton v-if="isMediaLoading" />
                  <img v-show="!isMediaLoading" :src="getFile(card.filename)" alt="Image">
                </div>
              </div>

              <div class="tutorial_text-wrapper">
                <h1 class="tutorial_title" v-html="card.title"></h1>
                <div class="tutorial_text">
                  <span v-for="(line, index) in parseMessage(card.description)" :key="index">
                    {{ line }}
                    <br v-if="index < parseMessage(card.description).length - 1"/>
                  </span>
                </div>
              </div>

            </div>
          </swiper-slide>
        </swiper>
      </div>
      <div class="tutorial_bottom">
        <div class="pagination">
            <span
                v-for="(_, i) in tutorialData"
                :key="i"
                :class="`bullet ${currentSlide === i ? 'active' : ''}`"
            ></span>
        </div>
        <div class="tutorial_buttons">
          <button
              v-if="currentSlide !== 0"
              class="tutorial_prev"
              @click="prevSlide"
          >
            Назад
          </button>
          <button
              :style="{background: miniAppStore.accentedColor}"
              class="tutorial_next"
              @click="nextSlide"
          >
            {{currentSlide !== tutorialData?.length - 1 ? t('general.buttons.next') : t('general.buttons.start')}}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from 'vue';
import {Swiper, SwiperSlide} from "swiper/vue";
import "swiper/scss";
import {useNavTabsStore} from "../../../store/tabsStore.js";
import {useI18n} from "vue-i18n";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {storeToRefs} from "pinia";
import {getFile, parseMessage} from "../../../helpers/index.js";
import OnboardingSkeleton from "../../../components/Skeletons/OnboardingSkeleton.vue";
import {useMediaLoader} from "../../../composable/useMediaLoader.js";

const slider = ref();
const onboardContainerRef = ref(null);
const currentSlide = ref(0);
const tabsStore = useNavTabsStore();
const miniAppStore = useMiniAppStore();
const {tutorialData} = storeToRefs(miniAppStore);
const {t} = useI18n();
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();

const getSlider = (swiper) => {
  slider.value = swiper;
};

const onSlideChange = () => {
  currentSlide.value = slider.value.activeIndex;
}

const nextSlide = () => {
  if (currentSlide.value === tutorialData.value?.length - 1) {
    tabsStore.updateActiveTab('main');

    localStorage.setItem('tutorial', JSON.stringify({status: 'passed'}));

    return;
  }

  slider.value.slideNext();
  currentSlide.value = slider.value.activeIndex;
}

const prevSlide = () => {
  slider.value.slidePrev();
  currentSlide.value = slider.value.activeIndex;
};

onMounted(async() => {
  await waitForMediaLoad(onboardContainerRef.value)
})
</script>
<style scoped lang="scss">
@use "../../../assets/styles/_main.scss" as *;

.tutorial {
  width: 100%;
  height: 100vh;
  background-color: #0B0B0B;
  color: $black;

  &_container {
    position: relative;
    height: calc(100vh - 1px);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding-bottom: 65px;

    & .pagination {
      width: fit-content;
      margin: auto 0 30px;
      display: flex;
      gap: 8px;
      align-self: center;
    }

    & .bullet {
      display: block;
      min-width: 8px;
      min-height: 8px;
      max-width: 8px;
      max-height: 8px;
      border-radius: 6px;
      background-color: rgba(255, 255, 255, 0.4);
      transition: min-width 0.2s ease;

      &.active {
        min-width: 18px;
        border-radius: 6px;
        background-color: $white;
      }
    }
  }

  &_content {
    position: relative;
    display: flex;
    flex-direction: column;
    border-radius: 8px;
    height: 100%;
  }

  &_top {
    width: 100%;
    height: 45vh;
    border-radius: 0 0 24px 24px;
    display: flex;
    overflow: hidden;
  }

  &_text-wrapper {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 0 24px;
  }

  &_title {
    color: $white;
    font-size: 32px;
    font-style: normal;
    font-weight: 400;
    line-height: 120%;
  }

  &_text {
    & span {
      color: $white;
      font-size: 17px;
      font-style: normal;
      font-weight: 400;
      line-height: 120%;
    }
  }

  &_swiper {
    &-outer {
      height: 100%;

      ::v-deep(.swiper) {
        height: 100%;
      }
    }
  }

  &_image-box {
    width: 100%;
    height: 100%;
    display: flex;

    & img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      object-position: top;
    }
  }

  &_register-box {
    display: flex;
    flex-direction: column;
    gap: 15px;
    width: 100%;
  }

  &_register-button {
    width: 100%;
    height: 48px;
    display: flex;
    align-items: center;
    border-radius: 10px;
    background-color: $white;
    padding-inline: 15px;

    & img {
      &:first-child {
        margin-right: 15px;
      }

      &:last-child {
        margin-left: auto;
        width: 16px;
      }
    }

    & span {
      @include fz-14;
      font-weight: 500;
      color: $black;
    }

    &.black {
      background-color: $dark-grey-1;

      & span {
        color: $white;
      }
    }
  }

  &_bottom {
    padding: 0 24px;
  }

  &_buttons {
    width: 100%;
    display: flex;
    gap: 12px;
    height: 45px;

    & button {
      border-radius: 12px;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 130%;
    }
  }

  &_prev {
    height: 100%;
    min-width: 100px;
    background-color: transparent;
    outline: none;
    border: 1px solid rgba(255, 255, 255, 0.17);
    color: $white;
  }

  &_next {
    background: $main-gradient;
    color: #FFF;
    width: 100%;
  }

  &_input {
    width: 100%;
    display: flex;
    gap: 8px;
    flex-direction: column;

    & input {
      width: 100%;
      background-color: rgba(255, 255, 255, 0.20);
      border-radius: 10px;
      padding: 16px;
      border: none;
      color: $white;
      font-family: 'SF Pro Display', sans-serif;
      font-size: 14px;
      font-style: normal;
      font-weight: 400;
      line-height: 100%;
      letter-spacing: -0.56px;

      &::placeholder {
        @include fz-14;
        color: rgba(255, 255, 255, 0.50);
      }

      &.warning {
        border: 1px solid #FF5E5E;
      }
    }

    & div {
      display: flex;
      justify-content: space-between;
      align-items: center;

      & .error {
        align-self: center;
      }
    }

    & .error {
      display: block;
      @include fz-12;
      color: #FF8181;
      align-self: flex-end;
    }

    & .info {
      @include fz-10;
      color: $dark-grey-2;
      max-width: 200px;
    }
  }

  &_discord-box {
    display: flex;
    flex-direction: column;
    width: 100%;
    gap: 12px;

    & .discord-code {
      width: 100%;
      height: 48px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      border-radius: 10px;
      background-color: $dark-grey-1;
      padding-inline: 15px;

      & span {
        @include fz-14;
        font-weight: 500;
        color: $white;
      }
    }
  }
}
</style>