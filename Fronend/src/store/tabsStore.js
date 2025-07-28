import { defineStore } from "pinia"
import { ref } from "vue"
import {cloneDeep} from "lodash";
import {useMiniAppStore} from "./miniAppStore.js";

export const useNavTabsStore = defineStore("tabs", () => {
  const miniAppStore = useMiniAppStore();

  const activeTab = ref('main');
  const isSupportOpen = ref(false);
  const isNavBarVisible = ref(true);
  const activeState = ref([]);
  const previousTab = ref(['main']);

  const updateActiveTab = value => {
    if (!miniAppStore.miniAppData) return;

    if (['main', 'favourites'].includes(value)) {
      previousTab.value = [];
    } else {
      previousTab.value.push(cloneDeep(activeTab.value));
    }

    activeTab.value = value;

    activeState.value = [];

    backButtonHandler();
  }

  const toggleSupport = () => {
    isSupportOpen.value = !isSupportOpen.value;
  };

  const toggleNavBarVisibility = () => {
    isNavBarVisible.value = !isNavBarVisible.value
  }

  const pageWithBackButton = [
    'profile',
    'selected_course',
    'language',
    'billing_history',
    'selected_lesson',
    'edit_profile'
  ];

  const backButtonHandler = () => {
    if (pageWithBackButton.includes(activeTab.value)) {
      window?.Telegram?.WebApp?.BackButton?.show();
    } else {
      window?.Telegram?.WebApp?.BackButton?.hide();
    }
  }

  const setActiveState = (val) => {
    activeState.value.push(val);

    backButtonHandler();
  };

  const setPreviousTab = () => {
    if (activeState.value.length) {
      activeState.value.pop();
    } else {
      const removedRoute = previousTab.value.pop();
      activeTab.value = cloneDeep(removedRoute || 'main');
    }

    backButtonHandler();
  }

  return {
    activeTab,
    isSupportOpen,
    isNavBarVisible,
    activeState,

    updateActiveTab,
    toggleSupport,
    toggleNavBarVisibility,
    setActiveState,
    setPreviousTab,
  }
})
