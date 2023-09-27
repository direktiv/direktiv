import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { getNamespaces } from "~/api/namespaces/query/get";
import { headers } from "e2e/utils/testutils";

const getNamespacesFromAPI = () =>
  getNamespaces({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
    },
    headers,
  });

/**
 * Our Playwright setup utilizes direktiv namespaces to isolate tests from each other.
 * Due to this design, it's important to note that the Playwright server will almost
 * always have multiple other namespaces that we have no control over. This means that
 * we cannot guarantee a testing environment where no namespaces exist.
 *
 * The server will very have other namespaces, but just to be sure that namespaces will
 * not be empty when we finished the delete namespace test, we create one extra namespace
 * that we will delete after this test.
 */
let plusOneNamespace = "";
test.beforeEach(async () => {
  plusOneNamespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(plusOneNamespace);
  plusOneNamespace = "";
});

test("it is possible to delete a namespace", async ({ page }) => {
  const namespace = await createNamespace();
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("btn-delete-namespace").click();
  const confirmButton = page.getByTestId("delete-namespace-confirm-btn");

  await expect(
    confirmButton,
    "confirmation buttons should be disabled before typing the namespace name"
  ).toBeDisabled();
  const confirmText = page.getByTestId("delete-namespace-confirm-input");

  const namespacesBeforeDelete = await getNamespacesFromAPI();

  expect(
    namespacesBeforeDelete.results.some(
      (nsFromServer) => nsFromServer.name === namespace
    ),
    "the api includes the current namespace in the namespace list"
  ).toBe(true);

  confirmText.type(namespace);

  await expect(
    confirmButton,
    "confirmation buttons should be enabled after typing the namespace name"
  ).toBeEnabled();

  await confirmButton.click();
  await page.waitForURL("/");

  const namespacesAfterDelete = await getNamespacesFromAPI();
  expect(
    namespacesAfterDelete.results.some((item) => item.name === namespace),
    "the api does not include the current namespace in the namespace list after deletion"
  ).toBe(false);

  const frontpageUrl = new URL(page.url());
  expect(
    frontpageUrl.pathname,
    "the user should be navigated to the landingpage for a very short time"
  ).toBe("/");

  await expect
    .poll(async () => {
      const currentPath = new URL(page.url()).pathname;
      const namespace = currentPath.split("/")[1];

      return (
        currentPath.endsWith("/explorer/tree") &&
        namespacesAfterDelete.results[0]?.name === namespace
      );
    }, "after the landingpage redirect, the user should be navigated to the explorer page of the the first namespace found in the api   response")
    .toBe(true);
});
