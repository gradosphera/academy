import {defineStore} from "pinia";
import {ref} from "vue";

const toastStatus = {success: "success", error: "error"};

const defaultTimeout = 2000;

const createToast = (text, status) => ({
    text,
    status,
    id: Math.random() * 1000,
});

export const useToastStore = defineStore('toast', () => {
    const toasts = ref([]);

    const updateState = (payload, status) => {
        const { text, timeout } = payload;

        const toast = createToast(text, status);

        toasts.value.push(toast);

        setTimeout(() => {
            toasts.value = toasts.value.filter((t) => t.id !== toast.id);
        }, timeout ?? defaultTimeout);
    };

    function success(payload) {
        updateState(payload, toastStatus.success);
    }

    function error(payload) {
        updateState(payload, toastStatus.error);
    }

    return {
        toasts,
        success,
        error,
    }
})
