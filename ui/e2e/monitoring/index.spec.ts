import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("It will show the logs on the monitoring page", async ({ page }) => {
  await page.goto(`/n/${namespace}/monitoring`);

  await expect(
    page.getByText("msg: updating namespace gateway"),
    "It will show a log message"
  ).toBeVisible();

  await expect(
    page.getByText("received 1 log entry"),
    "It will show the number of logs"
  ).toBeVisible();

  /**
   * move to the jq Playground and come back to the monitoring page.
   * This will invalidate the cache and cause reloading of the logs
   */
  await page.getByRole("link", { name: "jq Playground" }).click();
  await page.getByRole("link", { name: "Monitoring" }).click();

  // give the cache some time to update
  await new Promise((resolve) => setTimeout(resolve, 3000));

  await expect(
    page.getByText("msg: updating namespace gateway"),
    "When coming back to the monitoring page, it still shows the same log message"
  ).toBeVisible();

  await expect(
    page.getByText("received 1 log entry"),
    "When coming back to the monitoring page, it still shows the same number of logs"
  ).toBeVisible();
});
