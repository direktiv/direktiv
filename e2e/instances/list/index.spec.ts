import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import {
  parentWorkflow as parentWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "./utils";

import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { getInstances } from "~/api/instances/query/get";
import { headers } from "e2e/utils/testutils";
import moment from "moment";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";

type Instance = Awaited<ReturnType<typeof runWorkflow>>;

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
    headers,
  });

  await createWorkflow({
    payload: workflowThatFailsContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: workflowThatFails,
    },
    headers,
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
    headers,
  });

const createFailedInstance = async () =>
  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: workflowThatFails,
    },
    headers,
  });

test("it displays a note, when there are no instances yet.", async ({
  page,
}) => {
  await page.goto(`${namespace}/instances/`);
  await expect(
    page.getByTestId("instance-no-result"),
    "no result message should be visible"
  ).toBeVisible();
  await expect(
    page.getByTestId("instance-list-pagination"),
    "there is no pagination when there is no result"
  ).not.toBeVisible();
});

test("it renders the instance item correctly for failed and success status", async ({
  page,
}) => {
  const instances = [await createFailedInstance(), await createBasicInstance()];

  const checkInstanceRender = async (instance: Instance) => {
    const instancesList = await getInstances({
      urlParams: {
        baseUrl: process.env.VITE_DEV_API_DOMAIN,
        namespace,
        limit: 10,
        offset: 0,
      },
      headers,
    });

    const instanceDetail = instancesList.instances.results.find(
      (x) => x.id === instance.instance
    );

    const workflowName = instanceDetail?.as.split(":")[0];

    if (!workflowName) throw new Error("workflowName is not defined");

    await expect(
      page.getByTestId(`instance-row-workflow-${instance.instance}`),
      "the workflow name should be visible"
    ).toContainText(workflowName);

    const instanceItemRow = page.getByTestId(
      `instance-row-wrap-${instance.instance}`
    );
    const instanceItemId = page.getByTestId(
      `tooltip-copy-badge-${instance.instance}`
    );
    await expect(
      instanceItemId,
      "id column shows the first 8 digits of the id"
    ).toContainText(instance.instance.slice(0, 8));
    await instanceItemId.hover();
    const idTooltip = page.getByTestId(
      `tooltip-copy-badge-content-${instance.instance}`
    );
    await expect(
      idTooltip,
      "on hover, a tooltip reveals full id"
    ).toContainText(instance.instance);

    const invoker = page.getByTestId(
      `instance-row-invoker-${instance.instance}`
    );

    await expect(invoker, `invoker column shows "api"`).toContainText("api");

    const state = page.getByTestId(`instance-row-state-${instance.instance}`);

    if (!instanceDetail?.status) {
      throw new Error("instanceDetail?.status is not defined");
    }

    await expect(
      state,
      "the status columns should should be same status as the api response"
    ).toContainText(instanceDetail?.status.toString());

    if (instanceDetail?.status === "failed") {
      await state.hover();
      const errorTooltip = page.getByTestId(
        `instance-row-state-error-tooltip-${instance.instance}`
      );

      await expect(
        errorTooltip,
        "on hover, a tooltip reveals the error message"
      ).toContainText("this is my error message");
    }

    const createdRelTime = page.getByTestId(
      `instance-row-relative-created-time-${instance.instance}`
    );
    await expect(
      createdRelTime,
      `the "started at" column should display a relative time of the createdAt api response`
    ).toContainText(moment(instanceDetail.createdAt).fromNow(true));

    await page
      .getByRole("heading", { name: "Recently executed instances" })
      .click(); // click on table header to close all tooltips opened

    await createdRelTime.hover({ force: true });
    const createdAtTooltip = page.getByTestId(
      `instance-row-absolute-created-time-${instance.instance}`
    );
    await expect(
      createdAtTooltip,
      "on hover, the absolute time should appear"
    ).toContainText(instanceDetail.createdAt);

    const updatedRelTime = page.getByTestId(
      `instance-row-relative-updated-time-${instance.instance}`
    );
    await expect(
      updatedRelTime,
      `the "last updateed" column should display a relative time of the updatedAt api response`
    ).toContainText(moment(instanceDetail.updatedAt).fromNow(true));

    await page
      .getByRole("heading", { name: "Recently executed instances" })
      .click(); // click on table header to close all tooltips opened
    await updatedRelTime.hover({ force: true });
    const updatedAtTooltip = page.getByTestId(
      `instance-row-absolute-updated-time-${instance.instance}`
    );
    await expect(
      updatedAtTooltip,
      "on hover, the absolute time should appear"
    ).toContainText(instanceDetail.updatedAt);

    const workflowLink = page.getByTestId(
      `instance-row-workflow-${instance.instance}`
    );

    await workflowLink.click();
    await expect(
      page,
      "when the workflow name is clicked, page should navigate to the workflow page"
    ).toHaveURL(`/${namespace}/explorer/workflow/active${workflowName}`);

    await page.goBack();
    await instanceItemRow.click();
    await expect(
      page,
      "on click row, page should navigate to the instance detail page"
    ).toHaveURL(`/${namespace}/instances/${instance.instance}`);
    await page.goBack();
  };

  await page.goto(`${namespace}/instances/`);

  for (let i = 0; i < instances.length; i++) {
    const instance = instances[i];
    if (!instance) throw new Error("instance is not created properly");
    await checkInstanceRender(instance);
  }

  await expect(
    page.getByTestId("instance-list-pagination"),
    "no pagination is visible when there is only one page"
  ).not.toBeVisible();
});

