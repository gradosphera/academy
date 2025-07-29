<template>
  <div class="selected-course">
    <div class="selected-course_container">
      <SelectedCourseHead :locale="localeLang"/>
      <SelectedCourseChapters @close-modal="closePaymentModal" @open-modal="openPaymentModal"/>
      <SelectedCourseBonuses :data="bonuses"/>
    </div>
    <PayWall
        :is-pay-wall-open="isPaymentModalOpen"
        @close-pay-wall="closePaymentModal"
    />
  </div>
</template>
<script setup>
import {computed, onUnmounted, ref} from "vue"
import {useNavTabsStore} from "../../store/tabsStore"
import SelectedCourseHead from "./Head.vue"
import SelectedCourseChapters from "./Chapters.vue"
import {useSelectedCourse} from "../../store/selectedCourse.js";
import {useProductsStore} from "../../store/productsStore.js";
import SelectedCourseBonuses from "./SelectedCourseBonuses.vue";
import PayWall from "../UI/PayWall.vue";
import {submitAnalyticsData} from "../../helpers/index.js";


const navTabsStore = useNavTabsStore();
const productStore = useProductsStore();
const courseStore = useSelectedCourse();
const isPaymentModalOpen = ref(false);

const tariffs = computed(() => {
  const allTariffs = productStore.selectedProduct?.product_levels || [];

  if (productStore.paidLessons && productStore.paidLessons.length) {

    return allTariffs.filter(elem => {
      let isTariffPaid = false;

      for (let i = 0; i < elem?.product_level_lessons?.length; i++) {
        const findLessonInPaidArr = productStore.paidLessons.find(item => item.LessonID === elem?.product_level_lessons[i].lesson_id);

        if (findLessonInPaidArr) {
          isTariffPaid = true;
          break;
        }
      }

      return !isTariffPaid;
    });
  }

  return allTariffs;
})

const bonuses = computed(() => {
  const arr = [];

  tariffs.value.forEach(item => {
    const data = {title: '', bonuses: []};

    if (item.bonus?.length) {
      for (let i = 0; i < item.product_level_lessons?.length; i++) {
        const isLessonUnlocked = productStore.paidLessons.find(el => el.LessonID === item.product_level_lessons[i].lesson_id);

        if (isLessonUnlocked) {
          const titleObject = item.bonus.find(val => val.description === 'bonus title');


          if (titleObject) {
            data.title = titleObject.title;
          }

          data.bonuses = item.bonus.filter(item => item.id !== titleObject.id);
          arr.push(data);

          break;
        }
      }
    }
  })

  return arr;
})

const localeLang = computed(() => {
  const lang = localeLang.value;

  if (courseStore.selectedCourse) {
    return courseStore.selectedCourse.language[lang];
  }
})

const openPaymentModal = () => {
  isPaymentModalOpen.value = true;
  navTabsStore.isNavBarVisible = false;

  submitAnalyticsData('paywall_shown', {
    source: 'lesson_open',
    course_title: productStore.selectedProduct?.title || '',
  })
};
const closePaymentModal = () => {
  isPaymentModalOpen.value = false;
  navTabsStore.isNavBarVisible = true;

  submitAnalyticsData('paywall_closed', {
    source: 'lesson_open',
    course_title: productStore.selectedProduct?.title || '',
  })
};

onUnmounted(() => {
  navTabsStore.isNavBarVisible = true;
})
</script>
<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.selected-course {
  height: calc(100vh - 1px);
  overflow-y: auto;
  overflow-x: hidden;
  position: relative;
  background: var(--Neutral-01, #0a0a0a);
  padding-bottom: 100px;
  z-index: 6;

  &__modal {
    max-height: 100vh;
    overflow-y: scroll;
    width: 100%;
    border-radius: 32px 32px 0 0;
    background: #181818;
    padding: 24px 24px 42px;

    &_wrapper {
      position: fixed;
      bottom: 0;
      width: 100%;
      left: 0;
      z-index: 10;
      min-height: 100vh;
      background: rgba(0, 0, 0, 0.40);
      backdrop-filter: blur(2px);
      display: flex;
      align-items: flex-end;
      color: #FFF;
    }

    &_tariffs {
      margin-top: 16px;
      margin-bottom: 30px;
      display: flex;
      flex-direction: column;
      gap: 12px;
    }

    &_tariff {
      display: flex;
      flex-direction: column;
      gap: 24px;
      padding: 24px;
      border-radius: 12px;
      border: 1px solid;

      &.active {
        border-width: 2px;
      }

      & .title-wrapper {
        display: flex;
        justify-content: space-between;
        gap: 10px;
      }
    }

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;

      & button {
        width: 24px;
        height: 24px;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.07);
      }
    }

    &_info {
      height: 23px;
      border-radius: 100px;
      padding: 0 8px;
      display: flex;
      align-items: center;

      color: #FFF;
      font-size: 11px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_single {
      margin-top: 16px;

      & .selected-course {
        &__modal_benefits {
          margin: 24px 0 32px;
        }
      }
    }

    &_product {
      color: #FFF;
      font-size: 22px;
      font-style: normal;
      font-weight: 700;
      line-height: 125%;
      margin: 16px 0 0;
    }

    &_options {
      color: rgba(255, 255, 255, 0.50);
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 125%;
      margin-top: 32px;
    }

    &_title {
      color: #FFF;
      font-size: 18px;
      font-style: normal;
      font-weight: 600;
      line-height: 125%;
    }

    &_price {
      color: #FFF;
      font-size: 18px;
      font-style: normal;
      font-weight: 600;
      line-height: 125%;
    }

    &_benefits {
      display: flex;
      flex-direction: column;
      gap: 14px;
    }

    &_item {
      display: flex;
      align-items: center;
      gap: 8px;
      color: #FFF;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    &_conditions {
      margin-top: 15px;
      display: flex;
      align-items: center;
      justify-content: center;

      color: rgba(255, 255, 255, 0.50);
      font-size: 13px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }
  }
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.6s ease;
}

.slide-up-enter-from {
  transform: translateY(100%);
}

.slide-up-enter-to {
  transform: translateY(0);
}

.slide-up-leave-from {
  transform: translateY(0);
}

.slide-up-leave-to {
  transform: translateY(100%);
}
</style>
