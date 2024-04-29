import {
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "./utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeAll(async () => {
  namespace = await createNamespace();
});

test.afterAll(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to create and sync a mirror", async ({ page }) => {
  /* prepare test data */
  const mirrorName = createNamespaceName();

  /* open app and create namespace */
  await page.goto("/");

  await page.getByTestId("dropdown-trg-namespace").click();
  await page.getByTestId("new-namespace").click();
  await page.getByRole("tab", { name: "Mirror" }).click();

  await page.getByTestId("new-namespace-name").fill(mirrorName);
  await page
    .getByTestId("new-namespace-url")
    .fill("https://github.com/direktiv/e2e-mirror");
  await page.getByTestId("new-namespace-ref").fill("main");

  await page.getByTestId("new-namespace-submit").click();

  /* assert mirror page is rendered and sync is listed */
  await expect(page, "it redirects to the mirror route").toHaveURL(
    `/n/${mirrorName}/mirror/`
  );

  await expect(
    page.getByText("https://github.com/direktiv/e2e-mirror (main)"),
    "it renders the mirror url and ref"
  ).toBeVisible();

  await expect(
    page.getByTestId("sync-row"),
    "Initially, one sync is rendered in the list"
  ).toHaveCount(1);

  await expect(
    page.getByTestId("sync-row").getByRole("cell", { name: "complete" }),
    "The sync was successful"
  ).toBeVisible();

  await expect(
    page.getByTestId("sync-row").getByTestId("createdAt-relative"),
    "It renders the relative time"
  ).toContainText("seconds ago");

  /* update mirror to be invalid */
  await page.getByRole("button", { name: "Edit mirror" }).click();
  await page.getByTestId("new-namespace-ref").fill("invalid");
  await page.getByTestId("new-namespace-submit").click();

  await expect(
    page.getByText("https://github.com/direktiv/e2e-mirror (invalid)"),
    "it renders the updated mirror url and ref"
  ).toBeVisible();

  await page.getByRole("button", { name: "Sync" }).click();

  await expect(
    page.getByText("Warning: This will overwrite all files in this namespace"),
    "it renders a warning that files will be overwritten"
  ).toBeVisible();

  await page.getByRole("button", { name: "Sync" }).click();

  /* assert new sync is listed */
  await expect(
    page.getByTestId("sync-row"),
    "After syncing, two syncs are rendered in the list"
  ).toHaveCount(2);

  await expect(
    page.getByTestId("sync-row").first().getByRole("cell", { name: "failed" }),
    "The sync has failed"
  ).toBeVisible();

  await expect(
    page.getByTestId("sync-row").first().getByTestId("createdAt-relative"),
    "It renders the relative time"
  ).toContainText("seconds ago");

  /* update mirror to be valid */
  await page.getByRole("button", { name: "Edit mirror" }).click();
  await page.getByTestId("new-namespace-ref").fill("main");
  await page.getByTestId("new-namespace-submit").click();

  await expect(
    page.getByText("https://github.com/direktiv/e2e-mirror (main)"),
    "it renders the updated mirror url and ref"
  ).toBeVisible();

  await page.getByRole("button", { name: "Sync" }).click();
  await page.getByRole("button", { name: "Sync" }).click();

  /* assert new sync is listed */
  await expect(
    page.getByTestId("sync-row"),
    "After syncing, three syncs are rendered in the list"
  ).toHaveCount(3);

  await expect(
    page
      .getByTestId("sync-row")
      .first()
      .getByRole("cell", { name: "complete" }),
    "The sync has completed"
  ).toBeVisible();

  await expect(
    page.getByTestId("sync-row").first().getByTestId("createdAt-relative"),
    "It renders the relative time"
  ).toContainText("seconds ago");
});
