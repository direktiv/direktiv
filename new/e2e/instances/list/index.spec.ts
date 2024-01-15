import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import {
  parentWorkflow as parentWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "../utils/workflows";

import { createInstance } from "../utils";
import { createWorkflow } from "~/api/tree/mutate/createWorkflow";
import { faker } from "@faker-js/faker";
import { getInstances } from "~/api/instances/query/get";
import { headers } from "e2e/utils/testutils";
import moment from "moment";
import { runWorkflow } from "~/api/tree/mutate/runWorkflow";

type Instance = Awaited<ReturnType<typeof runWorkflow>>;

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");
const failingWorkflowName = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  // place some workflows in the namespace that we can use to create instances
  await createWorkflow({
    payload: simpleWorkflowContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: simpleWorkflowName,
    },
    headers,
  });

  await createWorkflow({
    payload: workflowThatFailsContent,
    urlParams: {
      baseUrl: process.env.VITE_DEV_API_DOMAIN,
      namespace,
      name: failingWorkflowName,
    },
    headers,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it displays a note, when there are no instances yet.", async ({
  page,
}) => {
  await page.goto(`${namespace}/instances/`);
  await expect(
    page.getByTestId("no-result"),
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
  const instances = [
    await createInstance({ namespace, path: simpleWorkflowName }),
    await createInstance({ namespace, path: failingWorkflowName }),
  ];

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

    if (!instanceDetail?.status) {
      throw new Error("instanceDetail?.status is not defined");
    }

    const workflowName = instanceDetail?.as.split(":")[0];

    if (!workflowName) throw new Error("workflowName is not defined");

    const instanceItemRow = page.getByTestId(
      `instance-row-${instance.instance}`
    );

    await expect(
      instanceItemRow.getByTestId(`instance-column-name`),
      "the workflow name should be visible"
    ).toContainText(workflowName);

    const instanceItemIdColumn =
      instanceItemRow.getByTestId("instance-column-id");

    await expect(
      instanceItemIdColumn.getByTestId(`tooltip-copy-trigger`),
      "id badge shows the first 8 digits of the id"
    ).toContainText(instance.instance.slice(0, 8));

    await instanceItemIdColumn.getByTestId(`tooltip-copy-trigger`).hover();

    await expect(
      instanceItemIdColumn.getByTestId("tooltip-copy-content"),
      "on hover, a tooltip reveals full id"
    ).toContainText(instance.instance);

    await page
      .getByRole("heading", { name: "Recently executed instances" })
      .click(); // click on header to close all tooltips opened

    await expect(
      instanceItemRow.getByTestId("instance-column-invoker"),
      'invoker column shows "api"'
    ).toContainText("api");

    await expect(
      instanceItemRow.getByTestId("instance-column-state"),
      "the status column should should be same status as the api response"
    ).toContainText(instanceDetail.status.toString());

    if (instanceDetail?.status === "failed") {
      await instanceItemRow
        .getByTestId("instance-column-state")
        .getByTestId("tooltip-copy-trigger")
        .hover();

      await expect(
        instanceItemRow
          .getByTestId("instance-column-state")
          .getByTestId("tooltip-copy-content"),
        "on hover, a tooltip reveals the error message"
      ).toContainText("this is my error message");
    }

    await page
      .getByRole("heading", { name: "Recently executed instances" })
      .click(); // click on header to close all tooltips opened

    await expect(
      instanceItemRow.getByTestId("instance-column-created-time"),
      `the "started at" column should display a relative time of the createdAt api response`
    ).toContainText(moment(instanceDetail.createdAt).fromNow(true));

    await instanceItemRow
      .getByTestId("instance-column-created-time")
      .getByTestId("tooltip-trigger")
      .hover(); // is force: true needed?

    await expect(
      instanceItemRow
        .getByTestId("instance-column-created-time")
        .getByTestId("tooltip-content"),
      "on hover, the absolute time should appear"
    ).toContainText(instanceDetail.createdAt);

    await page
      .getByRole("heading", { name: "Recently executed instances" })
      .click(); // click on header to close all tooltips opened

    await expect(
      instanceItemRow.getByTestId("instance-column-updated-time"),
      `the "last updated" column should display a relative time of the updatedAt api response`
    ).toContainText(moment(instanceDetail.updatedAt).fromNow(true));

    await instanceItemRow
      .getByTestId("instance-column-updated-time")
      .getByTestId("tooltip-trigger")
      .hover();

    await expect(
      instanceItemRow
        .getByTestId("instance-column-updated-time")
        .getByTestId("tooltip-content"),
      "on hover, the absolute time should appear"
    ).toContainText(instanceDetail.updatedAt);

    await instanceItemRow
      .getByTestId("instance-column-name")
      .getByRole("link")
      .click();

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
      childName: simpleWorkflowName,
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

  const instanceItemRow = page.getByTestId(`instance-row-${firstInstance.id}`);

  await expect(
    instanceItemRow.getByTestId("instance-column-id"),
    "the first row on page 3 should should be same as the api response"
  ).toContainText(firstInstance.id.slice(0, 8));
});

test("It will display child instances as well", async ({ page }) => {
  const parentWorkflow = faker.system.commonFileName("yaml");

  await createWorkflow({
    payload: parentWorkflowContent({
      childName: simpleWorkflowName,
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

  const instanceItemRow = page.getByTestId(
    `instance-row-${childInstanceDetail.id}`
  );

  await expect(
    instanceItemRow.getByTestId("instance-column-invoker"),
    `invoker is "instance"`
  ).toContainText("instance");

  await instanceItemRow
    .getByTestId("instance-column-invoker")
    .getByTestId("tooltip-copy-trigger")
    .hover();

  const expectedInvokerId = childInstanceDetail.invoker.split(":")[1] as string;

  await expect(
    instanceItemRow
      .getByTestId("instance-column-invoker")
      .getByTestId("tooltip-copy-content")
  ).toContainText(expectedInvokerId);
});
