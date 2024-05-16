import Backend from "i18next-http-backend";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import translationFilePath from "~/assets/locales/en/translation.json?url";

// Example: we could define translations here (or import them as a module)
// and pass them to the init object below. Our current setup uses the
// HTTP backend instead, with translation files in /public.
// const resources = {
//   en: {
//     translation: {
//       welcomeTo: "Welcome to",
//     }
//   }
// };

i18n
  .use(Backend)
  .use(initReactI18next) // passes i18n down to react-i18next
  .init({
    backend: {
      loadPath: translationFilePath,
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

export default i18n;
