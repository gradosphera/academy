<template>
  <div class="favorites-empty">
    <div class="favorites-empty_icon">
      <img src="../../assets/svg/favorites-empty-icon.svg" alt="" />
    </div>
    <p class="favorites-empty_title">{{t('general.favourites.emptyText')}}</p>
  </div>
</template>
<script setup>
import {useI18n} from "vue-i18n";

const {t} = useI18n();
</script>
<style scoped lang="scss">
.favorites-empty {
  height: calc(100% - 61px);
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 20px;
  align-items: center;

  &_icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 67.126px;
    height: 67.126px;
  }

  &_title {
    color: var(--White, var(--system-white, #fff));
    text-align: center;
    font-size: 18px;
    font-style: normal;
    font-weight: 500;
    line-height: 120%; /* 26.4px */
  }
}
</style>
