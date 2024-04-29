import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  workflowWithDelay as delayedWorkflowContent,
  workflowWithDelayBeforeLogging as loggingWorkflowContent,
  workflowWithManyLogs as scrollableWorkflowContent,
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
const loggingWorkflowName = faker.system.commonFileName("yaml");
const scrollableWorkflowName = faker.system.commonFileName("yaml");

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
    name: loggingWorkflowName,
    namespace,
    type: "workflow",
    yaml: loggingWorkflowContent,
  });

  await createFile({
    name: scrollableWorkflowName,
    namespace,
    type: "workflow",
    yaml: scrollableWorkflowContent,
  });

  await mockClipboardAPI(page);
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the header of the instance page shows the relevant data for the workflow", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).data.id;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

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
    header.getByText("finished at"),
    "It renders the category 'finished at'"
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
  const editURL = `/n/${namespace}/explorer/workflow/edit/${simpleWorkflowName}`;
  await expect(
    page,
    "the button 'Open Workflow' is clickable and links to the correct URL"
  ).toHaveURL(editURL);
});

test("the diagram panel on the instance page responds to user interaction", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
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
    "it shows the correct text when hovering over the resize button"
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
    "it shows the correct text when hovering over the resize button"
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

test("the input/output panel responds to user interaction", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
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
  const expectedInput = `{}`;
  const expectedOutput = `{    "result": "Hello world!"}`;
  const expectedOutputCopy = '{"result":"Hello world!"}';

  await resizeButton.hover();
  await expect(
    page.getByText("maximize output"),
    "it shows the correct text when hovering over the resize button"
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
    "it shows the correct text when hovering over the resize button"
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
  ).data.id;
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

test("The Logs panel displays the list of logs as expected", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: loggingWorkflowName,
    })
  ).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  // I did not find any other way to select this div, it does not take a testId because it is rendered later...
  const logsPanelParent = page.locator("div").filter({ has: logsPanel }).nth(6);

  await expect(logsPanel).toBeVisible();

  await expect(
    logsPanel.locator("h3"),
    "The headline of the logs is correct"
  ).toContainText(`Logs for /${loggingWorkflowName}`);

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("pending");

  const entriesCounter = logsPanel.getByTestId("instance-logs-entries-counter");

  await expect(
    entriesCounter,
    "While running the workflow there are 0 log entries"
  ).toContainText("received 0 log entries");

  await expect(
    entriesCounter.locator("span").nth(1),
    "There is a loading spinner"
  ).toHaveClass(/animate-ping/);

  const resizeButton = page
    .getByTestId("instance-logs-container")
    .getByRole("button")
    .nth(2);

  resizeButton.hover();
  await expect(
    page.getByText("maximize logs"),
    "It shows the correct text when hovering over the resize button"
  ).toBeVisible();

  const minimizedHeight = (await logsPanelParent.boundingBox())?.height;

  await resizeButton.click();

  const maximizedHeight = (await logsPanelParent.boundingBox())?.height;
  if (minimizedHeight === undefined || maximizedHeight === undefined) {
    throw new Error("could not get height of logs panel");
  }

  expect(
    maximizedHeight / minimizedHeight,
    "The panel is significantly bigger after maximizing"
  ).toBeGreaterThan(1.5);

  page.reload();

  const currentHeightAfterReload = (await logsPanelParent.boundingBox())
    ?.height;
  expect(
    currentHeightAfterReload,
    "After reloading the page, the panel is still maximized"
  ).toEqual(maximizedHeight);

  await resizeButton.hover();
  await expect(
    page.getByText("minimize logs"),
    "It shows the correct text when hovering over the resize button"
  ).toBeVisible();

  await expect(
    logsPanel.locator("pre").locator("span").nth(0),
    "It displays an initial log entry"
  ).toContainText("Running state logic");

  await expect(
    logsPanel.locator("pre").locator("span").nth(15),
    "It displays the log message from the log field in the workflow yaml"
  ).toContainText("log-message");

  await expect(
    logsPanel.locator("pre").locator("span").last(),
    "It displays a final log entry"
  ).toContainText("Workflow completed");

  await expect(
    entriesCounter,
    "When the workflow finished running there are 22 log entries"
  ).toContainText("received 22 log entries");
});

test("The Logs panel can be toggled between verbose and non verbose logs", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  await expect(
    logsPanel.locator("pre").last().locator("span").first(),
    "It does NOT display the state for the final log entry"
  ).not.toContainText("state: helloworld");

  await expect(
    logsPanel.locator("pre").last().locator("span").last(),
    "It displays the final log entry"
  ).toContainText("msg: Workflow completed");

  const verboseButton = page
    .getByTestId("instance-logs-container")
    .getByRole("button")
    .nth(0);

  await verboseButton.click();

  await expect(verboseButton, "the verbose button is active").toHaveAttribute(
    "data-state",
    "on"
  );

  await expect(
    logsPanel.locator("pre").last().locator("span").first(),
    "It displays the state for the final log entry"
  ).toContainText("state: helloworld");

  await expect(
    logsPanel.locator("pre").last().locator("span").last(),
    "It displays the final log entry"
  ).toContainText("msg: Workflow completed.");

  page.reload();

  await expect(
    verboseButton,
    "After reloading the page the verbose setting is remembered"
  ).toHaveAttribute("data-state", "on");

  await expect(
    logsPanel.locator("pre").last().locator("span").first(),
    "After reloading the page the verbose setting is remembered"
  ).toContainText("state: helloworld");
});

test("The Logs can be copied", async ({ page }) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  const entriesCounter = logsPanel.getByTestId("instance-logs-entries-counter");

  await expect(entriesCounter, "Waiting for log entries").not.toContainText(
    "received 0 log entries"
  );

  const copyButton = page
    .getByTestId("instance-logs-container")
    .getByRole("button")
    .nth(1);

  await copyButton.click();

  const copiedLogs = "yaml - helloworld - Running state logic";

  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain(
    copiedLogs
  );
});

test("When having many incoming logs, the scrollbar is following the newest logs", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: scrollableWorkflowName,
    })
  ).instance;
  await page.goto(`/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("pending");

  const header = page.getByTestId("instance-header-container");

  await expect(
    header.getByText("spawned0 instances"),
    "category spawned shows 0 instances"
  ).toBeVisible();

  const entriesCounter = page.getByTestId("instance-logs-entries-counter");

  await expect(entriesCounter, "Waiting for any log entries").not.toContainText(
    "received 0 log entries"
  );

  // Wait until 20 or more logs are visible
  await page.locator(':nth-match(:text("msg"), 20)').waitFor();

  await expect(
    entriesCounter,
    "Waiting for more than 20 log entries"
  ).toContainText("received /([2-9][0-9] log entries");

  // scrollbar = overflow container

  const scrollbar = page.getByTestId("instance-logs-scroll-container");

  // const scrollbar = page.getByRole("scrollbar");

  await expect(scrollbar, "is there").toBeDefined();

  // await expect(scrollbar, "is down").toHaveCSS("top", "100%");

  await expect(
    logsPanel.locator("pre").last().locator("span").last(),
    "The final log entry is in the view, so the page is scrolled down"
  ).toBeInViewport;

  page.mouse.wheel(0, -100000000);

  page.mouse.wheel(0, 100000000);

  // expect(await page.evaluate(() => navigator.clipboard.readText())).toContain(
  //   copiedLogs
  // );

  // scroll panel is at 0% height in the scrollbar

  // scroll up

  // see button "follow logs"

  // click

  // scroll panel is at 0% height in the scrollbar
});
