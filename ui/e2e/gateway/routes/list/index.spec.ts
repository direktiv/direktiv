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

test("The route list can be visited", async ({ page }) => {
  await page.goto(`/n/${namespace}/gateway/gatewayInfo`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-gateway"),
    "it renders the 'Gateway' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByTestId("breadcrumb-info"),
    "it renders the 'Routes Info' breadcrumb"
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
    mimeType: "application/yaml",
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

  await page.goto(`/n/${namespace}/gateway/gatewayInfo`, {
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

  await page.getByTestId("route-table").getByText("9 methods").hover();

  const methods = [
    "connect",
    "delete",
    "get",
    "head",
    "options",
    "patch",
    "post",
    "put",
    "trace",
  ];

  for (const method of methods) {
    await expect(
      page.getByTestId("route-table").getByText(method)
    ).toBeVisible();
  }

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
    page.getByTestId("route-table").getByText("public endpoint"),
    "it renders the correct label for 'allow anonymous'"
  ).toBeVisible();
});

test("Route list shows an error on no target plugin", async ({ page }) => {
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: routeWithAWarning,
    mimeType: "application/yaml",
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
    mimeType: "application/yaml",
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
    page.getByTestId("route-table").getByText("no valid http method"),
    "it shows the error detail on hover"
  ).toBeVisible();
});

test("Route list links the file name to the route file", async ({ page }) => {
  await createFile({
    name: "my-route.yaml",
    namespace,
    type: "endpoint",
    content: createRouteFile(),
    mimeType: "application/yaml",
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
