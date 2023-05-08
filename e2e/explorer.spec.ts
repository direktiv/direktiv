import {
  assertNamespaceExists,
  createNamespace,
  createNamespaceName,
  deleteNamespace,
} from "./utils/namespace";
import { expect, test } from "@playwright/test";

let namespace = "";

test.beforeEach(async () => {
  namespace = await createNamespace();
});

test.afterEach(async () => {
  await deleteNamespace(namespace);
  namespace = "";
});

test("it is possible to navigate to a namespace via breadcrumbs", async ({
  page,
}) => {
  // visit page
  await page.goto("/");

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a namespace is loaded in the explorer"
  ).toBeVisible();

  // at this point, any namespace may be loaded.
  // let's navigate to the test's namespace via breadcrumbs.

  await page.getByRole("main").getByTestId("dropdown-trg-namespace").click();
  await page.getByRole("menuitemradio", { name: namespace }).click();

  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
});

test("it is possible to navigate to a namespace via URL", async ({ page }) => {
  // visit url
  await page.goto(`/${namespace}/explorer/tree`);

  // make sure breadcrumb and url are correct after loading
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the namespace is reflected in the breadcrumbs"
  ).toHaveText(namespace);
  await expect(page, "the namespace is reflected in the url").toHaveURL(
    `/${namespace}/explorer/tree`
  );
});

test("it is possible to create a namespace via breadcrumbs", async ({
  page,
}) => {
  // visit page and make sure explorer is loaded
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a namespace is loaded in the explorer"
  ).toHaveText(namespace);

  // create new namespace
  const newNamespace = createNamespaceName();
  await page.getByTestId("dropdown-trg-namespace").click();
  await page.getByTestId("new-namespace").click();
  await page.getByTestId("new-namespace-name").fill(newNamespace);
  await page.getByTestId("new-namespace-submit").click();

  // make sure it has navigated to new namespace
  await expect(page, "it redirects to the new namespace's url").toHaveURL(
    `/${newNamespace}/explorer/tree`
  );

  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "the new namespace is reflected in the breadcrumbs"
  ).toHaveText(newNamespace);

  // make sure namespace exists in backend
  await assertNamespaceExists(expect, newNamespace);

  // cleanup
  await deleteNamespace(newNamespace);
});

test("it is possible to create and delete a directory", async ({ page }) => {
  // visit page and make sure explorer is loaded
  await page.goto(`/${namespace}/explorer/tree`);
  await expect(
    page.getByTestId("breadcrumb-namespace"),
    "a namespace is loaded in the explorer"
  ).toHaveText(namespace);

  // create folder
  await page.getByTestId("dropdown-trg-new").click();
  await page.getByTestId("new-dir").click();
  await page.getByPlaceholder("folder-name").fill("awesome-folder");
  await page.getByRole("button", { name: "Create" }).click();

  // it automatically navigates to the folder
  await page.getByText("/ awesome-folder").isVisible();
  await expect(
    page,
    "it creates a new folder and navigates to it automatically"
  ).toHaveURL(`/${namespace}/explorer/tree/awesome-folder`);

  // navigate back to tree root
  await page.getByTestId("tree-root").click();
  await expect(
    page,
    "when clicking the tree icon, it navigates back to the tree root"
  ).toHaveURL(`/${namespace}/explorer/tree`);
  await expect(page.getByTestId("breadcrumb-namespace")).toHaveText(namespace);

  // navigate to folder by clicking on it
  await page.getByRole("link", { name: "awesome-folder" }).click();
  await expect(
    page,
    "when clicking on the folder, it navigates to it"
  ).toHaveURL(`/${namespace}/explorer/tree/awesome-folder`);

  // navigate back by clicking on .. "folder"
  await page.getByRole("link", { name: ".." }).click();
  await expect(
    page,
    "when clicking .. it navigates back to the tree root"
  ).toHaveURL(`/${namespace}/explorer/tree`);

  // click delete and confirm
  await page.getByTestId("dropdown-trg-dir-actions").click();
  await page.getByTestId("dir-actions-delete").click();
  await page.getByRole("button", { name: "Delete" }).click();

  await expect(page.getByRole("dialog", { name: "Delete" })).toHaveCount(0);
  await expect(page).toHaveURL(`/${namespace}/explorer/tree`);
  await expect(
    page.getByRole("link", { name: "awesome-folder" }),
    "it deletes the folder"
  ).toHaveCount(0);
});
