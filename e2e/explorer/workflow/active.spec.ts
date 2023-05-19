import { Page, expect, test } from "@playwright/test";
import {
  createNamespace,
  deleteNamespace,
} from "../../utils/namespace";

import {
  createWorkflow,
} from "../../utils/node";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";
const defaultDescription = 'A simple \'no-op\' state that returns \'Hello world!\''

const testSaveWorkflow = async (page: Page) =>{
  const description = page.getByText(defaultDescription);
  await description.click()
  // type random text in that textarea which is for description
  const testText = faker.random.alphaNumeric(9);
  await page.type("textarea", testText);
  
  //now click on Save
  const saveButton = page.getByTestId("workflow-editor-btn-save");
  await saveButton.click();

  // save button should be disabled/enabled while/after the api call
  await expect(saveButton, 'save button should be disabled while the api call').toBeDisabled();
  await expect(saveButton, 'save button should be enabled after the api call').toBeEnabled();
  
  // after saving is completed screen should have those new changed text before/after the page reload
  await expect(page.getByText(testText), 'screen should have the text written').toBeVisible();
  await page.reload({waitUntil: 'load'});
  await expect(page.getByText(testText), 'screen should have the text written').toBeVisible();

  // check the text at the bottom left
  await expect(page.getByTestId("workflow-txt-updated"), "text should be Updated a few seconds").toHaveText("Updated a few seconds");
}

const testMakeRevision = async (page: Page) =>{
   const revisionTrigger = page.getByTestId("workflow-edit-trg-revision");
  await revisionTrigger.click();
  const makeRevisionButton = page.getByTestId("workflow-editor-btn-make-revision");
  await makeRevisionButton.click();

  // check the success toast action button
  const toastAction = page.getByTestId("make-revision-toast-success-action");
  await expect(toastAction, 'success toast should appear after make-revision button click').toBeVisible();

  await toastAction.click();
  await expect(toastAction, 'success toast should disappear after toast action click').toBeHidden();
  await expect(page, "page url should have revision param").toHaveURL(/revision=/);
  const revisionId = page.url().split('revision=')[1];
  await expect(page.getByText(revisionId), "revisionId should be in the revision list").toBeVisible();
  //go back to the workflow editor
  await page.getByTestId('workflow-tabs-trg-activeRevision').click()
}

test.beforeEach(async ({page}) => {
  namespace = await createNamespace();
  workflow = await createWorkflow(namespace, faker.git.shortSha() + '.yaml');
  test.setTimeout(120000)
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

  await page.getByTestId(`explorer-item-link-${workflow}`).click()
  await expect(page.getByTestId('workflow-tabs-trg-activeRevision'), 'screen should have activeRevision tab').toBeVisible()
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("is that possible to save the workflow", async ({
  page,
}) => {
  // click on the description so it can have input focus
  await testSaveWorkflow(page)
});

test("is that possible to make the revision", async ({
  page,
}) => {
  //now click on Revision Menu Trigger
  await testMakeRevision(page)
});

test("is that possible to revert the revision", async ({
  page,
}) => {
  await testMakeRevision(page)
  await testSaveWorkflow(page)
  //now click on Revision Menu Trigger
  const revisionTrigger = page.getByTestId("workflow-edit-trg-revision");
  await revisionTrigger.click();
  const revertRevisionButton = page.getByTestId("workflow-editor-btn-revert-revision");
  await revertRevisionButton.click();

  // check the success toast success
  const successToast = page.getByTestId("toast-success");
  await expect(successToast, 'success toast should appear after revert-revision button click').toBeVisible();
  await page.getByTestId("toast-close").click();
  await expect(successToast, 'success toast should disappear after click toast-close').toBeHidden();

  // check the description is reverted
  await expect(page.getByText(defaultDescription), "description should be reverted to the default").toBeVisible();
  // check the bottom left
  await expect(page.getByTestId("workflow-txt-updated"), "text should be Updated a few seconds").toHaveText("Updated a few seconds");

  //check both after page reload
  await page.reload({waitUntil: 'load'});
  await expect(page.getByText(defaultDescription), "description should be reverted to the default").toBeVisible();
  await expect(page.getByTestId("workflow-txt-updated"), "text should be Updated a few seconds").toHaveText("Updated a few seconds");
});

