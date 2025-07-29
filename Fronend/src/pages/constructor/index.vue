<template>
  <div class="constructor">
    <div class="constructor_content">
      <Loader @loaderComplete="completedLoading" v-if="!isLoaded" />
      <div v-else>
        <BackgroundBlueLayout />
        <Transition name="slide">
          <div class="constructor_pages" v-show="!authorStore.isShowAuthor">
            <Component :is="content[activeTab]" :key="activeTab" />
            <Support/>
            <NavBar v-if="isNavBarShown" />
          </div>
        </Transition>
      </div>
    </div>
  </div>
</template>
<script setup>
import {ref, computed} from "vue"

import Loader from "../../components/Loader.vue"
import NavBar from "../../components/NavBar.vue"

import Main from "./content/Main.vue"
import Payment from "./content/Payment.vue"
import Calendar from "./content/Calendar.vue"
import HomeWork from "./content/HomeWork.vue"
import Favourites from "./content/Favourites.vue"
import Notifications from "./content/Notifications.vue"
import Profile from "./content/Profile.vue"
import SelectedCourse from "../../components/SelectedCourse/Index.vue"
import SelectedLesson from "../../components/SelectedLesson/Index.vue"
import SelectedHomework from "../../components/SelectedHomework/Index.vue"
import BackgroundBlueLayout from "../../components/BackgroundBlueLayout.vue"

import { useNavTabsStore } from "../../store/tabsStore.js"
import { useAuthorStore } from "../../store/authorStore.js"
import Tutorial from "./content/Tutorial.vue";
import Support from "./content/Support.vue";
import BillingHistory from "./content/BillingHistory.vue";
import Language from "./content/Language.vue";
import {useRoute} from "vue-router";
import EditProfile from "../../components/Profile/EditProfile.vue";
import PaymentStatus from "./content/PaymentStatus.vue";

const route = useRoute();
const authorStore = useAuthorStore()
const navTabsStore = useNavTabsStore()
const activeTab = computed(() => navTabsStore.activeTab)

const isLoaded = ref(false)
const completedLoading = () => {
  isLoaded.value = true
}

const pagesWithoutNavBar = [
    'tutorial',
    'payment_status',
    'payment',
]

const isNavBarShown = computed(() => {
  return !pagesWithoutNavBar.includes(navTabsStore.activeTab) && navTabsStore.isNavBarVisible;
})

const content = {
  'main': Main,
  'calendar': Calendar,
  'homework': HomeWork,
  'favourites': Favourites,
  'selected_course': SelectedCourse,
  'selected_lesson': SelectedLesson,
  'notifications': Notifications,
  'selected_homework': SelectedHomework,
  'profile': Profile,
  'tutorial': Tutorial,
  'support': Main,
  'billing_history': BillingHistory,
  'language': Language,
  'edit_profile': EditProfile,
  'payment_status': PaymentStatus,
  'payment': Payment,
}
</script>
<style scoped lang="scss">
@use "../../assets/styles/_main.scss" as *;
.constructor {
  width: 100%;
  min-height: calc(100vh - 1px);
  max-height: 100vh;
  display: flex;
  flex-direction: column;

  &_content {
    position: relative;
    width: 100%;
    flex: 1;
  }

  &_pages {
    position: relative;
  }

}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.7s ease;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateY(100%); /* Move off-screen */
}

.slide-enter-to,
.slide-leave-from {
  transform: translateY(0); /* Bring back on-screen */
}
</style>