test("it provides a proper pagination", async ({ page }) => {
  const totalCount = 35;
  const pageSize = 15;

  const parentWorkflow = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: parentWorkflowContent({
      childName: simpleWorkflow,
      children: totalCount - 1,
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: parentWorkflow,
    },
    headers,
  });

  await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: parentWorkflow,
    },
    headers,
  });

  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });
  await expect(
    page.getByTestId("pagination-wrapper"),
    "there should be pagination component"
  ).toBeVisible();

  const btnPrev = page.getByTestId("pagination-btn-left");
  const btnNext = page.getByTestId("pagination-btn-right");
  const page1Btn = page.getByTestId(`pagination-btn-page-1`);

  // page number starts from  1
  await expect(
    page1Btn,
    "active button with the page number should have an aria-current attribute with a value of page"
  ).toHaveAttribute("aria-current", "page");

  await expect(
    btnPrev,
    "prev button should be disabled at page 1"
  ).toBeDisabled();

  await expect(
    btnNext,
    "next button should be enabled at page 1"
  ).toBeEnabled();

  // go to page 2 by clicking nextButton
  await btnNext.click();
  await expect(
    btnPrev,
    "prev button should be enabled at page 2"
  ).toBeEnabled();

  await expect(
    btnNext,
    "next button should be enabled at page 2"
  ).toBeEnabled();

  // go to page 3 by clicking number 3
  const btnNumber3 = page.getByTestId(`pagination-btn-page-3`);
  await btnNumber3.click();

  // check with api response
  const instancesListPage3 = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: pageSize,
      offset: 2 * pageSize,
    },
    headers,
  });

  const firstInstance = instancesListPage3.instances.results[0];
  if (!firstInstance) throw new Error("there should be at least one instance");
  const instanceItemId = page.getByTestId(
    `tooltip-copy-badge-${firstInstance.id}`
  );
  await expect(
    instanceItemId,
    "the first row on page 3 should should be same as the api response"
  ).toContainText(firstInstance.id.slice(0, 8));
});

test("It will display child instances as well", async ({ page }) => {
  const parentWorkflow = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: parentWorkflowContent({
      childName: simpleWorkflow,
      children: 1,
    }),
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: parentWorkflow,
    },
    headers,
  });

  const parentInstance = await runWorkflow({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      path: parentWorkflow,
    },
    headers,
  });

  await page.goto(`${namespace}/instances/`, { waitUntil: "networkidle" });

  const instancesList = await getInstances({
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      limit: 15,
      offset: 0,
    },
    headers,
  });

  const childInstanceDetail = instancesList.instances.results.find(
    (x) => x.id !== parentInstance.instance
  );

  if (!childInstanceDetail)
    throw new Error("there should be at least one child instance");

  const invoker = page.getByTestId(
    `instance-row-invoker-${childInstanceDetail.id}`
  );
  await expect(invoker, `invoker is "instance"`).toContainText("instance");

  const instanceItemId = page.getByTestId(
    `tooltip-copy-badge-${childInstanceDetail.id}`
  );
  await expect(
    instanceItemId,
    "id column shows the first 8 digits of the id"
  ).toContainText(childInstanceDetail.id.slice(0, 8));
  await instanceItemId.hover();
  const idTooltip = page.getByTestId(
    `tooltip-copy-badge-content-${childInstanceDetail.id}`
  );
  await expect(idTooltip, "on hover, a tooltip reveals full id").toContainText(
    childInstanceDetail.id
  );
});
