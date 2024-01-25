import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { createRedisServiceFile, serviceWithAnError } from "./utils";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { getServices } from "~/api/services/query/services";
import { headers } from "e2e/utils/testutils";
import { updateWorkflow } from "~/api/tree/mutate/updateWorkflow";

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

  // wait one second to make sure the service has been created and avoid flaky tests
  await page.waitForTimeout(1000);

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

  // message can be "Up 1 second" or "Up 2 seconds" or "Up Less than a second"
  await expect(
    page.getByTestId("service-row").getByText(/Up .* second/),
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

test("Service list will link the file name to the service file", async ({
  page,
}) => {
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

  await expect(
    page,
    "after clicking on the service file name, the user gets redirected to the file explorer page of the service file"
  ).toHaveURL(`/${namespace}/explorer/service/redis-service.yaml`);

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "it renders the 'Services' breadcrumb"
  ).toHaveText("redis-service.yaml");
});

test("Service list will link the row to the service details page", async ({
  page,
}) => {
  await createWorkflow({
    payload: createRedisServiceFile(),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-service.yaml",
    },
    headers,
  });

  const { data: services } = await getServices({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
    },
  });

  const createdService = services.find(
    (service) => service.filePath === "/redis-service.yaml"
  );
  if (!createdService) throw new Error("could not find service");

  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await page.getByTestId("service-row").click();

  await expect(
    page,
    "after clicking on the service row, the user gets redirected to the service details page"
  ).toHaveURL(`/${namespace}/services/${createdService.id}`);

  await expect(
    page
      .getByTestId("breadcrumb-services")
      .getByRole("link", { name: "Services" }),
    "it renders the 'Services' breadcrumb segment"
  ).toBeVisible();

  await expect(
    page.getByRole("link", {
      name: `${createdService.id}`,
    }),
    "it renders the filename breadcrumb segment"
  ).toBeVisible();
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
  await expect(
    page.getByRole("button", { name: "Rebuild" }),
    "it opens the context menu"
  ).toBeVisible();

  await page.getByRole("button", { name: "Rebuild" }).click();
  await expect(
    page.getByLabel("Rebuild service"),
    "it opens the rebuild service modal"
  ).toBeVisible();
  await page.getByRole("button", { name: "Rebuild" }).click();

  await expect(
    page.getByTestId("toast-success"),
    "it renders a confirmation toast after resending the event"
  ).toBeVisible();
});

test("Service list will highlight services that have errors", async ({
  page,
}) => {
  await createWorkflow({
    payload: serviceWithAnError,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "failed-service.yaml",
    },
    headers,
  });

  // wait one second to make sure the service has been created and avoid flaky tests
  await page.waitForTimeout(1000);

  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("service-row").locator("a").filter({ hasText: "Error" }),
    "it renders the Error status of the service"
  ).toBeVisible();

  await page
    .getByTestId("service-row")
    .locator("a")
    .filter({ hasText: "Error" })
    .hover();

  await expect(
    page.getByTestId("service-row").getByText("image pull, err:"),
    "it renders the uptime in a tooltip"
  ).toBeVisible();
});

test("Service list will update the services when refecth button is clicked", async ({
  page,
}) => {
  await createWorkflow({
    payload: createRedisServiceFile({
      scale: 1,
      size: "large",
    }),
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
      .getByRole("cell", { name: "1", exact: true }),
    "it will show the scale of 1"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "large" }),
    "it will show a size of large"
  ).toBeVisible();

  await updateWorkflow({
    payload: createRedisServiceFile({
      scale: 2,
      size: "small",
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: "redis-service.yaml",
    },
    headers,
  });

  await page.getByLabel("Refetch services").click();

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("cell", { name: "2", exact: true }),
    "it will have updated the scale to 1"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "small" }),
    "it will have updated the size to small"
  ).toBeVisible();
});
