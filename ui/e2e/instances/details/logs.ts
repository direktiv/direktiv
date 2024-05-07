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

test("the logs panel can be resized, it displays a log message from the workflow yaml, one initial and one final log entry", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: fewLogsWorkflowName,
    })
  ).data.id;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  const scrollContainer = page.getByTestId("instance-logs-scroll-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    logsPanel.locator("h3"),
    "The headline of the logs shows the name of the currently running workflow"
  ).toContainText(`Logs for /${fewLogsWorkflowName}`);

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("pending");

  const entriesCounter = page.getByTestId("instance-logs-entries-counter");

  await expect(
    entriesCounter.locator("span").nth(1),
    "There is a loading spinner"
  ).toHaveClass(/animate-ping/);

  await expect(
    entriesCounter,
    "While starting the workflow there are 6 log entries"
  ).toContainText("received 6 log entries");

  const resizeButton = page
    .getByTestId("instance-logs-container")
    .getByRole("button")
    .nth(2);

  resizeButton.hover();
  await expect(
    page.getByText("maximize logs"),
    "It shows the text 'maximize logs' when hovering over the resize button"
  ).toBeVisible();

  const minimizedHeight = (await logsPanel.boundingBox())?.height;

  await resizeButton.click();

  const maximizedHeight = (await logsPanel.boundingBox())?.height;
  if (minimizedHeight === undefined || maximizedHeight === undefined) {
    throw new Error("could not get height of logs panel");
  }

  expect(
    maximizedHeight / minimizedHeight,
    "The panel is significantly bigger after maximizing"
  ).toBeGreaterThan(1.5);

  page.reload();

  const currentHeightAfterReload = (await logsPanel.boundingBox())?.height;
  expect(
    currentHeightAfterReload,
    "After reloading the page, the panel is still maximized"
  ).toEqual(maximizedHeight);

  await resizeButton.hover();
  await expect(
    page.getByText("minimize logs"),
    "It shows the text 'minimize logs' when hovering over the resize button"
  ).toBeVisible();

  await expect(
    scrollContainer.locator("pre").locator("span").nth(0),
    "It displays an initial log entry"
  ).toContainText("Running state logic");

  await expect(
    scrollContainer.locator("pre").locator("span").nth(15),
    "It displays the log message from the log field in the workflow yaml"
  ).toContainText("log-message");

  await expect(
    scrollContainer.locator("pre").locator("span").last(),
    "It displays a final log entry"
  ).toContainText("Workflow completed");

  await expect(
    entriesCounter,
    "When the workflow finished running there are 22 log entries"
  ).toContainText("received 22 log entries");
});

test("the logs panel can be toggled between verbose and non verbose logs", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).data.id;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  const scrollContainer = page.getByTestId("instance-logs-scroll-container");

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  const twoNumbersAndTheLogMessage = /[0-9]{2}msg: Workflow completed\./;
  await expect(
    scrollContainer.getByText(twoNumbersAndTheLogMessage),
    "It does not display the state in the last log entry"
  ).toBeVisible();

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
    scrollContainer.locator("pre").last(),
    "It displays the state in the last log entry"
  ).toContainText("state: helloworldmsg: Workflow completed.");

  page.reload();

  await expect(
    verboseButton,
    "After reloading the page the verbose button is still active"
  ).toHaveAttribute("data-state", "on");

  await expect(
    scrollContainer.locator("pre").last(),
    "After reloading the page it still displays the state in the last log entry"
  ).toContainText("state: helloworldmsg: Workflow completed.");
});

test("the logs can be copied", async ({ page }) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).data.id;
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  const entriesCounter = page.getByTestId("instance-logs-entries-counter");

  await expect(entriesCounter, "Waiting for log entries").not.toContainText(
    "received 0 log entries"
  );

  const copyButton = page
    .getByTestId("instance-logs-container")
    .getByRole("button")
    .nth(1);

  await copyButton.click();

  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain(
    "yaml - helloworld - Running state logic"
  );
});

test("log entries will be automatically scrolled to the end", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: manyLogsWorkflowName,
    })
  ).data.id;

  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  const entriesCounter = page.getByTestId("instance-logs-entries-counter");

  await expect(entriesCounter, "Waiting for any log entries").not.toContainText(
    "received 0 log entries"
  );

  const scrollContainer = page.getByTestId("instance-logs-scroll-container");

  await expect(scrollContainer, "Container is scrollable").toBeDefined();

  await expect(
    scrollContainer.locator("pre").last().locator("span").last(),
    "The last log entry is in the view, so the page is scrolled down"
  ).toBeInViewport();

  const countLogsAfterScrolling = await scrollContainer.locator("pre").count();

  await expect(
    countLogsAfterScrolling,
    "With more than 20 logs the button appears"
  ).toBeGreaterThan(20);

  // click on first entry to scroll to the top of the list
  const currentFirstEntry = scrollContainer.locator("pre").first();
  currentFirstEntry.click();

  await expect(
    currentFirstEntry,
    "The first log entry is in the view, so the page is scrolled up"
  ).toBeInViewport();

  const followButton = page.getByRole("button", { name: "Follow logs" });

  await expect(
    followButton,
    "After scrolling up, a button appeared"
  ).toBeVisible();

  followButton.click();

  await expect(
    followButton,
    "After clicking it, the button disappeared"
  ).not.toBeVisible();

  await expect(
    scrollContainer.locator("pre").last(),
    "The last log entry is in the view, so the page was scrolled down"
  ).toBeInViewport();

  // scrolling up again
  scrollContainer.locator("pre").first().click();

  await expect(
    followButton,
    "The 'Follow Logs' button is visible"
  ).toBeVisible();

  const header = page.getByTestId("instance-header-container");

  await expect(async () => {
    expect(
      header.locator("div").first(),
      "The badge complete is visible"
    ).toContainText("complete");
  }).toPass();

  await expect(
    currentFirstEntry,
    "The page is still scrolled up"
  ).toBeInViewport();

  await expect(
    followButton,
    "The 'Follow Logs' button is not visible when the workflow has completed running"
  ).not.toBeVisible();
});
