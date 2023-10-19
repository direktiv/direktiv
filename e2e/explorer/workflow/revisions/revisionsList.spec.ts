import { Page, expect, test } from "@playwright/test";
import { actionDeleteRevision, actionWaitForSuccessToast } from "../utils";
import { createNamespace, deleteNamespace } from "../../../utils/namespace";

import { createWorkflow } from "../../../utils/node";
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

const actionCreateRevisionAndTag = async (page: Page) => {
  const name = `${faker.internet.domainWord()}.yaml`;
  const {
    revisionsReponse: [firstRev],
  } = await createWorkflowWithThreeRevisions(namespace, name);
  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);

  const firstItemMenuTrg = page.getByTestId(
    `workflow-revisions-item-menu-trg-${firstRev.revision.name}`
  );
  await firstItemMenuTrg.click();

  const createTagTrg = page.getByTestId(
    `workflow-revisions-trg-create-tag-dlg-${firstRev.revision.name}`
  );
  await createTagTrg.click();

  // type name and save & wait for the success toast
  const inputName = page.getByTestId("dialog-create-tag-input-name");
  const newTag = faker.random.alphaNumeric(9);
  await inputName.type(newTag);
  await page.getByTestId("dialog-create-tag-btn-submit").click();
  await actionWaitForSuccessToast(page);
  return [firstRev.revision.name, newTag] as const;
};

test("it is possible to navigate to the revisions tab", async ({ page }) => {
  const workflow = await createWorkflow(
    namespace,
    faker.internet.domainWord() + ".yaml"
  );
  await page.goto("/");
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for a namespace"
  ).toBeVisible();
  // at this point, any namespace may be loaded.
  // let's navigate to the test's namespace via breadcrumbs.
  await page.getByTestId("dropdown-trg-namespace").click();
  await page
    .getByRole("option", {
      name: namespace,
    })
    .click();
  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);

  await page.getByTestId(`explorer-item-link-${workflow}`).click();
  const revisionsTab = page.getByTestId("workflow-tabs-trg-revisions");
  await expect(
    revisionsTab,
    "screen should have activeRevision tab"
  ).toBeVisible();
  await revisionsTab.click();
  await expect(page, "the workflow is reflected in the url").toHaveURL(
    `${namespace}/explorer/workflow/revisions/${workflow}`
  );
});

test("latest is the only revision by default", async ({ page }) => {
  const workflow = await createWorkflow(
    namespace,
    faker.internet.domainWord() + ".yaml"
  );
  await page.goto(`${namespace}/explorer/workflow/revisions/${workflow}`);

  const revisions = page.getByTestId(/workflow-revisions-link-item-/);
  await expect(revisions, "revisions should have the name latest").toHaveText(
    "latest"
  );
  await expect(revisions, "number of revisions should be one").toHaveCount(1);
});

test("it is possible to revert to the previous the workflow in the revisions list", async ({
  page,
}) => {
  const name = faker.system.commonFileName("yaml");
  const {
    revisionsPayload: [firstContent],
    revisionsReponse: [firstRev],
  } = await createWorkflowWithThreeRevisions(namespace, name);
  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);

  const firstItemMenuTrg = page.getByTestId(
    `workflow-revisions-item-menu-trg-${firstRev.revision.name}`
  );
  await firstItemMenuTrg.click();

  await expect(
    page.getByTestId(
      `workflow-revisions-item-menu-content-${firstRev.revision.name}`
    ),
    "after click menu trigger, menu content should appear"
  ).toBeVisible();

  const revertTrg = page.getByTestId(
    `workflow-revisions-trg-revert-dlg-${firstRev.revision.name}`
  );
  await revertTrg.click();

  const btnRevert = page.getByTestId("dialog-revert-revision-btn-submit");
  await expect(
    btnRevert,
    "when the dialog appears, revert button should be enabled after loading"
  ).toBeEnabled();
  await btnRevert.click();
  await actionWaitForSuccessToast(page);
  await page.goto(`${namespace}/explorer/workflow/active/${name}`);

  const textArea = page.getByRole("textbox");
  await expect
    .poll(
      async () => await textArea.inputValue(),
      "it displays the reverted workflow content in the editor"
    )
    .toBe(firstContent);
});

