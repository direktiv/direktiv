import { expect, test } from "@playwright/test";

import { getStyle } from "./utils/testutils";

test("it is possible to switch light and dark mode", async ({ page }) => {
  await page.goto("/");
  const userMenuTrigger = page.getByTestId("dropdown-trg-user-menu");
  await userMenuTrigger.nth(1).click();
  const switchMenu = page.getByTestId("dropdown-item-switch-mode");
  const originalText = await switchMenu.textContent();
  const body = page.locator("body");
  const originBodyBGColor = await getStyle(body, "background-color");
  expect(
    originBodyBGColor,
    "light mode background color of Body element should be rgb(255,255,255)"
  );

  await switchMenu.click();
  const updatedBodyBGColor = await getStyle(body, "background-color");
  expect(
    updatedBodyBGColor,
    "dark mode background color of Body element should be rgb(0,0,0)"
  );

  await userMenuTrigger.nth(1).click();
  const updatedText = await switchMenu.textContent();
  expect(
    originalText,
    "The original mode and updated mode should be different after click"
  ).not.toBe(updatedText);
  expect(originalText, "the menu item should be 'switch to Dark mode'").toMatch(
    /switch to Dark mode/
  );
  expect(updatedText, "this mode has the pattern").toMatch(
    /switch to Light mode/
  );
  await switchMenu.click(); //switch back to light mode
  const secondUpdatedBGColor = await getStyle(body, "background-color");
  expect(
    secondUpdatedBGColor,
    "after switching back, the body background should be the original color "
  ).toBe(originBodyBGColor);
});
