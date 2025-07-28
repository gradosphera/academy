import {createComponent} from "../helpers/create-component.js";
import Toast from "../components/Toast.vue";
import {watch} from "vue";

export const DEFAULT_TOAST_OPTIONS = {
    type: 'success',
    duration: 3000, //3000
    hideOnRouteChange: true,
};

const open = (options) => {
    if (!options?.message) {
        return;
    }

    const props = { ...DEFAULT_TOAST_OPTIONS, ...options };
    const instance = createComponent(Toast, props, document.body, undefined, 'jup-toast');

    const dismiss = () => {
        (instance?.refs?.toast)?.parentElement?.remove();
    };

    setTimeout(dismiss, props.duration);

    if (props?.hideOnRouteChange) {
        const unwatch = watch(() => window.location.href, () => {
            dismiss();
            unwatch();
        });
    }

    return {
        dismiss,
    };
};

const success = (message, options) => {
    return open({
        ...options,
        message,
        type: 'success',
    });
};

const error = (message, options) => {
    return open({
        ...options,
        message,
        type: 'error',
    });
};

const wait = (message, options) => {
    return open({
        ...options,
        message,
        type: 'wait',
    });
};

export const useToast = () => {
    return {
        open,
        success,
        error,
        wait,
    };
};
