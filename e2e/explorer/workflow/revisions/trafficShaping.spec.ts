import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { createWorkflowWithThreeRevisions } from "../../../utils/revisions";
import { faker } from "@faker-js/faker";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("by default, traffic shaping is not enabled", async ({ page }) => {
  const name = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: basicWorkflow.data,
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
    "there is a hint that describes the traffic shaping"
  ).toHaveText(
    "Please select 2 different revisions to configure traffic shaping."
  );
});

test("it is possible to configure traffic shaping", async ({ page }) => {
  const name = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [firstRevision, secondRevision],
  } = await createWorkflowWithThreeRevisions(namespace, name);

  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);
  await page.getByTestId(`route-a-selector`).click();
  await page.getByTestId(firstRevision.revision.name).click();
  await page.getByTestId(`route-b-selector`).click();
  await page.getByTestId(secondRevision.revision.name).click();

  // move slider
  const slider = await page
    .getByTestId("traffic-shaping-slider")
    .getByRole("slider");
  await slider.dragTo(slider, {
    force: true,
    targetPosition: { x: 50, y: 0 },
  });
  const sliderValue = await slider.getAttribute("aria-valuenow");
  expect(sliderValue).not.toBe("0");

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "there is a hint that describes the traffic shaping"
  ).toHaveText(
    `The traffic will be split between ${firstRevision.revision.name} and ${
      secondRevision.revision.name
    } with a ratio of ${sliderValue} to ${100 - parseInt(sliderValue ?? "")} %`
  );

  // save
  await page.getByTestId("traffic-shaping-save-btn").click();

  // TODO: check for toast message
});
