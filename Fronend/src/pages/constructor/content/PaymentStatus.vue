<template>
  <teleport to="#payment">
    <div v-if="miniAppStore.afterPaymentData?.status === 'paid'" class="payment-status">
      <Transition name="slide-up">
        <div v-if="isContentVisible" class="payment-status__content">

          <div class="payment-status__image">
            <img v-show="product?.cover?.length" :src="getFile(product?.cover)" alt="Image">
          </div>

          <div class="payment-status__center">
            <div class="payment-status__title">{{infoData.title}}</div>
            <p class="payment-status__text">{{infoData.text}}</p>
          </div>

          <UIButton @cta="handleButtonClick" :bg="miniAppStore.accentedColor">
            {{ infoData.btn_text}}
          </UIButton>
        </div>
      </Transition>
    </div>
  </teleport>
</template>
<script setup>
import {useMiniAppStore} from "../../../store/miniAppStore.js";
import {computed, ref, watch} from "vue";
import {useI18n} from "vue-i18n";
import {useNavTabsStore} from "../../../store/tabsStore.js";
import {useProductsStore} from "../../../store/productsStore.js";
import UIButton from "../../../components/UI/UIButton.vue";
import {getFile, getProductData, isDatePassed} from "../../../helpers/index.js";

const {t} = useI18n();
const miniAppStore = useMiniAppStore();
const navBarStore = useNavTabsStore();
const productStore = useProductsStore();
const tg = window.Telegram.WebApp;
const isContentVisible = ref(true);
const product = ref(null);

const infoData = computed(() => {
  const obj = {title: "", text: "", btn_text: ""};

  if (!product.value) return obj;

  if (miniAppStore.afterPaymentData?.service === 'wayforpay') {
    let text = t('general.payment_status.text_success');
    text = text.replace('[course name]', product.value.title)

    obj.title = t('general.payment_status.title_success');
    obj.text = text;
    obj.btn_text = t('general.payment_status.button_start_learning');
  } else if (miniAppStore.afterPaymentData?.service === 'ton') {
    let text = t('general.payment_status.text_success_ton');
    text = text.replace('[course name]', product.value.title)

    obj.title = t('general.payment_status.title_success_ton');
    obj.text = text;
    obj.btn_text = t('general.payment_status.button_text_success');
  }


  return obj;
})

const handleButtonClick = async () => {

  if (miniAppStore.afterPaymentData?.service === 'ton') {
    miniAppStore.afterPaymentData = {status: '', product_id: '', service: ''};
    navBarStore.updateActiveTab('main');
  } else if (miniAppStore.afterPaymentData?.service === 'wayforpay') {
    if (!isDatePassed(product.value?.release_date)) {
      miniAppStore.afterPaymentData = {status: '', product_id: '', service: ''};

      return
    }

    miniAppStore.afterPaymentData = {status: '', product_id: '', service: ''};
    await getProductData(product.value?.id);
    navBarStore.updateActiveTab('selected_course');
  }
}

watch(() => miniAppStore.afterPaymentData?.status, (val) => {
  if (val === 'paid') {
    product.value = productStore.allProducts?.find(el => el.id === miniAppStore.afterPaymentData?.product_id) || '';

    setTimeout(() => {
      isContentVisible.value = true;
    }, 50)
  } else {
    product.value = null;
    isContentVisible.value = false;
  }
}, {immediate: true, deep: true})
</script>
<style scoped lang="scss">
@import "../../../assets/styles/main.scss";

.payment-status {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  min-height: 100vh;
  height: 100%;
  background: rgba(0, 0, 0, 0.50);
  backdrop-filter: blur(5px);
  display: flex;
  align-items: flex-end;
  color: #FFF;  z-index: 10000;

  &__content {
    max-height: 100vh;
    overflow-y: scroll;
    width: 100%;
    border-radius: 32px 32px 0 0;
    background: #181818;
    padding: 24px 24px 42px;
    display: flex;
    flex-direction: column;
  }

  &__image {
    height: 86px;
    border-radius: 6.378px;
    overflow: hidden;

    & img {
      height: 100%;
      width: 100%;
      object-fit: cover;

    }
  }

  &__center {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin: 31px 0 24px;
  }

  &__title {
    color: #68E355;
    font-size: 20px;
    font-style: normal;
    font-weight: 600;
    line-height: 125%;
    text-align: center;
  }

  &__text {
    color: #919191;
    font-size: 15px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
    text-align: center;
    white-space: pre-line;
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