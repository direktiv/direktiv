import { createNamespace, deleteNamespace } from "../../utils/namespace";
import { expect, test } from "@playwright/test";
import {
  parentWorkflow as parentWorkflowContent,
  simpleWorkflow as simpleWorkflowContent,
  workflowThatFails as workflowThatFailsContent,
} from "../utils/workflows";

import { createFile } from "e2e/utils/files";
import { createInstance } from "../utils";
import { faker } from "@faker-js/faker";
import { headers } from "e2e/utils/testutils";
import { runWorkflow } from "~/api/tree_obsolete/mutate/runWorkflow";

let namespace = "";
const simpleWorkflowName = faker.system.commonFileName("yaml");
const failingWorkflowName = faker.system.commonFileName("yaml");

test.beforeEach(async () => {
  namespace = await createNamespace();
  /* create workflows we can use to create instances later */
  await createFile({
    name: simpleWorkflowName,
    namespace,
    type: "workflow",
    yaml: simpleWorkflowContent,
  });

  await createFile({
    name: failingWorkflowName,
    namespace,
    type: "workflow",
    yaml: workflowThatFailsContent,
  });
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

/* create 3 instances with status: failed, 2 with status: completed */
const createStatusFilterInstances = async () => {
  await createInstance({ namespace, path: failingWorkflowName });
  await createInstance({ namespace, path: failingWorkflowName });
  await createInstance({ namespace, path: failingWorkflowName });
  await createInstance({ namespace, path: simpleWorkflowName });
  await createInstance({ namespace, path: simpleWorkflowName });
};

/* create 1 instance with trigger "api" and 2 with trigger "instance" */
const createTriggerFilterInstances = async () => {
  const parentWorkflowName = faker.system.commonFileName("yaml");

  await createFile({
    name: parentWorkflowName,
    namespace,
    type: "workflow",
    yaml: parentWorkflowContent({
      childPath: `/${simpleWorkflowName}`,
      children: 2,
    }),
  });

  await runWorkflow({
    urlParams: {
      baseUrl: process.env.PLAYWRIGHT_UI_BASE_URL,
      namespace,
      path: parentWorkflowName,
    },
    headers,
  });
};

test("it is possible to navigate to the instances list, it renders and paginates instances", async ({
  page,
}) => {
  await Promise.all(
    Array.from({ length: 17 }).map(() => {
      createInstance({ namespace, path: simpleWorkflowName });
    })
  );

  await page.goto(`${namespace}/`);

  await page
    .getByRole("navigation")
    .getByRole("link", { name: "Instances" })
    .click();

  await expect(
    page.getByTestId(/instance-row/),
    "it lists the expected number of instances on page 1"
  ).toHaveCount(15);

  await expect(page.getByLabel("Pagination")).toBeVisible();
  await expect(
    page.getByLabel("Pagination").getByRole("button", {
      name: "1",
    })
  ).toBeVisible();
  await expect(
    page.getByLabel("Pagination").getByRole("button", {
      name: "2",
    })
  ).toBeVisible();
  await page.getByTestId("pagination-btn-right").click();

  await expect(
    page.getByTestId(/instance-row/),
    "it lists the expected number of instances on page 2"
  ).toHaveCount(2);

  await page.getByTestId("pagination-btn-left").click();

  await expect(
    page.getByTestId(/instance-row/),
    "it lists the expected number of instances on page 1"
  ).toHaveCount(15);
});

test("it is possible to filter by date using created before", async ({
  page,
}) => {
  await createInstance({ namespace, path: failingWorkflowName });
  await createInstance({ namespace, path: failingWorkflowName });

  await page.goto(`${namespace}/instances/`);

  /* there should be 2 items initially */
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 2 instances"
  ).toHaveCount(2);

  /* filter createdAfter now should return 0 results */
  await page.getByTestId("filter-add").click();

  await page.getByRole("option", { name: "created before" }).click();

  const today = new Date().getDate();

  await page.getByText(today.toString(), { exact: true }).last().click();
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 0 rows when we filter before today"
  ).toHaveCount(0);

  /* remove the date filter */
  await page.getByTestId("filter-clear-BEFORE").click();
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 2 rows after removing the filter"
  ).toHaveCount(2);

  /* filter by created after (with date in the future) */
  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "created after" }).click();
  await page.getByLabel("Go to next month").click();
  await page.getByText("15", { exact: true }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 0 rows when filtering by created after with a future date"
  ).toHaveCount(0);

  await page.getByTestId("filter-clear-AFTER").click();
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 2 rows after removing the filter"
  ).toHaveCount(2);
});

test("it is possible to filter by trigger", async ({ page }) => {
  await createTriggerFilterInstances();
  await page.goto(`${namespace}/instances/`);

  /* there should be 3 items initially */
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 3 rows"
  ).toHaveCount(3);

  const btnPlus = page.getByTestId("filter-add");
  await btnPlus.click();
  await page.getByRole("option", { name: "trigger" }).click();
  await page.getByRole("option", { name: "instance" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 2 rows with filter trigger: instance"
  ).toHaveCount(2);

  /* change trigger filter to "api", expect 1 instance to be rendered */
  await page
    .getByTestId("filter-component")
    .getByRole("button", { name: "instance" })
    .click();
  await page.getByRole("option", { name: "api" }).click();
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 1 rows with filter trigger: api"
  ).toHaveCount(1);

  /* clear filter, expect 3 instances to be rendered */
  await page.getByTestId("filter-clear-TRIGGER").click();
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 3 rows when we cancel the filter"
  ).toHaveCount(3);
});

