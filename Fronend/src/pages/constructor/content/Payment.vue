<template>
  <div class="payment">
    <div class="payment__top">
      <div class="payment__title">{{t('general.payment_status.select_payment_service')}}</div>
      <SVGCross @click="backHandler" />
    </div>

    <div class="payment__items">
      <div
          v-if="miniAppStore.miniAppData?.active_payment_services?.includes('wayforpay')"
          class="payment__item"
          :class="{active: activeMethod === 'wayforpay'}"
          @click="activeMethod = 'wayforpay'"
      >
        <div class="payment__item_logo">
          <SVGWayForPay />
        </div>
        <div class="payment__item_wrap">
          <div class="payment__item_title">WayForPay</div>
          <div class="payment__item_description">
            <SVGPay />
            <SVGVisa />
            <SVGMasterCard />
          </div>
        </div>
      </div>
      <div
          v-if="miniAppStore.miniAppData?.active_payment_services?.includes('ton')"
          class="payment__item"
          :class="{active: activeMethod === 'tonconnect'}"
          @click="activeMethod = 'tonconnect'"
      >
        <div class="payment__item_logo">
          <SVGTon />
        </div>
        <div class="payment__item_wrap">
          <div class="payment__item_title">TON Connect</div>
          <div class="payment__item_description">
            TON, BTC, Blago, ...
          </div>
        </div>
      </div>
    </div>

    <div class="payment__buttons">
      <div
          v-if="activeMethod === 'wayforpay'"
          class="payment__buttons_text"
      >
        {{t('general.payment_status.external_page')}}
      </div>

      <UIButton @cta="continueHandler" :bg="miniAppStore.accentedColor">{{t('general.buttons.continue')}}</UIButton>
      <UIButton @cta="backHandler">{{t('general.buttons.back')}}</UIButton>
    </div>
    <PaymentStatus/>
  </div>
</template>

<script setup>
import UIButton from "../../../components/UI/UIButton.vue";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import SVGWayForPay from "../../../components/svg/SVGWayForPay.vue";
import SVGTon from "../../../components/svg/SVGTon.vue";
import SVGPay from "../../../components/svg/SVGPay.vue";
import SVGVisa from "../../../components/svg/SVGVisa.vue";
import SVGMasterCard from "../../../components/svg/SVGMasterCard.vue";
import SVGCross from "../../../components/svg/SVGCross.vue";
import {ref} from "vue";
import {useWalletStore} from "../../../store/walletStore.js";
import {useNavTabsStore} from "../../../store/tabsStore.js";
import {buyLessonsByTon, buyLessonsByWayForPay} from "../../../api/api.js";
import {useProductsStore} from "../../../store/productsStore.js";
import PaymentStatus from "./PaymentStatus.vue";
import {useI18n} from "vue-i18n";

const {t} = useI18n();
const navTabsStore = useNavTabsStore();
const miniAppStore = useMiniAppStore();
const walletStore = useWalletStore();
const productStore = useProductsStore();
const tg = window.Telegram.WebApp;

const activeMethod = ref(''); // wayforpay, tonconnect, crypto

const backHandler = () => {
  navTabsStore.setPreviousTab()
}

const continueHandler = async () => {
  let resp;

  switch (activeMethod.value) {
    case 'wayforpay':
      resp = await buyLessonsByWayForPay(productStore.selectedPaymentTariff?.id);

      if (resp.data) {
        if (resp.data.payment?.url?.length) {
          tg.openLink(resp.data.payment?.url, "_blank");
        }
      }
      break;
    case 'crypto':
      break;
    case 'tonconnect':
      resp = await buyLessonsByTon(productStore.selectedPaymentTariff?.id);

      await walletStore.transferBLG(resp.data.payment);
      break;
  }
}
</script>

<style scoped lang="scss">
@use "../../../assets/styles/main.scss" as *;

.payment {
  min-height: 100vh;
  background: $black;
  width: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px;
  &__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
    &::v-deep(svg) {
      cursor: pointer;
    }
  }
  &__title {
    color: #FFF;
    font-size: 22px;
    font-weight: 700;
    line-height: 125%;
  }
  &__items {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  &__item {
    cursor: pointer;
    display: flex;
    padding: 20px;
    align-items: stretch;
    gap: 24px;
    border-radius: 12px;
    border: 2px solid rgba(255, 255, 255, 0.12);
    transition: 0.3s;
    &.active {
      border: 2px solid #0061D2;
    }
    &_logo {
      display: flex;
      align-items: center;
      ::v-deep(svg) {
        width: 56px;
        height: 56px;
      }
    }
    &_wrap {
      display: flex;
      flex-direction: column;
      justify-content: space-between;
    }
    &_title {
      color: #FFF;
      font-size: 18px;
      font-weight: 600;
      line-height: 125%;
    }
    &_description {
      color: rgba(255, 255, 255, 0.80);
      font-size: 14px;
      font-weight: 500;
      display: flex;
      align-items: center;
      gap: 8px;
      height: 20px;
    }
  }
  &__buttons {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-top: auto;
    &_text {
      text-align: center;
      color: rgba(255, 255, 255, 0.50);
      font-size: 14px;
      font-weight: 500;
    }
  }
}
</style>
