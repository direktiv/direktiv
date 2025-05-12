import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import translation from "../src/assets/locales/en/translation.json";

i18n.use(initReactI18next).init({
  lng: "en",
  fallbackLng: "en",
  ns: ["translation"],
  defaultNS: "translation",
  resources: { en: { translation } },
});
