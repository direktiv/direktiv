import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { headers } from "e2e/utils/testutils";
import { workflowWithSecrets } from "./workflow";

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
  await page.goto(`/${namespace}/explorer/tree`);

  const notificaionBell = page.getByTestId("notification-bell").nth(1);

  await expect(
    notificaionBell,
    "it renders the Notification Bell"
  ).toBeVisible();

  await notificaionBell.click();
  const notificationText = page.getByTestId("notification-text");

  expect(
    await notificationText.textContent(),
    "the modal should now display 'You do not have any notifications.'"
  ).toMatch(/You do not have any notifications./);

  // create new Namespace

  // Notification Bell is visible

  // showIndicator = false

  //
  // text = {NoIssues}
});

test("Notification Bell shows an active indicator", async ({ page }) => {
  await createWorkflow({
    payload: workflowWithSecrets,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "worfklow-with-secrets.yaml",
    },
    headers,
  });

  // visit page
  await page.goto(`/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });
});

// setup with notification

/*

  const notificationMenuBtn = page.locator(
    ".mt-4 > .self-end > div > .inline-flex"
  );

*/
