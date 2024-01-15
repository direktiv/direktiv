import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "e2e/utils/testutils";
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
  await page.goto(`/${namespace}/explorer/tree`, {
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

test("Notification Bell updates depending on the count of Notification Messages", async ({
  page,
}) => {
  await createWorkflow({
    payload: workflowWithSecrets,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "worfklow-with-secrets.yaml",
    },
    headers,
  });

  await page.goto(`/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });

  const notificationBell = page.getByTestId("notification-bell").nth(1);

  await expect(
    notificationBell,
    "it renders the Notification Bell"
  ).toBeVisible();

  expect(
    page.getByTestId("notification-indicator").nth(1),
    "the indicator for new messages is visible"
  ).toBeVisible();

  await notificationBell.click();
  const notificationText = page.getByTestId("notification-text");

  expect(
    await notificationText.textContent(),
    "the modal should now display 'You have 2 uninitialized secrets.'"
  ).toMatch(/You have 2 uninitialized secrets./);

  await page.goto(`/${namespace}/settings`);

  const initialize_secret1 = page
    .getByRole("cell", { name: "ACCESS_KEY Initialize secret" })
    .getByRole("button");
  const initialize_secret2 = page
    .getByRole("cell", { name: "ACCESS_SECRET Initialize secret" })
    .getByRole("button");

  await initialize_secret1.click();

  await page.getByTestId("new-secret-editor").fill("abc");
  await page.getByTestId("secret-create-submit").click();
  await notificationBell.click();

  expect(
    page.getByTestId("notification-indicator").nth(1),
    "the indicator for new messages is visible"
  ).toBeVisible();

  expect(
    await notificationText.textContent(),
    "the modal should now display 'You have 1 uninitialized secret.'"
  ).toMatch(/You have 1 uninitialized secret./);

  await initialize_secret2.click();

  await page.getByTestId("new-secret-editor").fill("123");
  await page.getByTestId("secret-create-submit").click();

  expect(
    page.getByTestId("notification-indicator").nth(1),
    "the indicator for new messages is NOT visible"
  ).not.toBeVisible();

  await notificationBell.click();

  expect(
    await notificationText.textContent(),
    "the modal should now display 'You do not have any notifications.'"
  ).toMatch(/You do not have any notifications./);
});
