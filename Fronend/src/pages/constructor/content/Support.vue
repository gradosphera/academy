<template>
  <div v-if="navTabsStore.isSupportOpen" class="support__wrapper">
    <Transition name="slide-up">
      <div v-if="isSupportContentVisible" ref="refModal" class="support">
        <div class="support__container">
          <div @click="navTabsStore.toggleSupport()" class="support__close">
            <SVGClose/>
          </div>
          <div :style="{background: getLogoContainerBackground}" class="logo-container" :class="{noLogo: !miniAppStore.miniAppData?.logo?.length}">
            <div class="logo-container_inner">
              <div class="logo">
                <img v-if="miniAppStore.miniAppData?.logo?.length" :src="getLogo" alt="">
                <SVGNoSupport v-else />
              </div>
            </div>
          </div>
          <div class="support__text" v-html="supportText"></div>
          <button v-if="miniAppData?.support?.length" @click="openSupport" :style="{background: miniAppStore.accentedColor}" class="support__btn"><SVGTelegram/> {{t('general.buttons.sendSupport')}}</button>
        </div>
      </div>
    </Transition>
  </div>
</template>
<script setup>
import SVGTelegram from "../../../components/svg/SVGTelegram.vue";
import SVGClose from "../../../components/svg/SVGClose.vue";
import {useNavTabsStore} from "../../../store/tabsStore.js";
import {onClickOutside} from "@vueuse/core";
import {useI18n} from "vue-i18n";
import {computed, ref, watch} from "vue";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {storeToRefs} from "pinia";
import {getFile, submitAnalyticsData} from "../../../helpers/index.js";
import defaultLogo from "../../../assets/images/big-logo-for-branding.webp";
import SVGNoSupport from "../../../components/svg/SVGNoSupport.vue";

const miniAppStore = useMiniAppStore();
const {miniAppData} = storeToRefs(miniAppStore);
const navTabsStore = useNavTabsStore();
const refModal = ref(null);
const {t} = useI18n();
const isSupportContentVisible = ref(false);

onClickOutside(refModal, () => navTabsStore.toggleSupport())

const getLogoContainerBackground = computed(() => {
  if (!miniAppStore.miniAppData?.logo?.length) {
    return 'transparent'
  } else {
    return miniAppStore.accentedColor;
  }
})

const supportText = computed(() => {
  if (!miniAppData.value?.support?.length) {
    return t('general.support.empty_link')
  } else {
    return t('general.support.title')
  }
})

const getLogo = computed(() => {
  const app = miniAppStore.miniAppData;

  if (app?.logo?.length && app?.teacher_avatar?.length) {
    return getFile(app?.teacher_avatar)
  }

  if (app?.teacher_avatar?.length) {
    return getFile(app?.teacher_avatar);
  } else if (app?.logo?.length) {
    return getFile(app?.logo);
  } else {
    return defaultLogo;
  }
})

watch(() => navTabsStore.isSupportOpen, (val) => {
  if (val) {

    submitAnalyticsData('support_open_navigation', {
      location: 'navigation',
    })

    setTimeout(() => {
      isSupportContentVisible.value = true;
    }, 50)
  } else {
    isSupportContentVisible.value = false;
  }
})

const openSupport = () => {
  if (miniAppData.value?.support?.length) {
    submitAnalyticsData('support_open_navigation', {
      location: 'popup',
    })

    window.Telegram.WebApp.openLink(miniAppData.value?.support);
  }
}
</script>
<style scoped lang="scss">
@import "../../../assets/styles/main";

.support {
  width: 100%;
  display: flex;
  background-color: #181818;
  border-radius: 32px 32px 0 0;

  &__wrapper {
    position: fixed;
    bottom: 0;
    width: 100%;
    left: 0;
    z-index: 10000;
    min-height: 100vh;
    background: rgba(0, 0, 0, 0.40);
    backdrop-filter: blur(2px);
    display: flex;
    align-items: flex-end;
    color: #FFF;
  }

  &__container {
    position: relative;
    width: 100%;
    padding: 32px 24px 50px;
    display: flex;
    flex-direction: column;
    align-items: center;

    & .logo-container {
      position: relative;
      z-index: 2;
      width: 64px;
      height: 64px;
      background-size: contain;
      display: flex;
      justify-content: center;
      align-items: center;
      border-radius: 50%;
      padding: 3px;

      &_inner {
        width: 100%;
        height: 100%;
        border-radius: 50%;
        background: #0d0d0d;
        display: flex;
        justify-content: center;
        align-items: center;
      }

      .logo {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 58px;
        height: 58px;
        border-radius: 100px;
        background: #0d0d0d;
        overflow: hidden;

        & img {
          width: 100%;
          height: 100%;
          object-fit: contain;
        }
      }

      &.noLogo {
        background: #212121;
        border-radius: 50%;
        overflow: hidden;

        .logo {
          background: #212121;
        }
      }
    }
  }

  &__close {
    position: absolute;
    left: 50%;
    transform: translateX(-50%);
    top: -16px;
    min-width: 56px;
    min-height: 56px;
    max-height: 56px;
    max-width: 56px;
    border-radius: 50%;
    background-color: #181818;

    cursor: pointer;

    & svg {
      position: relative;
      left: 50%;
      top: 5px;
      transform: translateX(-50%);
    }
  }

  &__text {
    margin: 24px 0 32px;
    color: $white;
    @include fz-17;
    text-align: center;
  }

  &__btn {
    width: 100%;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: $main-gradient;

    color: #FFF;
    @include fz-16-500;
    line-height: 130%;

    ::v-deep(svg) {
      width: 18px;
      height: 18px;
    }
  }
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.6s ease;
}

.slide-up-enter-from {
  transform: translateY(100%);
}

.slide-up-enter-to {
  transform: translateY(0);
}

.slide-up-leave-from {
  transform: translateY(0);
}

.slide-up-leave-to {
  transform: translateY(100%);
}
</style>