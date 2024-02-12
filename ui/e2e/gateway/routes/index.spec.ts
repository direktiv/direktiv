import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import {
  createRedisRouteFile,
  findRouteWithApiRequest,
  routeWithAWarning,
  routeWithAnError,
} from "./utils";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Route list is empty by default", async ({ page }) => {
  await page.goto(`/${namespace}/gateway/routes`, {
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
  await createWorkflow({
    payload: createRedisRouteFile({
      path,
      targetType: "instant-response",
      targetConfigurationStatus: "202",
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-route.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/redis-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("route-table").locator("tr"),
    "it renders one row of routes"
  ).toHaveCount(1);

  await expect(
    page
      .getByTestId("route-table")
      .getByRole("link", { name: "/redis-route.yaml" }),
    "it renders the text for the file path"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-table").getByText("GET"),
    "it renders the text for the method"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-table").getByRole("textbox"),
    "it renders the text for the path"
  ).toHaveValue(`http://localhost:3333/ns/${namespace}/${path}`);

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

test("Route list shows a warning", async ({ page }) => {
  await createWorkflow({
    payload: routeWithAWarning,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-route.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/redis-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("route-table").locator("tr"),
    "it renders one row of routes"
  ).toBeVisible();

  await expect(
    page.locator("a").filter({ hasText: "1 warning" }),
    "there is a link with 1 warning"
  ).toHaveText("1 warning");

  await page.locator("a").filter({ hasText: "1 warning" }).hover();

  await expect(
    page.getByTestId("route-table").getByText("no target plugin set"),
    "it shows the warning details on hover"
  ).toBeVisible();
});

test("Route list shows an error", async ({ page }) => {
  await createWorkflow({
    payload: routeWithAnError,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-route.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findRouteWithApiRequest({
          namespace,
          match: (route) => route.file_path === "/redis-route.yaml",
        }),
      "the route was created and is available"
    )
    .toBeTruthy();

  await page.goto(`/${namespace}/gateway/routes`, {
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
      .getByText("plugin this-plugin-does-not-exist does not exist"),
    "it shows the error detail on hover"
  ).toBeVisible();
});

test("Route list links the file name to the route file", async ({ page }) => {
  await createWorkflow({
    payload: createRedisRouteFile(),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-route.yaml",
    },
    headers,
  });

  await page.goto(`/${namespace}/gateway/routes`, {
    waitUntil: "networkidle",
  });

  await page
    .getByTestId("route-table")
    .getByRole("link", { name: "/redis-route.yaml" })
    .click();

  await expect(
    page,
    "after clicking on the file name, the user gets redirected to the file explorer page of the service file"
  ).toHaveURL(`/${namespace}/explorer/endpoint/redis-route.yaml`);
});
