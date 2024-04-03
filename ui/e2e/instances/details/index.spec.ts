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

  const instanceID = (await newInstance).instance;

  await page.goto(`/${namespace}/instances/${instanceID}`);

  const header = page.getByTestId("instance-header-container");
  await expect(header, "the header is visible").toBeVisible();

  const instanceID_Header = header.locator("h3");
  await expect(instanceID_Header, "the instance ID is visible").toBeVisible();

  const instanceID_navLink = page.locator("ul").locator("a").nth(2);

  expect(instanceID_Header.innerText, "the instance IDs are the same").toEqual(
    instanceID_navLink.innerText
  );

  await expect(
    header.locator("div").first(),
    "the badge complete is visible"
  ).toContainText("complete");

  const trigger = header.locator('div.text-sm:has-text("trigger")');
  await expect(trigger, "the trigger is set to api").toContainText("api");

  // check visibility of the time categories but not the exact time stamp, because it is too divergent
  const startedAt = header.locator('div.text-sm:has-text("started at")');
  await expect(startedAt, "the category 'startedAt' is visible").toBeVisible();
  const lastUpdated = header.locator('div.text-sm:has-text("last updated")');
  await expect(
    lastUpdated,
    "the category 'lastUpdated' is visible"
  ).toBeVisible();

  const spawned = header.locator('div.text-sm:has-text("spawned")');
  await expect(spawned, "category spawned shows 0 instances").toContainText(
    "0 instances"
  );

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
