import { afterAll, beforeAll, vi } from "vitest";

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

beforeAll(() => {
  // avoid console.errors in tests
  vi.spyOn(console, "error").mockImplementation(() => null);
});

afterAll(() => {
  global.console.error = console.error;
});
