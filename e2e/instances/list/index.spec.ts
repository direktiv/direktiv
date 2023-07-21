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

interface Instance {
  namespace: string;
  instance: string;
}

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

test("it displays a note, when there are no instances yet.", async ({
  page,
}) => {
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
  const instances: Instance[] = [
    await createFailedInstance(),
    await createBasicInstance(),
  ];

  const checkInstanceRender = async (instance: Instance) => {
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

    const invoker = page.getByTestId(
      `instance-row-invoker-${instance.instance}`
    );
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
      `/${namespace}/explorer/workflow/active/${instanceDetail?.as.split(":")[0]
      }`
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
  };

  for (let i = 0; i < instances.length; i++) {
    const instance = instances[i];
    if (!instance) throw new Error("instance is not created properly");
    await checkInstanceRender(instance);
  }
});

test("it provides a proper pagination", async ({ page }) => {
  const TOTAL_COUNT = 35;
  const PAGE_SIZE = 15;

  await createWorkflow({
    payload: childWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "child.yaml",
    },
  });

  await createWorkflow({
    payload: parentWorkflowContent(TOTAL_COUNT - 1),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "parent.yaml",
    },
  });

  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: "parent.yaml",
    },
  });

  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("pagination-wrapper"),
    "there should be pagination component"
  ).toBeVisible();

  const btnPrev = page.getByTestId("pagination-btn-left");
  const btnNext = page.getByTestId("pagination-btn-right");

  // page number starts from  1
  const activeBtn = page.getByTestId(`pagination-btn-page-1`);
  await expect(
    activeBtn,
    "active button with the page number should have active attribute"
  ).toHaveAttribute("aria-current", "page");

  await expect(
    btnPrev,
    "prev button should be disabled at page 1"
  ).toBeDisabled();

  await expect(
    btnNext,
    "next button should be disabled at last page"
  ).toBeEnabled();

  // go to page 2 by nextButton
  await btnNext.click();

  // go to page 3 by clicking number 3
  const btnNumber3 = page.getByTestId(`pagination-btn-page-3`);
  await btnNumber3.click();

  //check with api response
  const instancesListOfPage = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: PAGE_SIZE,
      offset: 2 * PAGE_SIZE,
    },
  });

  const firstInstance = instancesListOfPage.instances.results[0];
  if (!firstInstance) throw new Error("there should be at least one instance");
  const instanceItemId = page.getByTestId(
    `instance-row-id-${firstInstance.id}`
  );
  await expect(instanceItemId, "ItemId should have the id").toContainText(
    firstInstance.id.slice(0, 8)
  );

  const invoker = page.getByTestId(`instance-row-invoker-${firstInstance?.id}`);
  await expect(
    invoker,
    "invoker should appear in the row, to be instance for this child instance"
  ).toContainText(firstInstance.invoker.split(":")[0] || "");
});

test("the child instance is invoked when you run the parent workflow", async ({
  page,
}) => {
  await createWorkflow({
    payload: childWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: "child.yaml",
    },
  });

  await createWorkflow({
    payload: parentWorkflowContent(1),
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

  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });

  const instancesList = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: 15,
      offset: 0,
    },
  });

  const childInstanceDetail = instancesList.instances.results.find(
    (x) => x.id !== parentInstance.instance
  );

  if (!childInstanceDetail)
    throw new Error("there should be at least one child instance");

  const revisionId = page.getByTestId(
    `instance-row-revision-id-${childInstanceDetail.id}`
  );

  await expect(
    revisionId,
    "revision id should appear as none in the row"
  ).toContainText("none");

  const invoker = page.getByTestId(
    `instance-row-invoker-${childInstanceDetail.id}`
  );
  await expect(
    invoker,
    "invoker should appear in the row, to be instance for this child instance"
  ).toContainText("instance");
});
