import { createNamespace, deleteNamespace } from "../utils/namespace";
import {
  createRedisServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
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

test("Service details page provides information about the service", async ({
  page,
}) => {
  await createWorkflow({
    payload: createRedisServiceFile({
      scale: 2,
    }),
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

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === "/redis-service.yaml",
  });

  if (!createdService) throw new Error("could not find service");

  await page.goto(`/${namespace}/services/${createdService.id}`);

  await expect(
    page.getByRole("heading", { name: createdService.id, exact: true }),
    "it renders the service id as a heading"
  ).toBeVisible();

  await expect(
    page.getByRole("link", { name: "/redis-service.yaml", exact: true }),
    "it renders a link to the service file"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-detail-header").getByText("imageredis"),
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
    page.getByTestId("service-detail-header").getByText("redis-server"),
    "it renders the service cmd"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .locator("a")
      .filter({ hasText: "UpAndReady" }),
    "it renders the UpAndReady status"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .locator("a")
      .filter({ hasText: "1 environment variable" }),
    "it renders the environment variable count"
  ).toBeVisible();

  await expect(
    page.getByText(`Logs for ${createdService.id}_1`),
    "it renders the headline for the first pods logs"
  ).toBeVisible();

  await expect(
    page.getByText("Ready to accept connections tcp"),
    "it renders the log entries"
  ).toBeVisible();

  await expect(
    page.getByText("received 8 log entries"),
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

  /**
   * since the logs from both pods are equal, we can only check, if changing
   * the tab will trigger a network request to the logs of the second pod.
   * We also check if the server responds with a 200 status code to make sure
   * that the logs are comming in.
   */
  await expect(
    await page.waitForResponse((response) => {
      const logsApiCall = `/api/v2/namespaces/${namespace}/services/${createdService.id}/pods/${createdService.id}_2/logs`;
      return response.status() === 200 && response.url().endsWith(logsApiCall);
    }),
    "after clicking on the second pod tab, a network request to the log was made"
  ).toBeTruthy();

  await expect(
    page.getByText(`Logs for ${createdService.id}_2`),
    "after clicking on the second pod tab, it renders the headline for the second pods logs"
  ).toBeVisible();

  await page.getByTestId("service-detail-header").getByRole("button").click();

  /**
   * since the logs from both pods are equal, we can only check, if changing
   * the tab will trigger a network request to the logs of the second pod.
   * We also check if the server responds with a 200 status code to make sure
   * that the logs are comming in.
   */
  await expect(
    await page.waitForResponse((response) => {
      const servicesApiCall = `/api/v2/namespaces/${namespace}/services`;
      return (
        response.status() === 200 && response.url().endsWith(servicesApiCall)
      );
    }),
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
            service.error !== null,
        }),
      "the service in the backend is in an Error state"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === "/redis-service.yaml",
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
