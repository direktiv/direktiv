import { expect, test } from "@playwright/test";

import { createNamespace } from "../utils/namespace";
import { getNamespaces } from "~/api/namespaces/query/get";
import { headers } from "e2e/utils/testutils";

type NamespaceResult = {
  createdAt: string;
  updatedAt: string;
  name: string;
  oid: string;
};
test("it is possible to delete namespaces", async ({ page }) => {
  const namespace = await createNamespace();
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("btn-delete-namespace").click();
  const confirmButton = page.getByTestId("delete-namespace-confirm-btn");

  await expect(
    confirmButton,
    "confirmation buttons should be disabled before typing the namespace name"
  ).toBeDisabled();
  const confirmText = page.getByTestId("delete-namespace-confirm-input");

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

  const namespaces = await getNamespaces({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
    },
    headers,
  });
  const findResult = namespaces.results.find(
    (item: NamespaceResult) => item.name === namespace
  );
  expect(
    findResult,
    "check result should be undefined as the namespace doesn't exist anymore"
  ).toBe(undefined);
});
