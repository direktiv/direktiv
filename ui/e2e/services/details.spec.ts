import { createNamespace, deleteNamespace } from "../utils/namespace";
import {
  createRedisServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
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

test("Service details page provides information about the service", async ({
  page,
}) => {
  await createFile({
    yaml: createRedisServiceFile({
      scale: 2,
    }),

    namespace,
    name: "redis-service.yaml",
    type: "service",
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
    page.getByText("log entries"),
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

  await expect(
    page.getByText(`Logs for ${createdService.id}_2`),
    "after clicking on the second pod tab, it renders the headline for the second pods logs"
  ).toBeVisible();

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
  await createFile({
    yaml: serviceWithAnError,
    namespace,
    name: "redis-service.yaml",
    type: "service",
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
