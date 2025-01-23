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

test("Info page default view", async ({ page }) => {
  await page.goto(`/n/${namespace}/gateway/info`, {
    waitUntil: "networkidle",
  });

  await expect(
    page.getByTestId("breadcrumb-info"),
    "it renders the 'Info' breadcrumb"
  ).toBeVisible();

  await expect(page.getByText("Gateway Info")).toBeVisible();

  await expect(
    page.getByText(namespace).first(),
    "it displays the current namespace in the breadcrumb"
  ).toBeVisible();

  await expect(
    page.getByText(namespace).nth(1),
    "it displays the current namespace in the title"
  ).toBeVisible();

  await expect(
    page.getByText(namespace).nth(2),
    "it displays the current namespace in the info section"
  ).toBeVisible();

  await expect(
    page.getByText("Version").nth(0),
    "it displays the gateway version in the title"
  ).toBeVisible();
  await expect(
    page.getByText("Version").nth(1),
    "it displays the gateway version in the title"
  ).toBeVisible();

  await expect(
    page.getByText("1.0").nth(0),
    "it displays the gateway version in the info section"
  ).toBeVisible();

  await expect(
    page.getByText("1.0").nth(1),
    "it displays the gateway version in the info section"
  ).toBeVisible();
});
