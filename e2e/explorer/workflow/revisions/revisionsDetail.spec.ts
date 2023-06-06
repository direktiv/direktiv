import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { createWorkflowWithThreeRevisions } from "../../../utils/revisions";
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

test("it is possible to navigate from the revision list to the details and back", async ({
  page,
}) => {
  const workflow = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [, secondRevision],
  } = await createWorkflowWithThreeRevisions(namespace, workflow);

  const secondRevisionName = secondRevision.revision.name;

  // revisions list
  await page.goto(`/${namespace}/explorer/workflow/revisions/${workflow}`);

  // open details page
  await page
    .getByTestId(`workflow-revisions-link-item-${secondRevisionName}`)
    .click();

  await expect(
    page.getByTestId("revisions-detail-title"),
    "it navigated to the revisions detail and shows the title"
  ).toContainText(secondRevisionName);

  // go back to list page
  await page.getByTestId(`revisions-detail-back-link`).click();

  // find the link in the list again to make sure we are back on the list page
  await expect(
    page.getByTestId(`workflow-revisions-link-item-${secondRevisionName}`),
    "it navigated back to the revisions list and finds the link again"
  ).toBeVisible();
});
