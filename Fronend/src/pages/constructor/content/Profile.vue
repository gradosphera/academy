<template>
  <div class="profile">
    <div class="profile_container">
      <div>
        <p class="profile_title">{{t('general.profile.title')}}</p>
        <div class="profile_user user">
          <div class="user_image">
            <img v-if="userData?.avatar?.length" :src="getFile(userData?.avatar)" alt="" />
            <SVGUserProfile v-else />
          </div>
          <div class="user_info">
            <p class="user_name">{{userData?.first_name?.length ? userData.first_name : `user_${userData?.telegram_id}`}}</p>
            <p v-if="userData?.telegram_username?.length" class="user_username">{{userData?.telegram_username ? `@${userData?.telegram_username}` : ''}}</p>
          </div>
          <div class="user__buttons">
            <button @click="openEditProfile">
              <SVGEdit/>
            </button>
          </div>
        </div>
        <div class="profile_settings-list">
          <div v-for="[key, value] in Object.entries(menuObj)" :key="key" @click="navTabsStore.updateActiveTab(key)" class="profile_settings-item settings-item">
            <p class="settings-item_text">{{value.name}}</p>
            <div>
              <span v-if="value.code?.length">{{value.code}}</span>
              <SVGArrowShortRight/>
            </div>
          </div>
        </div>
      </div>
    </div>
    <BgBlurSquare />
  </div>
</template>
<script setup>
import {computed, onMounted} from "vue"
import { useNavTabsStore } from "../../../store/tabsStore.js"
import BgBlurSquare from "../../../components/BgBlurSquare.vue"
import SVGArrowShortRight from "../../../components/svg/SVGArrowShortRight.vue";
import {useI18n} from "vue-i18n";
import SVGUserProfile from "../../../components/svg/SVGUserProfile.vue";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {getFile} from "../../../helpers/index.js";
import SVGEdit from "../../../components/svg/SVGEdit.vue";

const {t, locale} = useI18n();
const {userData} = useMiniAppStore();
const navTabsStore = useNavTabsStore();

const openEditProfile = () => {
  navTabsStore.updateActiveTab('edit_profile');
}

const menuObj = computed(() => {
  return {
    // 'language': {name: t('general.profile.menu.language.title'), code: ''},
    'billing_history': {name: t('general.profile.menu.billing_history.title')},
  }
})

onMounted(() => {
  menuObj.value.language.code = locale.value;
})
</script>
<style scoped lang="scss">
.profile {
  color: #fff;
  background: #0a0a0a;
  height: calc(100vh - 1px);
  overflow-y: auto;
  padding: 20px 24px 100px;

  &_container {
    position: relative;
    z-index: 2;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  &_title {
    margin-bottom: 16px;
    opacity: 0.5;
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  .user {
    padding-bottom: 15px;
    display: flex;
    align-items: center;
    gap: 12px;

    &__buttons {
      display: flex;
      align-items: center;
      height: 40px;
      margin-left: auto;

      & button {
        height: 100%;
        width: 40px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        background: rgba(255, 255, 255, 0.15);
      }
    }

    &_image {
      width: 46px;
      height: 46px;
      border-radius: 50%;
      overflow: hidden;
      display: flex;
      align-items: center;
      justify-content: center;

      img {
        object-fit: cover;
        height: 100%;
        width: 100%;
      }
    }

    &_info {
      display: flex;
      flex-direction: column;
      gap: 4px;
      color: var(--White, var(--system-white, #fff));
    }

    &_name {
      font-size: 18px;
      font-style: normal;
      font-weight: 600;
      line-height: normal;
    }

    &_username {
      opacity: 0.4;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }
  }

  &_settings-list {
    & > :not(:last-child) {
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    }

    .settings-item {
      padding: 8px 0;
      display: flex;
      align-items: center;
      justify-content: space-between;
      cursor: pointer;

      & div {
        display: flex;
        align-items: center;
      }

      & span {
        text-transform: uppercase;
        color: #FFF;
        font-size: 14px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
        opacity: 0.5;
      }

      &_text {
        font-size: 17px;
        font-style: normal;
        font-weight: 500;
        line-height: 125%;
        padding: 11px 0;
      }
    }
  }

  &_delete-btn {
    width: 100%;
    display: flex;
    height: 44px;
    justify-content: center;
    align-items: center;
    gap: 8px;
    border-radius: 8px;
    background: rgba(235, 79, 87, 0.16);

    span {
      color: #eb4f57;
      text-align: center;
      font-size: 15px;
      font-style: normal;
      font-weight: 500;
      line-height: 19px;
    }
  }
}

.switch {
  position: relative;
  display: inline-block;
  width: 51px;
  height: 30px;

  input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: var(--Secondary-Blue-Gradient, linear-gradient(180deg, #0061d2 5%, #00a7e4 105%));
    transition: background 0.3s;
    border-radius: 15px;

    &::before {
      position: absolute;
      content: "";
      height: 26px;
      width: 26px;
      left: 2px;
      bottom: 2px;
      background-color: white;
      transition: transform 0.3s;
      border-radius: 50%;
      box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    }
  }

  input:checked + .slider {
    background: var(--Secondary-Blue-Gradient, linear-gradient(180deg, #0061d2 5%, #00a7e4 105%));
  }

  input:checked + .slider::before {
    transform: translateX(21px);
  }
}
</style>
