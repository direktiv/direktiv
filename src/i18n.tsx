import Backend from "i18next-http-backend";
import env from "./config/env";
import i18n from "i18next";
import { initReactI18next } from "react-i18next";

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
      loadPath: `${
        process.env.VITE?.VITE_BASE ?? "/"
      }locales/{{lng}}/{{ns}}.json`,
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
