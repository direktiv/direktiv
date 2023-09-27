import { expect, test } from "@playwright/test";

import { createNamespace } from "../utils/namespace";
import { getNamespaces } from "~/api/namespaces/query/get";
import { headers } from "e2e/utils/testutils";

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

  const namespacesBeforeDelete = await getNamespaces({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
    },
    headers,
  });

  const namespaceIsInResults = namespacesBeforeDelete.results.some(
    (item) => item.name === namespace
  );

  expect(
    namespaceIsInResults,
    "the api includes the current namespace in the namespace list"
  ).toBe(true);

  confirmText.type(namespace);

  await expect(
    confirmButton,
    "confirmation buttons should be enabled before typing the namespace name"
  ).toBeEnabled();

  await confirmButton.click();

  await page.waitForURL("/");

  const namespacesAfterDelete = await getNamespaces({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
    },
    headers,
  });

  const deletedNamespaceIsInResults = namespacesAfterDelete.results.some(
    (item) => item.name === namespace,
    "the api does not include the current namespace in the namespace list after deletion"
  );
  expect(deletedNamespaceIsInResults).toBe(false);

  const frontpageUrl = new URL(page.url());
  expect(frontpageUrl.pathname, "url should navigate to the landingpage").toBe(
    "/"
  );

  await expect
    .poll(async () => {
      const currentPath = new URL(page.url()).pathname;
      const namespace = currentPath.split("/")[1];
      return (
        currentPath.endsWith("/explorer/tree") &&
        namespacesAfterDelete.results.some((ns) => ns.name === namespace)
      );
    }, "redirected url should match with namespace path regexp")
    .toBe(true);
});
