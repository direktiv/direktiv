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

test("the diagram panel on the instance page responds to user interaction", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).instance;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const diagramPanel = page.getByTestId("rf__wrapper");
  await expect(diagramPanel, "It renders the diagram panel").toBeVisible();

  const startNode = diagramPanel.getByTestId("rf__node-startNode");
  const endNode = diagramPanel.getByTestId("rf__node-endNode");

  await expect(startNode, "It renders the start node").toBeVisible();
  await expect(endNode, "It renders the end node").toBeVisible();

  const resizeButton = page.getByTestId("resizeDiagram");
  await expect(resizeButton, "It renders the maximize button").toBeVisible();

  await resizeButton.hover();
  await expect(
    page.getByText("maximize diagram"),
    "it shows the text 'maximize diagram' when hovering over the resize button"
  ).toBeVisible();

  const minimizedWidth = (await diagramPanel.boundingBox())?.width;
  await resizeButton.click();
  const maximizedWidth = (await diagramPanel.boundingBox())?.width;
  if (minimizedWidth === undefined || maximizedWidth === undefined) {
    throw new Error("could not get width of diagram panel");
  }
  expect(
    maximizedWidth / minimizedWidth,
    "The panel is significantly bigger after maximizing"
  ).toBeGreaterThan(1.5);

  await resizeButton.hover();
  await expect(
    page.getByText("minimize diagram"),
    "it shows the text 'minimize diagram' when hovering over the resize button"
  ).toBeVisible();

  await page.reload();

  const currentWidthAfterReload = (await diagramPanel.boundingBox())?.width;
  expect(
    currentWidthAfterReload,
    "after reloading the page, the panel is still maximized"
  ).toEqual(maximizedWidth);

  await resizeButton.click();

  const currentWidthAfterMinimize = (await diagramPanel.boundingBox())?.width;
  expect(currentWidthAfterMinimize, "the panel can be minimized again").toEqual(
    minimizedWidth
  );
});

test("the diagram on the instance page changes appearance dynamically", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: delayedWorkflowName,
    })
  ).instance;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const diagramPanel = page.getByTestId("rf__wrapper");
  await expect(diagramPanel, "It renders the diagram panel").toBeVisible();

  // resize screen to see the nodes better
  await page.getByTestId("resizeDiagram").click();

  const startNode = diagramPanel.getByTestId("rf__node-startNode");
  const endNode = diagramPanel.getByTestId("rf__node-endNode");
  const actionNode = page.getByText("delay").first();
  const startLine = page
    .getByTestId("rf__edge-startNode-delay")
    .locator("path")
    .first();
  const endLine = page
    .getByTestId("rf__edge-delay-endNode")
    .locator("path")
    .first();

  await expect(startNode, "It renders the start node").toBeVisible();
  await expect(endNode, "It renders the end node").toBeVisible();

  const header = page.getByTestId("instance-header-container");
  await expect(
    header.locator("div").first(),
    "the badge pending is visible"
  ).toContainText("pending");

  expect(
    startNode.locator("div").first(),
    "the start node is grey"
  ).toHaveClass(/ring-gray-8/);

  expect(endNode.locator("div").first(), "the end node is grey").toHaveClass(
    /ring-gray-8/
  );

  await expect(actionNode, "the action node is green").toHaveClass(
    /text-success-9/
  );

  await expect(startLine, "the start line is green").toHaveCSS(
    "stroke",
    "rgb(48, 164, 108)"
  );

  await expect(startLine, "the start line is dashed").toHaveCSS(
    "stroke-dasharray",
    "5px"
  );

  await expect(endLine, "the end line is NOT green").not.toHaveCSS(
    "stroke",
    "rgb(48, 164, 108)"
  );

  await expect(endLine, "the end line is NOT dashed").not.toHaveCSS(
    "stroke-dasharray",
    "5px"
  );

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  expect(
    startNode.locator("div").first(),
    "the start node is green"
  ).toHaveClass(/ring-success-9/);

  expect(endNode.locator("div").first(), "the end node is green").toHaveClass(
    /ring-success-9/
  );

  await expect(endLine, "the end line is green").toHaveCSS(
    "stroke",
    "rgb(48, 164, 108)"
  );

  await expect(endLine, "the end line is dashed").toHaveCSS(
    "stroke-dasharray",
    "5px"
  );
});
