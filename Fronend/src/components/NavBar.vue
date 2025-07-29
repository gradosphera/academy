<script setup>
import { useNavTabsStore } from "../store/tabsStore.js"
import {useI18n} from "vue-i18n";
import favourites from "./svg/SVGFavourites.vue";
import home from "./svg/SVGHome.vue";
import support from "./svg/SVGRobot.vue";
import {computed} from "vue";

const navTabsStore = useNavTabsStore()
const {t} = useI18n();

const navTabs = computed(() => {
  return [
    {
      value: 'favourites',
      title: t('general.nav_bar.favorites'),
      icon: favourites,
    },
    {
      value: 'main',
      title: {en: "Хоум", uk: "Хоум"},
      icon: home,
    },
    {
      value: 'support',
      title: t('general.nav_bar.support'),
      icon: support,
    }
  ]
})

const toggleNewTab = tab => {
  if (tab === 'support') {
    navTabsStore.toggleSupport();

    return;
  }

  navTabsStore.updateActiveTab(tab)
}
</script>

<template>
  <div class="nav-bar">
    <ul class="nav-list">
      <li
          v-for="(tab, i) in navTabs"
          :key="i"
          :class="`nav-item ${navTabsStore.activeTab === tab.value ? 'active' : ''}`"
          @click="toggleNewTab(tab.value)"
      >
        <div :class="['nav-icon', {'main': tab.value === 'main'}]">
          <component :is="tab.icon" :class="['svg']"/>
        </div>
        <span v-if="tab.value !== 'main'" class="nav-text">{{ tab.title }}</span>
      </li>
    </ul>
  </div>
</template>

<style scoped lang="scss">
@use "../assets/styles/_main.scss" as *;

.nav {
  &-bar {
    height: 88px;
    position: fixed;
    z-index: 1000;
    bottom: 0;
    width: 100%;
    padding: 8px 0 20px;
    border-top: 1px solid rgba(255, 255, 255, 0.04);
    background: var(--Neutral-01, #0a0a0a);
    backdrop-filter: blur(20px);
    user-select: none;
  }

  &-list {
    display: flex;
    justify-content: space-around;
    width: 100%;
  }

  &-item {
    cursor: pointer;
    display: flex;
    gap: 4px;
    flex-direction: column;
    align-items: center;

    &.active {
      & .nav-text {
        color: #FFF;
      }

      & .nav-icon {
        ::v-deep(.svg) {
          path {
            stroke: #FFF;
          }
        }
      }
    }
  }

  &-icon {
    display: flex;
    justify-content: center;
    align-items: center;

    &.main {
      position: relative;
      min-width: 44px;
      max-width: 44px;
      min-height: 44px;
      max-height: 44px;
      border-radius: 50%;
      background-color: rgba(255, 255, 255, 0.12);
      transform: translateY(-2px);

      ::v-deep(.svg) {
        path {
          stroke: transparent !important;
          fill: #FFF;
        }
      }
    }
  }

  &-text {
    color: #7f7f7f;
    font-size: 11px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }
}
</style>
