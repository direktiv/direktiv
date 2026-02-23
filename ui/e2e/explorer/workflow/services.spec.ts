import {} from "~/util/helpers";

import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { faker } from "@faker-js/faker";
import { simpleWorkflow } from "e2e/utils/workflows";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("workflow service list is empty by default", async ({ page }) => {
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