test("it is possible to delete the revision", async ({ page }) => {
  const name = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [firstRev],
  } = await createWorkflowWithThreeRevisions(namespace, name);
  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);

  const firstItemMenuTrg = page.getByTestId(
    `workflow-revisions-item-menu-trg-${firstRev.revision.name}`
  );
  await firstItemMenuTrg.click();

  await expect(
    page.getByTestId(
      `workflow-revisions-item-menu-content-${firstRev.revision.name}`
    ),
    "after click menu trigger, menu content should appear"
  ).toBeVisible();

  // click on the delete button to show the delete dialog
  const deleteTrg = page.getByTestId(
    `workflow-revisions-trg-delete-dlg-${firstRev.revision.name}`
  );
  await deleteTrg.click();
  const deleteDialog = page.getByTestId("dialog-delete-revision");
  await expect(
    deleteDialog,
    "after click delete menu, it should show the delete confirm dialog"
  ).toBeVisible();
  const submitButton = page.getByTestId("dialog-delete-revision-btn-submit");
  await submitButton.click();

  await actionWaitForSuccessToast(page);

  // after delete success, confirm that the revision item isn't visible anymore
  const revisionItem = page.getByTestId(
    `workflow-revisions-link-item-${firstRev.revision.name}`
  );
  await expect(
    revisionItem,
    "revision item should not be visible in the page"
  ).not.toBeVisible();
});

test("it is possible to create and delete tags", async ({ page }) => {
  // make revision and create a tag for that
  const [revision, tag] = await actionCreateRevisionAndTag(page);

  // validate all appears as expectation
  const tagItem = page.getByTestId(`workflow-revisions-link-item-${tag}`);
  const revisionItem = page.getByTestId(
    `workflow-revisions-link-item-${revision}`
  );
  await expect(
    revisionItem,
    "revision item should appear in the revisions list"
  ).toBeVisible();
  await expect(
    tagItem,
    "tag item should appear in the revisions list"
  ).toBeVisible();
  await page.reload();
  await expect(
    tagItem,
    "after reload, the new revision item should still be visible"
  ).toBeVisible();

  await actionDeleteRevision(page, tag);
  await actionWaitForSuccessToast(page);

  await expect(
    tagItem,
    "after deleting, tag item should not exist"
  ).not.toBeVisible();
  await expect(revisionItem, "revision item should still exist").toBeVisible();
});

test("it is possible to delete the tag by deleting the base revision", async ({
  page,
}) => {
  // create a revision, and a tag from that revision
  const [revision, tag] = await actionCreateRevisionAndTag(page);

  // delete the revision
  await actionDeleteRevision(page, revision);
  await actionWaitForSuccessToast(page);

  // both the revision and the tag should disappear from the list
  const revisionItem = page.getByTestId(
    `workflow-revisions-link-item-${revision}`
  );
  await expect(
    revisionItem,
    "revision item should not be visible in the page"
  ).not.toBeVisible();
  const tagItem = page.getByTestId(`workflow-revisions-link-item-${tag}`);
  await expect(
    tagItem,
    "tag item should not be visible in the page"
  ).not.toBeVisible();

  await page.reload();
  await expect(
    revisionItem,
    "after reload, revision item should not be visible in the page"
  ).not.toBeVisible();
  await expect(
    tagItem,
    "after reload, tag item should not be visible in the page"
  ).not.toBeVisible();
});

test('the context menu is not shown in the list for the "latest" revision', async ({
  page,
}) => {
  const workflow = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [, secondRevision],
  } = await createWorkflowWithThreeRevisions(namespace, workflow);

  await page.goto(`/${namespace}/explorer/workflow/revisions/${workflow}`);

  await expect(
    page
      .getByTestId(
        `workflow-revisions-item-last-row-${secondRevision.revision.name}`
      )
      .getByTestId(
        `workflow-revisions-item-menu-trg-${secondRevision.revision.name}`
      ),
    "the last row of the second revision has a button to show the context menu"
  ).toBeVisible();

  expect(
    await page
      .getByTestId(
        `workflow-revisions-item-last-row-${secondRevision.revision.name}`
      )
      .innerText(),
    'the last row of the "latest" revision has no content at all'
  ).toBe("");
});
