<template>
  <div class="homework-empty">
    <div class="homework-empty_icon">
      <img src="../../assets/svg/homework-empty-icon.svg" alt="" />
    </div>
    <p class="homework-empty_title" v-html="t('general.homeworks.emptyText')"></p>
    <p class="homework-empty_subtitle">{{t('general.homeworks.addText')}}</p>
  </div>
</template>
<script setup>
import {useI18n} from "vue-i18n";

const {t} = useI18n();
</script>
<style scoped lang="scss">
.homework-empty {
  height: 90%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;

  &_icon {
    margin-bottom: 25px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 80px;
    height: 80px;
  }

  &_title {
    margin-bottom: 16px;
    color: var(--White, var(--system-white, #fff));
    text-align: center;
    font-size: 22px;
    font-style: normal;
    font-weight: 500;
    line-height: 120%; /* 26.4px */
  }

  &_subtitle {
    margin-bottom: 16px;
    color: var(--70-white, rgba(255, 255, 255, 0.7));
    text-align: center;
    font-size: 17px;
    font-style: normal;
    font-weight: 400;
    line-height: 130%;
    letter-spacing: 0.17px;
  }
}
</style>
