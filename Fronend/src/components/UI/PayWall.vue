<template>
  <div v-if="isPayWallOpen" class="selected-course__modal_wrapper">
    <Transition name="slide-up">
      <div v-if="isPayWallContentVisible" class="selected-course__modal">
        <div class="selected-course__modal_top">
          <div :style="{background: miniAppStore.accentedColor}" class="selected-course__modal_info">
            {{ getTariffAccess() }}
          </div>
          <button @click="emits('closePayWall')">
            <SVGClose/>
          </button>
        </div>
        <div class="selected-course__modal_product">
          {{ productStore.selectedProduct?.title }}
        </div>
        <div v-if="tariffs?.length === 1" class="selected-course__modal_single">
            <span
                class="selected-course__modal_price">{{
                `${getCurrencySymbol(tariffs[0].currency)} ${tariffs[0].price}`
              }}</span>
          <div class="selected-course__modal_benefits">
            <div class="selected-course__modal_item">
              <SVGCopied/>
              {{ `${t('general.main.tariff_access_to')} ${getTariffText(tariffs[0], 'lessons')}` }}
            </div>
            <div v-if="tariffs[0].bonus?.length" class="selected-course__modal_item">
              <SVGCopied/>
              {{
                `${t('general.main.tariff_access_to')} ${tariffs[0].bonus} ${t('general.main.tariff_bonus_content')}`
              }}
            </div>
          </div>
        </div>
        <template v-else>
          <div class="selected-course__modal_options">
            {{ t('general.main.select_tariff') }}
          </div>
          <div class="selected-course__modal_tariffs">
            <div
                v-for="tariff in tariffs"
                :key="tariff.id"
                @click="handleTariffSelect(tariff)"
                :style="{borderColor: tariff.id === selectedTariff?.id ? selectedTariffBorderStyle : 'rgba(255, 255, 255, 0.12)'}"
                :class="['selected-course__modal_tariff', {active: tariff.id === selectedTariff?.id}]"
            >
              <div class="title-wrapper">
                <div class="selected-course__modal_title">
                  {{ tariff?.name === 'one price' ? productStore.selectedProduct?.title : tariff.name }}
                </div>
                <span
                    class="selected-course__modal_price">{{
                    `${getCurrencySymbol(tariff?.currency)} ${tariff?.price}`
                  }}</span>
              </div>
              <div class="selected-course__modal_benefits">
                <div class="selected-course__modal_item">
                  <SVGCopied/>
                  {{ `${t('general.main.tariff_access_to')} ${getTariffText(tariff, 'lessons')}` }}
                </div>
                <div v-if="tariff.bonus?.length" class="selected-course__modal_item">
                  <SVGCopied/>
                  {{
                    `${t('general.main.tariff_access_to')} ${getTariffText(tariff, 'bonus')} ${t('general.main.tariff_bonus_content')}`
                  }}
                </div>
              </div>
            </div>
          </div>
        </template>
        <UIButton @cta="buyLessonsFunc" :bg="miniAppStore.accentedColor">
          {{ t('general.main.tariff_buy_now') }}
        </UIButton>
        <a v-if="miniAppStore.miniAppData?.tos && miniAppStore.miniAppData?.tos.length"
           :href="getAgreementLink(miniAppStore.miniAppData?.tos[0])" class="selected-course__modal_conditions">
          <svg xmlns="http://www.w3.org/2000/svg" width="17" height="16" viewBox="0 0 17 16" fill="none">
            <path
                d="M6.33333 8L7.77778 9.5L10.6667 6.5M3 8.66667V4.27468C3 4.00711 3.15998 3.76546 3.4063 3.66096L8.23963 1.61046C8.40603 1.53986 8.59397 1.53986 8.76037 1.61046L13.5937 3.66096C13.84 3.76546 14 4.00711 14 4.27469V8.66667C14 11.7042 11.5376 14.1667 8.5 14.1667C5.46243 14.1667 3 11.7042 3 8.66667Z"
                stroke="white" stroke-opacity="0.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          {{ t('general.author.terms_of_use') }}
        </a>
      </div>
    </Transition>
    <PaymentStatus/>
  </div>
