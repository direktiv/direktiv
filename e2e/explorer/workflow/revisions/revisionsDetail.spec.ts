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
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflow,
    },
  });

  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=${revision}`,
    { waitUntil: "networkidle" }
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
    "it navigated to the revision details and shows the title"
  ).toContainText(secondRevisionName);

  // go back to list page
  await page.getByTestId(`revisions-detail-back-link`).click();

  await expect(
    page.getByTestId(`workflow-revisions-link-item-${secondRevisionName}`),
    "it navigated back to the revisions list and finds the link again"
  ).toBeVisible();
});

test("it is possible to revert a revision within the details page", async ({
  page,
}) => {
  const workflow = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [, secondRevision, latestRevisions],
  } = await createWorkflowWithThreeRevisions(namespace, workflow);

  const secondRevisionName = secondRevision.revision.name;

  // check the content of the latest revision
  await page.goto(`/${namespace}/explorer/workflow/active/${workflow}`);
  await expect(
    page.getByTestId("workflow-editor"),
    "it displays the latest workflow content in the editor"
  ).toContainText(atob(latestRevisions?.revision?.source).replace(/\n/g, ""));

  // open the details page of the second revision
  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=${secondRevisionName}`,
    { waitUntil: "networkidle" }
  );
  await expect(
    page.getByTestId("revisions-detail-editor"),
    "it displays the reverted workflow content in the editor"
  ).toContainText(atob(secondRevision?.revision?.source).replace(/\n/g, ""), {
    timeout: 10000,
  });
  // open and submit revert dialog
  await page.getByTestId(`revisions-detail-revert-btn`).click();
  await page.getByTestId(`dialog-revert-revision-btn-submit`).click();

  // click the toast button to open the editor
  await page.getByTestId("workflow-revert-revision-toast-action").click();

  await expect(
    page.getByTestId("workflow-editor"),
    "it displays the reverted workflow content in the editor"
  ).toContainText(atob(secondRevision?.revision?.source).replace(/\n/g, ""), {
    timeout: 10000,
  });
});

test('it does not show the actions button on the revision details of the "latest" revision', async ({
  page,
}) => {
  const workflow = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [, secondRevision],
  } = await createWorkflowWithThreeRevisions(namespace, workflow);

  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=${secondRevision.revision.name}`
  );

  await expect(
    page.getByTestId("revisions-detail-revert-btn"),
    "revisions actions button is visible on the non-latest revision"
  ).toBeVisible();

  await page.goto(
    `/${namespace}/explorer/workflow/revisions/${workflow}?revision=latest`,
    {
      // wait for all data to be loaded before checking for something to be
      // not visible because in a very early render process the button would
      // not be rendered yet and this test would accidentally pass
      waitUntil: "networkidle",
    }
  );

  await expect(
    page.getByTestId("revisions-detail-revert-btn"),
    "revisions actions button is not visible"
  ).not.toBeVisible();
});
