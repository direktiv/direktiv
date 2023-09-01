import {
  MimeTypeSchema,
  mimeTypeToLanguageDict,
} from "~/pages/namespace/Settings/Variables/MimeTypeSelect";
import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { createWorkflowVariables } from "e2e/utils/workflow";
import { faker } from "@faker-js/faker";
import { headers } from "e2e/utils/testutils";

const { options } = MimeTypeSchema;

let namespace = "";
let workflow = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
  workflow = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: basicWorkflow.data,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflow,
    },
    headers,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to the workflow settings page and to see the pagination working", async ({
  page,
}) => {
  await createWorkflowVariables(namespace, workflow, 15);
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);
  const rows = page.getByTestId(/wf-settings-var-row-/);

  // created 15 items and shows 10 items with the pagination
  await expect(
    rows,
    "there should be 10 variables in the first page"
  ).toHaveCount(10);
  const btnNext = page.getByTestId("pagination-btn-right");

  await expect(btnNext, "next button should be enabled").toBeEnabled();
  await btnNext.click();
  await expect(
    rows,
    "there should be 5 variables in the last page"
  ).toHaveCount(5);
});

test("it is possible to create variables", async ({ page }) => {
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);
  const rows = page.getByTestId(/wf-settings-var-row-/);

  const createBtn = page.getByTestId("variable-create");
  await createBtn.click();
  await expect(
    page.getByTestId("wf-form-create-variable"),
    "create variable form should be visible"
  ).toBeVisible();

  const newVariable = {
    name: faker.internet.domainWord(),
    value: faker.random.words(20),
    mimeType: options[Math.floor(Math.random() * options.length)] || options[0],
  };

  await page.getByTestId("new-variable-name").fill(faker.lorem.word());

  const editor = page.locator(".view-lines");
  await editor.click();
  await editor.type(newVariable.value);

  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${newVariable.mimeType}`).click();
  await page.getByTestId("variable-create-submit").click();

  const successToast = page.getByTestId("toast-success");
  await expect(successToast, "a success toast appears").toBeVisible();
  await expect(rows, "there should be 1 variable in the list").toHaveCount(1);
});

test("it is possible to update variables", async ({ page }) => {
  await createWorkflowVariables(namespace, workflow, 1);
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);
  const rows = page.getByTestId(/wf-settings-var-row-/);

  const itemName = await page.getByTestId("item-name").textContent();
  const menuTrigger = page.getByTestId(`dropdown-trg-item-${itemName}`);
  await menuTrigger.click();

  const editMenu = page.getByTestId("dropdown-actions-edit");
  await editMenu.click();

  await expect(
    page.getByTestId("wf-form-edit-variable"),
    "edit form should be opened"
  ).toBeVisible();

  const newVariable = {
    name: faker.internet.domainWord(),
    value: faker.random.words(20),
    mimeType: options[Math.floor(Math.random() * options.length)] || options[0],
  };

  const editor = page.locator(".view-lines");
  await editor.click();
  await editor.type("add-new-value");

  await page.getByTestId("variable-trg-mimetype").click();
  await page.getByTestId(`var-mimetype-${newVariable.mimeType}`).click();
  await page.getByTestId("var-edit-submit").click();

  const successToast = page.getByTestId("toast-success");
  await expect(successToast, "a success toast appears").toBeVisible();

  await expect(rows, "there should be 1 variable in the list").toHaveCount(1);

  await expect(successToast, "wait till the toast disappear").toBeHidden({
    timeout: 10000,
  });
  //we open up the edit box again confirm that's what we previously updated
  await menuTrigger.click();
  await editMenu.click();

  const current_mimetype = await page
    .getByTestId(`variable-trg-mimetype`)
    .textContent();
  expect(
    current_mimetype?.toLowerCase().replace(/\s/g, ""),
    "the mimetype selected before should be shown"
  ).toBe(mimeTypeToLanguageDict[newVariable.mimeType].toLowerCase());
  await expect(editor, "editor should have the updated value").toContainText(
    "add-new-value"
  );
});
