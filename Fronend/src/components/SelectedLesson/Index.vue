<template>
  <div ref="scrollableContainer" class="selected-lesson">
    <div class="selected-lesson_container">
      <SkeletonOpenedLesson v-if="isMediaLoading" />
      <div v-show="!isMediaLoading">
        <div class="selected-lesson_nav-bar nav-bar">
          <ul v-if="productStore.lessonTabs.length > 1" class="nav-bar_list">
            <template v-for="(tab, i) in navTabs" :key="tab.value">
              <li
                  v-if="productStore.lessonTabs.includes(tab.value)"
                  :class="`nav-bar_item ${productStore.selectedLessonTab === tab.value ? 'active' : ''}`"
                  @click="toggleNewLessonTab(tab.value)"
              >
                <p class="nav-bar_text">{{ tab.title }}</p>
              </li>
            </template>
          </ul>
        </div>
        <component :is="lessonContent[productStore.selectedLessonTab]" />
        <div class="selected-lesson__btns">
          <LessonButtonsControl @open-modal="openModal" @open-pay-wall="openPayWall"/>
        </div>
      </div>
      <Modal :is-open="isModalOpen" @close-modal="isModalOpen = false" :title="t('general.main.rate_the_lesson')">
        <template #default>
          <div class="selected-lesson__modal">
            <div class="selected-lesson__star-wrapper">
              <SVGEmptyStar @click="ratedStar = i + 1" v-for="(_, i) in 5" :key="i" :class="['selected-lesson__star', {rated: i + 1 <= ratedStar}]"/>
            </div>
            <div class="selected-lesson__modal_btns">
              <UIButton @cta="isModalOpen = false" bg="rgba(255, 255, 255, 0.12)">{{t('general.buttons.cancel')}}</UIButton>
              <UIButton @cta="rateLessonFunc" :bg="miniAppStore.accentedColor">{{t('general.buttons.confirm')}}</UIButton>
            </div>
          </div>
        </template>
      </Modal>
      <PayWall :is-pay-wall-open="isPayWallOpen" @close-pay-wall="closePayWall"/>
    </div>
  </div>
</template>
<script setup>
import {onMounted, onUnmounted, onUpdated, ref, watch} from "vue"
import Content from "./Content/Index.vue"
import Materials from "./Materials.vue"
import Homework from "./Homework.vue"
import {useI18n} from "vue-i18n";
import {useProductsStore} from "../../store/productsStore.js";
import LessonButtonsControl from "./LessonButtonsControl.vue";
import Modal from "../UI/Modal.vue";
import SVGEmptyStar from "../svg/SVGEmptyStar.vue";
import UIButton from "../UI/UIButton.vue";
import {useMiniAppStore} from "../../store/miniAppStore.js";
import {rateLesson} from "../../api/api.js";
import {getProductData, submitAnalyticsData} from "../../helpers/index.js";
import SkeletonOpenedLesson from "../Skeletons/SkeletonOpenedLesson.vue";
import {useMediaLoader} from "../../composable/useMediaLoader.js";
import PayWall from "../UI/PayWall.vue";
import {useNavTabsStore} from "../../store/tabsStore.js";

const {t} = useI18n();
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();
const isPayWallOpen = ref(false);
const navTabs = [
  {
    value: 'content',
    title: t("general.lessonTabs.content"),
  },
  {
    value: 'materials',
    title: t("general.lessonTabs.materials"),
  },
  {
    value: 'homework',
    title: t("general.lessonTabs.homework"),
  },
]

const productStore = useProductsStore();
const miniAppStore = useMiniAppStore();
const ratedStar = ref(0);
const isModalOpen = ref(false);
const navTabsStore = useNavTabsStore();

const toggleNewLessonTab = tab => {
  productStore.setLessonTab(tab);
}

const lessonContent = {
  'content': Content,
  'materials': Materials,
  'homework': Homework,
}

const scrollableContainer = ref(null)

const rateLessonFunc = async() => {
  const score = 10000 / 5 * ratedStar.value;
  const data = {score, text: ''};

  await rateLesson(productStore.selectedLesson?.id, data);
  productStore.isLessonReviewed = true;
  isModalOpen.value = false;
  ratedStar.value = 0;

  await getProductData(productStore.selectedProduct?.id);
}

const openModal = () => {
  isModalOpen.value = true;
}

const openPayWall = () => {
  isPayWallOpen.value = true;
  navTabsStore.isNavBarVisible = false;
}

const closePayWall = () => {
  isPayWallOpen.value = false;
  navTabsStore.isNavBarVisible = true;

  submitAnalyticsData('paywall_closed', {
    source: 'next_lesson_button',
    course_title: productStore.selectedProduct?.title || '',
  })
}

watch(() => productStore.selectedLesson, () => {
  if (productStore.selectedLesson && scrollableContainer.value) {
    scrollableContainer.value.scrollTo({ top: 0, behavior: "smooth" })
  }

  if (productStore.selectedLesson) {
    const isLessonCompleted = productStore.productProgress?.find(item => item.lesson_id === productStore.selectedLesson?.id);

    if (isLessonCompleted && !productStore.isLessonReviewed) {
      isModalOpen.value = true;
    }
  }
})

onMounted(async() => {
  await waitForMediaLoad(scrollableContainer.value);

  productStore.setLessonTab(navTabs[0].value);
})

onUpdated(() => {
  const container = document.querySelector(".selected-lesson_container");
  if (container) {
    container.scrollTo({top: 0, behavior: 'smooth'});
  }
})

onUnmounted(() => {
  productStore.resetLessonData();
})
</script>
<style scoped lang="scss">
.selected-lesson {
  width: 100%;
  height: calc(100vh - 1px);
  overflow-y: auto;
  overflow-x: hidden;
  background: var(--Neutral-01, #0a0a0a);
  padding: 16px 24px 90px;
  position: relative;
  z-index: 6;

  &_container {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: scroll;
  }

  &__btns {
    margin-top: auto;
    position: absolute;
    bottom: 104px;
    width: 100%;

    height: 48px;
    left: 0;
    padding: 0 24px;
    z-index: 5;
  }

  &__modal {
    display: flex;
    flex-direction: column;
    gap: 32px;
    margin-top: 32px;

    &_btns {
      display: flex;
      gap: 8px;
    }
  }

  &__star {
    cursor: pointer;
    &.rated {
      ::v-deep(path) {
        fill: #D6B72E;
        fill-opacity: 1;
      }
    }

    &-wrapper {
      display: flex;
      gap: 16px;
      margin: 0 auto;
    }
  }

  .nav-bar {
    position: relative;
    z-index: 2;
    margin-bottom: 24px;
    border-radius: 8px;
    background: rgba(118, 118, 128, 0.2);

    &_list {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 2px;
    }

    &_item {
      cursor: pointer;
      display: flex;
      justify-content: center;
      width: 100%;

      &.active {
        border-radius: 7px;
        border: 0.5px solid rgba(0, 0, 0, 0.04);
        background: #454548;
      }
    }

    &_text {
      text-align: center;
      padding: 6px;
      width: 100%;
      color: var(--Label-Color-Dark-Primary, var(--system-white, #fff));
      font-size: 13px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }
  }
}
</style>
