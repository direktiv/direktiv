import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

/**
 * planed tests
 * - by default both
 *   - both dropdowns are empty,
 *   - button is disabled
 *   - note is shown
 * - select two tags
 *  - slider is enabled
 *  - button is enabled
 *  - save and reload
 *  - note shows details about the config
 * - selecting the same tag twice, will still show the defailt note and keeps the button disabled
 */

test("be default, traffic shaping is not enabled", async ({ page }) => {
  // visit page
  await page.goto("/");

  expect(false).toBe(true);
});
