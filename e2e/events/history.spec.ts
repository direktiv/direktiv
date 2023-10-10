import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to the events page and between the sub pages", async ({
  page,
}) => {
  await page.goto(`/${namespace}`);

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for the test namespace"
  ).toHaveText(namespace);

  page.getByRole("link", { name: "Events" }).click();

  await expect(
    page,
    "it is possible to navigate to events/history via the main navigation menu"
  ).toHaveURL(`/${namespace}/events/history`);

  await expect(
    page.getByTestId("breadcrumb-event-history"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();

  page.getByTestId("event-tabs-trg-listeners").click();

  await expect(
    page,
    "it is possible to navigate to events/listeners via the tab menu"
  ).toHaveURL(`/${namespace}/events/listeners`);

  await expect(
    page.getByTestId("breadcrumb-event-listeners"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();

  page.getByTestId("event-tabs-trg-history").click();

  await expect(
    page,
    "it is possible to navigate to events/history via the tab menu"
  ).toHaveURL(`/${namespace}/events/history`);

  await expect(
    page.getByTestId("breadcrumb-event-history"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();
});
