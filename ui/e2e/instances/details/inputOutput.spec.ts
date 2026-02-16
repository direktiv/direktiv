import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { delayWorkflow1s, simpleWorkflow } from "e2e/utils/workflows";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";
import { mockClipboardAPI } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async ({ page }) => {
  namespace = await createNamespace();
  await mockClipboardAPI(page);
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the input/output panel responds to user interaction", async ({
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

  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
    })
  ).data.id;
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
  const expectedInput = "null";
  const expectedOutput = `{    "message": "Hello world!"}`;
  const expectedOutputCopy = '{"message":"Hello world!"}';

  await resizeButton.hover();
  await expect(
    page.getByText("maximize output"),
    "it shows the text 'maximize output' when hovering over the resize button"
  ).toBeVisible();

  const minimizedHeight = (await inputOutputPanel.boundingBox())?.height;

  await resizeButton.click();

  const maximizedHeight = (await inputOutputPanel.boundingBox())?.height;
  if (minimizedHeight === undefined || maximizedHeight === undefined) {
    throw new Error("could not get width of input/output panel");
  }
  expect(
    maximizedHeight / minimizedHeight,
    "The panel is significantly bigger after maximizing"
  ).toBeGreaterThan(1.5);

  await resizeButton.hover();
  await expect(
    page.getByText("minimize output"),
    "it shows the text 'minimize output' when hovering over the resize button"
  ).toBeVisible();

  await page.reload();

  const currentHeightAfterReload = (await inputOutputPanel.boundingBox())
    ?.height;
  expect(
    currentHeightAfterReload,
    "after reloading the page, the panel is still maximized"
  ).toEqual(maximizedHeight);

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
  const workflowName = faker.system.commonFileName("wf.ts");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: delayWorkflow1s,
    mimeType: "application/x-typescript",
  });
  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
    })
  ).data.id;
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
  const header = page.getByTestId("instance-header-container");
  const textarea = inputOutputPanel.locator(".view-lines");

  await expect(
    inputButton,
    "the input tab was selected initially"
  ).toHaveAttribute("data-state", "active");

  await outputButton.click();

  await expect(
    inputOutputPanel,
    "The output shows a note that the workflow is still running"
  ).toContainText("The workflow is still running");

  // TODO in TDI-219: remove manual reloads after streaming updates have been restored
  await page.waitForTimeout(1000);
  await page.reload();

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    textarea,
    "When the workflow finished the generated output is shown in the panel"
  ).toHaveText(`{    "message": "Hello world!"}`);
});

test("after a running instance finishes, the output tab is automatically selected", async ({
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
  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
    })
  ).data.id;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const inputOutputPanel = page.getByTestId("inputOutputPanel");

  await expect(inputOutputPanel).toBeVisible();

  const outputButton = inputOutputPanel
    .getByRole("tablist")
    .locator("button")
    .nth(1);

  const textarea = inputOutputPanel.locator(".view-lines");
  const header = page.getByTestId("instance-header-container");

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    outputButton,
    "the output tab was selected automatically"
  ).toHaveAttribute("data-state", "active");

  await expect(textarea, "the text shows the expected output").toHaveText(
    `{    "message": "Hello world!"}`
  );
});
