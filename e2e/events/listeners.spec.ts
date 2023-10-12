import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to send a new event", async ({ page }) => {
  await page.goto(`/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/${namespace}/events/listeners`
  );

  await expect(
    page.getByTestId("breadcrumb-event-listeners"),
    "it renders the 'Event Listeners' breadcrumb"
  ).toBeVisible();
});
