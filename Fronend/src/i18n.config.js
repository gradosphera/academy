import { createI18n } from "vue-i18n";
//en
import enGeneral from './locales/en/en-general.json';
//ru
import ruGeneral from './locales/ru/ru-general.json';

export const i18n = createI18n({
    legacy: false,
    locale: 'ru',
    locales: [
        { code: 'en', name: 'English', iso: 'en-US', file: 'en.js' },
        { code: 'ru', name: 'Русский', iso: 'ru-RU', file: 'ru.js' },
    ],
    messages: {
        en: {
            general: enGeneral,
        },
        ru: {
            general: ruGeneral,
        }
    }
});
