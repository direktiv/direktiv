import { checkIfNamespaceExists, createNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

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
  const regex = /^http:\/\/localhost:3333\/[a-zA-Z0-9]+\/explorer\/tree$/;
  await expect
    .poll(
      async () => page.url(),
      "redirected url should match with namespace path regexp"
    )
    .toMatch(regex);
  const ifExists = await checkIfNamespaceExists(namespace);
  expect(
    ifExists,
    "check result should be false as the namespace doesn't exist anymore"
  ).toBe(false);
});
