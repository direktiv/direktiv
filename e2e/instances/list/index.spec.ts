import {
  childWorkflow as childWorkflowContent,
  parentWorkflow as parentWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "./utils";
import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { getInstances } from "~/api/instances/query/get";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";

let namespace = "";
const simpleWorkflow = faker.system.commonFileName("yaml");
const workflowThatFails = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  // place some workflows in the namespace that we can use to create instances
  await createWorkflow({
    payload: simpleWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: simpleWorkflow,
    },
  });

  await createWorkflow({
    payload: workflowThatFailsContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflowThatFails,
    },
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

const createBasicInstance = async () =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: simpleWorkflow,
    },
  });

const createFailedInstance = async () =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: workflowThatFails,
    },
  });

test("this is an example test", async ({ page }) => {
  // const failedInstance = await createFailedInstance();
  // await createBasicInstance();
  // await createBasicInstance();

  // const instancesList = await getInstances({
  //   urlParams: {
  //     baseUrl: process.env.VITE_DEV_API_DOMAIN,
  //     namespace,
  //     limit: 10,
  //     offset: 0,
  //   },
  // });

  // const failedInstanceServerRes = instancesList.instances.results.find(
  //   (x) => x.id === failedInstance.instance
  // );

  // console.log("🚀", failedInstanceServerRes);

  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
});

test("there is no result", async ({ page }) => {
  // await createFailedInstance();
  // await createBasicInstance();
  // await createBasicInstance();
  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("instance-no-result"),
    "no result message should be visible"
  ).toBeVisible();
  await expect(
    page.getByTestId("instance-list-pagination"),
    "there is no pagination when there is no result"
  ).not.toBeVisible();
});

test("renders the instance item correctly for failed and success status", async ({
  page,
}) => {
  const instance =
    Math.random() > 0.5
      ? await createFailedInstance()
      : await createBasicInstance();
  const instancesList = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: 10,
      offset: 0,
    },
  });

  const instanceDetail = instancesList.instances.results.find(
    (x) => x.id === instance.instance
  );

  // { waitUntil: "networkidle" } might not be necessary, I just added
  // it so that running this example will show a nicely loaded page
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });

  const instanceItemRow = page.getByTestId(
    `instance-row-wrap-${instance.instance}`
  );
  const instanceItemId = page.getByTestId(
    `instance-row-id-${instance.instance}`
  );
  await expect(instanceItemId, "ItemId should have the id").toContainText(
    instance.instance.slice(0, 8)
  );
  await instanceItemId.hover();
  const idTooltip = page.getByTestId(
    `instance-row-id-full-${instance.instance}`
  );
  await expect(
    idTooltip,
    "on hover, there should be a tooltip that contains full id"
  ).toContainText(instance.instance);

  const revisionId = page.getByTestId(
    `instance-row-revision-id-${instance.instance}`
  );
  await expect(
    revisionId,
    "revision id should appear in the row"
  ).toContainText("latest");

  const invoker = page.getByTestId(`instance-row-invoker-${instance.instance}`);
  await expect(
    invoker,
    "invoker should appear in the row, to be api for this instance"
  ).toContainText("api");

  const state = page.getByTestId(`instance-row-state-${instance.instance}`);

  await expect(
    state,
    "state should appear in the row, to be same status from the api response"
  ).toContainText(instanceDetail?.status.toString() || "pending");

  if (instanceDetail?.status === "failed") {
    await state.hover();
    const errorTooltip = page.getByTestId(
      `instance-row-state-error-tooltip-${instance.instance}`
    );
    await expect(
      errorTooltip,
      "on hover the failed badge, error tooltip should appear"
    ).toBeVisible();
  }

  const createdReltime = page.getByTestId(
    `instance-row-relative-created-time-${instance.instance}`
  );
  await expect(createdReltime, "createAt should be visible").toBeVisible();

  await page
    .getByRole("heading", { name: "Recently executed instances" })
    .click(); // click on table header to close all tooltips opened
  await createdReltime.hover({ force: true });
  const createdAtTooltip = page.getByTestId(
    `instance-row-absolute-created-time-${instance.instance}`
  );
  await expect(
    createdAtTooltip,
    "on hover, the absolute time should appear"
  ).toBeVisible();

  const updatedReltime = page.getByTestId(
    `instance-row-relative-updated-time-${instance.instance}`
  );
  await expect(updatedReltime, "updateAd should be visible").toBeVisible();

  await page
    .getByRole("heading", { name: "Recently executed instances" })
    .click(); // click on table header to close all tooltips opened
  await updatedReltime.hover({ force: true });
  const updatedAtTooltip = page.getByTestId(
    `instance-row-absolute-updated-time-${instance.instance}`
  );
  await expect(
    updatedAtTooltip,
    "on hover, the absolute time should appear"
  ).toBeVisible();

  const workflowLink = page.getByTestId(
    `instance-row-workflow-${instance.instance}`
  );
  await workflowLink.click();
  await expect(
    page,
    "on click workflow, page should navigate to the workflow page"
  ).toHaveURL(
    `/${namespace}/explorer/workflow/active/${instanceDetail?.as.split(":")[0]}`
  );

  await page.goBack();
  await instanceItemRow.click();
  await expect(
    page,
    "on click row, page should navigate to the instance detail page"
  ).toHaveURL(`/${namespace}/instances/${instance.instance}`);

  //no pagination to be visibe
  await expect(
    page.getByTestId("instance-list-pagination"),
    "there is no pagination when there aren't many instances"
  ).not.toBeVisible();
});

