<template>
  <div class="billing-history">
    <span class="billing-history__title">{{t('general.profile.menu.billing_history.title_inside')}}</span>

    <div v-if="!miniAppStore.payments?.length" class="billing-history__empty">
      <h3>{{t('general.profile.menu.billing_history.empty_state')}}</h3>
      <span>{{t('general.profile.menu.billing_history.empty_state_text')}}</span>
    </div>

    <div v-else class="billing-history__list">
      <div v-for="(bill, i) in miniAppStore.payments" :key="i" class="billing-history__content_block">
        <div class="billing-history__bill_top">
          <span class="billing-history__bill_title">{{getProductTitle(bill.product_level)}}</span>
          <span class="billing-history__bill_amount">{{getBillAmount(bill)}}</span>
        </div>
        <div class="billing-history__bill_center">
          <div class="billing-history__bill_date">
            {{formatDate(bill.updated_at).full_date}}
            <div class="divider"></div>
            {{formatDate(bill.updated_at).time}}
          </div>
          <div :class="['billing-history__bill_status', bill.status]">{{getBillingStatusLocale(bill.status)}}</div>
        </div>
        <UIButton @cta="tryAgain(bill)" v-if="bill.status === 'failed'" class="billing-history__bill_btn" :bg="miniAppStore.accentedColor">{{t('general.buttons.try_again')}}</UIButton>
      </div>
    </div>
  </div>
</template>
<script setup>
import UIButton from "../../../components/UI/UIButton.vue";
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {onMounted} from "vue";
import {getPayments} from "../../../api/api.js";
import {useToastStore} from "../../../store/toastStore.js";
import {useProductsStore} from "../../../store/productsStore.js";
import {useI18n} from "vue-i18n";
import {getCurrencySymbol} from "../../../helpers/index.js";

const {t} = useI18n();
const tg = window.Telegram.WebApp
const toastStore = useToastStore();
const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();

const getProductTitle = (level) => {
  if (level) {
    const product = productStore.allProducts?.find((product) => product.id === level.product_id);

    if (product) {
      if (level.name === 'one price') {
        return product.title
      }

      return product.title + '. ' + level.name;
    }
  }

  return '';
}

const getBillingStatusLocale = (status) => {
  switch (status) {
    case 'failed':
      return t('general.profile.billing_history.failed');
    case 'pending':
      return t('general.profile.billing_history.pending');
    case 'completed':
      return t('general.profile.billing_history.completed');
  }
}

const getBillAmount = (bill) => {
  const currencySymbol = getCurrencySymbol(bill.currency);

  return currencySymbol + bill.amount;
}

const tryAgain = async(tariff) => {
  if (tariff?.product_level) {
  }
}

const formatDate = (date) => {
  if (date) {
    const dateFull = new Date(date);
    const obj = {full_date: '', time: ''}

    const hours = String(dateFull.getHours()).padStart(2, '0');
    const minutes = String(dateFull.getMinutes()).padStart(2, '0');
    obj.time = `${hours}:${minutes}`;

    let day = String(dateFull.getDate());
    day = day.length === 1 ? '0' + day : day;

    let month = String(dateFull.getMonth() + 1);
    month = month.length === 1 ? '0' + month : month;

    const year = dateFull.getFullYear();

    obj.full_date = `${day}.${month}.${year}`

    return obj;
  }

  return '';
}
onMounted(async() => {
  try {
    const respPayments = await getPayments({status: [], limit: 0, offset: 0});

    if (respPayments.data) miniAppStore.setPayments(respPayments.data.payments);
  } catch (error) {
    toastStore.error({text: t('general.toast_notifications.something_went_wrong')});
  }
})
</script>
<style scoped lang="scss">
.billing-history {
  color: #fff;
  background: #0a0a0a;
  height: calc(100vh - 1px);
  overflow-y: auto;
  padding: 16px 24px 100px;
  min-height: 100%;

  &__content_block {
    padding: 16px;
    border-radius: 16px;
    background: rgba(255, 255, 255, 0.07);
    backdrop-filter: blur(12px);
    width: 100%;
  }

  &__empty {
    position: absolute;
    width: 100%;
    top: calc(50% - 44px);
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 0 24px;

    & h3 {
      color: #FFF;
      font-size: 18px;
      font-style: normal;
      font-weight: 500;
      line-height: 120%;
      margin: 20px 0 16px;
    }

    & span {
      text-align: center;
      color: rgba(255, 255, 255, 0.70);
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.17px;
    }
  }

  &__title {
    color: rgba(255, 255, 255, 0.50);
    font-size: 14px;
    font-style: normal;
    font-weight: 500;
    line-height: normal;
  }

  &__list {
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  &__bill {
    &_center {
      display: flex;
      gap: 8px;
      justify-content: space-between;
      align-items: center;
      margin: 16px 0 0;
    }

    &_status {
      height: 24px;
      padding: 0 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 10px;
      background: rgba(255, 156, 24, 0.07);

      color: #FF9C18;
      font-size: 13px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
      text-transform: capitalize;

      &.completed {
        color: #68E355;
        background: rgba(104, 227, 85, 0.07);
      }

      &.failed {
        color: #FA4851;
        background: rgba(250, 72, 81, 0.07);
      }
    }

    &_top {
      display: flex;
      align-items: center;
      justify-content: space-between;
      color: #FFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: 125%;
    }

    &_title {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 240px;
    }

    &_date {
      color: #919191;
      font-size: 13px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
      display: flex;
      align-items: center;
      gap: 6px;

      & .divider {
        max-height: 4px;
        min-height: 4px;
        aspect-ratio: 1/1;
        border-radius: 50%;
        background: #919191;
      }
    }

    &_btn {
      margin-top: 20px;
    }
  }
}
</style>