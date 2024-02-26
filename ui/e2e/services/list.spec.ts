import {
  createHttpServiceFile,
  createRedisServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { encode } from "js-base64";
import { headers } from "e2e/utils/testutils";
import { patchFile } from "~/api/files/mutate/patchFile";

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
  const serviceFile = await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createHttpServiceFile(),
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === serviceFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "ConfigurationsReady" && c.status === "True"
            ),
        }),
      "the service in the backend is in state ConfigurationsReady"
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
      .getByRole("link", { name: serviceFile.data.path }),
    "it renders the link to the service file"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").filter({ hasText: "ConfigurationsReady" }),
    "it renders the ConfigurationsReady status of the service"
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
    page.getByTestId("service-row").getByRole("cell", {
      name: "gcr.io/direktiv/functions/http-request:1.0",
      exact: true,
    }),
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
});

test("Service list links the file name to the service file", async ({
  page,
}) => {
  await createFile({
    name: "redis-service.yaml",
    namespace,
    type: "service",
    yaml: createRedisServiceFile(),
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
  await createFile({
    name: "redis-service.yaml",
    namespace,
    type: "service",
    yaml: createRedisServiceFile(),
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
  const serviceFile = await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createHttpServiceFile(),
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === serviceFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "ConfigurationsReady" && c.status === "True"
            ),
        }),
      "the service in the backend is in state ConfigurationsReady"
    )
    .toBeTruthy();

  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("service-row").filter({ hasText: "ConfigurationsReady" }),
    "it renders the ConfigurationsReady status of the service"
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
  const serviceFile = await createFile({
    name: "failed-service.yaml",
    namespace,
    type: "service",
    yaml: serviceWithAnError,
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === serviceFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "ConfigurationsReady" && c.status === "False"
            ),
        }),
      "the service in the backend is in an error state"
    )
    .toBeTruthy();

  await page.goto(`/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await page
    .getByTestId("service-row")
    .locator("a")
    .filter({ hasText: "ConfigurationsReady" })
    .hover();

  await await expect(
    page
      .getByTestId("service-row")
      .getByText("failed with message: Unable to fetch image"),
    "it renders the error in a tooltip"
  ).toBeVisible();
});

test("Service list will update the services when refetch button is clicked", async ({
  page,
}) => {
  await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createHttpServiceFile({
      scale: 1,
      size: "large",
    }),
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

  await patchFile({
    payload: {
      data: encode(
        createHttpServiceFile({
          scale: 2,
          size: "small",
        })
      ),
    },
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: "/http-service.yaml",
    },
    headers,
  });

  await page.waitForTimeout(1000);

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
