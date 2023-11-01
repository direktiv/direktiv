import { createNamespace, deleteNamespace } from "./utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

/*

Test2:

Create a secret without initializing. Check if "showIndicator=true" and NotificationMessage is mounted, see if 

Test3:

Initialize the secret. Check if "showIndicator=false"

*/

test("Notification Bell has an inactive state by default", async ({ page }) => {
  // visit page
  await page.goto("/");

  const bodyTag = page.locator("body");

  const userMenuBtn = page.getByTestId("dropdown-trg-user-menu");

  await expect(
    // page.getByTestId("notification-bell"),
    page.getByTestId("notification-bell"),
    "it renders the Notification Bell"
  ).toBeVisible();

  const notificationMenuBtn = page.getByTestId("notification-bell");
  await notificationMenuBtn.nth(1).click();

  const notificationText = page.getByTestId("notification-text");

  expect(
    await notificationText.textContent(),
    "the modal should now display 'You do not have any notifications.'"
  ).toMatch(/You do not have any notifications./);

  // create new Namespace

  // Notification Bell is visible

  // showIndicator = false

  // text = {NoIssues}
});

/*

  const notificationMenuBtn = page.locator(
    ".mt-4 > .self-end > div > .inline-flex"
  );

*/
