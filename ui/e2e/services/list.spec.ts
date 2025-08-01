import { createFile, deleteFile } from "e2e/utils/files";
import { createNamespace, deleteNamespace } from "../utils/namespace";
import {
  createRequestServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
import { expect, test } from "@playwright/test";

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
  await page.goto(`/n/${namespace}/services`, {
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
    yaml: createRequestServiceFile(),
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === serviceFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "Available" && c.status === "True"
            ),
        }),
      "the service in the backend is in state Available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/services`, {
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
    page.getByTestId("service-row").filter({ hasText: "Available" }),
    "it renders the Available status of the service"
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
      name: "direktiv/request:v4",
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
  const serviceFile = await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createRequestServiceFile(),
  });

  await page.goto(`/n/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await page
    .getByTestId("service-row")
    .getByRole("link", { name: serviceFile.data.path })
    .click();

  await expect(
    page,
    "after clicking on the file name, the user gets redirected to the file explorer page of the service file"
  ).toHaveURL(`/n/${namespace}/explorer/service/http-service.yaml`);

  await expect(
    page.getByTestId("breadcrumb-segment"),
    "it renders the filename in the breadcrumb"
  ).toHaveText("http-service.yaml");
});

test("Service list links the row to the service details page", async ({
  page,
}) => {
  const serviceFile = await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createRequestServiceFile(),
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) => service.filePath === serviceFile.data.path,
        }),
      "the service was mounted in the backend"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === serviceFile.data.path,
  });

  if (!createdService) throw new Error("could not find service");

  await page.goto(`/n/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await page.getByTestId("service-row").click();

  await expect(
    page,
    "after clicking on the service row, the user gets redirected to the service details page"
  ).toHaveURL(`/n/${namespace}/services/${createdService.id}`);

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
    yaml: createRequestServiceFile(),
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === serviceFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "Available" && c.status === "True"
            ),
        }),
      "the service in the backend is in state Available"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("service-row").filter({ hasText: "Available" }),
    "it renders the Available status of the service"
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
              (c) => c.type === "Available" && c.status === "False"
            ),
        }),
      "the service in the backend is in an error state"
    )
    .toBeTruthy();

  await page.goto(`/n/${namespace}/services`, {
    waitUntil: "networkidle",
  });

  await page
    .getByTestId("service-row")
    .locator("a")
    .filter({ hasText: "Available" })
    .hover();

  await await expect(
    page
      .getByTestId("service-row")
      .getByText("Deployment does not have minimum availability"),
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
    yaml: createRequestServiceFile({
      scale: 1,
      size: "large",
    }),
  });

  await page.goto(`/n/${namespace}/services`, {
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
        createRequestServiceFile({
          scale: 2,
          size: "small",
        })
      ),
    },
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
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

test.describe("system namespace", () => {
  const systemNamespaceName = "system";
  const systemServiceName = "http-service.yaml";
  let cleanUpSystemNamespace = true;

  test.beforeAll(async () => {
    try {
      await createNamespace(systemNamespaceName);
    } catch (e) {
      cleanUpSystemNamespace = false;
    }

    await createFile({
      name: systemServiceName,
      namespace: systemNamespaceName,
      type: "service",
      yaml: createRequestServiceFile(),
    });
  });

  test.afterAll(async () => {
    await deleteFile({
      namespace: systemNamespaceName,
      path: systemServiceName,
    });

    if (cleanUpSystemNamespace) {
      await deleteNamespace(systemNamespaceName);
    }
  });

  test("services will also be listed in the system namespace", async ({
    page,
  }) => {
    await page.goto(`/n/${systemNamespaceName}/services`, {
      waitUntil: "networkidle",
    });

    await expect(
      page.getByTestId("service-row").getByText(systemServiceName),
      "it renders the service name"
    ).toBeVisible();
  });
});