test("test the pagination", async ({ page }) => {
  test.setTimeout(200000);

  await createWorkflow({
    payload: childWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "child.yaml",
    },
  });

  await createWorkflow({
    payload: parentWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "parent.yaml",
    },
  });

  const parentInstance = await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: "parent.yaml",
    },
  });

  // there needs to be a discussion to check the total count
  // expect(instancesList.instances.pageInfo.total, "length of the instance list should be 5").toBe(5);
  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });

  await expect(
    page.getByTestId("pagination-wrapper"),
    "there should be pagination component"
  ).toBeVisible();

  const btnPrev = page.getByTestId("pagination-btn-left");
  const btnNext = page.getByTestId("pagination-btn-right");

  //loop through all pages by clicking Next button
  for (let p = 1; p < Math.ceil(170 / 15) + 1; p++) {
    const activeBtn = page.getByTestId(`pagination-btn-page-${p}`);
    await expect(
      activeBtn,
      "active button with the page number should have active class"
    ).toHaveClass(/z-10 bg-gray-3 dark:bg-gray-dark-3/);

    if (p === 1) {
      await expect(
        btnPrev,
        "prev button should be disabled at page 1"
      ).toBeDisabled();
    } else if (p === Math.ceil(170 / 15)) {
      await expect(
        btnNext,
        "next button should be disabled at last page"
      ).toBeDisabled();
    } else {
      await expect(
        btnNext,
        "next button should be enabled except at page 1"
      ).toBeEnabled();
      await expect(
        btnPrev,
        "next button should be enabled except at last page"
      ).toBeEnabled();
    }
    if (p !== Math.ceil(170 / 15)) {
      await btnNext.click();
    }
  }

  // we are at the last page now
  // get the the first row which should be child and test it
  const instancesList = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: 0,
      offset: 165,
    },
  });

  const childInstanceDetail = instancesList.instances.results.find(
    (x) => x.id !== parentInstance.instance
  );

  const revisionId = page.getByTestId(
    `instance-row-revision-id-${childInstanceDetail?.id}`
  );

  await expect(
    revisionId,
    "revision id should appear as none in the row"
  ).toContainText("none");

  const invoker = page.getByTestId(
    `instance-row-invoker-${childInstanceDetail?.id}`
  );
  await expect(
    invoker,
    "invoker should appear in the row, to be instance for this child instance"
  ).toContainText("instance");
});
