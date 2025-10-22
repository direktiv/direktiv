import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import translation from "~/assets/locales/en/translation.json";

i18n.use(initReactI18next).init({
  resources: {
    en: {
      translation,
    },
  },
  lng: "en",
  fallbackLng: "en",
  debug: true,
  react: {
    transKeepBasicHtmlNodesFor: ["br", "strong", "i", "p", "b"],
  },
  interpolation: {
    escapeValue: false, // react already safes from xss
  },
});
