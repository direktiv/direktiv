import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  workflowWithDelay as delayedWorkflowContent,
  workflowWithFewLogs as fewLogsWorkflowContent,
  workflowWithManyLogs as manyLogsWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
} from "../utils/workflows";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";
import { mockClipboardAPI } from "e2e/utils/testutils";

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");
const delayedWorkflowName = faker.system.commonFileName("yaml");
const fewLogsWorkflowName = faker.system.commonFileName("yaml");
const manyLogsWorkflowName = faker.system.commonFileName("yaml");

test.beforeEach(async ({ page }) => {
  namespace = await createNamespace();
  /* create workflows we can use to create instances later */
  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });

  await createFile({
    name: delayedWorkflowName,
    namespace,
    type: "workflow",
    yaml: delayedWorkflowContent,
  });

  await createFile({
    name: fewLogsWorkflowName,
    namespace,
    type: "workflow",
    yaml: fewLogsWorkflowContent,
  });

  await createFile({
    name: manyLogsWorkflowName,
    namespace,
    type: "workflow",
    yaml: manyLogsWorkflowContent,
  });

  await mockClipboardAPI(page);
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the input/output panel responds to user interaction", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).instance;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const inputOutputPanel = page.getByTestId("inputOutputPanel");

  await expect(
    inputOutputPanel,
    "It renders the input/output panel"
  ).toBeVisible();

  const copyButton = inputOutputPanel.locator("button").nth(0);
  const resizeButton = inputOutputPanel.locator("button").nth(1);

  const inputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(0);
  const outputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(1);

  const textarea = inputOutputPanel.locator(".view-lines");
  const expectedInput = `{}`;
  const expectedOutput = `{    "result": "Hello world!"}`;
  const expectedOutputCopy = '{"result":"Hello world!"}';

  await resizeButton.hover();
  await expect(
    page.getByText("maximize output"),
    "it shows the text 'maximize output' when hovering over the resize button"
  ).toBeVisible();

  const minimizedWidth = (await inputOutputPanel.boundingBox())?.width;

  await resizeButton.click();

  const maximizedWidth = (await inputOutputPanel.boundingBox())?.width;
  if (minimizedWidth === undefined || maximizedWidth === undefined) {
    throw new Error("could not get width of input/output panel");
  }
  expect(
    maximizedWidth / minimizedWidth,
    "The panel is significantly bigger after maximizing"
  ).toBeGreaterThan(1.5);

  await resizeButton.hover();
  await expect(
    page.getByText("minimize output"),
    "it shows the text 'minimize output' when hovering over the resize button"
  ).toBeVisible();

  await page.reload();

  const currentWidthAfterReload = (await inputOutputPanel.boundingBox())?.width;
  expect(
    currentWidthAfterReload,
    "after reloading the page, the panel is still maximized"
  ).toEqual(maximizedWidth);

  await resizeButton.click();
  await inputButton.click();

  await expect(textarea, "the text shows the expected input").toHaveText(
    expectedInput
  );

  await outputButton.click();

  await expect(textarea, "the text shows the expected output").toHaveText(
    expectedOutput
  );

  await copyButton.click();

  expect(await page.evaluate(() => navigator.clipboard.readText())).toEqual(
    expectedOutputCopy
  );

  await inputButton.click();
  await copyButton.click();

  expect(await page.evaluate(() => navigator.clipboard.readText())).toEqual(
    expectedInput
  );
});

test("the output is shown when the workflow finished running", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: delayedWorkflowName,
    })
  ).instance;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const inputOutputPanel = page.getByTestId("inputOutputPanel");

  await expect(inputOutputPanel).toBeVisible();

  const outputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(1);

  const header = page.getByTestId("instance-header-container");
  const textarea = inputOutputPanel.locator(".view-lines");

  const runningInstanceOutput = "The workflow is still running";
  const expectedOutput = `{    "result": "finished"}`;

  await outputButton.click();

  await expect(
    inputOutputPanel,
    "The output shows a note that the workflow is still running"
  ).toContainText(runningInstanceOutput);

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    textarea,
    "When the workflow finished the generated output is shown in the panel"
  ).toHaveText(expectedOutput);
});

test("after a running instance finishes, the output tab is automatically selected", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: delayedWorkflowName,
    })
  ).instance;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const inputOutputPanel = page.getByTestId("inputOutputPanel");

  await expect(inputOutputPanel).toBeVisible();

  const inputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(0);
  const outputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(1);

  const textarea = inputOutputPanel.locator(".view-lines");
  const expectedOutput = `{    "result": "finished"}`;
  const header = page.getByTestId("instance-header-container");

  await expect(
    inputButton,
    "the input tab was selected initially"
  ).toHaveAttribute("data-state", "active");

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    outputButton,
    "the output tab was selected automatically"
  ).toHaveAttribute("data-state", "active");

  await expect(textarea, "the text shows the expected output").toHaveText(
    expectedOutput
  );
});
