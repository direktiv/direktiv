import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test('it is possible to open the revision details of the "latest" revision', async ({
  page,
}) => {
  const workflow = faker.system.commonFileName("yaml");
  const revision = "latest";
  await createWorkflow({
    payload: basicWorkflow.data,
    headers: undefined,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflow,
    },
  });

  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=${revision}`
  );

  await expect(
    page.getByTestId("revisions-detail-title"),
    "it displays the revision title"
  ).toContainText(revision);

  await expect(
    page.getByTestId("revisions-detail-editor"),
    "it displays the workflow content in the editor"
  ).toContainText(basicWorkflow.data.replace(/\n/g, ""));
});
