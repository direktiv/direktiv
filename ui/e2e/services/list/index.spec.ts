import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createRedisServiceFile } from "./utils";
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

test("Service list is empty by default", async ({ page }) => {
  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-services"),
    "it renders the 'Services' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByText("No services exist yet"),
    "it renders an empy list of services"
  ).toBeVisible();
});

test("Service list will list available services", async ({ page }) => {
  await createWorkflow({
    payload: createRedisServiceFile(),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-service.yaml",
    },
    headers,
  });

  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("service-row"),
    "it renders one row of services"
  ).toHaveCount(1);

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("link", { name: "/redis-service.yaml" }),
    "it renders the link to the service file"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .locator("a")
      .filter({ hasText: "UpAndReady" }),
    "it renders the status of the service"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("cell", { name: "redis", exact: true }),
    "it renders the image name of the service"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "1" }),
    "it renders the scale of the service"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "small" }),
    "it renders the size of the service"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "redis-server" }),
    "it renders the cmd of the service"
  ).toBeVisible();
});
