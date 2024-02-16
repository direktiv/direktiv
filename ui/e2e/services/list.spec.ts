import { createNamespace, deleteNamespace } from "../utils/namespace";
import {
  createRedisServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { encode } from "js-base64";
import { headers } from "e2e/utils/testutils";
import { patchNode } from "~/api/filesTree/mutate/patchNode";

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

test("Service list shows all available services", async ({ page }) => {
  await createWorkflow({
    payload: createRedisServiceFile(),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "redis-service.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === "/redis-service.yaml" &&
            (service.conditions ?? []).some((c) => c.type === "UpAndReady"),
        }),
      "the service in the backend is in an UpAndReady state"
    )
    .toBeTruthy();

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
    "it renders the environment variable count"
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

test("Service list links the file name to the service file", async ({
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
    "after clicking on the file name, the user gets redirected to the file explorer page of the service file"
  ).toHaveURL(`/${namespace}/explorer/service/redis-service.yaml`);

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "it renders the filename in the breadcrumb"
  ).toHaveText("redis-service.yaml");
});

test("Service list links the row to the service details page", async ({
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

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) => service.filePath === "/redis-service.yaml",
        }),
      "the service was mounted in the backend"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === "/redis-service.yaml",
  });

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
    "it renders the service id breadcrumb segment"
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

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === "/redis-service.yaml" &&
            (service.conditions ?? []).some((c) => c.type === "UpAndReady"),
        }),
      "the service in the backend is in an UpAndReady state"
    )
    .toBeTruthy();

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

test("Service list highlights services that have errors", async ({ page }) => {
  await createWorkflow({
    payload: serviceWithAnError,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "failed-service.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === "/failed-service.yaml" &&
            service.error !== null,
        }),
      "the service in the backend is in an error state"
    )
    .toBeTruthy();

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
    "it renders the error in a tooltip"
  ).toBeVisible();
});

test("Service list will update the services when refetch button is clicked", async ({
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
    "it shows the service's scale"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "large" }),
    "it shows the service's size"
  ).toBeVisible();

  await patchNode({
    payload: {
      data: encode(
        createRedisServiceFile({
          scale: 2,
          size: "small",
        })
      ),
    },
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
    "it has updated the rendered scale to the new value"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "small" }),
    "it has updated the rendered size to the new value"
  ).toBeVisible();
});
