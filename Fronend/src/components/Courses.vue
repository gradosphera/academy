<template>
  <div class="courses">
    <p class="courses_title">{{ t('general.main.products') }}</p>
    <div v-if="!productsStore.availableProducts?.length" class="courses_empty_list">
      <SVGEmptyProducts/>
      <span>{{t('general.main.product_empty_list_title')}}</span>
      <p>{{t('general.main.product_empty_list')}}</p>
    </div>
    <div v-show="productsStore.availableProducts?.length" ref="mainProductsRef" class="courses_list">
      <template v-if="isMediaLoading">
        <ProductSkeleton v-for="product in productsStore.availableProducts" :key="product.id" />
      </template>
      <template v-else>
        <div v-for="course in productsStore.availableProducts" :key="course.id" class="courses_course course"
             @click="openCourse(course, course?.release_date)">
        <div v-if="productStatusFunc(course)" class="course_product_date">
          <component v-if="productStatusFunc(course).svg" :is="productStatusFunc(course).svg" />
          <span>{{ productStatusFunc(course).text }}</span>
        </div>
        <div v-if="course?.deleted_at" class="course_ban"></div>
        <div class="course_image" :class="productStatusFunc(course)?.type">
          <img v-if="course.cover?.length" :src="getFile(course.cover)" alt=""/>
        </div>
        <div class="course_body">
          <div class="course_type">{{ "#" + t('general.main.product_type') }}</div>
          <p class="course_title">{{ course.title }}</p>
          <div class="course_bottom">
            <div v-if="course?.product_levels?.length" class="course_item price">
              <div class="course_icon">
                <SVGCoin/>
              </div>
              {{ formatPriceRange(course.product_levels)}}
            </div>
            <div v-else class="course_item">
              <div class="course_icon">
                <SVGTag/>
              </div>
              {{ t('general.main.product_free') }}
            </div>
            <button v-if="productStatusFunc(course)?.type !== 'ban' && productStatusFunc(course)?.type !== 'release_date'" class="course_explore">
              {{ course?.deleted_at ? t('general.main.product_view_disabled') : t('general.main.start_product') }}
              <SVGArrowRightLong v-if="!course?.deleted_at"/>
            </button>
          </div>
        </div>
        </div>
      </template>
    </div>

    <Modal :is-open="modalDetails.is_open" position="bottom" @close-modal="closeModal" :title="modalDetails.title">
      <template #bottom>
        <ProductBanNotification :product-name="modalDetails.course_name" @close="closeModal" />
      </template>
    </Modal>
    <PaymentStatus />
  </div>
</template>
<script setup>
import {useNavTabsStore} from "../store/tabsStore.js"
import {useI18n} from "vue-i18n";
import {useProductsStore} from "../store/productsStore.js";
import SVGCoin from "./svg/SVGCoin.vue";
import SVGTag from "./svg/SVGTag.vue";
import SVGArrowRightLong from "./svg/SVGArrowRightLong.vue";
import {formatPriceRange, formatProductDate, getFile, getProductData, isDatePassed} from "../helpers/index.js";
import SVGCalendar from "./svg/SVGCalendar.vue";
import {onMounted, ref} from "vue";
import SVGLock from "./svg/SVGLock.vue";
import Modal from "./UI/Modal.vue";
import ProductSkeleton from "./Skeletons/ProductSkeleton.vue";
import {useMediaLoader} from "../composable/useMediaLoader.js";
import ProductBanNotification from "./UI/ProductBanNotification.vue";
import SVGEmptyProducts from "./svg/SVGEmptyProducts.vue";
import PaymentStatus from "../pages/constructor/content/PaymentStatus.vue";

const {t, locale} = useI18n();
const productsStore = useProductsStore();
const navTabsStore = useNavTabsStore();
const {isMediaLoading, waitForMediaLoad} = useMediaLoader();
const mainProductsRef = ref(null);
const modalDetails = ref({
  is_open: false,
  type: '',
  title: '',
  text: '',
  btn_text: '',
  course_name: ''
})

const closeModal = () => modalDetails.value.is_open = false;

const openCourse = async (course, date) => {
  if (!isDatePassed(date)) return;
  if (course?.deleted_at) {
    modalDetails.value.is_open = true;
    modalDetails.value.type = 'ban';
    modalDetails.value.title = t('general.main.access_removed');
    modalDetails.value.text = t('general.main.ban_text');
    modalDetails.value.btn_text = t('general.main.got_it');
    modalDetails.value.course_name = course.title;

    return;
  }

  await getProductData(course.id);
  navTabsStore.updateActiveTab('selected_course');
}

