import { createNamespace, deleteNamespace } from "../utils/namespace";
import {
  createRedisServiceFile,
  findServiceViaApi,
  serviceWithAnError,
} from "./utils";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
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
        await findServiceViaApi({
          namespace,
          searchFn: (service) =>
            service.filePath === "/redis-service.yaml" &&
            (service.conditions ?? []).some((c) => c.type === "UpAndReady"),
        }),
      "the service in the backend is in an UpAndReady state"
    )
    .toBeTruthy();

  const createdService = await findServiceViaApi({
    namespace,
    searchFn: (service) => service.filePath === "/redis-service.yaml",
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
    "it renders the service image"
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
    page.getByTestId("service-detail-header").getByText("UpAndReady"),
    "it renders the UpAndReady status"
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

  await page.getByRole("tab", { name: "Pod 2 of 2" }).click();

  await expect(
    page.getByText(`Logs for ${createdService.id}_2`),
    "after clicking on the second pod tab, it renders the headline for the second pods logs"
  ).toBeVisible();
});
