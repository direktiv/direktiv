import { expect, test } from "@playwright/test";

test("it is possible to switch light and dark mode", async ({ page }) => {
  await page.goto("/");
  const userMenuTrigger = page.getByTestId("dropdown-trg-user-menu");
  await userMenuTrigger.nth(1).click();
  const switchMenu = page.getByTestId("dropdown-item-switch-mode");
  const originalText = await switchMenu.textContent();
  await switchMenu.click();
  await userMenuTrigger.nth(1).click();
  const updatedText = await switchMenu.textContent();
  expect(
    originalText,
    "The original mode and updated mode should be different after click"
  ).not.toBe(updatedText);
  expect(originalText, "this mode has the pattern").toMatch(
    /switch to \w+ mode/
  );
  expect(updatedText, "this mode has the pattern").toMatch(
    /switch to \w+ mode/
  );
});
