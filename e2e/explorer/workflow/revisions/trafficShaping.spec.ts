import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";

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
 * - select two tags
 *  - slider is enabled
 *  - button is enabled
 *  - save and reload
 *  - note shows details about the config
 * - selecting the same tag twice, will still show the defailt note and keeps the button disabled
 */

test("by default, traffic shaping is not enabled", async ({ page }) => {
  const name = "workflow.yaml";

  await createWorkflow({
    payload: basicWorkflow,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name,
    },
  });

  /**
   * TODO: to be discussed. Please remove this in the PR review.
   * In some places I use texts from the translation files and I wonder if we can keep it that way.
   * because generally I would prefer to use a test id, but I have some cases where this is not possible
   * or would introduce dom notes just for testing, which I don't like either. The downside is, that
   * the tests break when when we change some translations but it's also a very easy fix.
   *
   * examples where I use text:
   * "Select Revision..." this is the eaysiest way to find out if the dropdown is unused. There is no data attribute or something else
   * "Please select 2 different revisions to configure traffic shaping." - I would prefer this instead of introducing a new dom element
   */

  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);
  await expect(
    page.getByTestId("traffic-shaping-container"),
    "it renders the traffic shaping component"
  ).toBeVisible();

  await expect(
    page.getByTestId("route-a-selector"),
    "route a selector is empty"
  ).toHaveText("Select Revision...");

  await expect(
    page.getByTestId("route-b-selector"),
    "route b selector is empty"
  ).toHaveText("Select Revision...");

  await expect(
    page.getByTestId("traffic-shaping-save-btn"),
    "button is disabled"
  ).toBeDisabled();

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "route b selector is empty"
  ).toHaveText(
    "Please select 2 different revisions to configure traffic shaping."
  );
});
