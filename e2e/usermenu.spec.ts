import { expect, test } from "@playwright/test";

import { getStyle } from "./utils/testutils";

test("it is possible to switch between light and dark mode", async ({
  page,
}) => {
  await page.goto("/");
  const bodyTag = page.locator("body");
  const userMenuBtn = page.getByTestId("dropdown-trg-user-menu");
  await userMenuBtn.nth(1).click();
  const themeSwitchBtn = page.getByTestId("dropdown-item-switch-theme");

  // user is in light mode
  expect(
    await getStyle(bodyTag, "background-color"),
    "the user is in light mode and the background color of the page should be white"
  ).toBe("rgb(255, 255, 255)");

  expect(
    await themeSwitchBtn.textContent(),
    "the theme button should display 'switch to dark mode' initially"
  ).toMatch(/Switch to dark mode/);

  // switch to dark mode
  await themeSwitchBtn.click();
  await userMenuBtn.nth(1).click();

  // user is now in dark mode
  expect(
    await getStyle(bodyTag, "background-color"),
    "the user is in dark mode and the background color of the page should be black"
  ).toBe("rgb(0, 0, 0)");

  expect(
    await themeSwitchBtn.textContent(),
    "the theme button should now display 'switch to light mode'"
  ).toMatch(/Switch to light mode/);

  // back to light mode again
  await themeSwitchBtn.click();
  await userMenuBtn.nth(1).click();

  // user is now in light mode again
  expect(
    await themeSwitchBtn.textContent(),
    "the theme button should display 'switch to Dark mode' initially"
  ).toMatch(/Switch to dark mode/);

  expect(
    await getStyle(bodyTag, "background-color"),
    "the user is in light mode and the background color of the page should be white"
  ).toBe("rgb(255, 255, 255)");
});
