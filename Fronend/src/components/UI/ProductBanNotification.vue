<template>
  <div ref="refModal" :class="`product-ban ${tgPlatform}`">
    <div class="product-ban__container">
      <div @click="emits('close')" class="product-ban__close">
        <SVGClose/>
      </div>
      <div class="logo-container">
        <SVGBlockLesson/>
      </div>
      <div class="product-ban__text" v-html="supportText"></div>
      <span>{{ t('general.main.ban_text_span') }}</span>
      <button :style="{background: miniAppStore.accentedColor}" v-if="miniAppData?.support?.length" @click="openSupport" class="product-ban__btn">
        {{ t('general.main.ban_contact_support') }}
      </button>
    </div>
  </div>
</template>
<script setup>
import SVGClose from "../../components/svg/SVGClose.vue";
import {useNavTabsStore} from "../../store/tabsStore.js";
import {onClickOutside} from "@vueuse/core";
import {useI18n} from "vue-i18n";
import {computed, onMounted, ref} from "vue";
import {useMiniAppStore} from "../../store/miniAppStore.js";
import {storeToRefs} from "pinia";
import SVGBlockLesson from "../svg/SVGBlockLesson.vue";

const props = defineProps({
  productName: {type: String, default: ''},
})

const emits = defineEmits(['close']);
const tg = window.Telegram.WebApp;
const miniAppStore = useMiniAppStore();
const {miniAppData} = storeToRefs(miniAppStore);
const navTabsStore = useNavTabsStore();
const refModal = ref(null);
const {t} = useI18n();
const tgPlatform = ref('');

onClickOutside(refModal, () => navTabsStore.toggleSupport())

const supportText = computed(() => {
  let text = t('general.main.ban_text');

  text = text.replace('[product title]', props.productName);

  return text;
})

const openSupport = () => {
  if (miniAppData.value?.support?.length) {
    tg.openLink(miniAppData.value?.support);
  }
}

onMounted(() => {
  tgPlatform.value = tg.platform;
})
</script>
<style scoped lang="scss">
@import "../../assets/styles/main";

.product-ban {
  position: fixed;
  z-index: 10000;
  bottom: 0;
  left: 0;
  width: 100%;
  max-height: 88%;
  display: flex;
  background-color: #181818;
  border-radius: 32px 32px 0 0;

  &.ios {
    .product-ban {
      &__container {
        padding-bottom: 50px;
      }
    }
  }

  &__container {
    position: relative;
    width: 100%;
    padding: 32px 24px 25px;
    display: flex;
    flex-direction: column;
    align-items: center;

    & span {
      color: #D0D0D0;
      text-align: center;
      font-size: 10px;
      font-style: normal;
      font-weight: 500;
      line-height: 125%;
      display: block;
      margin-bottom: 8px;
    }

    & .logo-container {
      position: relative;
      z-index: 2;
      width: 64px;
      height: 64px;
      background: #1B1B1B;
      display: flex;
      justify-content: center;
      align-items: center;
      border-radius: 50%;

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
          object-fit: cover;
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
    margin: 24px 0;
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
    background: #FFFFFF1F;

    color: #FFF;
    @include fz-16-500;
    line-height: 130%;

    ::v-deep(svg) {
      width: 18px;
      height: 18px;
    }
  }
}
</style>