test("it is possible to filter by status", async ({ page }) => {
  await createStatusFilterInstances();
  await page.goto(`${namespace}/instances/`);

  /* there should be 5 items initially */
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 5 rows"
  ).toHaveCount(5);

  const btnPlus = page.getByTestId("filter-add");

  /* filter by status "complete", expect 2 results to be rendered */
  await btnPlus.click();
  await page.getByRole("option", { name: "status" }).click();
  await page.getByRole("option", { name: "complete" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 2 rows with filter status: complete"
  ).toHaveCount(2);

  /* change filter to status "failed", expect 3 results to be rendered */
  await page.getByRole("button", { name: "complete" }).click();
  await page.getByRole("option", { name: "failed" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 3 rows with filter status: failed"
  ).toHaveCount(3);

  /* clear filter, expect 5 results to be rendered */
  await page.getByTestId("filter-clear-STATUS").click();

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 5 rows when we clear the filter"
  ).toHaveCount(5);
});

test("it is possible to filter by AS (name)", async ({ page }) => {
  const workflowNames = [
    "workflow1.yaml",
    "workflow2.yaml",
    "workflow3.yaml",
    "test.yaml",
  ];

  await Promise.all(
    workflowNames.map((name) =>
      createFile({
        yaml: simpleWorkflowContent,
        namespace,
        name,
        type: "workflow",
      })
    )
  );

  await Promise.all(
    workflowNames.map((path) =>
      createInstance({
        path,
        namespace,
      })
    )
  );

  await page.goto(`${namespace}/instances/`);

  /* there should be 4 items initially */
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 4 rows"
  ).toHaveCount(4);

  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "name" }).click();

  await page.getByPlaceholder("filename.yaml").type("workflow");
  await page.getByPlaceholder("filename.yaml").press("Enter");

  /* filter by name "workflow", result should be 3 */
  await expect(
    page.getByTestId(/instance-row/),
    "there should be 3 rows with filter name: workflow"
  ).toHaveCount(3);

  /* change the filter to name "test", result should be 1 */
  await page.getByRole("button", { name: "workflow" }).click();
  await page.getByPlaceholder("filename.yaml").fill("test");
  await page.getByPlaceholder("filename.yaml").press("Enter");

  await expect(
    page.getByTestId(/instance-row/),
    "there should be 1 rows with filter name: test"
  ).toHaveCount(1);

  /* clear filter */
  await page.getByTestId("filter-clear-AS").click();
  await expect(
    page.getByTestId(/instance-row/),
    "after clearing the filter, there should be 4 results again"
  ).toHaveCount(4);
});

test("it is possible to apply multiple filters", async ({ page }) => {
  /* set up test data */
  createTriggerFilterInstances();

  const workflowNames = [
    "five-a.yaml",
    "five-b.yaml",
    "five-c.yaml",
    "five-d.yaml",
    "five-f.yaml",
  ];

  await Promise.all(
    workflowNames.map((name) =>
      createFile({
        yaml: simpleWorkflowContent,
        name,
        namespace,
        type: "workflow",
      })
    )
  );

  await Promise.all(
    workflowNames.map((path) => {
      createInstance({ namespace, path });
    })
  );

  const failingInstances = Array.from({ length: 3 }).map(() =>
    createInstance({ namespace, path: failingWorkflowName })
  );

  await Promise.all(failingInstances);

  /* visit page and test initial state */
  await page.goto(`${namespace}/instances/`);

  await expect(
    page.getByTestId(/instance-row/),
    "initially, it renders all instances"
  ).toHaveCount(11);

  /* add filter by "trigger": "api" */
  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "trigger" }).click();
  await page.getByRole("option", { name: "api" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(9);

  /* add second filter, by "status": "failed" */
  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "status" }).click();
  await page.getByRole("option", { name: "failed" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(3);

  /* change filter by "trigger" to "instance" */
  await page.getByTestId("filter-add").click();
  await page
    .getByTestId("filter-component")
    .getByRole("button", { name: "api" })
    .click();
  await page.getByRole("option", { name: "instance" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(0);

  /* change filter by "status" to "complete" */
  await page.getByTestId("filter-add").click();
  await page
    .getByTestId("filter-component")
    .getByRole("button", { name: "failed" })
    .click();
  await page.getByRole("option", { name: "complete" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(2);

  /* remove filter by "trigger" */
  await page.getByTestId("filter-clear-TRIGGER").click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(8);

  /* add second filter by "trigger": "api" (filter) */
  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "trigger" }).click();
  await page.getByRole("option", { name: "api" }).click();

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(6);

  /* add third filter by "name": "five" (filter) */
  await page.getByTestId("filter-add").click();
  await page.getByRole("option", { name: "name" }).click();
  await page.getByPlaceholder("filename.yaml").fill("five");
  await page.getByPlaceholder("filename.yaml").press("Enter");

  await expect(
    page.getByTestId(/instance-row/),
    "it renders the expected number of results"
  ).toHaveCount(5);
});
