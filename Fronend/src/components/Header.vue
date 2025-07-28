<template>
  <header class="header">
    <div class="header_container">
      <div class="header_left" @click="authorStore.toggleAuthorVisibility()">
        <div class="logo-container">
          <div class="logo">
            <img :src="getLogo" alt="">
          </div>
        </div>
        <span>{{t('general.author.greetings')}}</span>
      </div>
      <div class="header_right">
        <div @click="navigateToProfile" class="header_profile">
          <SVGProfile/>
        </div>
      </div>
    </div>
  </header>
</template>
<script setup>
import { useNavTabsStore } from "../store/tabsStore.js"
import { useAuthorStore } from "../store/authorStore.js"
import { storeToRefs } from "pinia"
import {useI18n} from "vue-i18n";
import {useMiniAppStore} from "../store/miniAppStore.js";
import {computed} from "vue";
import {getFile} from "../helpers/index.js";
import SVGProfile from "./svg/SVGProfile.vue";
import defaultLogo from "../assets/images/big-logo-for-branding.png";

const miniAppStore = useMiniAppStore();
const authorStore = useAuthorStore();
const { isShowAuthor } = storeToRefs(authorStore);
const navTabsStore = useNavTabsStore();
const {t} = useI18n();

function navigateToNotifications() {
  if (isShowAuthor.value) authorStore.toggleAuthorVisibility()
  navTabsStore.updateActiveTab('notifications')
}

function navigateToProfile() {
  if (isShowAuthor.value) authorStore.toggleAuthorVisibility()
  navTabsStore.updateActiveTab('profile')
}

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
</script>
<style scoped lang="scss">
@import "../assets/styles/main.scss";

.header {
  &_container {
    display: flex;
    margin-bottom: 40px;
    justify-content: space-between;
  }

  &_left {
    background: inherit;
    outline: none;
    display: flex;
    align-items: center;
    gap: 8px;
    -webkit-tap-highlight-color: transparent;
    cursor: pointer;

    &:focus,
    &:focus-visible,
    &:active {
      background-color: inherit;
    }

    .logo-container {
      width: 40px;
      height: 40px;
      background: url("../assets/images/logo-circle.webp");
      background-size: contain;
      display: flex;
      justify-content: center;
      align-items: center;

      .logo {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 32px;
        height: 32px;
        border-radius: 100.313px;
        background: #0d0d0d;
        overflow: hidden;

        & img {
          width: 100%;
          height: 100%;
          object-fit: contain;
        }
      }
    }

    span {
      color: var(--system-white, #fff);
      font-feature-settings: "case" on;
      font-family: "Inter Tight", sans-serif;
      font-size: 17px;
      font-style: normal;
      font-weight: 500;
      line-height: 22px;
    }
  }

  &_profile {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.15);
  }
}
</style>
