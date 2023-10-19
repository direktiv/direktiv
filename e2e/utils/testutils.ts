import { Locator } from "@playwright/test";

const token = process.env.VITE_E2E_API_TOKEN;
export const headers: { "Direktiv-Token"?: string } = token
  ? {
      "Direktiv-Token": token,
    }
  : {};

export const getStyle = async (
  locator: Locator,
  property: string
): Promise<string> =>
  locator.evaluate(
    (el, property) => window.getComputedStyle(el).getPropertyValue(property),
    property
  );

/**
 * This is a workaround for a problem with clicking on some elements in Webkit.
 * The .click() method should be used whenever possible. But in some cases,
 * it does not work reliably in Webkit. In these cases, dispatchEvent("click")
 * works. The downside is that it triggers the click event without simulating
 * a mouse click, so it is not a reliable test that the UI works.
 *
 * It seems that this problem occurs only with or within Radix UI elements (dropdowns,
 * popups, toasts, etc.).
 *
 * @param browserName is available in tests with test("...", { page, browserName })
 * @param locator e.g. page.getByTestid("test-id")
 */
export const radixClick = async (
  browserName: "chromium" | "firefox" | "webkit",
  locator: Locator
) => {
  if (browserName === "webkit") {
    await locator.dispatchEvent("click");
  } else {
    await locator.click();
  }
};
