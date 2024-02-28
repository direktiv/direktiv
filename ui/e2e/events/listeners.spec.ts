import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

import { createFile } from "e2e/utils/files";

let namespace = "";

const createListener = async (name: string) => {
  const yaml = `direktiv_api: workflow/v1
description: This workflow spawns an event listener as soon as the file is created
start:
  type: event
  event:
    type: fake.event.one
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!`;

  await createFile({
    name,
    namespace,
    type: "workflow",
    yaml,
  });
};

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it renders event listeners", async ({ page }) => {
  /* set up test data */
  const workflowNames = Array.from(
    { length: 4 },
    (_, index) => `workflow${index}.yaml`
  );

  await Promise.all(workflowNames.map((name) => createListener(name)));

  /* visit page and assert a list of listeners is rendered */
  await page.goto(`/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/${namespace}/events/listeners`
  );

  await expect(
    page.getByTestId("breadcrumb-event-listeners"),
    "it renders the 'Event Listeners' breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByRole("cell", { name: "start workflow" }),
    'type "start workflow" is rendered for every event listener'
  ).toHaveCount(4);

  await Promise.all(
    workflowNames.map((name, index) =>
      expect(
        page.getByRole("cell", { name }),
        `each workflow name is rendered in the list (${index} of ${workflowNames.length})`
      ).toBeVisible()
    )
  );

  await expect(
    page.getByRole("cell", { name: "simple" }),
    'mode "simple" is rendered for every event listener'
  ).toHaveCount(4);

  await expect(
    page.getByTestId("receivedAt-tooltip-trigger"),
    "the time tooltip is rendered for every event listener"
  ).toHaveCount(4);

  await expect(
    page.getByRole("cell", { name: "fake.event.one" }),
    "the event type is rendered for every event listener"
  ).toHaveCount(4);

  await page.getByRole("link", { name: workflowNames[2] }).click();

  await expect(
    page,
    "when clicking on the workflow name, it navigates to the workflow page"
  ).toHaveURL(`${namespace}/explorer/workflow/edit/${workflowNames[2]}`);
});

test("it paginates event listeners", async ({ page }) => {
  /* set up test data */
  const workflowNames = Array.from(
    { length: 13 },
    (_, index) => `workflow${index}.yaml`
  );

  await Promise.all(workflowNames.map((name) => createListener(name)));

  /* visit page and assert a list of listeners is rendered */
  await page.goto(`/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/${namespace}/events/listeners`
  );

  await expect(
    page.getByRole("cell", { name: "start workflow" }),
    "it renders the expected number of items on page 1"
  ).toHaveCount(10);

  await expect(page.getByLabel("Pagination")).toBeVisible();
  await expect(page.getByTestId("pagination-btn-page-1")).toBeVisible();
  await expect(page.getByTestId("pagination-btn-page-2")).toBeVisible();
  await expect(page.getByTestId("pagination-btn-page-3")).not.toBeVisible();

  await page.getByTestId("pagination-btn-right").click();

  await expect(
    page.getByRole("cell", { name: "start workflow" }),
    "it navigates to page 2 and renders the expected number of items"
  ).toHaveCount(3);

  await page.getByTestId("pagination-btn-left").click();

  await expect(
    page.getByRole("cell", { name: "start workflow" }),
    "it navigates to page 1 and renders the expected number of items"
  ).toHaveCount(10);
});
