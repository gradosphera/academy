<template>
  <div class="language__wrapper">
    <div class="language__inner">
      <div @click="isLangListVisible = true" class="language__current">
        {{i18n.locale.value}}
      </div>
      <div ref="langListRef" :class="['language__list', {'visible': isLangListVisible}]">
        <div
            :class="['language__item', {'active': lang.code === i18n.locale.value}]"
            v-for="lang in languages"
            :key="lang.code"
            @click="setLang(lang.code)"
        >
          {{lang.name}}
          <SVGLanguageSelect v-if="lang.code === i18n.locale.value"/>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import SVGLanguageSelect from "./svg/SVGLanguageSelect.vue";
import {ref, onMounted} from "vue";
import {onClickOutside} from "@vueuse/core";
import {useI18n} from "vue-i18n";

const i18n = useI18n();
const isLangListVisible = ref(false);
const langListRef = ref(null);
const languages = [{code: 'ru', name: 'Русский'}];

onClickOutside(langListRef, () => {isLangListVisible.value = false})

const setLang = (lang) => {
  i18n.locale.value = lang;

  localStorage.setItem("academy-cryptomannn::lang", lang);

  isLangListVisible.value = false;
}

onMounted(() => {
  const lang = localStorage.getItem("academy-cryptomannn::lang");
  if (lang) i18n.locale.value = lang;
})
</script>
<style scoped lang="scss">
.language {
  &__wrapper {
    position: fixed;
    z-index: 5;
    top: 20px;
    right: 20px;
  }

  &__inner {
    position: relative;
  }

  &__current {
    width: 40px;
    height: 40px;
    border-radius: 12px;
    background: rgba(255, 255, 255, 0.15);
    display: flex;
    justify-content: center;
    align-items: center;

    color: #FFF;
    font-size: 13.5px;
    font-style: normal;
    font-weight: 600;
    line-height: 115%;
    text-transform: uppercase;

    cursor: pointer;
  }

  &__list {
    display: none;
    position: absolute;
    min-width: 124px;
    top: calc(100% + 6px);
    right: 0;
    flex-direction: column;
    border-radius: 16px;
    border: 1px solid rgba(255, 255, 255, 0.12);
    background: #464D88;
    backdrop-filter: blur(12px);

    &.visible {
      display: flex;
    }
  }

  &__item {
    padding: 16px;
    display: flex;
    justify-content: space-between;

    color: rgba(255, 255, 255, 0.50);
    font-size: 13px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
    cursor: pointer;

    &.active {
      color: #FFF;
    }

    &:last-child {
      border-top: 1px solid rgba(255, 255, 255, 0.12);
    }
  }
}
</style>