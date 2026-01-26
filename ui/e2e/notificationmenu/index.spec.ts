import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { workflowWithSecrets } from "./utils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("Notification Bell has an inactive state by default", async ({ page }) => {
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });

  const notificationBell = page.getByTestId("notification-bell").nth(1);

  await expect(
    notificationBell,
    "it renders the Notification Bell"
  ).toBeVisible();

  await notificationBell.click();
  const notificationText = page.getByTestId("notification-text");

  expect(
    await notificationText.textContent(),
    "the modal should display 'You do not have any notifications.'"
  ).toMatch(/You do not have any notifications./);
});

test("Notification Bell shows dot for uninitialized secrets", async ({
  page,
}) => {
  await createFile({
    name: "worfklow-with-secrets.yaml",
    namespace,
    type: "workflow",
    content: workflowWithSecrets,
    mimeType: "application/x-typescript",
  });

  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("notification-bell").nth(1),
    "it renders the Notification Bell"
  ).toBeVisible();

  // TODO in TDI-219: Implement polling, remove manual reload
  await page.waitForTimeout(2000);
  await page.reload({ waitUntil: "networkidle" });

  expect(
    page.getByTestId("notification-indicator").nth(1),
    "the indicator for new messages is visible"
  ).toBeVisible();

  await page.getByTestId("notification-bell").nth(1).click();

  expect(
    await page.getByTestId("notification-text").textContent(),
    "the modal should now display 'You have 2 uninitialized secrets.'"
  ).toMatch(/You have 2 uninitialized secrets./);

  await page.goto(`/n/${namespace}/settings`);

  await page
    .getByRole("cell", { name: "one Initialize secret" })
    .getByRole("button")
    .click();

  await page.locator("textarea").fill("abc");
  await page.getByRole("button", { name: "Save" }).click();

  // TODO in TDI-219: Implement polling, remove manual reload
  await page.waitForTimeout(2000);
  await page.reload({ waitUntil: "networkidle" });

  await page.getByTestId("notification-bell").nth(1).click();
  expect(
    page.getByTestId("notification-indicator").nth(1),
    "the indicator for new messages is visible"
  ).toBeVisible();
  expect(
    await page.getByTestId("notification-text").textContent(),
    "the modal should now display 'You have 1 uninitialized secret.'"
  ).toMatch("You have 1 uninitialized secret.");

  await page
    .getByRole("cell", { name: "two Initialize secret" })
    .getByRole("button")
    .click();

  await page.locator("textarea").fill("123");
  await page.getByRole("button", { name: "Save" }).click();

  await expect(
    page.getByTestId("notification-indicator"),
    "the indicator for new messages is NOT visible"
  ).toHaveCount(0);

  // TODO in TDI-219: Implement polling, remove manual reload
  await page.waitForTimeout(2000);
  await page.reload({ waitUntil: "networkidle" });

  await page.getByTestId("notification-bell").nth(1).click();

  expect(
    await page.getByTestId("notification-text").textContent(),
    "the modal should now display 'You do not have any notifications.'"
  ).toMatch(/You do not have any notifications./);
});
