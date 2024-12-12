import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createWorkflowVariables } from "../../utils/variables";
import { faker } from "@faker-js/faker";
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
    yaml: "direktiv_api: workflow/v1\nstates:\n- id: noop\n  type: noop",
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
});

test("bulk delete workflow variables", async ({ page }) => {
  // Create 3 test variables
  await createWorkflowVariables(namespace, workflow, 3);

  await page.goto(`/n/${namespace}/explorer/workflow/settings/${workflow}`);

  // Check first checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").first().check();
  await page.getByRole("button", { name: "Delete" }).click();
  await expect(
    page.getByText(`Are you sure you want to delete variable`, { exact: false })
  ).toBeVisible();
  await page.getByRole("button", { name: "Cancel" }).click();

  // Check second checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").nth(0).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(1).check();
  await page.getByRole("button", { name: "Delete" }).click();
  await expect(
    page.getByText("Are you sure you want to delete 2 variables?", {
      exact: true,
    })
  ).toBeVisible();
  await page.getByRole("button", { name: "Cancel" }).click();

  // Check third checkbox and click delete
  await page.getByTestId("item-name").getByRole("checkbox").nth(0).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(1).check();
  await page.getByTestId("item-name").getByRole("checkbox").nth(2).check();
  await page.getByRole("button", { name: "Delete" }).click();
  await expect(
    page.getByText("Are you sure you want to delete all variables?", {
      exact: true,
    })
  ).toBeVisible();

  // Confirm deletion
  await page.getByRole("button", { name: "Delete" }).click();
  await waitForSuccessToast(page);

  await expect(page.getByTestId("item-name")).toHaveCount(0);
});
