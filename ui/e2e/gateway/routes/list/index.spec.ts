import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import {
  createRouteFile,
  findRouteWithApiRequest,
  routeWithAWarning,
  routeWithAnError,
} from "../utils";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Route list is empty by default", async ({ page }) => {
  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-routes"),
    "it renders the 'Routes' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByText("No routes exist yet"),
    "it renders an empty list of routes"
  ).toBeVisible();
});

test("Route list shows all available routes", async ({ page }) => {
  const path = "newPath";
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: createRouteFile({
      path,
      targetType: "instant-response",
      targetConfigurationStatus: "202",
    }),
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/my-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("route-table").locator("tr"),
    "it renders one row of routes"
  ).toHaveCount(1);

  await expect(
    page
      .getByTestId("route-table")
      .getByRole("link", { name: "/my-route.yaml" }),
    "it renders the text for the file path"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-table").getByRole("textbox"),
    "it renders the text for the path"
  ).toHaveValue(
    `${process.env.PLAYWRIGHT_UI_BASE_URL}/ns/${namespace}/${path}`
  );

  await expect(
    page.getByTestId("route-table").getByText("GET", { exact: true }),
    "it renders the text for the method"
  ).toBeVisible();

  await page.getByTestId("route-table").getByText("+7").hover();

  await expect(
    page
      .getByTestId("route-table")
      .getByText("OPTIONSPUTPOSTHEADCONNECTPATCHTRACE"),
    'it shows more methods when hovering over the "+7"'
  ).toBeVisible();

  // hover over somethiing else to make the overlay disappear
  await page.getByTestId("route-table").getByText("yes").hover();

  await expect(
    page.getByTestId("route-table").getByRole("cell", { name: "1 plugin" }),
    "it renders the correct number for plugins"
  ).toBeVisible();

  await page
    .getByTestId("route-table")
    .getByRole("cell", { name: "1 plugin" })
    .hover();

  await expect(
    page.getByTestId("route-table").getByText("1 target plugin"),
    "it shows the plugin details on hover"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-table").getByText("yes"),
    "it renders the correct label for 'allow anonymous'"
  ).toBeVisible();
});

test("Route list shows an error on no target plugin", async ({ page }) => {
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: routeWithAWarning,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/my-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("route-table").locator("tr"),
    "it renders one row of routes"
  ).toBeVisible();

  await expect(
    page.locator("a").filter({ hasText: "1 error" }),
    "there is a link with 1 error"
  ).toHaveText("1 error");

  await page.locator("a").filter({ hasText: "1 error" }).hover();

  await expect(
    page.getByTestId("route-table").getByText("no target plugin found"),
    "it shows the warning details on hover"
  ).toBeVisible();
});

test("Route list shows an error", async ({ page }) => {
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: routeWithAnError,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/my-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("route-table").locator("tr"),
    "it renders one row of routes"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("route-table").locator("a").filter({ hasText: "1 error" }),
    "there is a link with 1 error"
  ).toBeVisible();

  await page
    .getByTestId("route-table")
    .locator("a")
    .filter({ hasText: "1 error" })
    .hover();

  await expect(
    page
      .getByTestId("route-table")
      .getByText("plugin 'this-plugin-does-not-exist' err: doesn't exist"),
    "it shows the error detail on hover"
  ).toBeVisible();
});

test("Route list links the file name to the route file", async ({ page }) => {
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: createRouteFile(),
  });

  await page.goto(`/n/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await page
    .getByTestId("route-table")
    .getByRole("link", { name: "/my-route.yaml" })
    .click();

  await expect(
    page,
    "after clicking on the file name, the user gets redirected to the file explorer page of the service file"
  ).toHaveURL(`/n/${namespace}/explorer/endpoint/my-route.yaml`);
});
