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

test("Service list will show all available services", async ({ page }) => {
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
    "it renders the UpAndReady status of the service"
  ).toBeVisible();

  await page
    .getByTestId("service-row")
    .locator("a")
    .filter({ hasText: "UpAndReady" })
    .hover();

  await expect(
    page.getByTestId("service-row").getByText(/Up \d+ second/),
    "it renders the uptime in a tooltip"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .locator("a")
      .filter({ hasText: "1 environment variable" }),
    "it renders one environment variable"
  ).toBeVisible();

  await page
    .getByTestId("service-row")
    .locator("a")
    .filter({ hasText: "1 environment variable" })
    .hover();

  await expect(
    page.getByText("MY_ENV_VAR=env-var-value"),
    "it shows the environment variable in a tooltip"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("cell", { name: "redis", exact: true }),
    "it renders the image name of the service"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("cell", { name: "1", exact: true }),
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

test("Service list will link to the service file", async ({ page }) => {
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

  await page
    .getByTestId("service-row")
    .getByRole("link", { name: "/redis-service.yaml" })
    .click();

  await expect(page, "after clicking the service, the user was").toHaveURL(
    `/${namespace}/explorer/service/redis-service.yaml`
  );

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "it renders the 'Services' breadcrumb"
  ).toHaveText("redis-service.yaml");
});

test("Service list lets the user rebuild a service", async ({ page }) => {
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
    page
      .getByTestId("service-row")
      .locator("a")
      .filter({ hasText: "UpAndReady" }),
    "it renders the UpAndReady status of the service"
  ).toBeVisible();

  await page.getByTestId("service-row").getByRole("button").click();
  await page.getByRole("button", { name: "Rebuild" }).click(); // select the rebuild option from the context menu
  await page.getByRole("button", { name: "Rebuild" }).click(); // click the confirm button

  await expect(
    page
      .getByTestId("service-row")
      .locator("a")
      .filter({ hasText: "1 environment variable" }),
    "it renders one environment variable"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-row")
      .locator("a")
      .filter({ hasText: "UpAndReady" }),
    "but it does not render the UpAndReady status of the service anymore"
  ).not.toBeVisible();
});
