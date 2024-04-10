import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import {
  simpleWorkflow as simpleWorkflowContent,
  workflowWithDelay as workflowWithDelayContent,
} from "../utils/workflows";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");
const workflowWithDelayName = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  /* create workflows we can use to create instances later */
  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });

  await createFile({
    name: workflowWithDelayName,
    namespace,
    type: "workflow",
    yaml: workflowWithDelayContent,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the header of the instance page shows the relevant data for the workflow", async ({
  page,
}) => {
  const newInstance = createInstance({ namespace, path: simpleWorkflowName });
  await expect(newInstance, "wait until process was completed").toBeDefined();
  const instanceId = (await newInstance).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const header = page.getByTestId("instance-header-container");
  await expect(header, "It renders the header").toBeVisible();

  const instanceIdShort = instanceId.slice(0, 8);
  await expect(
    header.locator("h3"),
    "It renders the instance ID in the header"
  ).toHaveText(instanceIdShort);

  await expect(
    page.locator("ul").locator("a").nth(2),
    "It renders the instance ID in the breadcrumb navigation"
  ).toHaveText(instanceIdShort);

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    header.getByText("triggerapi"),
    "It renders the instance trigger"
  ).toBeVisible();

  // check visibility of the time categories but not the exact time stamp, because it is too divergent
  await expect(
    header.getByText("started at"),
    "It renders the category 'started at'"
  ).toBeVisible();
  await expect(
    header.getByText("last updated"),
    "It renders the category 'last updated'"
  ).toBeVisible();

  await expect(
    header.getByText("spawned0 instances"),
    "category spawned shows 0 instances"
  ).toBeVisible();

  await expect(
    header.getByRole("button").locator("svg.lucide-xcircle"),
    "the button for cancelling the workflow is disabled"
  ).toBeDisabled();

  await header.getByRole("link", { name: "Open workflow" }).click();
  const editURL = `${namespace}/explorer/workflow/edit/${simpleWorkflowName}`;
  await expect(
    page,
    "the button 'Open Workflow' is clickable and links to the correct URL"
  ).toHaveURL(editURL);
});

test("the diagram panel on the instance page responds to user interaction", async ({
  page,
}) => {
  const newInstance = createInstance({ namespace, path: simpleWorkflowName });
  await expect(newInstance, "wait until process was completed").toBeDefined();
  const instanceId = (await newInstance).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

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
    "it shows the correct text when hovering over the resize button"
  ).toBeVisible();
  // hover over something else to make the overlay disappear
  page.getByTestId("rf__wrapper").hover();

  await resizeButton.click();
  await expect(
    diagramPanel,
    "It maximizes the diagram to full width of the screen"
  ).toHaveCSS("width", "1032px");
  await resizeButton.hover();

  await expect(
    page.getByText("minimize diagram"),
    "it shows the correct text when hovering over the resize button"
  ).toBeVisible();
  // hover over something else to make the overlay disappear
  page.getByTestId("rf__wrapper").hover();

  await page.reload();
  await expect(
    diagramPanel,
    "After a page reload the setting for maximizing the width is remembered"
  ).toHaveCSS("width", "1032px");

  await resizeButton.click();
  await expect(
    diagramPanel,
    "It minimizes the diagram to the initial width of the screen"
  ).toHaveCSS("width", "506px");
});

test("the diagram on the instance page changes appearance dynamically", async ({
  page,
}) => {
  const newInstance = createInstance({
    namespace,
    path: workflowWithDelayName,
  });
  await expect(newInstance, "wait until process was completed").toBeDefined();
  const instanceId = (await newInstance).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const diagramPanel = page.getByTestId("rf__wrapper");
  await expect(diagramPanel, "It renders the diagram panel").toBeVisible();

  const resizeButton = page.getByTestId("resizeDiagram");
  await expect(resizeButton, "It renders the maximize button").toBeVisible();
  // resize screen to see the nodes better
  await resizeButton.click();

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
