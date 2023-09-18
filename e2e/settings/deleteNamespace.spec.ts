import { expect, test } from "@playwright/test";

import { createNamespace } from "../utils/namespace";

test("it is possible to delete namespaces", async ({ page }) => {
  const namespace = await createNamespace();
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("btn-delete-namespace").click();
  const confirmButton = page.getByTestId("namespace-delete-confirm");
  await expect(
    confirmButton,
    "confirmation buttons should be disabled before typing the namespace name"
  ).toBeDisabled();
  const confirmText = page.getByTestId("inp-delete-namespace-confirm");
  confirmText.type(namespace);

  await expect(
    confirmButton,
    "confirmation buttons should be disabled before typing the namespace name"
  ).toBeEnabled();
  await confirmButton.click();

  await page.waitForURL("/");
  expect(page.url(), "url should navigate to the very initial state").toBe(
    "http://localhost:3333/"
  );
  // this shows the onboarding
  //on this page, we are navigating again to the first namespace in the list, which happens with in a sec
  // this test user has the namespace "sebxian" at the top
  await page.waitForURL("/sebxian/explorer/tree");
  expect(page.url(), "url should navigate to the very initial state").toBe(
    "http://localhost:3333/sebxian/explorer/tree"
  );
});
