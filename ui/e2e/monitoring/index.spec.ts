import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createInstance } from "e2e/instances/utils";
import { createWorkflow } from "e2e/utils/workflow";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("It will show the logs on the monitoring page", async ({ page }) => {
  const workflowName = "workflow.yaml";
  await createWorkflow(namespace, workflowName);
  await createInstance({ namespace, path: workflowName });

  await page.goto(`/n/${namespace}/monitoring`, {
    waitUntil: "domcontentloaded",
  });

  await expect(
    page.getByText("msg: Workflow has been triggered"),
    "It will show a log message"
  ).toBeVisible();

  await expect(
    page.getByText("received 2 log entries"),
    "It will show the number of logs"
  ).toBeVisible();

  /**
   * move to the jq Playground and come back to the monitoring page.
   * This will invalidate the cache and cause reloading of the logs
   */
  await page.getByRole("link", { name: "jq Playground" }).click();
  await page.getByRole("link", { name: "Monitoring" }).click();

  // give the cache some time to update
  await page.waitForTimeout(3000);

  await expect(
    page.getByText("msg: Workflow has been triggered"),
    "When coming back to the monitoring page, it still shows the same log message"
  ).toBeVisible();

  await expect(
    page.getByText("received 1 log entry"),
    "When coming back to the monitoring page, it still shows the same number of logs"
  ).toBeVisible();
});

test("it renders an error when the api response returns an error", async ({
  page,
}) => {
  await page.route(`/api/v2/namespaces/${namespace}/logs`, async (route) => {
    if (route.request().method() === "GET") {
      const json = {
        error: { code: 422, message: "oh no!" },
      };
      await route.fulfill({ status: 422, json });
    } else route.continue();
  });

  await page.goto(`/n/${namespace}/monitoring`);

  await expect(
    page.getByText("The API returned an unexpected error: oh no!")
  ).toBeVisible();
});