</template>
<script setup>
import {getAgreementLink, getCurrencySymbol} from "../../helpers/index.js";
import {computed, ref, watch} from "vue";
import {useMiniAppStore} from "../../store/miniAppStore.js";
import SVGClose from "../svg/SVGClose.vue";
import {useProductsStore} from "../../store/productsStore.js";
import {useI18n} from "vue-i18n";
import UIButton from "./UIButton.vue";
import SVGCopied from "../svg/SVGCopied.vue";
import {buyLessonsByTon, buyLessonsByWayForPay} from "../../api/api.js";
import {useNavTabsStore} from "../../store/tabsStore.js";
import {useWalletStore} from "../../store/walletStore.js";
import PaymentStatus from "../../pages/constructor/content/PaymentStatus.vue";

const props = defineProps({
  isPayWallOpen: Boolean,
})

const emits = defineEmits(["closePayWall"]);

const walletStore = useWalletStore();
const tg = window.Telegram.WebApp;
const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();
const navTabsStore = useNavTabsStore();
const {t} = useI18n();
const isPayWallContentVisible = ref(false);
const selectedTariff = ref(null);

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

const handleTariffSelect = (tariff) => {
  selectedTariff.value = tariff;
  productStore.setSelectedPaymentTariff(tariff);
}

const selectedTariffBorderStyle = computed(() => {
  return "#" + miniAppStore.miniAppData?.color_theme?.accent_color?.split('/')[0];
})

const getTariffText = (tariff, type = 'lessons') => {
  if (!tariff) return;
  const allLessons = productStore.selectedProduct?.lessons;

  if (type === 'lessons') {
    return allLessons.length === tariff.product_level_lessons?.length ? t('general.main.tariff_all_lessons') : `${tariff.product_level_lessons?.length} ${t('general.main.tariff_part_lessons')}`;
  }

  if (type === 'bonus') {
    return tariff.bonus?.length - 1;
  }
}

const getTariffAccess = () => {
  const product = productStore.selectedProduct;

  if (product) {
    if (product.access_time) {
      const access = t('general.main.tariff_access_for');
      const period = Object.keys(product.access_time);
      let finalPeriodText = '';

      if (product.access_time[period[0]] === 1) {
        finalPeriodText = t(`general.main.tariff_period.${period[0]}.${1}`)
      } else if (product.access_time[period[0]] >= 2 && product.access_time[period[0]] <= 4) {
        finalPeriodText = t(`general.main.tariff_period.${period[0]}.${2}`)
      } else if (product.access_time[period[0]] >= 5) {
        finalPeriodText = t(`general.main.tariff_period.${period[0]}.${5}`)
      }

      return access + ' ' + product.access_time[period[0]] + ' ' + finalPeriodText;
    } else {
      return t('general.main.tariff_get_access');
    }
  }

  return ''
}



const buyLessonsFunc = async () => {
  const activePaymentServices = miniAppStore.miniAppData?.active_payment_services || [];

  if (tariffs.value?.length === 1) {
    productStore.setSelectedPaymentTariff(tariffs.value[0]);
    selectedTariff.value = tariffs.value[0];
  }

  if (!selectedTariff.value) return;

  let resp;

  if (activePaymentServices.length > 1) {
    navTabsStore.updateActiveTab('payment');

    return;
  }

  if (activePaymentServices.length === 1) {
    if (activePaymentServices[0] === 'ton') {
      resp = await buyLessonsByTon(selectedTariff.value.id);

      await walletStore.transferBLG(resp.data.payment);
    } else if (activePaymentServices[0] === 'wayforpay') {
      resp = await buyLessonsByWayForPay(selectedTariff.value.id);

      if (resp.data) {
        if (resp.data.payment?.url?.length) {
          tg.openLink(resp.data.payment?.url, "_blank");
        }
      }
    }
  }
}

watch(() => props.isPayWallOpen, (val) => {
  if (val) {
    setTimeout(() => {
      isPayWallContentVisible.value = true;
    }, 50)
  } else {
    isPayWallContentVisible.value = false;
  }
})
</script>
<style scoped lang="scss">
@use "../../assets/styles/main.scss" as *;

.selected-course {
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