import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

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

test("it is possible to navigate to the code editor ", async ({ page }) => {
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
    page.getByTestId("workflow-tabs-trg-editor"),
    "screen should have code editor tab"
  ).toBeVisible();

  await expect(page, "the workflow is reflected in the url").toHaveURL(
    `${namespace}/explorer/workflow/edit/${workflow}`
  );
});

test("it is possible to save the workflow", async ({ page }) => {
  await page.goto(`${namespace}/explorer/workflow/edit/${workflow}`);

  const editorElement = page.getByText(defaultDescription);
  await editorElement.click();

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
});
