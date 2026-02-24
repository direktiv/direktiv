import {} from "~/util/helpers";

import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import { simpleWorkflow, workflowWithService } from "e2e/utils/workflows";

import { createFile } from "e2e/utils/files";
import { createInstance } from "e2e/instances/utils";
import { faker } from "@faker-js/faker";
import { findServiceWithApiRequest } from "e2e/services/utils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Workflow service list is empty by default", async ({ page }) => {
  const workflowName = faker.system.commonFileName("wf.ts");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: simpleWorkflow,
    mimeType: "application/x-typescript",
  });

  await page.goto(
    `/n/${namespace}/explorer/workflow/services/list/${workflowName}`
  );

  await expect(
    page.getByText("No services exist yet"),
    "it renders an empy list of services"
  ).toBeVisible();
});

test("Workflow service list shows all services mounted by the workflow", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("wf.ts");
  const workflowFile = await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: workflowWithService,
    mimeType: "application/x-typescript",
  });

  await createInstance({
    namespace,
    path: workflowName,
  });

  await expect
    .poll(
      async () =>
        await findServiceWithApiRequest({
          namespace,
          match: (service) =>
            service.filePath === workflowFile.data.path &&
            (service.conditions ?? []).some(
              (c) => c.type === "Available" && c.status === "True"
            ),
        }),
      {
        timeout: 50000,
        message: "the service in the backend is in state Available",
      }
    )
    .toBeTruthy();

  await page.goto(
    `/n/${namespace}/explorer/workflow/services/list/${workflowName}`,
    {
      waitUntil: "networkidle",
    }
  );

  await expect(
    page.getByTestId("service-row"),
    "it renders one row of services"
  ).toHaveCount(1);

  await expect(
    page
      .getByTestId("service-row")
      .getByRole("link", { name: workflowFile.data.path }),
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
      .getByRole("cell", { name: "0", exact: true }),
    "it renders the scale of the service"
  ).toBeVisible();

  await expect(
    page.getByTestId("service-row").getByRole("cell", { name: "small" }),
    "it renders the size of the service"
  ).toBeVisible();
});
