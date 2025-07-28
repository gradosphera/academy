import {defineStore} from "pinia";
import {computed, ref} from "vue";

export const useMiniAppStore = defineStore('miniapp', () => {
    const miniAppData = ref(null);
    const userData = ref(null);
    const isAuthPageOpen = ref(true);
    const tutorialData = ref([]);
    const payments = ref([]);
    const isMobileDevice = ref(false);
    const afterPaymentData = ref({status: '', product_id: '', service: ''});
    const tariffId = ref('');

    const setMiniAppData = (data) => miniAppData.value = data;
    const setUserData = (data) => userData.value = data;
    const toggleAuthPage = () => isAuthPageOpen.value = false;
    const setTutorialData = (data) => {
        if (data) {
            tutorialData.value = data
        }
    };
    const setPayments = (data) => payments.value = data;
    const setPaymentStatus = (status) => {
        if (status) {
            const splitStatus = status.split('=');
            if (splitStatus.length > 1) {
                afterPaymentData.value = {status: 'paid', product_id: splitStatus[1], service: 'wayforpay'}
            }
            // if (status.startsWith('payment_failed')) {
            //     isPaymentSuccess.value = false;
            //
            //     const id = status.split('payment_failed_id_');
            //     tariffId.value = id[1];
            //
            // } else {
            //     isPaymentSuccess.value = true;
            // }
        }
    }

    const accentedColor = computed(() => {
        if (miniAppData.value?.color_theme?.accent_color) {
            const colors = miniAppData.value?.color_theme?.accent_color.split('/');

            if (colors.length > 1) {
                return `linear-gradient(180deg, #${colors[0]}, #${colors[1]})`;
            } else {
                return `#${colors[0]}`;
            }
        } else {
            return 'linear-gradient(180deg, #046AE2 56.69%, #046AE2 61.8%, #066CE2 65.95%, #086DE2 69.3%, #0B70E1 71.99%, #0E73E1 74.17%, #1276E1 75.99%, #177AE0 77.58%, #1B7EE0 79.11%, #2082DF 80.7%, #2586DF 82.52%, #2A8ADE 84.7%, #2F8EDE 87.39%, #3492DD 90.74%, #3996DD 94.89%, #3D9ADC 100%)'
        }
    })


    return {
        miniAppData,
        accentedColor,
        userData,
        isAuthPageOpen,
        tutorialData,
        payments,
        isMobileDevice,
        afterPaymentData,
        tariffId,

        setMiniAppData,
        setUserData,
        toggleAuthPage,
        setTutorialData,
        setPayments,
        setPaymentStatus,
    }
})