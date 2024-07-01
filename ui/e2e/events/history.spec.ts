import { createNamespace, deleteNamespace } from "../utils/namespace";
import { expect, test } from "@playwright/test";

import { createEvents } from "e2e/utils/events";

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
  await page.goto(`/n/${namespace}`);

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "it renders the breadcrumb for the test namespace"
  ).toHaveText(namespace);

  page.getByRole("link", { name: "Events" }).click();

  await expect(
    page,
    "it is possible to navigate to events/history via the main navigation menu"
  ).toHaveURL(`/n/${namespace}/events/history`);

  await expect(
    page.getByTestId("breadcrumb-event-history"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();

  page.getByTestId("event-tabs-trg-listeners").click();

  await expect(
    page,
    "it is possible to navigate to events/listeners via the tab menu"
  ).toHaveURL(`/n/${namespace}/events/listeners`);

  await expect(
    page.getByTestId("breadcrumb-event-listeners"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();

  page.getByTestId("event-tabs-trg-history").click();

  await expect(
    page,
    "it is possible to navigate to events/history via the tab menu"
  ).toHaveURL(`/n/${namespace}/events/history`);

  await expect(
    page.getByTestId("breadcrumb-event-history"),
    "it renders the 'Event History' breadcrumb"
  ).toBeVisible();
});

test("it is possible to send a new event", async ({ page }) => {
  await page.goto(`/n/${namespace}/events/history`);

  await expect(page, "it is possible to visit events/history ").toHaveURL(
    `/n/${namespace}/events/history`
  );

  await expect(
    page.getByTestId("no-result").getByText("No events found")
  ).toBeVisible();
  await expect(page.getByTestId("event-row")).toHaveCount(0);

  page.getByRole("button", { name: "Send new event" }).click();

  await expect(page.getByTestId("send-event-form")).toBeVisible();

  page.getByRole("button", { name: "Send" }).click();

  await expect(page.getByTestId("no-result")).not.toBeVisible();

  await expect(page.getByTestId("event-row")).toHaveCount(1);
});

test("it renders, filters, and paginates events", async ({ page }) => {
  const events = await createEvents(namespace);

  /**
   * Visit the page, test pagination.
   */

  await page.goto(`/n/${namespace}/events/history`);

  await expect(page, "it is possible to visit events/history").toHaveURL(
    `/n/${namespace}/events/history`
  );

  const paginationWrapper = page.getByTestId("pagination-wrapper");

  await expect(
    paginationWrapper.getByRole("button", { name: "1" })
  ).toBeVisible();
  await expect(
    paginationWrapper.getByRole("button", { name: "2" })
  ).toBeVisible();
  await expect(
    paginationWrapper.getByRole("button", { name: "3" })
  ).toBeVisible();
  await expect(
    paginationWrapper.getByRole("button", { name: "4" })
  ).not.toBeVisible();

  await expect(page.getByTestId("event-row")).toHaveCount(10);

  await paginationWrapper.getByRole("button", { name: "2" }).click();
  await expect(page.getByTestId("event-row")).toHaveCount(10);

  await paginationWrapper.getByRole("button", { name: "3" }).click();
  await expect(page.getByTestId("event-row")).toHaveCount(2);

  /**
   * Test the select options for page size
   */

  const selectPagesize = page.getByRole("combobox");
  await expect(selectPagesize).toBeVisible();
  expect(selectPagesize).toHaveText("Show 10 rows");

  selectPagesize.click();
  page.getByLabel("Show 30 rows").click();

  await expect(page.getByTestId("event-row")).toHaveCount(22);

  selectPagesize.click();
  page.getByLabel("Show 20 rows").click();

  await expect(page.getByTestId("event-row")).toHaveCount(20);

  await paginationWrapper.getByRole("button", { name: "2" }).click();
  await expect(page.getByTestId("event-row")).toHaveCount(2);

  /**
   * Filter by event type and expect a subset of the events to be returned.
   */

  await page.getByRole("button", { name: "Filter" }).click();
  await page.getByRole("option", { name: "type" }).click();

  await page.getByPlaceholder("cloud.event.type").fill("foo.bar.alpha");
  await page.getByPlaceholder("cloud.event.type").press("Enter");

  await expect(
    page.getByTestId("event-row"),
    "when filtering by event type, it shows the correct number of events"
  ).toHaveCount(5);

  await page.getByRole("button", { name: "foo.bar.alpha" }).click();

  await page.getByPlaceholder("cloud.event.type").fill("foo.bar.delta");
  await page.getByPlaceholder("cloud.event.type").press("Enter");

  await expect(
    page.getByTestId("event-row"),
    "when filtering by event type, it shows the correct number of events"
  ).toHaveCount(7);

  await expect(page.getByTestId("pagination-wrapper")).not.toBeVisible();

  /**
   * Additionally filter by content, expect a smaller subset of results.
   * The content filter will search everywhere in the event, so in this case,
   * we search for the url in the "source" property.
   */

  await page.getByTestId("add-filter").click();
  await page.getByRole("option", { name: "content contains" }).click();

  await page
    .getByPlaceholder("search cloudevent content")
    .fill("http://example.two");

  await page.getByPlaceholder("search cloudevent content").press("Enter");

  await expect(
    page.getByTestId("event-row"),
    "when filtering by event type + content, it shows the correct number of events"
  ).toHaveCount(4);

  await page.getByTestId("clear-filter-typeContains").click();

  await expect(
    page.getByTestId("event-row"),
    "after removing the type filter, it shows the correct number of events"
  ).toHaveCount(9);

  /**
   * Filter by date. Since all events are generated at roughly the same time it would
   * be too complicated to isolate a specific group of events. So this is just a
   * smoke test: we select a "created after" date in the future and expect that
   * no results are returned.
   */

  await page.getByTestId("add-filter").click();
  await page.getByRole("option", { name: "received after" }).click();
  await page.getByLabel("Go to next month").click();
  await page.getByText("15", { exact: true }).click();

  await expect(page.getByTestId("event-row")).toHaveCount(0);

  await expect(page.getByTestId("no-result")).toContainText(
    "No events found with these filter criteria"
  );

  await page.getByTestId("clear-filter-receivedAfter").click();
  await expect(
    page.getByTestId("event-row"),
    "after removing the date filter, it shows the correct number of events"
  ).toHaveCount(9);

  /**
   * Filter by event content to find a unique event,
   * assert that all columns are rendered correctly
   */

  const subject = events[7] as { type: string; source: string; data: string };

  await page.getByRole("button", { name: "http://example.two" }).click();

  await page.getByPlaceholder("search cloudevent content").fill(subject.data);
  await page.getByPlaceholder("search cloudevent content").press("Enter");

  await expect(
    page.getByTestId("event-row"),
    "when filtering by content (unique), it renders exactly one event"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("event-row"),
    "... and it renders the event's type"
  ).toContainText(subject.type);
  await expect(
    page.getByTestId("event-row"),
    "... and it renders the event's source"
  ).toContainText(subject.source);

  /* id is not known in test data, so just check it has correct length */
  const renderedId = await page.locator('td[headers="event-id"]').textContent();
  await expect(renderedId, "... and it renders the ID").toHaveLength(8);

  await expect(
    page.getByTestId("event-row"),
    "... and it renders the time string"
  ).toContainText("seconds ago");

  /**
   * View and resend the event
   */
  await page.getByTestId("event-row").click();

  await expect(
    page.getByRole("heading", { name: "View event" }),
    "it renders the event view dialog"
  ).toBeVisible();

  await page.getByRole("button", { name: "Cancel" }).click();

  await expect(
    page.getByRole("heading", { name: "View event" }),
    "it closes the event view dialog when clicking cancel"
  ).not.toBeVisible();

  await page.getByTestId("event-row").click();

  /*
   * Comparing the full content of the monaco editor would be complex,
   * so as a smoke test, this just checks for one part of the expected information.
   */
  await expect(
    page.getByTestId("event-view-card").getByText(subject.data),
    "in the event view dialog, it renders the event data"
  ).toBeVisible();

  await page.getByRole("button", { name: "Retrigger" }).click();

  await expect(
    page.getByRole("heading", { name: "View event" }),
    "it closes the event view dialog after retriggering the event"
  ).not.toBeVisible();

  await expect(
    page.getByTestId("toast-success"),
    "it renders a confirmation toast after resending the event"
  ).toBeVisible();

  await page.getByTestId("toast-close").click();

  await expect(
    page.getByTestId("toast-success"),
    "it is possible to close the toast"
  ).not.toBeVisible();

  /*
   * After resending the event, all the values will be the same,
   * so there is no point in adding additional assertions.
   */
  await expect(
    page.getByTestId("event-row"),
    "after resending the event, it still renders only one event"
  ).toHaveCount(1);
});
