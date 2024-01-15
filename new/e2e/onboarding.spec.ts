import { createNamespaceName, deleteNamespace } from "./utils/namespace";
import { expect, test } from "@playwright/test";

test("if no namespaces exist, it renders the onboarding page", async ({
  page,
}) => {
  const namespace = createNamespaceName();
  // mock namespaces endpoint with empty results
  await page.route(`/api/namespaces`, async (route) => {
    if (route.request().method() === "GET") {
      const json = {
        pageInfo: {
          order: [],
          filter: [],
          limit: 0,
          offset: 0,
          total: 0,
        },
        results: [],
      };
      await route.fulfill({ json });
    }
  });

  // visit page
  await page.goto("/");

  // check that elements exist
  await expect(
    page.getByRole("button", { name: "Create namespace" })
  ).toBeVisible();

  await expect(
    page.getByRole("link", { name: "Getting started" })
  ).toBeVisible();

  await expect(page.getByRole("link", { name: "Slack" })).toBeVisible();

  await expect(page.getByRole("link", { name: "GitHub" })).toBeVisible();

  // should always be true if checked too early, thus check after elements have rendered
  await expect(page, "the url should not point to a namespace").toHaveURL("/");

  await expect(page).toHaveTitle("direktiv.io");

  // create a namespace - this will not trigger the mocked endpoint above
  await page.getByRole("button", { name: "Create namespace" }).click();
  await page.getByPlaceholder("new-namespace-name").fill(namespace);
  await page.getByRole("button", { name: "Create" }).click();

  await await expect(
    page,
    "it should redirect to namespace/explorer/tree"
  ).toHaveURL(`/${namespace}/explorer/tree`);

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the breadcrumb shows the new namespace"
  ).toHaveText(namespace);

  // cleanup
  await deleteNamespace(namespace);
});
