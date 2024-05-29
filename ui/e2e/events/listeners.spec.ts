import {
  contextFiltersListenerYaml,
  createListener,
  simpleListenerYaml,
} from "./utils";
import { createNamespace, deleteNamespace } from "e2e/utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

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

  const yaml = simpleListenerYaml;

  await Promise.all(
    workflowNames.map((name) => createListener({ name, namespace, yaml }))
  );

  /* visit page and assert a list of listeners is rendered */
  await page.goto(`/n/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/n/${namespace}/events/listeners`
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
  ).toHaveURL(`/n/${namespace}/explorer/workflow/edit/${workflowNames[2]}`);
});

test("it paginates event listeners", async ({ page }) => {
  /* set up test data */
  const workflowNames = Array.from(
    { length: 13 },
    (_, index) => `workflow${index}.yaml`
  );

  const yaml = simpleListenerYaml;

  await Promise.all(
    workflowNames.map((name) => createListener({ name, namespace, yaml }))
  );

  /* visit page and assert a list of listeners is rendered */
  await page.goto(`/n/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/n/${namespace}/events/listeners`
  );

  await expect(
    page.getByRole("cell", { name: "start workflow" }),
    "it renders the expected number of items on page 1"
  ).toHaveCount(10, { timeout: 10000 });

  const paginationWrapper = page.getByTestId("pagination-wrapper");

  await expect(
    paginationWrapper.getByRole("button", { name: "1" })
  ).toBeVisible();
  await expect(
    paginationWrapper.getByRole("button", { name: "2" })
  ).toBeVisible();
  await expect(
    paginationWrapper.getByRole("button", { name: "3" })
  ).not.toBeVisible();

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

test("it renders event context filters", async ({ page }) => {
  /* set up test data */
  const yaml = contextFiltersListenerYaml;
  await createListener({
    name: "listener.yaml",
    namespace,
    yaml,
  });

  /* visit page and assert filters rendered */
  await page.goto(`/n/${namespace}/events/listeners`);

  await expect(page, "it is possible to visit events/listeners ").toHaveURL(
    `/n/${namespace}/events/listeners`
  );

  await expect(
    page.getByRole("cell", { name: "listener.yaml" }),
    "it renders a row for the event listener"
  ).toHaveCount(1);

  await page.getByRole("cell", { name: "2 filters " }).hover();

  const popup = page.getByTestId("context-filter-popup");
  await expect(popup).toBeVisible();

  await expect(
    popup.getByRole("cell", { name: "fake.event.one" })
  ).toBeVisible();
  await expect(
    popup.getByRole("cell", { name: "somekey: somevalue" })
  ).toBeVisible();
  await expect(popup.getByRole("cell", { name: "more: stuff" })).toBeVisible();
  await expect(
    popup.getByRole("cell", { name: "fake.event.two" })
  ).toBeVisible();
  await expect(
    popup.getByRole("cell", { name: "anotherkey: anothervalue" })
  ).toBeVisible();
});
