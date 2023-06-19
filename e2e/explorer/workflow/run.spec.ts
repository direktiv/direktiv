import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "../../utils/node";
import { faker } from "@faker-js/faker";

let namespace = "";
let workflow = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
  workflow = await createWorkflow(namespace, faker.git.shortSha() + ".yaml");
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible top open the run workflow modal from the editor and the header of the workflow page", async ({
  page,
}) => {
  await page.goto(`${namespace}/explorer/workflow/active/${workflow}`);

  // open modal via editor button
  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog from the editor button"
  ).toBeVisible();
  await page.getByTestId("run-workflow-cancel-btn").click();
  expect(await page.getByTestId("run-workflow-dialog")).not.toBeVisible();

  // open modal via header button
  await page.getByTestId("workflow-header-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog from the header button"
  ).toBeVisible();
  await page.getByTestId("run-workflow-cancel-btn").click();
  expect(await page.getByTestId("run-workflow-dialog")).not.toBeVisible();
});

test("it is possible to run the workflow by setting an input JSON via tha editor", async ({
  page,
}) => {
  await page.goto(`${namespace}/explorer/workflow/active/${workflow}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "the submit button is enabled by default"
  ).toEqual(true);

  expect(
    await page
      .getByTestId("run-workflow-json-tab-btn")
      .getAttribute("aria-selected"),
    "the json tab is selected by default"
  ).toBe("true");

  expect(
    await page
      .getByTestId("run-workflow-form-tab-btn")
      .getAttribute("aria-selected"),
    "the form tab is not selected"
  ).toBe("false");

  await page.type("textarea", "some invalid json");

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "submit button is disaled when the json is invalid"
  ).toEqual(false);

  await page.getByTestId("run-workflow-editor").click();
  await page.keyboard.press("Control+A");
  await page.keyboard.press("Backspace");
  await page.keyboard.type(`{"cool": true}`);

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "submit is enabled when the json is valid"
  ).toEqual(true);

  // run-workflow-editor

  // run-workflow-dialog
  await page.getByTestId("run-workflow-submit-btn").click();

  // url should be
  await expect(page).toHaveURL(new RegExp(`${namespace}/instances/`));
});
