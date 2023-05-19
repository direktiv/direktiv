import {
  checkIfNamespaceExists,
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "../../utils/namespace";
import {
  checkIfNodeExists,
  createDirectory,
  createWorkflow,
  deleteNode,
  workflowExamples,
} from "../../utils/node";
import { expect, test } from "@playwright/test";

// add tests for /namespace/explorer/workflow/active tab here.
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("is that possible to save the workflow", async ({
  page,
}) => {
  // visit page
  test.setTimeout(120000)
  await page.goto("/");
  const workflow = await createWorkflow(namespace, faker.git.shortSha() + '.yaml');
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

  // click on the description so it can have input focus
  const description = page.getByText('A simple \'no-op\' state that returns \'Hello world!\'');
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
  await expect(page.getByTestId("workflow-txt-updated"), "text should be Updated a few seconds ago").toHaveText("Updated a few seconds ago");
});