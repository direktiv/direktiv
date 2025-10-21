import {} from "~/util/helpers";

import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { faker } from "@faker-js/faker";
import { getInstanceInput } from "~/api/instances/query/input";
import { headers } from "e2e/utils/testutils";
import { simpleWorkflow } from "e2e/instances/utils/workflows";
import { testDiacriticsWorkflow } from "./utils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to open and close the run workflow modal from the editor and the header of the workflow page", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("wf.ts");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: simpleWorkflow,
    mimeType: "application/x-typescript",
  });

  await page.goto(`/n/${namespace}/explorer/workflow/edit/${workflowName}`);

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

test("it is possible to run the workflow with an input JSON via the editor", async ({
  page,
  browserName,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: simpleWorkflow,
    mimeType: "application/x-typescript",
  });

  await page.goto(`/n/${namespace}/explorer/workflow/edit/${workflowName}`);

  await page.getByTestId("workflow-editor-btn-run").click();
  expect(
    await page.getByTestId("run-workflow-dialog"),
    "it opens the dialog"
  ).toBeVisible();

  expect(
    await page.getByTestId("run-workflow-submit-btn").isEnabled(),
    "the submit button is enabled by default"
  ).toEqual(true);

  await page.getByTestId("run-workflow-editor").click();
  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  const userInputString = `{"string":"1","integer":1,"boolean":true,"array":[1,2,3],"object":{"key":"value"}}`;
  await page.keyboard.type(userInputString);

  // submit to run the workflow
  await page.getByTestId("run-workflow-submit-btn").click();

  const reg = new RegExp(`/n/${namespace}/instances/(.*)`);
  await expect(
    page,
    "workflow was triggered with our input and user was redirected to the instances page"
  ).toHaveURL(reg);
  const instanceId = page.url().match(reg)?.[1];

  if (!instanceId) {
    throw new Error("instanceId not found");
  }

  // check the server state of the input
  const res = await getInstanceInput({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      instanceId,
      namespace,
    },
    headers,
  });

  const inputResponseString = res.data.input ?? "";

  expect(
    inputResponseString,
    "the server result is the same as the input that was sent"
  ).toBe(userInputString);
});

test("it is possible to run a workflow with input data containing special characters", async ({
  page,
  browserName,
}) => {
  const name = "test-diacritics.yaml";

  await createFile({
    name,
    namespace,
    type: "workflow",
    content: testDiacriticsWorkflow,
    mimeType: "application/x-typescript",
  });

  await page.goto(`/n/${namespace}/explorer/workflow/edit/${name}`);

  await expect(
    page.locator(".view-lines"),
    "The editor renders special characters correctly"
  ).toContainText("A workflow for testing characters like îèüñÆ");

  await page.getByTestId("workflow-editor-btn-run").click();

  await page.keyboard.press(browserName === "webkit" ? "Meta+A" : "Control+A");
  await page.keyboard.press("Backspace");
  const userInputString = `{"name":"Kateřina Horáčková"}`;
  await page.keyboard.type(userInputString);

  await page.getByTestId("run-workflow-submit-btn").click();

  await expect(
    page.getByTestId("inputOutputPanel").locator(".view-lines"),
    "The text from the input is rendered correctly in the workflow output"
  ).toContainText(`"result": "Hello Kateřina Horáčková"`, {
    useInnerText: true,
  });
});

test("it is not possible to run the workflow when the editor has unsaved changes", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: simpleWorkflow,
    mimeType: "application/x-typescript",
  });

  await page.goto(`/n/${namespace}/explorer/workflow/edit/${workflowName}`);

  await expect(page.getByTestId("workflow-header-btn-run")).not.toBeDisabled();
  await expect(page.getByTestId("workflow-editor-btn-run")).not.toBeDisabled();

  await page.type("textarea", faker.string.alphanumeric(1));

  await expect(page.getByTestId("workflow-header-btn-run")).toBeDisabled();
  await expect(page.getByTestId("workflow-editor-btn-run")).toBeDisabled();

  await page.locator("textarea").press("Backspace");

  await expect(
    page.getByTestId("workflow-header-btn-run"),
    "when the text input is equal to the saved data the run button is active"
  ).not.toBeDisabled();
  await expect(
    page.getByTestId("workflow-editor-btn-run"),
    "when the text input is equal to the saved data the run button is active"
  ).not.toBeDisabled();
});
