import { createNamespace, deleteNamespace } from "./utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeAll(async () => {
  namespace = await createNamespace();
});

test.afterAll(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("the 404 error page shows, when the user opens a url that does not exist", async ({
  page,
}) => {
  await page.goto("/this/page/does/not/exist", { waitUntil: "networkidle" });
  await expect(page.getByTestId("error-title")).toHaveText("404");
  await expect(page.getByTestId("error-message")).toContainText(
    "The resource you are trying to access does not exist."
  );
});

test("the back button on the 404 error page navigates the user back to the previous page", async ({
  page,
}) => {
  await page.goto(`/n/${namespace}/explorer/tree`, {
    waitUntil: "networkidle",
  });

  await expect(page).toHaveURL(`/n/${namespace}/explorer/tree`);

  await page.goto("/this/page/does/not/exist");
  await expect(page).not.toHaveURL(`/n/${namespace}/explorer/tree`);

  await page.getByTestId("error-back-btn").click();

  await expect(page).toHaveURL(`/n/${namespace}/explorer/tree`);
});

test("the reload button on the error page reloads the current page", async ({
  page,
}) => {
  await page.goto("/this/page/does/not/exist", { waitUntil: "networkidle" });
  await page.getByTestId("error-reload-btn").click();

  /**
   * to test if the page will be reloaded, we wait for the request
   * to the version endpoint that should be made on every page load
   */
  await page.waitForRequest("/api/v2/status");
});

test("the home button on the error page navigates the user to the home page", async ({
  page,
}) => {
  await page.goto("/this/page/does/not/exist");
  await expect(
    page.getByRole("link", { name: "Go to homepage" }),
    "the home button links to /"
  ).toHaveAttribute("href", "/");
  /**
   * note: the user may be redirected to the first existing namespace
   * when visiting /. This behavior is tested in onboarding.spec.ts
   */
});
