import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { NamespaceListSchemaType } from "~/api/namespaces/schema";
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

test("it is possible to delete a namespace and it will immediately redirect to a different namespace if available", async ({
  page,
}) => {
  const namespace = await createNamespace();
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("btn-delete-namespace").click();
  const confirmButton = page.getByTestId("delete-namespace-confirm-btn");

  await expect(
    confirmButton,
    "confirmation buttons should be disabled before typing the namespace name"
  ).toBeDisabled();
  const confirmInput = page.getByTestId("delete-namespace-confirm-input");

  const namespacesBeforeDelete = await getNamespacesFromAPI();

  expect(
    namespacesBeforeDelete.results.some(
      (nsFromServer) => nsFromServer.name === namespace
    ),
    "the api includes the current namespace in the namespace list"
  ).toBe(true);

  confirmInput.type(namespace);

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

test("it is possible to delete the last namespace and it will redirect to the landingpage", async ({
  page,
}) => {
  const namespace = await createNamespace();
  await page.goto(`/${namespace}/settings`);
  await page.getByTestId("btn-delete-namespace").click();

  const confirmButton = page.getByTestId("delete-namespace-confirm-btn");
  const confirmInput = page.getByTestId("delete-namespace-confirm-input");

  confirmInput.type(namespace);
  await confirmButton.click();

  /**
   * when the api is now called for the namespaces, we return
   * an empty  list to act like there are no namespaces left
   */
  const mockedNamespace: NamespaceListSchemaType = {
    results: [],
  };
  await page.route("**/api/namespaces", (route, reg) => {
    if (reg.method() !== "GET") {
      return route.continue();
    }
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(mockedNamespace),
    });
  });

  await page.waitForURL("/");
  /**
   * When other namespaces are available, the redirect would happen immediately
   * we wait for a second and make sure we are still on the landingpage
   */
  await page.waitForTimeout(1000);
  const frontpageUrl = new URL(page.url());
  expect(frontpageUrl.pathname, "should stay on the landingpage").toBe("/");
});
