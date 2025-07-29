<template>
  <div class="app">
    <div class="incorrect-page">
      <div class="incorrect-page__inner">
        <svg xmlns="http://www.w3.org/2000/svg" width="80" height="80" viewBox="0 0 80 80" fill="none">
          <path d="M13.3712 69.7266H66.6305C67.4956 69.7264 68.3458 69.5018 69.098 69.0748C69.8503 68.6478 70.4789 68.0329 70.9224 67.2902C71.3659 66.5475 71.6092 65.7025 71.6284 64.8377C71.6476 63.9729 71.4422 63.1179 71.0321 62.3562L44.404 12.9031C42.5149 9.39687 37.4868 9.39687 35.5977 12.9031L8.96961 62.3562C8.55954 63.1179 8.35408 63.9729 8.37331 64.8377C8.39253 65.7025 8.63579 66.5475 9.07932 67.2902C9.52284 68.0329 10.1514 68.6478 10.9037 69.0748C11.656 69.5018 12.5062 69.7264 13.3712 69.7266Z" stroke="white" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M39.1022 30.5245L39.999 49.587L40.8944 30.5323C40.8999 30.4105 40.8805 30.2888 40.8374 30.1747C40.7943 30.0606 40.7285 29.9566 40.6438 29.8688C40.5591 29.7811 40.4574 29.7115 40.345 29.6644C40.2325 29.6173 40.1116 29.5936 39.9897 29.5948C39.8699 29.5959 39.7515 29.6211 39.6416 29.6688C39.5317 29.7164 39.4324 29.7856 39.3497 29.8723C39.267 29.959 39.2025 30.0613 39.1599 30.1733C39.1174 30.2853 39.0978 30.4047 39.1022 30.5245Z" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M40 61.1953C39.555 61.1953 39.12 61.0634 38.75 60.8161C38.38 60.5689 38.0916 60.2175 37.9213 59.8063C37.751 59.3952 37.7064 58.9428 37.7932 58.5064C37.8801 58.0699 38.0943 57.669 38.409 57.3543C38.7237 57.0397 39.1246 56.8254 39.561 56.7385C39.9975 56.6517 40.4499 56.6963 40.861 56.8666C41.2722 57.0369 41.6236 57.3253 41.8708 57.6953C42.118 58.0653 42.25 58.5003 42.25 58.9453C42.25 59.542 42.0129 60.1143 41.591 60.5363C41.169 60.9583 40.5967 61.1953 40 61.1953Z" fill="white"/>
        </svg>
        <div class="incorrect-page__title">{{t('general.incorrect_page.title')}}</div>
        <div class="incorrect-page__text">{{t('general.incorrect_page.text')}}</div>
      </div>
    </div>
  </div>
</template>
<script setup>
import {useI18n} from "vue-i18n";

const {t} = useI18n();
</script>
<style scoped lang="scss">
@use "../assets/styles/_main.scss" as *;
.app {
  width: 100%;
  min-height: calc(100vh - 1px);
  max-height: 100vh;
  display: flex;
  flex-direction: column;
}

.incorrect-page {
  background: #0A0A0A;
  height: 100vh;
  width: 100%;
  display: flex;

  &__inner {
    height: 100%;
    width: 1000%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0 24px;
    gap: 20px;
  }

  &__title {
    color: #FFF;
    font-size: 22px;
    font-style: normal;
    font-weight: 700;
    line-height: 125%;
    text-align: center;
  }

  &__text {
    color: rgba(255, 255, 255, 0.80);
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }
}
</style>
