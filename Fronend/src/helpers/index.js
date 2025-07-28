import CookieManager from "./cookieManager.js";
import {useProductsStore} from "../store/productsStore.js";
import {getMiniApp, getProduct} from "../api/api.js";
import {useMiniAppStore} from "../store/miniAppStore.js";
import {BASE_MEDIA_URL} from "../constants/const.js";
import {event} from "vue-gtag";
import {useGtm} from "@gtm-support/vue-gtm";

const productStore = useProductsStore();

export function decodeJwt(token) {
  const base64Url = token.split('.')[1]; // Get payload part
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  return JSON.parse(atob(base64));
}

export const getMiniAppFunc = async () => {
  const miniAppStore = useMiniAppStore();
  const productsStore = useProductsStore();

  try {
    const miniAppResp = await getMiniApp();
    if (miniAppResp.data) {
      miniAppStore.setMiniAppData(miniAppResp.data.mini_app)
      productsStore.setAllProducts(miniAppResp.data.mini_app?.products)
    }
  } finally {
  }
}

export const getFile = (item) => {
  if (item) {
    const token = CookieManager.getItem("student_access_token");

    return BASE_MEDIA_URL + '/' + item + '?jwt=' + token;
  }
};

export const getRandomNumber = () => {
  return Math.floor(10000 + Math.random() * 90000);
}

export const timeSince = (dateString) => {
  const pastDate = new Date(dateString);
  const now = new Date();

  let year = now.getFullYear() - pastDate.getFullYear();
  let months = now.getMonth() - pastDate.getMonth();
  let days = now.getDate() - pastDate.getDate();

  if (months < 0) {
    year--;
    months += 12;
  }

  if (days < 0) {
    months--;
    const prevMonth = new Date(now.getFullYear(), now.getMonth(), 0);
    days += prevMonth.getDate();
  }

  return {year, months, days};
}

export const getAvgScore = (avgScore) => {
  const maxScore = 10000;
  const scoreScaleOnPage = 5;

  const displayScore = (avgScore / maxScore) * scoreScaleOnPage;

  return displayScore.toFixed(1);
}

export const getCompletedLessons = (progress, totalLessons) => {
  const maxProgress = 10000;
  const progressPerLesson = maxProgress / totalLessons;

  // Calculate completed lessons
  return Math.floor(progress / progressPerLesson);
};

export const formatFileSize = (bytes) => {
  if (bytes < 1024) return `${bytes} bytes`;
  if (bytes < 1024 ** 2) return `${(bytes / 1024).toFixed(2)} KB`;
  if (bytes < 1024 ** 3) return `${(bytes / (1024 ** 2)).toFixed(2)} MB`;

  return `${(bytes / (1024 ** 3)).toFixed(2)} GB`;
};

export const getProductData = async (id) => {
  const {setSelectedProduct, setProductProgress, setProductReviews, setPaidLessons, setProductInvite} = useProductsStore();
  const resp = await getProduct(id);

  if (resp.data) {
    setSelectedProduct(resp.data.product);
    setProductProgress(resp.data.progress)
    setProductReviews(resp.data.reviews);
    setPaidLessons(resp.data.unlocked_lessons);
    setProductInvite(resp.data.invite);
  }
}

export const checkLessonStatus = (elem) => {
  elem.value = productStore.productProgress.find(item => item.lesson_id === productStore.selectedLesson?.id);
}

export const checkLessonReview = (elem) => {
  if (productStore.selectedLesson) {
    elem.value = productStore.productReviews.find(review => review.lesson_id === productStore.selectedLesson.id);
  }
}

export const formatProductDate = (isoString, locale) => {
  const date = new Date(isoString);

  const str = date.toLocaleString(locale, {
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });

  return str.replace(' at', ',')
};

export const isDatePassed = (date) => {
  if (!date) return true;

  const now = new Date().getTime();
  const subjectDate = new Date(date).getTime();

  return now > subjectDate;
}

export const parseMessage = (messageText) => {
  if (!messageText) return []
  return messageText.split('\n')
}

export const getAgreementLink = (obj) => {
  if (obj.url?.length) return obj.url;
  if (obj.filename?.length) return getFile(obj.filename);

  return ''
}

export const getCurrencySymbol = (currency) => {
  if (!currency || !currency.length) return '';

  switch (currency) {
    case 'USD':
      return '$';
    case 'EUR':
      return '€';
    case 'UAH':
      return '₴';
  }
}

export function formatPriceRange(levels) {
  if (!levels.length) return '';

  const sorted = [...levels].sort((a, b) => Number(a.price) - Number(b.price));
  const minPrice = sorted[0].price;
  const maxPrice = sorted[sorted.length - 1].price;
  const currency = sorted[0].currency;

  if (levels.length === 1) {
    return `${getCurrencySymbol(currency)}${minPrice}`;
  }

  return `${getCurrencySymbol(currency)}${minPrice} - ${maxPrice}`;
}

export function submitAnalyticsData(action, parameters) {
  if (!action?.length || !parameters) return;

  //google analytics
  event(action, parameters)

  //facebook pixel
  window.fbq?.('trackCustom', action, parameters)

  const gtm = useGtm();

  if (!gtm) return;

  gtm.trackEvent({
    event: 'material_click',
    ...parameters,
  });
}

export async function initializeGoogleAnalytics(id) {
  if (window.gtag || !id) return;

  // Create the first script tag for gtag.js
  const script1 = document.createElement('script');
  script1.async = true;
  script1.src = `https://www.googletagmanager.com/gtag/js?id=${id}`;
  document.head.appendChild(script1);

  // Create the second script tag with the dataLayer and gtag functions
  const script2 = document.createElement('script');
  const inlineScriptContent = `
    window.dataLayer = window.dataLayer || [];
    function gtag(){dataLayer.push(arguments);}
    gtag('js', new Date());
    gtag('config', '${id}');
  `;
  script2.appendChild(document.createTextNode(inlineScriptContent)); // Changed this line
  document.head.appendChild(script2);
}