<template>
  <div class="loader">
    <div class="loader_container">
      <div v-if="isLoaderVisible" class="loader_logo-wrapper">
        <img :src="getLogo" class="loader_logo-animation" alt="">
      </div>
    </div>
    <BgBlurSquare/>
  </div>
</template>
<script setup>
import {computed, getCurrentInstance, onMounted, ref} from "vue"
import {useNavTabsStore} from "../store/tabsStore.js"
import BgBlurSquare from "./BgBlurSquare.vue"
import {getFile} from "../helpers/index.js";
import {useMiniAppStore} from "../store/miniAppStore.js";
import {getMe, getMiniApp, signIn} from "../api/api.js";
import {useProductsStore} from "../store/productsStore.js";
import {useToken} from "../composable/useToken.js";
import CookieManager from "../helpers/cookieManager.js";
import {useRoute, useRouter} from "vue-router";
import {useToastStore} from "../store/toastStore.js";
import {useI18n} from "vue-i18n";
import defaultLogo from "../assets/images/default_logo.webp"
import {createGtag} from 'vue-gtag'
import {initFacebookPixel} from "../helpers/facebookPixel.js";
import { createGtm } from '@gtm-support/vue-gtm';

const emit = defineEmits(["loaderComplete"])
const navTabsStore = useNavTabsStore();
const miniAppStore = useMiniAppStore();
const productStore = useProductsStore();
const toastStore = useToastStore();
const tg = window.Telegram.WebApp;
const initData = tg.initData;
const {setToken} = useToken();
const route = useRoute();
const isLoaderVisible = ref(false);
const {locale} = useI18n();
const router = useRouter();
const app = getCurrentInstance().appContext.app;

const toggleTab = value => {
  navTabsStore.updateActiveTab(value)
}

const getLogo = computed(() => {
  if (miniAppStore.miniAppData?.logo?.length) {
    return getFile(miniAppStore.miniAppData?.logo)
  }

  return defaultLogo;
})

const getAppData = async () => {
  const start_param = tg.initDataUnsafe.start_param;

  try {
    const miniAppResp = await getMiniApp();
    if (miniAppResp.data) {
      miniAppStore.setMiniAppData(miniAppResp.data.mini_app);
      miniAppStore.setTutorialData(miniAppResp.data.mini_app?.slides)
      productStore.setAllProducts(miniAppResp.data.mini_app.products);
    }

    isLoaderVisible.value = true;

    if (!miniAppStore.userData) {
      const userDataResp = await getMe();

      if (userDataResp.data) {
        miniAppStore.setUserData(userDataResp.data.user)
      }
    }

    const newDate = new Date();
    const ISOStandard = newDate.toISOString();
    const timeUTC = ISOStandard.split('T')[1].slice(0, 5);

    const miniAppLang = miniAppResp.data.mini_app.language;
    const userLanguage = miniAppStore.userData?.language;

    if (userLanguage?.length) {
      locale.value = userLanguage;
    } else if (miniAppLang?.length) {
      locale.value = miniAppLang;
    }

    const tutorial = localStorage.getItem("tutorial");

    if (start_param && start_param.startsWith('product_name')) {
      miniAppStore.setPaymentStatus(start_param);
    } else if (tutorial || !miniAppResp.data?.mini_app?.slides?.length) {
      toggleTab('main');
    } else {
      toggleTab('tutorial');
    }

    if (miniAppResp.data?.mini_app?.analytics) {
      const analytics = miniAppResp.data?.mini_app?.analytics;

      if (analytics.facebook_pixel?.length) {
        await initFacebookPixel(analytics.facebook_pixel);
      }

      if (analytics.google_analytics?.length) {
        const gtag = createGtag({
          tagId: analytics.google_analytics
        })

        app.use(gtag);
      }

      if (analytics.g_tag?.length) {
        app.use(createGtm({
          id: analytics.g_tag,
          debug: false, // Set to false in production
          vueRouter: null, // If you have Vue Router, pass your router instance here: vueRouter: router
          loadScript: true, // Recommended to load the GTM script
          defer: false, // Set to true to defer script loading
          compatibility: false, // Set to true for Vue 2 compatibility mode if needed
          enabled: true, // Set to false to disable GTM entirely (e.g., in development without a GTM ID)
          // trackOnNextTick: false, // If you're using Vue Router, set to true if you want to track page views after the next tick
        }))
      }
    }

  } catch (e) {
    if (e.response?.data === 'mini app not found') {
      await submitAuth();
    }
  } finally {
    setTimeout(() => {
      emit("loaderComplete")
    }, 2000)
  }
}


const submitAuth = async (start_param) => {
  CookieManager.removeItem("student_access_token");
  CookieManager.removeItem("student_refresh_token");

  const data = {mini_app_name: '', init_data: initData};
  const appName = route.params;
  if (appName?.name?.length) {
    data.mini_app_name = appName?.name;
  } else {
    toastStore.error({text: t('general.toast_notifications.failed_get_mini_app')});

    return;
  }

  let signResp;
  try {
    if (start_param && start_param.length) {
      if (start_param.startsWith('invite_')) {
        data.invite_id = start_param.split('invite_')[1];

        signResp = await signIn(data, start_param);
      } else {
        signResp = await signIn(data)
      }
    } else {
      signResp = await signIn(data)
    }
  } catch (e) {
    if (e.response?.data === 'invite already claimed' || e.response?.data === 'failed to claim invite') {
      signResp = await signIn(data)
    }
  } finally {
    if (signResp?.data) {
      const {jwt_info, user} = signResp.data;
      miniAppStore.setUserData(user);
      setToken(jwt_info?.access_token, jwt_info?.refresh_token);

      await getAppData()
    }
  }
}

onMounted(async () => {
  const start_param = tg.initDataUnsafe.start_param;
  miniAppStore.isMobileDevice = tg.platform === "android" || tg.platform === "ios";

  await submitAuth(start_param);
})
</script>
<style scoped lang="scss">
@use "../assets/styles/_main.scss" as *;

.loader {
  background: #000;

  &_container {
    position: relative;
    z-index: 2;
    height: 100vh;
    width: 100vw;
    display: flex;
    flex-direction: column;
    gap: 24px;
    justify-content: center;
    align-items: center;
  }

  &_auth {
    padding: 0 30px;

    & input {
      margin-top: 16px;
      width: 100%;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      border-radius: 12px;
      background: #212121;
      height: 50px;
      padding: 16px;

      color: #FFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 400;
      line-height: normal;
    }

    & span {
      color: #FFF;
      font-size: 14px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }

    & button {
      width: 100%;
      margin-top: 40px;
      border: none;
      outline: none;
      border-radius: 12px;
      height: 42px;
      background: linear-gradient(180deg, #0061D2 5%, #00A7E4 105%);

      color: #FFF;
      font-size: 16px;
      font-style: normal;
      font-weight: 500;
      line-height: normal;
    }
  }

  &_logo {
    &-wrapper {
      width: 200px;
      height: 200px;
      display: flex;
      align-items: center;
      justify-content: center;

      & img {
        width: 100%;
      }
    }

    &-animation {
      animation: fade-in 2s ease 0s infinite reverse both;
    }
  }
}

@keyframes fade-in {
  0% {
    opacity: 1;
  }

  100% {
    opacity: 0;
  }
}
</style>
