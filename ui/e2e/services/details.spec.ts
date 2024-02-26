import {
  createHttpServiceFile,
  findServiceWithApiRequest,
  serviceWithAnError,
} from "./utils";
import { createNamespace, deleteNamespace } from "../utils/namespace";
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
  const serviceFile = await createFile({
    name: "http-service.yaml",
    namespace,
    type: "service",
    yaml: createHttpServiceFile({ scale: 2 }),
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

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === serviceFile.data.path,
  });

  if (!createdService) throw new Error("could not find service");

  await page.goto(`/${namespace}/services/${createdService.id}`);

  await expect(
    page.getByRole("heading", { name: createdService.id, exact: true }),
    "it renders the service id as a heading"
  ).toBeVisible();

  await expect(
    page.getByRole("link", { name: serviceFile.data.path, exact: true }),
    "it renders a link to the service file"
  ).toBeVisible();

  await expect(
    page
      .getByTestId("service-detail-header")
      .getByText("gcr.io/direktiv/functions/http-request:1.0"),
    "it renders the service image name"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-detail-header").getByText("2 / 2"),
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
    page.getByText("Logs for"),
    "it renders the headline for the pods logs"
  ).toBeVisible();

  const firstPodLogHeadline = await page.getByText("Logs for").innerText();

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

  await expect(
    page.getByText("Logs for"),
    "it renders the headline for the pods logs"
  ).toBeVisible();

  await expect(
    await page.getByText("Logs for").innerText(),
    "after clicking on the second pod button, it has updated the headline for the corresponding pod logs"
  ).not.toEqual(firstPodLogHeadline);

  const waitForRefresh = page.waitForResponse((response) => {
    const servicesApiCall = `/api/v2/namespaces/${namespace}/services`;
    return (
      response.status() === 200 && response.url().endsWith(servicesApiCall)
    );
  });

  await page
    .getByTestId("service-detail-header")
    .getByRole("button")
    .nth(1)
    .click();

  await expect(
    await waitForRefresh,
    "after clicking on the refetch button, a network request to the services was made"
  ).toBeTruthy();
});

test("Service details page renders no logs when the service did not mount", async ({
  page,
}) => {
  const serviceFile = await createFile({
    name: "error-service.yaml",
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
      "the service in the backend is in an Error state"
    )
    .toBeTruthy();

  const createdService = await findServiceWithApiRequest({
    namespace,
    match: (service) => service.filePath === serviceFile.data.path,
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
