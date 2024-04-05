import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils/index";
import { faker } from "@faker-js/faker";
import { simpleWorkflow as simpleWorkflowContent } from "../utils/workflows";

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  /* create workflows we can use to create instances later */
  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the header of the instance page shows the relevant data for the workflow", async ({
  page,
}) => {
  const newInstance = createInstance({ namespace, path: simpleWorkflowName });

  await expect(newInstance, "wait until process was completed").toBeDefined();

  const instanceId = (await newInstance).instance;

  await page.goto(`/${namespace}/instances/${instanceId}`);

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
    header.getByText("last updated"),
    "It renders the category 'last updated'"
  ).toBeVisible();

  await expect(
    header.getByText("spawned0 instances"),
    "category spawned shows 0 instances"
  ).toBeVisible();

  await expect(
    header.getByRole("button").locator("svg.lucide-xcircle"),
    "the button for cancelling the workflow is disabled"
  ).toBeDisabled();

  await header.getByRole("link", { name: "Open workflow" }).click();
  const editURL = `${namespace}/explorer/workflow/edit/${simpleWorkflowName}`;
  await expect(
    page,
    "the button 'Open Workflow' is clickable and links to the correct URL"
  ).toHaveURL(editURL);
});
