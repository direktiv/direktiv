import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  delayWorkflow1s,
  selectStateThroughInputWorkflow,
  simpleWorkflow,
} from "e2e/utils/workflows";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test.skip("the diagram panel on the instance page responds to user interaction", async ({
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

test.skip("the diagram on the instance page changes appearance dynamically", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");
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

test("the diagram shows the visited path as green and animated", async ({
  page,
  browserName,
}) => {
  const workflowName = faker.system.commonFileName("wf.ts");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: selectStateThroughInputWorkflow,
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
  const userInputString = `{"data":"B"}`;
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

  const logsScrollContainer = page.getByTestId(
    "instance-logs-scroll-container"
  );

  await expect(logsScrollContainer, "stateA was visited").toContainText(
    "transitioning to 'stateA'"
  );

  await expect(
    page.getByTestId("rf__edge-startNode-stateA"),
    "edge to stateA is animated"
  ).toHaveClass(/(^|\s)animated(\s|$)/);

  await expect(logsScrollContainer, "stateB was visited").toContainText(
    "transitioning to 'stateB'"
  );

  await expect(
    page.getByTestId("rf__edge-stateA-stateB"),
    "edge to stateB is animated"
  ).toHaveClass(/(^|\s)animated(\s|$)/);

  await expect(logsScrollContainer, "stateC was NOT visited").not.toContainText(
    "transitioning to 'stateC'"
  );

  await expect(
    page.getByTestId("rf__edge-stateA-stateC"),
    "the edge to stateC is NOT animated because it was not visited"
  ).not.toHaveClass(/(^|\s)animated(\s|$)/);

  // testing the other path by reloading the page and submitting the workflow with different input

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
  const otherUserInputString = `{"data":"C"}`;
  await page.keyboard.type(otherUserInputString);

  // submit to run the workflow
  await page.getByTestId("run-workflow-submit-btn").click();

  await expect(logsScrollContainer, "stateA was visited").toContainText(
    "transitioning to 'stateA'"
  );

  await expect(
    page.getByTestId("rf__edge-startNode-stateA"),
    "edge to stateA is animated"
  ).toHaveClass(/(^|\s)animated(\s|$)/);

  await expect(logsScrollContainer, "stateC was visited").toContainText(
    "transitioning to 'stateC'"
  );

  await expect(
    page.getByTestId("rf__edge-stateA-stateC"),
    "edge to stateC is animated"
  ).toHaveClass(/(^|\s)animated(\s|$)/);

  await expect(logsScrollContainer, "stateB was NOT visited").not.toContainText(
    "transitioning to 'stateB'"
  );

  await expect(
    page.getByTestId("rf__edge-stateA-stateB"),
    "the edge to stateB is NOT animated because it was not visited"
  ).not.toHaveClass(/(^|\s)animated(\s|$)/);
});
