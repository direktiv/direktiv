import { Page, expect, test } from "@playwright/test";
import {
  actionMakeRevision,
  actionRevertRevision,
  actionWaitForSuccessToast,
} from "./utils";
import { createNamespace, deleteNamespace } from "../../utils/namespace";

import { createWorkflow } from "../../utils/node";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";
const defaultDescription = "A simple 'no-op' state that returns 'Hello world!'";

test.beforeEach(async () => {
  namespace = await createNamespace();
  workflow = await createWorkflow(
    namespace,
    faker.internet.domainWord() + ".yaml"
  );
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

const actionNavigateToActiveWorkflow = async (page: Page) => {
  await page.goto(`${namespace}/explorer/workflow/active/${workflow}`);
};

const testSaveWorkflow = async (page: Page) => {
  const description = page.getByText(defaultDescription);
  await description.click();
  // type random text in that textarea which is for description
  const testText = faker.random.alphaNumeric(9);
  await page.type("textarea", testText);

  // now click on Save
  const saveButton = page.getByTestId("workflow-editor-btn-save");
  await saveButton.click();

  // Commented out since this is not a critical step, but maybe we can enable
  // it again at some point after learning more about the following problem:
  // These steps fail locally, but work with a remote API. They works locally
  // with throttling enabled in devtools. This implies the request is completed
  // so fast there is not enough time to detect the inactive button.
  // await expect(
  //   saveButton,
  //   "save button should be disabled during the api call"
  // ).toBeDisabled();
  // await expect(
  //   saveButton,
  //   "save button should be enabled after the api call"
  // ).toBeEnabled();

  // after saving is completed screen should have those new changed text before/after the page reload
  await expect(
    page.getByText(testText),
    "after saving, screen should have the updated text"
  ).toBeVisible();
  await page.reload({ waitUntil: "networkidle" });
  await expect(
    page.getByText(testText),
    "after reloading, screen should have the updated text"
  ).toBeVisible();

  // check the text at the bottom left
  await expect(
    page.getByTestId("workflow-txt-updated"),
    "text should be Updated a few seconds ago"
  ).toHaveText("Updated a few seconds ago");
};

const testMakeRevision = async (page: Page) => {
  await actionMakeRevision(page);
  // check the success toast action button
  const toastAction = page.getByTestId("make-revision-toast-success-action");
  await expect(
    toastAction,
    "success toast should appear after make-revision button click"
  ).toBeVisible();

  await toastAction.click();
  await expect(
    toastAction,
    "success toast should disappear after toast action click"
  ).toBeHidden();
  await expect(page, "page url should have revision param").toHaveURL(
    /revision=/
  );
  const revisionId = page.url().split("revision=")[1];

  if (!revisionId) throw new Error("revisionId should be present in the url");

  await expect(
    page.getByText(revisionId),
    "revisionId should be in the revision list"
  ).toBeVisible();
  // go back to the workflow editor
  await page.getByTestId("workflow-tabs-trg-activeRevision").click();
};

test("it is possible to navigate to the active revision", async ({ page }) => {
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

  await expect(
    page.getByTestId("workflow-tabs-trg-activeRevision"),
    "screen should have activeRevision tab"
  ).toBeVisible();

  await expect(page, "the workflow is reflected in the url").toHaveURL(
    `${namespace}/explorer/workflow/active/${workflow}`
  );
});

test("it is possible to save the workflow", async ({ page }) => {
  await actionNavigateToActiveWorkflow(page);
  await testSaveWorkflow(page);
});

test("it is possible to make the revision", async ({ page }) => {
  await actionNavigateToActiveWorkflow(page);
  await testMakeRevision(page);
});

test("it is possible to revert the revision", async ({ page }) => {
  await actionNavigateToActiveWorkflow(page);
  await testMakeRevision(page);
  await testSaveWorkflow(page);
  await actionRevertRevision(page);
  await actionWaitForSuccessToast(page);

  // check the description is reverted
  await expect(
    page.getByText(defaultDescription),
    "description should be reverted to the default"
  ).toBeVisible();

  // check the bottom left
  await expect(
    page.getByTestId("workflow-txt-updated"),
    "text should be Updated a few seconds ago"
  ).toHaveText("Updated a few seconds ago");

  // check both after page reload
  await page.reload({ waitUntil: "networkidle" });

  await expect(
    page.getByText(defaultDescription),
    "description should be reverted to the default"
  ).toBeVisible();

  await expect(
    page.getByTestId("workflow-txt-updated"),
    "text should be Updated a few seconds ago"
  ).toHaveText("Updated a few seconds ago");
});
