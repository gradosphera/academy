<template>
  <div class="language">
    <span class="language_title">{{t('general.profile.menu.language.title_inside')}}</span>
    <ul class="language_list">
      <li
          v-for="lang in languages"
          :key="lang.id"
          :class="['language_item', {'active': lang.code === selectedLang}]"
          @click="selectedLang = lang.code"
      >
        {{ lang.lang }}
        <SVGCopied v-if="lang.code === selectedLang"/>
      </li>
    </ul>
    <div class="language_btn">
      <UIButton
          :bg="miniAppStore.accentedColor"
          @cta="saveLang"
      >
        {{t('general.buttons.save')}}
      </UIButton>
    </div>
  </div>
</template>
<script setup>
import {onMounted, ref} from "vue";
import SVGCopied from "../../../components/svg/SVGCopied.vue";
import UIButton from "../../../components/UI/UIButton.vue";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {useToastStore} from "../../../store/toastStore.js";
import {useI18n} from "vue-i18n";
import {editUserProfile} from "../../../api/api.js";

const toastStore = useToastStore();
const miniAppStore = useMiniAppStore();
const {t, locale} = useI18n();
const languages = [
  {id: 0, code: 'en', lang: 'English'},
  {id: 1, code: 'ru', lang: 'Русский'},
  // {id: 2, code: 'de', lang: 'Deutsch'},
  // {id: 3, code: 'kk', lang: 'Казақша'},
  // {id: 4, code: 'en', lang: 'Русский'},
  // {id: 5, code: 'fr', lang: 'Français'},
  // {id: 6, code: 'es', lang: 'Español'},
  // {id: 7, code: 'it', lang: 'Italiano'},
  // {id: 8, code: 'pt', lang: 'Português'},
]
const selectedLang = ref(languages[0]);

const saveLang = async () => {

  if (!miniAppStore.userData) return;

  const newLang = {...miniAppStore.userData, language: selectedLang.value};
  const formData = new FormData();
  formData.append('user', JSON.stringify(newLang));

  const resp = await editUserProfile(formData);
  if (resp.data) {
    miniAppStore.setUserData(resp.data.user);
    locale.value = selectedLang.value;
  }

  toastStore.success({text: t('general.toast_notifications.changes_saved')})
}

onMounted(() => {
  if (miniAppStore.userData?.language?.length) {
    selectedLang.value = miniAppStore.userData?.language;
  } else if (miniAppStore.miniAppData?.language?.length) {
    selectedLang.value = miniAppStore.miniAppData?.language;
  }
})
</script>
<style scoped lang="scss">
@import "../../../assets/styles/main";

.language {
  height: 100vh;
  padding: 16px 24px 155px;
  background: #0A0A0A;

  &_title {
    color: rgba(255, 255, 255, 0.5);

    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &_list {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &_item {
    position: relative;
    height: 43px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    color: rgba(255, 255, 255, 0.50);
    font-size: 17px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
    transition: color .3s ease;
    cursor: pointer;

    &.active {
      color: $white;
    }

    &:hover {
      color: $white;
    }

    &::after {
      position: absolute;
      content: '';
      display: block;
      height: 1px;
      width: 100%;
      background-color: rgba(255, 255, 255, 0.05);
      left: 0;
      bottom: -8px;
    }

    &:last-child {
      &::after {
        display: none;
      }
    }
  }

  &_btn {
    position: fixed;
    left: 0;
    bottom: 108px;
    width: 100%;
    padding: 0 24px;
  }
}
</style>