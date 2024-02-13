import {
  createHttpServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
import { createNamespace, deleteNamespace } from "../utils/namespace";
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

test("Service details page provides information about the service", async ({
  page,
}) => {
  await createWorkflow({
    payload: createHttpServiceFile({
      scale: 2,
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "http-service.yaml",
    },
    headers,
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === "/http-service.yaml" &&
            (service.conditions ?? []).some(
              (c) => c.type === "ConfigurationsReady"
            ),
        }),
      "the service in the backend is in state ConfigurationsReady"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === "/http-service.yaml",
  });

  if (!createdService) throw new Error("could not find service");

  await page.goto(`/${namespace}/services/${createdService.id}`);

  await expect(
    page.getByRole("heading", { name: createdService.id, exact: true }),
    "it renders the service id as a heading"
  ).toBeVisible();

  await expect(
    page.getByRole("link", { name: "/http-service.yaml", exact: true }),
    "it renders a link to the service file"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .getByText("gcr.io/direktiv/functions/http-request:1.0"),
    "it renders the service image name"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-detail-header").getByText("scale2"),
    "it renders the service scale"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-detail-header").getByText("small"),
    "it renders the service size"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .filter({ hasText: "ConfigurationsReady" }),
    "it renders the ConfigurationsReady status"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .locator("a")
      .filter({ hasText: "1 environment variable" }),
    "it renders the environment variable count"
  ).toBeVisible();

  await expect(
    page.getByText("Serving HTTP request at http://[::]:8080"),
    "it renders the log entries"
  ).toBeVisible();

  await expect(
    page.getByText(/received [0-9]+ log (entry|entries)/),
    "it renders the log summary"
  ).toBeVisible();

  await expect(
    page.getByRole("tab", { name: "Pod 1 of 2" }),
    "it renders the the first pod tab"
  ).toBeVisible();

  await expect(
    page.getByRole("tab", { name: "Pod 2 of 2" }),
    "it renders the the second pod tab"
  ).toBeVisible();

  await page.click("text=Pod 2 of 2");

  const waitForRefresh = page.waitForResponse((response) => {
    const servicesApiCall = `/api/v2/namespaces/${namespace}/services`;
    return (
      response.status() === 200 && response.url().endsWith(servicesApiCall)
    );
  });

  await page.getByTestId("service-detail-header").getByRole("button").click();

  await expect(
    await waitForRefresh,
    "after clicking on the refetch button, a network request to the services was made"
  ).toBeTruthy();
});

test("Service details page renders no logs when the service did not mount", async ({
  page,
}) => {
  await createWorkflow({
    payload: serviceWithAnError,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "http-service.yaml",
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
            service.error !== null,
        }),
      "the service in the backend is in an Error state"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === "/http-service.yaml",
  });

  if (!createdService) throw new Error("could not find service");

  await page.goto(`/${namespace}/services/${createdService.id}`);

  await expect(
    page.getByRole("heading", { name: createdService.id, exact: true }),
    "it renders the service id as a heading"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .locator("a")
      .filter({ hasText: "Error" }),
    "it renders the Error status"
  ).toBeVisible();

  await expect(
    page.getByText("No running pods"),
    "it renders a message that no pods are running"
  ).toBeVisible();
});
