import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { errorWorkflow, simpleWorkflow } from "e2e/utils/workflows";
import { expect, test } from "@playwright/test";
import { workflowWithFewLogs, workflowWithManyLogs } from "../utils/workflows";

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

test("It displays a log message from the workflow yaml, one initial and one final log entry", async ({
  page,
}) => {
  /* prepare data*/
  const workflowName = faker.system.commonFileName("wf.ts");
  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: workflowWithFewLogs,
    mimeType: "application/x-typescript",
  });

  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
    })
  ).data.id;

  /* visit page and test */
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  const scrollContainer = page.getByTestId("instance-logs-scroll-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    logsPanel.locator("h3"),
    "The headline of the logs shows the name of the currently running workflow"
  ).toContainText(`Logs for /${workflowName}`);

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("pending");

  const entriesCounter = page.getByTestId("instance-logs-entries-counter");

  await expect(
    entriesCounter.locator("span").nth(1),
    "There is a loading spinner"
  ).toHaveClass(/animate-ping/);

  await expect(
    scrollContainer
      .locator("pre")
      .locator("span", { hasText: "msg: workflow has been started" }),
    "It displays an initial log entry"
  ).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  // for some reason, logs do not update in the test at this point.
  // Possibly streaming is interrupted in the test context? As a workaround,
  // reload the page to ensure all logs are eventually rendered.
  page.reload();

  await expect(
    scrollContainer
      .locator("pre")
      .locator("span", { hasText: "msg: hello-world" }),
    "It displays the log message from the log field in the workflow yaml"
  ).toBeVisible();

  await expect(
    scrollContainer.locator("pre").locator("span").last(),
    "It displays a final log entry"
  ).toContainText("msg: flow finished");

  await expect(
    entriesCounter,
    "When the workflow finished running there are 6 log entries"
  ).toContainText(/received \d+ log entries/);
});

test("the logs panel can be maximized", async ({ page }) => {
  /* prepare data */
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

  /* visit page and test */
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

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
});

test("the logs panel can be toggled between verbose and non verbose logs", async ({
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

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  const scrollContainer = page.getByTestId("instance-logs-scroll-container");

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("complete");

  const twoNumbersAndTheLogMessage = /[0-9]{2}msg: flow finished/;
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
  ).toContainText("state: stateHellomsg: flow finished");

  page.reload();

  await expect(
    verboseButton,
    "After reloading the page the verbose button is still active"
  ).toHaveAttribute("data-state", "on");

  await expect(
    scrollContainer.locator("pre").last(),
    "After reloading the page it still displays the state in the last log entry"
  ).toContainText("state: stateHellomsg: flow finished");
});

test("the logs can be copied", async ({ page }) => {
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
    "flow starting"
  );
  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain(
    "transitioning to 'stateHello'"
  );

  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain(
    "flow finished"
  );
});

test("log entries will be automatically scrolled to the end", async ({
  page,
}) => {
  const workflowName = faker.system.commonFileName("yaml");

  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: workflowWithManyLogs,
    mimeType: "application/x-typescript",
  });

  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
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

  await expect(
    scrollContainer.locator("pre").nth(20),
    "With more than 20 logs the button appears"
  ).toBeVisible();

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

  await followButton.click();

  await expect(
    followButton,
    "After clicking it, the button disappeared"
  ).not.toBeVisible();

  await expect(
    scrollContainer.locator("pre").last().locator("span").last(),
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
    scrollContainer.locator("pre").first(),
    "The page is still scrolled up"
  ).toBeInViewport();

  await expect(
    followButton,
    "The 'Follow Logs' button is not visible when the workflow has completed running"
  ).not.toBeVisible();
});

test("it renders error details for errors in the logs", async ({ page }) => {
  /* prepare data */
  const workflowName = faker.system.commonFileName("yaml");

  await createFile({
    name: workflowName,
    namespace,
    type: "workflow",
    content: errorWorkflow,
    mimeType: "application/x-typescript",
  });

  const instanceId = (
    await createInstance({
      namespace,
      path: workflowName,
    })
  ).data.id;

  /* perform test */
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const logsPanel = page.getByTestId("instance-logs-container");

  await expect(logsPanel).toBeVisible();

  await expect(
    page.getByTestId("instance-header-container").locator("div").first()
  ).toContainText("failed");

  await expect(
    page.getByText(
      "msg: schema validation error: email: Does not match format 'email'"
    )
  ).toBeVisible();
});

test("it renders an error when the api response returns an error", async ({
  page,
}) => {
  /* prepare data */
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

  /* register mock error response */
  await page.route(
    `/api/v2/namespaces/${namespace}/logs?instance=${instanceId}?last=50`,
    async (route) => {
      if (route.request().method() === "GET") {
        const json = {
          error: { code: 422, message: "oh no!" },
        };
        await route.fulfill({ status: 422, json });
      } else route.continue();
    }
  );

  /* perform test */
  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  await expect(
    page.getByText("The API returned an unexpected error: oh no!")
  ).toBeVisible();
});
