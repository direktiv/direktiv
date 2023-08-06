import { createNamespace, deleteNamespace } from "../../../utils/namespace";
import { expect, test } from "@playwright/test";

import { noop as basicWorkflow } from "~/pages/namespace/Explorer/Tree/NewWorkflow/templates";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { createWorkflowWithThreeRevisions } from "../../../utils/revisions";
import { faker } from "@faker-js/faker";
import { headers } from "e2e/utils/testutils";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("by default, traffic shaping is not configured", async ({ page }) => {
  const name = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: basicWorkflow.data,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name,
    },
    headers,
  });

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

test("it is not possible to save traffic shaping when the same revision is selected twice", async ({
  page,
}) => {
  const name = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: basicWorkflow.data,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name,
    },
    headers,
  });

  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);

  // select the latest revision for both routes
  await page.getByTestId(`route-a-selector`).click();
  await page.getByTestId("latest").click();
  await page.getByTestId(`route-b-selector`).click();
  await page.getByTestId("latest").click();

  // double check if the selection was actually made
  await expect(page.getByTestId(`route-a-selector`)).toHaveText("latest");
  await expect(page.getByTestId(`route-b-selector`)).toHaveText("latest");

  await expect(
    page.getByTestId("traffic-shaping-save-btn"),
    "button is disabled"
  ).toBeDisabled();
});

test("it is possible to configure traffic shaping", async ({ page }) => {
  const name = faker.system.commonFileName("yaml");
  const {
    revisionsReponse: [firstRevision, secondRevision],
  } = await createWorkflowWithThreeRevisions(namespace, name);

  const firstRevisionName = firstRevision.revision.name;
  const secondRevisionName = secondRevision.revision.name;

  await page.goto(`/${namespace}/explorer/workflow/revisions/${name}`);
  await page.getByTestId(`route-a-selector`).click();
  await page.getByTestId(firstRevisionName).click();
  await page.getByTestId(`route-b-selector`).click();
  await page.getByTestId(secondRevisionName).click();

  // move slider
  const slider = await page
    .getByTestId("traffic-shaping-slider")
    .getByRole("slider");
  await slider.dragTo(slider, {
    force: true,
    targetPosition: { x: 50, y: 0 },
  });
  const sliderValue = await slider.getAttribute("aria-valuenow");
  const secondWeight = 100 - parseInt(sliderValue ?? "");
  expect(sliderValue).not.toBe("0");

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "there is a hint that describes the traffic shaping"
  ).toHaveText(
    `The traffic will be split between ${firstRevisionName} and ${secondRevisionName} with a ratio of ${sliderValue} to ${secondWeight} %`
  );

  // save
  await page.getByTestId("traffic-shaping-save-btn").click();

  // TODO: waiting for DIR-576 to get fixed
  // since returned routes are random, we serialize the data in aphabetical order
  // this might lead to swapped routes in the UI after a reload, but this will be
  // constistent from now on (otherwise it would always be random)
  const isInAlphabeticalOrder =
    firstRevisionName.localeCompare(secondRevisionName) === -1;

  const revisionA = isInAlphabeticalOrder
    ? firstRevisionName
    : secondRevisionName;
  const revisionB = isInAlphabeticalOrder
    ? secondRevisionName
    : firstRevisionName;
  const weightA = isInAlphabeticalOrder ? sliderValue : secondWeight;
  const weightB = isInAlphabeticalOrder ? secondWeight : sliderValue;

  const firstRevisionRow = await page.getByTestId(
    `revisions-list-${revisionA}`
  );
  await expect(
    firstRevisionRow.getByTestId("traffic-distribution-primary")
  ).toHaveText(`${weightA} % of traffic distribution`);

  const secondRevisionRow = await page.getByTestId(
    `revisions-list-${revisionB}`
  );

  await expect(
    secondRevisionRow.getByTestId("traffic-distribution-secondary")
  ).toHaveText(`${weightB} % of traffic distribution`);

  // reload page
  await page.reload();

  await expect(
    page.getByTestId("route-a-selector"),
    "after reloading, route a selector is prefilled with the previously selected revision"
  ).toHaveText(revisionA.slice(0, 8));

  await expect(
    page.getByTestId("route-b-selector"),
    "after reloading, route b selector is prefilled with the previously selected revision"
  ).toHaveText(revisionB.slice(0, 8));

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "after reloading, there is a hint that describes the traffic shaping"
  ).toHaveText(
    `The traffic will be split between ${revisionA} and ${revisionB} with a ratio of ${weightA} to ${weightB} %`
  );
});
