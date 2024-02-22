import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/components/modals/CreateNew/Workflow/templates";
import { createFile } from "e2e/utils/files";
import { createWorkflowVariables } from "e2e/utils/variables";
import { faker } from "@faker-js/faker";
import { headers } from "e2e/utils/testutils";
import { setVariable } from "~/api/tree/mutate/setVariable";
import { waitForSuccessToast } from "./utils";

let namespace = "";
let workflow = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
  workflow = faker.system.commonFileName("yaml");

  await createFile({
    name: workflow,
    namespace,
    type: "workflow",
    yaml: basicWorkflow.data,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to the workflow settings page and use pagination", async ({
  page,
}) => {
  await createWorkflowVariables(namespace, workflow, 15);

  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);
  await expect(
    page.getByTestId("variable-row"),
    "there should be 10 variables on the first page"
  ).toHaveCount(10);

  await expect(page.getByTestId("pagination-wrapper")).toBeVisible();
  await expect(
    page.getByTestId("pagination-wrapper").getByRole("button", { name: "1" })
  ).toBeVisible();
  await expect(
    page.getByTestId("pagination-wrapper").getByRole("button", { name: "2" })
  ).toBeVisible();

  await page.getByTestId("pagination-btn-right").click();
  await expect(
    page.getByTestId("variable-row"),
    "there should be 5 variables on the second page"
  ).toHaveCount(5);

  await page.getByTestId("pagination-btn-left").click();
  await expect(
    page.getByTestId("variable-row"),
    "it is possible to go back to page 1"
  ).toHaveCount(10);
});

test("it is possible to create a variable", async ({ page }) => {
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);

  const subject = {
    name: "workflow-variable",
    value: "this variable will be created via the form",
    mimeType: "plaintext",
  };

  await page.getByTestId("variable-create").click();

  await expect(
    page.getByRole("heading", { name: "Add a workflow variable" }),
    "create variable form should be visible"
  ).toBeVisible();

  await page.getByLabel("Name").fill(subject.name);

  await page.locator(".view-lines").click();
  await page.locator(".view-lines").type(subject.value);

  await page.getByLabel("Mimetype").click();
  await page.getByLabel(subject.mimeType).click();
  await page.getByRole("button", { name: "Create" }).click();

  const successToast = page.getByTestId("toast-success");
  await expect(successToast, "a success toast appears").toBeVisible();
  await expect(
    page.getByTestId("variable-row"),
    "there should be 1 variable in the list"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("variable-row"),
    "there should be 1 variable in the list"
  ).toContainText(subject.name);
});

test("it is possible to update variables", async ({ page }) => {
  /* set up test data */
  const subject = await setVariable({
    payload: "edit me",
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: workflow,
      name: "editable-var",
    },
    headers: {
      ...headers,
      "content-type": "text/plain",
    },
  });

  if (!subject) {
    throw new Error("error setting up test data");
  }

  /* visit page and edit variable */
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);

  await page.getByTestId(`dropdown-trg-item-${subject.key}`).click();
  await page.getByRole("button", { name: "edit" }).click();

  await expect(
    page.getByRole("heading", { name: `Edit ${subject.key}` }),
    "it opens the edit form"
  ).toBeVisible();

  await expect(page.getByLabel("Mimetype")).toContainText(subject.mimeType);
  await page.getByLabel("Mimetype").click();
  await page.getByLabel("JSON").click();

  await expect(
    page.locator(".view-lines"),
    "the variable's content is rendered in the editor"
  ).toContainText("edit me");

  await page.locator(".view-lines").click();
  for (let i = 0; i < 7; i++) {
    await page.locator(".view-lines").press("Backspace");
  }
  await page.locator(".view-lines").type('{"foo": "bar"}');

  /* save changes and assert they have been persisted */
  await page.getByRole("button", { name: "Save" }).click();

  await waitForSuccessToast(page);
  await page.reload();

  await expect(
    page.getByTestId("variable-row"),
    "there should be 1 variable in the list"
  ).toHaveCount(1);

  await page.getByTestId(`dropdown-trg-item-${subject.key}`).click();
  await page.getByRole("button", { name: "edit" }).click();

  await expect(page.getByLabel("Mimetype")).toContainText("application/json");
  await expect(
    page.locator(".view-lines"),
    "editor should have the updated value"
  ).toContainText('{"foo": "bar"}');
});

test("it is possible to delete variables", async ({ page }) => {
  /* set up test data */
  const variables = await createWorkflowVariables(namespace, workflow, 4);
  const subject = variables[2];

  if (!subject) {
    throw new Error("error setting up test data");
  }

  /* visit page and delete variable */
  await page.goto(`/${namespace}/explorer/workflow/settings/${workflow}`);

  await page.getByTestId(`dropdown-trg-item-${subject.key}`).click();
  await page.getByRole("button", { name: "delete" }).click();

  await expect(
    page.getByLabel("Confirmation required").getByText(subject.key),
    "it renders the confirmation dialog"
  ).toBeVisible();
  await page.getByRole("button", { name: "Delete" }).click();
  await waitForSuccessToast(page);

  await expect(
    page.getByTestId("variable-row"),
    "the variable is no longer rendered in the list"
  ).toHaveCount(variables.length - 1);
});
