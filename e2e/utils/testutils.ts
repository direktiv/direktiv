import { Locator, Page } from "@playwright/test";

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

/**
 * this will mock the browsers clipboard API, since it might not be available in the test environment
 * due to invalid permissions. It's recommended to use this function in the beforeAll or beforeEach hook
 * of the test to inject the mock into the page very early. It will e.g. not work if it's called after
 * page.goto() has been called.
 */
export const mockClipboardAPI = async (page: Page) =>
  await page.addInitScript(() => {
    // create a mock of the clipboard API
    const mockClipboard = {
      clipboardData: "",
      writeText: async (text: string) => {
        mockClipboard.clipboardData = text;
      },
      readText: async () => mockClipboard.clipboardData,
    };

    // override the native clipboard API
    Object.defineProperty(navigator, "clipboard", {
      value: mockClipboard,
      writable: false,
      enumerable: true,
      configurable: true,
    });
  });
