import { createNamespace, deleteNamespace } from "../../utils/namespace";
import {
  workflowWithDelay as delayedWorkflowContent,
  workflowWithFewLogs as fewLogsWorkflowContent,
  workflowWithManyLogs as manyLogsWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
} from "../utils/workflows";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";
import { mockClipboardAPI } from "e2e/utils/testutils";

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");
const delayedWorkflowName = faker.system.commonFileName("yaml");
const fewLogsWorkflowName = faker.system.commonFileName("yaml");
const manyLogsWorkflowName = faker.system.commonFileName("yaml");

test.beforeEach(async ({ page }) => {
  namespace = await createNamespace();
  /* create workflows we can use to create instances later */
  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });

  await createFile({
    name: delayedWorkflowName,
    namespace,
    type: "workflow",
    yaml: delayedWorkflowContent,
  });

  await createFile({
    name: fewLogsWorkflowName,
    namespace,
    type: "workflow",
    yaml: fewLogsWorkflowContent,
  });

  await createFile({
    name: manyLogsWorkflowName,
    namespace,
    type: "workflow",
    yaml: manyLogsWorkflowContent,
  });

  await mockClipboardAPI(page);
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the header of the instance page shows the relevant data for the workflow", async ({
  page,
}) => {
  const instanceId = (
    await createInstance({
      namespace,
      path: simpleWorkflowName,
    })
  ).data.id;

  await page.goto(`/n/${namespace}/instances/${instanceId}`);

  const header = page.getByTestId("instance-header-container");
  await expect(header, "It renders the header").toBeVisible();

  const instanceIdShort = instanceId.slice(0, 8);
  await expect(
    header.locator("h3"),
    "It renders the instance ID in the header"
  ).toHaveText(instanceIdShort);

  await expect(
    page.locator("ul").locator("a").nth(2),
    "It renders the instance ID in the breadcrumb navigation"
  ).toHaveText(instanceIdShort);

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  await expect(
    header.getByText("triggerapi"),
    "It renders the instance trigger"
  ).toBeVisible();

  // check visibility of the time categories but not the exact time stamp, because it is too divergent
  await expect(
    header.getByText("started at"),
    "It renders the category 'started at'"
  ).toBeVisible();
  await expect(
    header.getByText("finished at"),
    "It renders the category 'finished at'"
  ).toBeVisible();

  await expect(
    header.getByText("spawned0 instances"),
    "category spawned shows 0 instances"
  ).toBeVisible();

  await expect(
    header.getByTestId("cancel-workflow"),
    "the button for cancelling the workflow is disabled"
  ).toBeDisabled();

  await header.getByRole("link", { name: "Open workflow" }).click();
  const editURL = `/n/${namespace}/explorer/workflow/edit/${simpleWorkflowName}`;
  await expect(
    page,
    "the button 'Open Workflow' is clickable and links to the URL to edit this workflow"
  ).toHaveURL(editURL);
});