const productStatusFunc = (product) => {
  if (!product) return null;

  const obj = {id: '0', svg: null, text: '', type: ''};

  if (isProductBan(product)) {
    obj.text = t('general.main.access_removed');
    obj.svg = SVGLock;
    obj.type = 'ban'
    obj.id = product.id;

    return obj;
  } else if (product?.release_date && !isDatePassed(product?.release_date)) {
    obj.text = formatProductDate(product.release_date, locale.value);
    obj.svg = SVGCalendar;
    obj.id = product.id;
    obj.type = 'release_date'

    return obj;
  }

  return null;
}

const isProductBan = (product) => {
  const findProduct = productsStore.productsAccess?.find((item) => item.product_id === product.id);

  if (findProduct) {
    return findProduct.deleted_at;
  }
}

onMounted(async() => {
  await waitForMediaLoad(mainProductsRef.value)
})
</script>
<style scoped lang="scss">
.courses {
  height: 100%;
  &_title {
    margin-bottom: 16px;
    color: #CECECE;
    font-size: 17px;
    font-style: normal;
    font-weight: 500;
    line-height: 125%;
  }

  &_empty_list {
    height: calc(100% - 88px);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;

    & span {
      display: block;
      margin: 20px 0 16px;

      color: #FFF;
      font-size: 18px;
      font-style: normal;
      font-weight: 600;
      line-height: 120%;
    }

    & p {
      text-align: center;
      color: rgba(255, 255, 255, 0.70);
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.16px;
    }
  }

  &_list {
    position: relative;
    z-index: 2;
    display: flex;
    flex-direction: column;
    gap: 16px;

    .course {
      position: relative;
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.07);
      backdrop-filter: blur(12px);
      cursor: pointer;
      overflow: hidden;

      &_ban {
        position: absolute;
        z-index: 5;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;

        border-radius: 16px;
        opacity: 0.5;
        background: rgba(0, 0, 0, 0.5);
        backdrop-filter: blur(12px);
      }

      &_image {
        position: relative;
        height: 86px;
        background: rgba(255, 255, 255, 0.12);

        & img {
          height: 100%;
          width: 100%;
          object-fit: cover;
        }

        &.release_date {
          opacity: .5;
        }
      }

      &_product_date {
        position: absolute;
        z-index: 6;
        top: 8px;
        right: 10px;
        padding: 4px 8px;
        display: flex;
        gap: 4px;
        background-color: #FFF;
        border-radius: 10px;
        align-items: center;

        & span {
          color: #0A0A0A;

          font-size: 11px;
          font-style: normal;
          font-weight: 500;
          line-height: normal;
        }

        & svg {
          width: 12px;
          height: 12px;

          ::v-deep(path) {
            stroke: #0A0A0A;
          }
        }
      }

      &_body {
        padding: 16px;
      }

      &_type {
        margin-bottom: 8px;
        display: inline-block;
        border-radius: 10px;
        background: #0f0f0f;
        color: #a5a5a5;
        padding: 4px 6px;
        font-size: 11px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
      }

      &_title {
        margin-bottom: 30px;
        color: var(--White, var(--system-white, #fff));
        font-size: 15px;
        font-style: normal;
        font-weight: 600;
        line-height: 120%; /* 18px */
      }

      &_bottom {
        margin-top: 15px;
        display: flex;
      }

      &_item {
        display: flex;
        gap: 4px;

        color: #FFF;

        font-size: 14px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;

        &.price {
          margin-right: 25px;
        }
      }

      &_icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 20px;
        height: 20px;
        border-radius: 50%;
        background: rgba(255, 255, 255, 0.07);

        & svg {
          width: 70%;
        }
      }

      &_explore {
        margin-left: auto;
        background-color: transparent;
        color: rgba(255, 255, 255, 0.70);
        font-size: 13px;
        font-style: normal;
        font-weight: 500;
        line-height: normal;
        display: flex;
        align-items: center;
        gap: 4px;
      }
    }
  }

  &_modal {
    margin-top: 16px;
    & p {
      margin-bottom: 24px;

      color: rgba(255, 255, 255, 0.80);
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: 130%;
      letter-spacing: 0.16px;
    }
  }
}
</style>
