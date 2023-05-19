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

test("you can't save, traffic shaping with two of the same revisions", async ({
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
  const secondShaping = 100 - parseInt(sliderValue ?? "");
  expect(sliderValue).not.toBe("0");

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "there is a hint that describes the traffic shaping"
  ).toHaveText(
    `The traffic will be split between ${firstRevisionName} and ${secondRevisionName} with a ratio of ${sliderValue} to ${secondShaping} %`
  );

  // save
  await page.getByTestId("traffic-shaping-save-btn").click();

  const firstRevisionRow = await page.getByTestId(
    `revisions-list-${firstRevisionName}`
  );
  await expect(
    firstRevisionRow.getByTestId("traffic-distribution-primary")
  ).toHaveText(`${sliderValue} % of traffic distribution`);

  const secondRevisionRow = await page.getByTestId(
    `revisions-list-${secondRevisionName}`
  );
  await expect(
    secondRevisionRow.getByTestId("traffic-distribution-secondary")
  ).toHaveText(`${secondShaping} % of traffic distribution`);

  // reload page
  await page.reload();

  // TODO: waiting for DIR-576 to get fixed
  // since returned routes are random, we serialize the data in aphabetical order
  // this might lead to swapped routes in the UI after a reload, but this will be
  // constistent from now on (otherwise it would always be random)
  const isInAlphabeticalOrder =
    firstRevisionName.localeCompare(secondRevisionName) === -1;

  await expect(
    page.getByTestId("traffic-shaping-note"),
    "there is a hint that describes the traffic shaping"
  ).toHaveText(
    isInAlphabeticalOrder
      ? `The traffic will be split between ${firstRevisionName} and ${secondRevisionName} with a ratio of ${sliderValue} to ${secondShaping} %`
      : `The traffic will be split between ${secondRevisionName} and ${firstRevisionName} with a ratio of ${secondShaping} to ${sliderValue} %`
  );
});
